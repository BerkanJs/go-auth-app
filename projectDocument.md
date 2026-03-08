# Go Kisi API — Proje Dokümantasyonu

Bu doküman, projenin güncel mimarisini, katmanlarını, veri modellerini, akışlarını ve tasarım kararlarını ayrıntılı biçimde açıklar.

---

## 1. Genel Bakış

Proje; kullanıcı yönetimi, rol tabanlı erişim kontrolü, JWT kimlik doğrulama ve blog yönetimi sunan tam yığın (full-stack) bir web uygulamasıdır. Backend Go ile yazılmış, frontend ise sunucu taraflı render edilen Bootstrap 5 şablonlarından oluşmaktadır.

### Teknoloji Yığını

| Katman | Teknoloji |
|---|---|
| Programlama dili | Go |
| Web sunucusu | `net/http` (framework yok) |
| Veritabanı | SQLite (`people.db`) — ORM yok, doğrudan SQL |
| Kimlik doğrulama | JWT (`golang-jwt/jwt/v5`) + bcrypt |
| Şifre güvenliği | `golang.org/x/crypto/bcrypt` |
| Arayüz | HTML şablonları (`html/template`) + Bootstrap 5 |
| API Dokümantasyonu | Swagger (`swaggo/swag`, `swaggo/http-swagger`) |
| Profiling / Monitoring | Go yerleşik pprof (`net/http/pprof`) |

### Port

Uygulama **`:8081`** portunda çalışır.

---

## 2. Klasör Yapısı

```
go-auth-app/
├── main.go                          # Giriş noktası, global middleware zinciri
├── people.db                        # SQLite veritabanı
├── app.log                          # Yapılandırılmış JSON uygulama logları
├── projectDocument.md               # Bu doküman
├── routes/
│   ├── routes.go                    # RegisterRoutes — AppBuilder'a delege eder
│   └── app_builder.go               # Builder Pattern: repo→service→handler→route inşası
├── handlers/
│   ├── auth_handler.go              # AuthHandler: Login, Refresh, Logout, JwtAuthMiddleware
│   ├── person_handler.go            # PersonHandler: kişi CRUD API endpoint'leri
│   ├── web_handler.go               # WebHandler: Login, Register, Update, Delete + PageRenderer'lar
│   ├── blog_handler.go              # BlogHandler: blog CRUD + sayfa renderer'ları
│   ├── page_renderer.go             # Template Method: PageRenderer interface + RenderPage()
│   ├── registration_pipeline.go     # Chain of Responsibility: kayıt pipeline'ı
│   ├── role_middleware.go           # AdminMiddleware, EditorMiddleware, RoleChecker
│   ├── auth_handler_test.go         # AuthHandler ve JwtAuthMiddleware unit testleri
│   └── registration_pipeline_test.go# Pipeline handler'larının unit testleri
├── service/
│   ├── interfaces.go                # AuthService, BlogService, PersonService, AuthorizationStrategy
│   ├── auth_service.go              # authService implementasyonu
│   ├── auth_service_test.go         # AuthService unit testleri (mock repo)
│   ├── blog_service.go              # blogService implementasyonu
│   ├── authorization.go             # Strategy Pattern: AdminOnly, OwnerOrAdmin
│   └── person_service.go            # personService implementasyonu
├── repository/
│   ├── interfaces.go                # PersonRepository, AuthRepository, BlogRepository
│   ├── person_repo.go               # SQLitePersonRepo implementasyonu + paket seviyesi wrapper'lar
│   ├── auth_repo.go                 # SQLiteAuthRepo implementasyonu + paket seviyesi wrapper'lar
│   ├── blog_repo.go                 # SQLiteBlogRepo implementasyonu + paket seviyesi wrapper'lar
│   └── photo_repo.go                # Dosya yükleme yardımcıları (UploadPhoto, DeletePhoto)
├── models/
│   ├── person.go                    # Person, CreatePersonRequest, PersonResponse, LoginRequest, TokenResponse
│   └── blog.go                      # Blog, CreateBlogRequest, UpdateBlogRequest, BlogResponse, Role
├── shared/
│   ├── auth.go                      # JWT üretimi ve doğrulaması (ParseAccessToken)
│   ├── config.go                    # Config struct ve environment variable okuma
│   ├── cors.go                      # CorsMiddleware
│   ├── errors.go                    # CustomError, ErrorType, hata yardımcıları
│   ├── errors_test.go               # CustomError testleri
│   ├── logging.go                   # LogInfo, LogError, LoggingMiddleware (JSON format)
│   ├── rate_limiter.go              # IP tabanlı fixed-window RateLimiter
│   ├── rate_limiter_test.go         # RateLimiter testleri
│   ├── validation.go                # Validator: email, şifre, isim, yaş, telefon doğrulama
│   ├── validation_test.go           # Validator testleri
│   └── web_helpers.go               # TemplateData, GetTemplateData (cookie → JWT → kullanıcı)
├── db/
│   └── db.go                        # db.Init(), global db.DB, tablo şemaları
├── queries/
│   └── queries.go                   # Tüm SQL sorguları sabit olarak
├── docs/
│   ├── docs.go                      # Swagger otomatik üretilen doküman
│   ├── swagger.json
│   └── swagger.yaml
├── tests/
│   └── integration/
│       ├── setup_test.go            # Test sunucusu kurulumu
│       ├── auth_test.go             # Login/Refresh/Logout entegrasyon testleri
│       └── health_test.go           # /api/health endpoint testi
└── static/                          # CSS, JS, görseller
    └── uploads/                     # Yüklenen kullanıcı fotoğrafları ve blog görselleri
```

---

## 3. Mimari: 3 Katmanlı Yapı

```
HTTP İsteği
    ↓
[Handlers] — HTTP parse, auth kontrol, response yazma
    ↓  ctx = r.Context() ile geçirilir
[Services] — İş mantığı, yetkilendirme kararları
    ↓  ctx iletilir
[Repositories] — QueryContext/ExecContext ile DB sorguları
    ↓
[SQLite DB]
```

### context.Context Akışı

Her HTTP isteğinde `r.Context()` handler'dan başlayarak servis ve repository katmanlarına iletilir. Bu sayede:
- İstemci bağlantıyı kestiğinde devam eden DB sorguları iptal edilir
- Timeout ayarlanabilir
- Request-scoped değerler taşınabilir

DB'ye dokunan tüm methodlar `ctx context.Context` alır. Sadece JWT işlemi yapan methodlar (`GenerateAccessToken`, `ParseRefreshToken`) ctx almaz çünkü bunlar DB'ye dokunmaz.

---

## 4. Veri Modelleri

### Person

```go
// İç model — PasswordHash asla JSON'a sızmaz
type Person struct {
    ID           int    `json:"id"`
    Name         string `json:"name"`
    Surname      string `json:"surname"`
    Email        string `json:"email"`
    Age          int    `json:"age"`
    Phone        string `json:"phone"`
    PhotoPath    string `json:"photoPath"`
    Role         string `json:"role"`
    PasswordHash string `json:"-"`
}

// API giriş DTO
type CreatePersonRequest struct {
    Name      string `json:"name"`
    Surname   string `json:"surname"`
    Email     string `json:"email"`
    Age       int    `json:"age"`
    Phone     string `json:"phone"`
    PhotoPath string `json:"photoPath"`
    Role      string `json:"role"`
    Password  string `json:"password"` // Hash'lenerek saklanır
}

// API çıkış DTO — şifre hash'i içermez
type PersonResponse struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Surname   string `json:"surname"`
    Email     string `json:"email"`
    Age       int    `json:"age"`
    Phone     string `json:"phone"`
    PhotoPath string `json:"photoPath"`
    Role      string `json:"role"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type TokenResponse struct {
    AccessToken  string `json:"accessToken"`
    RefreshToken string `json:"refreshToken"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refreshToken"`
}
```

### Blog

```go
type Role string
const (
    RoleAdmin  Role = "admin"
    RoleEditor Role = "editor"
)

type Blog struct {
    ID         int       `json:"id"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    Summary    string    `json:"summary"`
    ImagePath  string    `json:"imagePath"`
    AuthorID   int       `json:"authorId"`
    AuthorName string    `json:"authorName"`
    Published  bool      `json:"published"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}

type BlogResponse struct { /* Blog ile aynı alanlar */ }
```

---

## 5. Arayüzler (Interfaces)

### Repository Katmanı (`repository/interfaces.go`)

```go
type PersonRepository interface {
    AddPerson(ctx context.Context, p models.Person) (int64, error)
    GetAllPeople(ctx context.Context) ([]models.Person, error)
    GetPersonByID(ctx context.Context, id int) (models.Person, error)
    GetPersonByEmail(ctx context.Context, email string) (models.Person, error)
    EmailExists(ctx context.Context, email string) (bool, error)
    DeletePerson(ctx context.Context, id int) error
    UpdatePerson(ctx context.Context, p models.Person) error
}

type AuthRepository interface {
    SaveRefreshToken(ctx context.Context, userID int, token string) error
    IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
    RevokeRefreshToken(ctx context.Context, token string) error
}

type BlogRepository interface {
    CreateBlog(ctx context.Context, blog models.Blog) (int64, error)
    GetAllBlogs(ctx context.Context) ([]models.Blog, error)
    GetPublishedBlogs(ctx context.Context) ([]models.Blog, error)
    GetBlogByID(ctx context.Context, id int) (models.Blog, error)
    GetBlogsByAuthor(ctx context.Context, authorID int) ([]models.Blog, error)
    UpdateBlog(ctx context.Context, blog models.Blog) error
    DeleteBlog(ctx context.Context, id int) error
    UpdateBlogPublishStatus(ctx context.Context, id int, published bool) error
}
```

### Servis Katmanı (`service/interfaces.go`)

```go
type AuthService interface {
    Login(ctx context.Context, email, password string) (models.Person, error)
    GenerateAccessToken(userID int) (string, error)                             // DB yok → ctx yok
    GenerateRefreshToken(ctx context.Context, userID int) (string, error)
    IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
    ParseRefreshToken(token string) (int, error)                                // DB yok → ctx yok
    RevokeRefreshToken(ctx context.Context, token string) error
}

type BlogService interface {
    GetBlogsForUser(ctx context.Context, userRole string, userID int) ([]models.Blog, error)
    CreateBlog(ctx context.Context, title, content, summary, imagePath string, published bool, authorID int, authorName string) (int64, error)
    UpdateBlog(ctx context.Context, blogID int, title, content, summary, imagePath string, published bool, userRole string, userID int) error
    DeleteBlog(ctx context.Context, blogID int, userRole string, userID int) error
}

type PersonService interface {
    UpdatePerson(ctx context.Context, req UpdatePersonRequest) error
}

// Strategy Pattern arayüzü
type AuthorizationStrategy interface {
    IsAuthorized(userRole string, userID int, resourceOwnerID int) bool
}
```

### Handler Katmanı

```go
// Template Method Pattern arayüzü (handlers/page_renderer.go)
type PageRenderer interface {
    RequiresAuth() bool
    Title() string
    TemplateName() string
    LoadData(ctx context.Context, data *shared.TemplateData, userID int) error
}

// Chain of Responsibility Pattern arayüzü (handlers/registration_pipeline.go)
type RegistrationHandler interface {
    SetNext(RegistrationHandler) RegistrationHandler
    Handle(ctx context.Context, regCtx *registrationContext, repo repository.PersonRepository) error
}
```

---

## 6. Tasarım Desenleri

### A. Builder Pattern — `routes/app_builder.go`

Tüm bağımlılıkların inşasını merkezi bir yapıda toplar. Fluent interface ile konfigürasyon yapılır.

```go
// Kullanım
builder := routes.NewAppBuilder(db.DB).
    WithLoginRateLimit(10, time.Minute)
builder.Build()

// İç akış
func (b *AppBuilder) Build() {
    repos    := b.buildRepos()    // repo'lar oluşturulur
    services := b.buildServices() // repo'lar inject edilir
    handlers := b.buildHandlers() // service'ler inject edilir
    b.registerRoutes(handlers)    // handler'lar route'lara bağlanır
}
```

### B. Chain of Responsibility Pattern — `handlers/registration_pipeline.go`

Kayıt akışı her biri tek sorumluluk üstlenen handler zinciriyle yürütülür. Herhangi bir adım hata dönerse zincir durur.

```
EmailCheckHandler → PersonBuildHandler → PersonSaveHandler
```

```go
// Zincir kurulumu
func NewRegistrationChain() RegistrationHandler {
    emailCheck  := &EmailCheckHandler{}
    personBuild := &PersonBuildHandler{}
    personSave  := &PersonSaveHandler{}
    emailCheck.SetNext(personBuild).SetNext(personSave)
    return emailCheck
}

// Her handler ctx'i bir sonrakine iletir
func (h *EmailCheckHandler) Handle(ctx context.Context, regCtx *registrationContext, repo repository.PersonRepository) error {
    exists, err := repo.EmailExists(ctx, regCtx.Req.Email)
    if err != nil { return err }
    if exists    { return errEmailAlreadyExists }
    return h.handleNext(ctx, regCtx, repo) // bir sonraki handler
}
```

| Handler | Sorumluluk |
|---|---|
| `EmailCheckHandler` | Email daha önce kayıtlı mı? |
| `PersonBuildHandler` | Şifreyi bcrypt'le hash'le, Person struct'ı oluştur |
| `PersonSaveHandler` | Kişiyi DB'ye kaydet, dönen ID'yi set et |

### C. Strategy Pattern — `service/authorization.go`

Yetkilendirme kararı, runtime'da değiştirilebilir bir strateji nesnesine delege edilir.

```go
type AuthorizationStrategy interface {
    IsAuthorized(userRole string, userID int, resourceOwnerID int) bool
}

// Sadece admin yetkilidir
type AdminOnlyStrategy struct{}
func (s *AdminOnlyStrategy) IsAuthorized(userRole string, _ int, _ int) bool {
    return userRole == string(models.RoleAdmin)
}

// Admin veya kaynağın sahibi yetkilidir
type OwnerOrAdminStrategy struct{}
func (s *OwnerOrAdminStrategy) IsAuthorized(userRole string, userID int, resourceOwnerID int) bool {
    return userRole == string(models.RoleAdmin) || userID == resourceOwnerID
}
```

`blogService` içinde bu strateji field olarak tutulur ve `UpdateBlog`/`DeleteBlog` çağrılarında kullanılır.

### D. Template Method Pattern — `handlers/page_renderer.go`

Sayfa render akışının iskeleti `RenderPage()` fonksiyonunda sabitlenmiştir. Her sayfa kendi `LoadData()` metodunu implement eder.

```go
// Sabit iskelet
func RenderPage(w http.ResponseWriter, r *http.Request, renderer PageRenderer) {
    ctx  := r.Context()
    data := shared.GetTemplateData(r)

    if renderer.RequiresAuth() && !data.IsAuthenticated {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    data.Title = renderer.Title()
    // userID parse ...
    renderer.LoadData(ctx, &data, userID)  // değişen kısım
    renderTemplate(w, renderer.TemplateName(), data)
}

// Her sayfa LoadData'yı farklı implement eder
type homePageRenderer struct{ blogRepo repository.BlogRepository }
func (p *homePageRenderer) LoadData(ctx context.Context, data *shared.TemplateData, userID int) error {
    blogs, err := p.blogRepo.GetPublishedBlogs(ctx)
    data.Blogs = models.ToBlogResponseList(blogs)
    return err
}
```

| Renderer | Şablon | Veri |
|---|---|---|
| `homePageRenderer` | `home.html` | Yayınlanmış bloglar |
| `adminPageRenderer` | `admin.html` | Tüm kullanıcılar |
| `blogPageRenderer` | `blog.html` | Kullanıcıya göre bloglar |
| `editorPageRenderer` | `editor.html` | Kullanıcıya göre bloglar |

### E. Repository Pattern — `repository/`

Veritabanı işlemleri arayüz arkasına gizlenir. Test sırasında gerçek DB yerine mock repo kullanılabilir.

```go
// Üretim
repo := repository.NewPersonRepo(db)

// Test
repo := &mockPersonRepo{emailExists: false, addPersonID: 42}
```

### F. Dependency Injection — `routes/app_builder.go`

Tüm bağımlılıklar constructor injection ile dışarıdan verilir. Hiçbir servis veya handler kendi bağımlılığını oluşturmaz.

```go
personRepo := repository.NewPersonRepo(db)
authRepo   := repository.NewAuthRepo(db)

authSvc    := service.NewAuthService(authRepo, personRepo, cfg)
blogSvc    := service.NewBlogService(blogRepo)

authHandler   := handlers.NewAuthHandler(authSvc)
personHandler := handlers.NewPersonHandler(personRepo)
```

---

## 7. HTTP Endpoint'leri

### REST API (JSON)

| Method | Path | Handler | Auth |
|---|---|---|---|
| `POST` | `/api/login` | `LoginHandler` | — |
| `POST` | `/api/refresh` | `RefreshHandler` | — |
| `POST` | `/api/logout` | `LogoutHandler` | JWT |
| `GET` | `/api/health` | `HealthHandler` | — |
| `POST` | `/api/add` | `AddPersonHandler` | — |
| `GET` | `/api/all` | `GetAllPeopleHandler` | JWT |
| `GET` | `/api/get?id=` | `GetPersonByIDHandler` | JWT |
| `GET` | `/api/delete?id=` | `DeletePersonHandler` | JWT |

### Web UI (Form/HTML)

| Method | Path | Açıklama | Auth |
|---|---|---|---|
| `GET` | `/` | Ana sayfa (yayınlanmış bloglar) | — |
| `GET` | `/login` | Giriş sayfası | — |
| `GET` | `/register` | Kayıt sayfası | — |
| `POST` | `/web-login` | Giriş işlemi (cookie set) | — |
| `POST` | `/web-register` | Kayıt işlemi (fotoğraf yükleme) | — |
| `GET` | `/web-logout` | Çıkış (cookie sil) | — |
| `GET` | `/admin` | Admin paneli | Admin |
| `POST` | `/user/update` | Kullanıcı güncelleme | Admin |
| `GET` | `/web-delete?id=` | Kullanıcı silme | Admin |
| `GET` | `/blogs` | Blog yönetimi | Editor/Admin |
| `GET` | `/editor` | Editor paneli | Editor/Admin |
| `POST` | `/blog/create` | Blog oluşturma | Editor/Admin |
| `POST` | `/blog/update` | Blog güncelleme | Editor/Admin |
| `GET` | `/blog/delete?id=` | Blog silme | Editor/Admin |

### Diğer

| Path | Açıklama |
|---|---|
| `/static/*` | Statik dosyalar |
| `/uploads/*` | Yüklenen fotoğraflar |
| `/swagger/*` | Swagger API dokümantasyonu |
| `/debug/pprof/*` | Go pprof profiling |

---

## 8. Middleware Zinciri

```
İstek → LoggingMiddleware → CorsMiddleware → mux
                                                ↓
                                    JwtAuthMiddleware (korumalı route'lar)
                                                ↓
                                    AdminMiddleware / EditorMiddleware
                                                ↓
                                    RateLimiter.Middleware (login için)
                                                ↓
                                    Handler
```

| Middleware | Dosya | Görev |
|---|---|---|
| `LoggingMiddleware` | `shared/logging.go` | Her isteği JSON olarak loglar |
| `CorsMiddleware` | `shared/cors.go` | Cross-origin header'ları ekler |
| `JwtAuthMiddleware` | `handlers/auth_handler.go` | Bearer token doğrular, userID context'e yazar |
| `AdminMiddleware` | `handlers/role_middleware.go` | Admin rolü zorunlu kılar |
| `EditorMiddleware` | `handlers/role_middleware.go` | Editor veya admin rolü zorunlu kılar |
| `RateLimiter.Middleware` | `shared/rate_limiter.go` | IP başına istek sınırlar |

---

## 9. Veritabanı Şeması

```sql
-- Kullanıcılar
CREATE TABLE people (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    surname       TEXT,
    email         TEXT NOT NULL UNIQUE,
    age           INTEGER,
    phone         TEXT,
    photo_path    TEXT,
    role          TEXT NOT NULL DEFAULT 'editor',
    password_hash TEXT NOT NULL
);

-- Refresh token'lar
CREATE TABLE refresh_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    token      TEXT NOT NULL UNIQUE,
    revoked    INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    revoked_at TEXT
);

-- Bloglar
CREATE TABLE blogs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    summary     TEXT,
    image_path  TEXT,
    author_id   INTEGER NOT NULL,
    author_name TEXT NOT NULL,
    published   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES people(id)
);
```

---

## 10. Kimlik Doğrulama ve Yetkilendirme Akışı

### Giriş (Login)

```
1. POST /api/login → email + password
2. authService.Login(ctx, email, password)
   a. repo.GetPersonByEmail(ctx, email) → Person
   b. bcrypt.CompareHashAndPassword(hash, password)
3. GenerateAccessToken(userID)  → imzalı JWT (15 dk)
4. GenerateRefreshToken(ctx, userID) → imzalı JWT (7 gün) + DB'ye kaydedilir
5. Response: { accessToken, refreshToken }
```

### Token Yenileme (Refresh)

```
1. POST /api/refresh → refreshToken
2. IsRefreshTokenValid(ctx, token) → DB'de revoked değil mi?
3. ParseRefreshToken(token) → userID
4. GenerateAccessToken(userID) → yeni access token
5. Response: { accessToken }
```

### Çıkış (Logout)

```
1. POST /api/logout → refreshToken
2. IsRefreshTokenValid(ctx, token) → geçerli mi?
3. RevokeRefreshToken(ctx, token) → DB'de revoked=1
4. Response: 200 OK
```

### Web Oturumu

Web route'larında token cookie'de (`auth_token`) taşınır. `GetTemplateData()` her istekte cookie'yi okur, JWT'yi parse eder ve kullanıcı bilgisini yükler.

---

## 11. Validasyon

`shared/validation.go` içindeki `Validator` struct fluent interface sunar:

```go
v := shared.NewValidator()
v.RequireNonEmpty(req.Name, "name").
  ValidateName(req.Name, "name").
  ValidateEmail(req.Email).
  ValidatePassword(req.Password).
  ValidateAge(req.Age).
  ValidatePhone(req.Phone)

if v.HasErrors() {
    // v.Errors() → []string
}
```

| Kural | Detay |
|---|---|
| Email | Standart format kontrolü |
| Şifre | 6-50 karakter, en az 1 büyük harf, en az 1 rakam |
| İsim/Soyisim | 2-50 karakter |
| Yaş | 0-150 |
| Telefon | 10-20 karakter |

---

## 12. Dosya Yükleme

`repository/photo_repo.go` — kullanıcı fotoğrafı ve blog görseli yükleme:

- İzin verilen uzantılar: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`
- Güvenli dosya adı: UUID tabanlı rastgele isim (path traversal koruması)
- Konum: `static/uploads/`
- Güncelleme/silme sırasında eski dosya diskten kaldırılır

---

## 13. Loglama

`shared/logging.go` — yapılandırılmış JSON formatı:

```go
shared.LogInfo("BLOG_CREATED", "Blog created successfully", map[string]interface{}{
    "author_id": claims.UserID,
})
shared.LogError("BLOG_DELETE_ERROR", "Failed to delete blog", map[string]interface{}{
    "blog_id": blogID,
    "error":   err.Error(),
})
```

- Her log satırı JSON objesidir
- `app.log` dosyasına yazılır
- `LoggingMiddleware` tüm HTTP isteklerini (method, path, status, süre) loglar

---

## 14. Rate Limiting

`shared/rate_limiter.go` — IP tabanlı fixed-window:

```go
limiter := shared.NewRateLimiter(10, time.Minute) // 10 istek/dakika
mux.Handle("/api/login", limiter.Middleware(loginHandler))
```

- Her IP için ayrı sayaç tutulur
- Pencere dolunca `429 Too Many Requests` döner
- Süresi dolmuş kayıtlar arka planda temizlenir

---

## 15. Test Yapısı

### Unit Testler

| Dosya | Test Sayısı | Kapsam |
|---|---|---|
| `handlers/auth_handler_test.go` | 11 | LoginHandler, RefreshHandler, LogoutHandler, JwtAuthMiddleware |
| `handlers/registration_pipeline_test.go` | 11 | EmailCheck, PersonBuild, PersonSave, zincir davranışı, runRegistrationPipeline |
| `service/auth_service_test.go` | 16 | Login, token üretme/parse, refresh token CRUD |
| `shared/errors_test.go` | — | CustomError tipleri |
| `shared/validation_test.go` | — | Validator kuralları |
| `shared/rate_limiter_test.go` | — | Limit aşımı, pencere sıfırlama |

### Entegrasyon Testleri (`tests/integration/`)

| Dosya | Kapsam |
|---|---|
| `auth_test.go` | Gerçek HTTP istekleriyle Login/Refresh/Logout akışı |
| `health_test.go` | `/api/health` endpoint'i |
| `setup_test.go` | Test sunucusu başlatma ve teardown |

### Mock Yapısı

Testlerde gerçek DB yerine interface'i implement eden mock struct'lar kullanılır:

```go
// handlers/registration_pipeline_test.go
type mockPersonRepo struct {
    emailExists    bool
    emailExistsErr error
    addPersonErr   error
    addPersonID    int64
}
func (m *mockPersonRepo) EmailExists(_ context.Context, email string) (bool, error) { ... }
func (m *mockPersonRepo) AddPerson(_ context.Context, p models.Person) (int64, error) { ... }
// ... diğer interface metodları

// service/auth_service_test.go
type mockAuthRepo struct {
    saveErr      error
    isValidResult bool
    revokeErr    error
}
```

---

## 16. Konfigürasyon

`shared/config.go` — environment variable'lardan okunur:

| Değişken | Varsayılan | Açıklama |
|---|---|---|
| `JWT_ACCESS_SECRET` | `default-access-secret` | Access token imza anahtarı |
| `JWT_REFRESH_SECRET` | `default-refresh-secret` | Refresh token imza anahtarı |
| `ACCESS_TOKEN_TTL` | `900` (15 dk) | Access token geçerlilik süresi (saniye) |
| `REFRESH_TOKEN_TTL` | `604800` (7 gün) | Refresh token geçerlilik süresi (saniye) |
| `SERVER_PORT` | `:8081` | Sunucu portu |
| `DATABASE_PATH` | `people.db` | SQLite dosya yolu |
| `ENVIRONMENT` | `development` | `production` modunda güçlü secret zorunlu |

---

## 17. Uygulama Başlangıcı (`main.go`)

```go
func main() {
    db.Init()                          // SQLite başlat, tabloları oluştur
    repository.SetDB(db.DB)            // Paket seviyesi wrapper'lar için

    var handler http.Handler = http.DefaultServeMux
    handler = shared.CorsMiddleware(handler)
    handler = shared.LoggingMiddleware(handler)

    routes.RegisterRoutes()            // AppBuilder ile tüm route'ları kaydet

    http.ListenAndServe(":8081", handler)
}
```

---

## 18. SQL Sorguları (`queries/queries.go`)

Tüm SQL sorguları tek dosyada sabit olarak tutulur. Handler veya service içinde ham SQL yoktur.

```go
const (
    SelectAllPeople    = `SELECT id, name, surname, email, age, phone, photo_path, role, password_hash FROM people`
    SelectPersonByID   = `SELECT ... FROM people WHERE id = ?`
    InsertPerson       = `INSERT INTO people (name, surname, email, age, phone, photo_path, role, password_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    // ... diğer sorgular
)
```

---

## 19. Güvenlik Özeti

| Alan | Önlem |
|---|---|
| Şifreler | bcrypt ile hash'lenir, düz metin asla saklanmaz |
| JWT | Access (15 dk) + Refresh (7 gün) ayrı secret'larla imzalanır |
| Refresh Token | DB'de saklanır, logout'ta revoke edilir |
| Dosya yükleme | UUID isim, uzantı kontrolü, path traversal koruması |
| SQL | Parametreli sorgular (SQL injection yok) |
| Rate limiting | IP başına login denemesi sınırlandırılır |
| Rol kontrolü | Her korumalı route'ta middleware seviyesinde kontrol |
| JSON çıktısı | `PasswordHash` asla serialize edilmez (`json:"-"`) |
| context.Context | İstemci bağlantıyı kesince DB sorguları iptal edilir |
