# Go Kisi API — Proje Dokümantasyonu

Bu doküman, projenin güncel mimarisini, katmanlarını, veri modellerini, akışlarını ve gelecek planlarını ayrıntılı biçimde açıklar.

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
├── routes/
│   └── routes.go                    # Repo/service/handler oluşturulur ve wire edilir
├── handlers/
│   ├── auth_handler.go              # AuthHandler struct: Login, Refresh, Logout, JwtAuthMiddleware
│   ├── person_handler.go            # PersonHandler struct: kişi CRUD API endpoint'leri
│   ├── web_handler.go               # WebHandler struct: Home, Login, Register, Admin, Update, Delete + renderTemplate
│   ├── blog_handler.go              # BlogHandler struct: Blog/Editor sayfası + CRUD
│   ├── editor_handler.go            # (stub — EditorPageHandler BlogHandler'a taşındı)
│   ├── user_update_handler.go       # (stub — UpdateUserHandler WebHandler'a taşındı)
│   ├── registration_pipeline.go     # Kayıt akışı pipeline (Chain of Responsibility)
│   ├── role_middleware.go           # Rol tabanlı erişim middleware'leri
│   └── health_handler.go            # Sağlık kontrolü endpoint'i
├── service/
│   ├── interfaces.go                # AuthService, BlogService, PersonService interface'leri
│   ├── auth_service.go              # authService: Login, token üretimi/doğrulaması
│   ├── blog_service.go              # blogService: blog CRUD + yetki kontrolü
│   └── person_service.go            # personService: kullanıcı güncelleme iş mantığı
├── repository/
│   ├── interfaces.go                # AuthRepository, BlogRepository, PersonRepository interface'leri
│   ├── person_repo.go               # SQLitePersonRepo + wrapper fonksiyonlar + fotoğraf yükleme
│   ├── blog_repo.go                 # SQLiteBlogRepo + wrapper fonksiyonlar
│   └── auth_repo.go                 # SQLiteAuthRepo + wrapper fonksiyonlar
├── models/
│   ├── person.go                    # Person entity'si ve DTO'ları
│   └── blog.go                      # Blog entity'si ve DTO'ları
├── queries/
│   └── queries.go                   # Merkezi SQL sorgu sabitleri
├── db/
│   └── db.go                        # SQLite bağlantısı ve tablo kurulumu
├── shared/
│   ├── auth.go                      # JWT parse (dışa açık)
│   ├── config.go                    # Ortam değişkeni tabanlı yapılandırma
│   ├── cors.go                      # CORS middleware
│   ├── errors.go                    # CustomError sistemi, HTTP helper'lar, uygulama loglama
│   ├── logging.go                   # HTTP istek loglama middleware
│   ├── validation.go                # Girdi doğrulama katmanı
│   ├── rate_limiter.go              # IP bazlı rate limiting
│   └── web_helpers.go               # TemplateData yapısı ve GetTemplateData
├── templates/
│   ├── layout.html                  # Tüm sayfaların temel şablonu (navbar, alert'ler)
│   ├── home.html                    # Yayınlanmış blog'ların gösterildiği ana sayfa
│   ├── login.html                   # Giriş formu
│   ├── register.html                # Kayıt formu
│   ├── admin.html                   # Admin paneli (kullanıcı listesi + blog yönetimi bağlantısı)
│   ├── editor.html                  # Editor paneli (kendi blog'larını yönetir)
│   └── blog.html                    # Blog yönetim sayfası (admin + editor)
├── static/
│   ├── css/style.css                # Özel stiller
│   └── js/app.js                    # Kullanıcı silme ve düzenleme JS fonksiyonları
├── uploads/                         # Yüklenen fotoğraflar (çalışma zamanında oluşur)
├── docs/                            # Swagger tarafından otomatik üretilen dosyalar
├── people.db                        # SQLite veritabanı dosyası
└── app.log                          # Uygulama log dosyası
```

---

## 3. Veritabanı Katmanı (`db/db.go`)

`Init()` fonksiyonu uygulama başlarken çağrılır ve üç tabloyu `CREATE TABLE IF NOT EXISTS` ile oluşturur.

### 3.1. `people` tablosu

| Sütun | Tip | Özellik |
|---|---|---|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL |
| surname | TEXT | — |
| email | TEXT | NOT NULL UNIQUE |
| age | INTEGER | — |
| phone | TEXT | — |
| photo_path | TEXT | Yüklenen fotoğrafın URL yolu |
| role | TEXT | NOT NULL DEFAULT 'editor' |
| password_hash | TEXT | NOT NULL — bcrypt hash |

Email kolonu **UNIQUE** olduğundan aynı e-posta ile birden fazla kayıt hem kod hem de veritabanı seviyesinde engellenir.

### 3.2. `refresh_tokens` tablosu

| Sütun | Tip | Özellik |
|---|---|---|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| user_id | INTEGER | NOT NULL |
| token | TEXT | NOT NULL UNIQUE |
| revoked | INTEGER | NOT NULL DEFAULT 0 |
| created_at | TEXT | NOT NULL |
| revoked_at | TEXT | — (logout anında set edilir) |

### 3.3. `blogs` tablosu

| Sütun | Tip | Özellik |
|---|---|---|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| title | TEXT | NOT NULL |
| content | TEXT | NOT NULL |
| summary | TEXT | — |
| image_path | TEXT | Yüklenen görselin URL yolu |
| author_id | INTEGER | NOT NULL, FOREIGN KEY → people(id) |
| author_name | TEXT | NOT NULL |
| published | INTEGER | NOT NULL DEFAULT 0 (0=taslak, 1=yayında) |
| created_at | TEXT | NOT NULL |
| updated_at | TEXT | NOT NULL |

---

## 4. Model Katmanı

### 4.1. `models/person.go`

**`Person`** — Veritabanı entity'si:
- `ID`, `Name`, `Surname`, `Email`, `Age`, `Phone`, `PhotoPath`, `Role`
- `PasswordHash string json:"-"` — JSON serialization'dan gizlenir, yalnızca DB'de taşınır

**`CreatePersonRequest`** — Yeni kullanıcı oluşturma DTO'su:
- `Name`, `Surname`, `Email`, `Age`, `Phone`, `PhotoPath`, `Role`, `Password` (düz metin)

**`PersonResponse`** — Dışa döndürülen DTO:
- `ID`, `Name`, `Surname`, `Email`, `Age`, `Phone`, `PhotoPath`, `Role`
- Şifre bilgisi içermez

**`LoginRequest`** — Giriş DTO'su: `Email`, `Password`

**`TokenResponse`** — Token yanıt DTO'su: `AccessToken`, `RefreshToken`

**`RefreshTokenRequest`** — Token yenileme DTO'su: `RefreshToken`

Yardımcı fonksiyonlar:
- `ToPersonResponse(p Person) PersonResponse` — tek dönüşüm
- `ToPersonResponseList(people []Person) []PersonResponse` — liste dönüşümü

---

### 4.2. `models/blog.go`

**`Blog`** — Veritabanı entity'si:
- `ID`, `Title`, `Content`, `Summary`, `ImagePath`
- `AuthorID int`, `AuthorName string`
- `Published bool`
- `CreatedAt time.Time`, `UpdatedAt time.Time`

**`CreateBlogRequest`** — Blog oluşturma DTO'su: `Title`, `Content`, `Summary`, `ImagePath`, `Published`

**`UpdateBlogRequest`** — Blog güncelleme DTO'su: yukarıdakine ek `ID`

**`BlogResponse`** — Dışa döndürülen DTO (Blog ile birebir aynı yapı)

Yardımcı fonksiyonlar:
- `ToBlogResponse(blog Blog) BlogResponse`
- `ToBlogResponseList(blogs []Blog) []BlogResponse`

---

## 5. Yapılandırma Katmanı (`shared/config.go`)

Tüm yapılandırma ortam değişkenlerinden okunur. Değişken tanımlı değilse varsayılan değer kullanılır.

| Ortam Değişkeni | Varsayılan | Açıklama |
|---|---|---|
| JWT_ACCESS_SECRET | `super-secret-access-key` | Access token imzalama anahtarı |
| JWT_REFRESH_SECRET | `super-secret-refresh-key` | Refresh token imzalama anahtarı |
| ACCESS_TOKEN_TTL | `900` (15 dakika) | Access token geçerlilik süresi (saniye) |
| REFRESH_TOKEN_TTL | `604800` (7 gün) | Refresh token geçerlilik süresi (saniye) |
| SERVER_PORT | `:8081` | Sunucu portu |
| DB_PATH | `people.db` | SQLite veritabanı dosya yolu |
| ENVIRONMENT | `development` | Ortam tipi |

**Production güvenlik koruması:** `ENVIRONMENT=production` iken JWT secret'ları varsayılan değerlerde bırakılırsa uygulama başlamaz ve panic verir.

---

## 6. Kimlik Doğrulama (`shared/auth.go` + `service/auth_service.go` + `handlers/auth_handler.go`)

### 6.1. JWT Yapısı

`shared/auth.go`'da `jwtClaims` (unexported):
```go
type jwtClaims struct {
    UserID int `json:"userId"`
    jwt.RegisteredClaims
}
```

`service/auth_service.go`'da `authClaims` (aynı yapı, refresh token parse için):
```go
type authClaims struct {
    UserID int `json:"userId"`
    jwt.RegisteredClaims
}
```

`RegisteredClaims` aracılığıyla `ExpiresAt` ve `IssuedAt` alanları otomatik yönetilir.

### 6.2. AuthService (`service/auth_service.go`)

Token üretimi ve kimlik doğrulama iş mantığı `authService` struct'ında toplanmıştır:

| Metod | Açıklama |
|---|---|
| `Login(email, password)` | Kullanıcıyı DB'den çeker, bcrypt ile şifreyi doğrular |
| `GenerateAccessToken(userID)` | Kısa ömürlü access token üretir (varsayılan 15 dk) |
| `GenerateRefreshToken(userID)` | Uzun ömürlü refresh token üretir, DB'ye kaydeder (varsayılan 7 gün) |
| `IsRefreshTokenValid(token)` | Token'ın DB'de geçerli olup olmadığını kontrol eder |
| `ParseRefreshToken(token)` | Refresh JWT'yi doğrular, userID döner |
| `RevokeRefreshToken(token)` | Token'ı geçersiz kılar |

JWT secret'ları ve TTL değerleri `NewAuthService()` constructorında bir kez okunur, struct field'larında saklanır.

### 6.3. Token Doğrulaması

- `shared.ParseAccessToken(tokenStr)` — dışa açık, web handler'lar ve middleware tarafından cookie/header'dan token okumak için kullanılır
- `authService.ParseRefreshToken(token)` — yalnızca refresh endpoint'inde kullanılır

### 6.4. İki Farklı Auth Mekanizması

| Mekanizma | Kullanım Yeri | Detay |
|---|---|---|
| Cookie (`auth_token`) | Web UI | HttpOnly cookie, 1 saatlik access token |
| Bearer Token | REST API (`/api/*`) | `Authorization: Bearer <token>` header'ı |

### 6.5. `JwtAuthMiddleware`

API endpoint'leri için JWT doğrulama middleware'i:
1. `Authorization` header'ı okunur
2. `Bearer <token>` formatı kontrol edilir
3. `ParseAccessToken` ile doğrulanır
4. Hata varsa `401 Unauthorized` döner
5. Geçerliyse `userID` context'e eklenir (`userIDContextKey`) ve `next` handler çağrılır

---

## 7. Rol Tabanlı Erişim Kontrolü (`handlers/role_middleware.go`)

Proje iki rol destekler:

| Rol | Erişim |
|---|---|
| `admin` | Tüm sayfalar: kullanıcı listesi, kullanıcı ekleme/düzenleme/silme, tüm blog'lar |
| `editor` | Yalnızca kendi blog'ları: ekleme, düzenleme, silme |

### Middleware'ler

```
RoleMiddleware(allowedRoles ...string)  — genel rol kontrolü
AdminMiddleware(next)                   — sadece admin
EditorMiddleware(next)                  — admin + editor
```

`RoleMiddleware` akışı:
1. `auth_token` cookie'si okunur → yoksa `/login`'e yönlendir
2. Token parse edilir → geçersizse `/login`'e yönlendir
3. `GetPersonByID` ile kullanıcı DB'den alınır
4. Rol listesinde var mı kontrol edilir → yoksa `403 Forbidden`

---

## 8. Mimari: Katmanlı Yapı ve Dependency Injection

### 8.1. Katmanlar

```
Handler  →  Service  →  Repository  →  DB
```

Her katman bir üst katmanın **interface**'ini kullanır, concrete implementasyona bağımlı değildir.

### 8.2. Interface'ler

**`repository/interfaces.go`** — DB bağımsızlığı sağlar:
- `AuthRepository` — token CRUD
- `BlogRepository` — blog CRUD
- `PersonRepository` — kişi CRUD

**`service/interfaces.go`** — İş mantığı bağımsızlığı sağlar:
- `AuthService` — kimlik doğrulama + token yönetimi
- `BlogService` — blog iş kuralları (yetki kontrolü dahil)
- `PersonService` — kişi güncelleme iş kuralları

### 8.3. Implementasyonlar

| Interface | SQLite Implementasyonu |
|---|---|
| `AuthRepository` | `SQLiteAuthRepo` |
| `BlogRepository` | `SQLiteBlogRepo` |
| `PersonRepository` | `SQLitePersonRepo` |

Her implementasyon için `NewXxxRepo()` constructor'ı vardır ve dönüş tipi interface'tir (concrete tip sızdırmaz). Ayrıca `role_middleware.go` ve `shared/web_helpers.go` gibi inject almayan çağrılar için **paket düzeyinde wrapper fonksiyonlar** korunur.

### 8.4. Handler Struct'ları

| Struct | Bağımlılıklar |
|---|---|
| `AuthHandler` | `AuthService` |
| `BlogHandler` | `BlogService` |
| `PersonHandler` | `PersonRepository` |
| `WebHandler` | `AuthService` + `PersonRepository` + `BlogRepository` + `PersonService` |

### 8.5. Wire (Bağlama) — `routes/routes.go`

```go
authRepo   := repository.NewAuthRepo()
blogRepo   := repository.NewBlogRepo()
personRepo := repository.NewPersonRepo()

authSvc    := service.NewAuthService(authRepo, personRepo)
blogSvc    := service.NewBlogService(blogRepo)
personSvc  := service.NewPersonService(personRepo)

authH   := handlers.NewAuthHandler(authSvc)
blogH   := handlers.NewBlogHandler(blogSvc)
personH := handlers.NewPersonHandler(personRepo)
webH    := handlers.NewWebHandler(authSvc, personRepo, blogRepo, personSvc)
```

### 8.6. PostgreSQL'e Geçiş

Tüm DB değişikliği sadece 3 satır — handler, service ve shared paketi dokunulmaz:

```go
authRepo   := repository.NewPostgresAuthRepo(db)
blogRepo   := repository.NewPostgresBlogRepo(db)
personRepo := repository.NewPostgresPersonRepo(db)
```

---

## 9. Route Haritası (`routes/routes.go`)

### Web Form İşlemleri

| Method | URL | Handler | Koruma |
|---|---|---|---|
| POST | `/web-login` | `WebLoginHandler` | Herkese açık |
| POST | `/web-register` | `WebRegisterHandler` | Herkese açık |
| GET | `/web-logout` | `WebLogoutHandler` | Herkese açık |
| POST | `/user/update` | `UpdateUserHandler` | Admin |
| GET | `/web-delete` | `WebDeletePersonHandler` | Admin |

### Blog İşlemleri

| Method | URL | Handler | Koruma |
|---|---|---|---|
| POST | `/blog/create` | `CreateBlogHandler` | Editor + Admin |
| POST | `/blog/update` | `UpdateBlogHandler` | Editor + Admin |
| GET | `/blog/delete` | `DeleteBlogHandler` | Editor + Admin |

### Statik Dosyalar

| URL | Açıklama |
|---|---|
| `/static/*` | CSS, JS dosyaları |
| `/uploads/*` | Yüklenen fotoğraflar |

### REST API Endpoint'leri

| Method | URL | Handler | Koruma |
|---|---|---|---|
| POST | `/api/login` | `LoginHandler` | Herkese açık |
| POST | `/api/refresh` | `RefreshHandler` | Herkese açık |
| POST | `/api/logout` | `LogoutHandler` | JWT |
| GET | `/api/health` | `HealthHandler` | Herkese açık |
| POST | `/api/add` | `AddPersonHandler` | Herkese açık |
| GET | `/api/all` | `GetAllPeopleHandler` | JWT |
| GET | `/api/get` | `GetPersonByIDHandler` | JWT |
| GET | `/api/delete` | `DeletePersonHandler` | JWT |

### Web Sayfaları

| URL | Handler | Koruma | Açıklama |
|---|---|---|---|
| `/` | `HomeHandler` | Herkese açık | Yayınlanmış blog'lar |
| `/login` | `LoginPageHandler` | Herkese açık | Giriş formu |
| `/register` | `RegisterPageHandler` | Herkese açık | Kayıt formu |
| `/admin` | `AdminPageHandler` | Admin | Kullanıcı yönetimi |
| `/editor` | `EditorPageHandler` | Editor + Admin | Kendi blog'larını yönetir |
| `/blogs` | `BlogPageHandler` | Editor + Admin | Blog yönetim sayfası |
| `/swagger/` | Swagger UI | Herkese açık | API dokümantasyonu |
| `/debug/pprof/` | pprof UI | Herkese açık* | Profiling ve monitoring |

> *pprof endpoint'i teknik olarak koruma altında değildir. Production'a taşınırsa `AdminMiddleware` ile korunması önerilir.

---

## 10. Handler Katmanı

### 10.1. Kayıt Pipeline'ı (`registration_pipeline.go`)

Kayıt işlemi **Chain of Responsibility** deseniyle üç adımda gerçekleşir:

```
1. ensureEmailNotExistsStep  →  Email daha önce kullanıldı mı?
2. buildPersonStep           →  Şifreyi bcrypt ile hashle, Person oluştur
3. savePersonStep            →  Veritabanına kaydet, ID'yi context'e yaz
```

Herhangi bir adımda hata olursa pipeline durur ve hata yukarı taşınır.

### 10.2. Web Handler'ları (`web_handler.go` — `WebHandler` struct)

**`WebLoginHandler` (POST /web-login)**
1. Email ile kullanıcıyı bul
2. bcrypt ile şifreyi doğrula
3. Access token üret
4. `auth_token` cookie'sini set et (HttpOnly, 1 saat)
5. Role göre yönlendir: `editor` → `/editor`, diğerleri → `/admin`

**`LoginPageHandler` / `RegisterPageHandler`**
- Zaten giriş yapılmışsa role göre yönlendirir: `editor` → `/editor`, diğerleri → `/admin`

**`WebRegisterHandler` (POST /web-register)**
1. `ParseMultipartForm(32MB)` ile form'u parse et
2. DTO'yu doldur ve validator'dan geçir
3. Varsa fotoğrafı `uploads/` klasörüne yükle
4. Kayıt pipeline'ını çalıştır
5. Başarılıysa: giriş yapılmışsa `/admin`'e, yapılmamışsa `/login?registered=true`'ya yönlendir

**`WebDeletePersonHandler` (GET /web-delete)**
1. Query'den `id` al
2. `DeletePerson(id)` çağır
3. JSON `{"success": true}` yanıtı döndür (AJAX ile çağrılır, sayfa yenilenir)

**`AdminPageHandler` (GET /admin)**
1. Tüm kullanıcıları DB'den çek
2. `TemplateData.Users`'a ata
3. `admin.html` şablonunu render et

### 10.3. Blog Handler'ları (`blog_handler.go` — `BlogHandler` struct)

**`BlogPageHandler` (GET /blogs)**
- Admin ise: tüm blog'ları getirir
- Editor ise: yalnızca kendi blog'larını getirir (`GetBlogsByAuthor`)

**`EditorPageHandler` (GET /editor)**
- `BlogPageHandler` ile aynı mantık, `editor.html` şablonunu render eder

**`CreateBlogHandler` (POST /blog/create)**
1. Form parse et (multipart)
2. Cookie'den token'ı parse ederek `AuthorID` ve `AuthorName` al
3. Varsa görsel yükle
4. Blog'u DB'ye kaydet
5. Başarı cookie'si set edip `/blogs`'a yönlendir

**`UpdateBlogHandler` (POST /blog/update)**
1. `blog_id` ile mevcut blog'u çek
2. Yetki kontrolü: admin her blog'u, editor yalnızca kendi blog'unu düzenleyebilir
3. Yeni görsel varsa yükle ve eski görseli diskten sil, yoksa mevcut yolu koru
4. Blog'u güncelle, `/blogs`'a yönlendir

**`DeleteBlogHandler` (GET /blog/delete)**
1. `id` ile blog'u bul
2. Yetki kontrolü (admin her blog'u, editor yalnızca kendi blog'unu silebilir)
3. Blog'u sil, `/blogs`'a yönlendir

### 10.4. Auth API Handler'ları (`auth_handler.go` — `AuthHandler` struct)

**`LoginHandler` (POST /api/login)** — Swagger dokümantasyonlu
1. JSON body'den `LoginRequest` oku
2. Email/şifre `ValidateRequired` ile doğrula
3. `authSvc.Login` ile kimlik doğrula
4. Access + refresh token üret
5. `WriteSuccess` ile `TokenResponse` döndür

**`RefreshHandler` (POST /api/refresh)**
1. `refreshToken` oku ve DB'de geçerli olup olmadığını kontrol et
2. JWT ile parse ederek `userID` al
3. Yeni access token üret ve döndür

**`LogoutHandler` (POST /api/logout)**
1. Body'den refresh token oku
2. DB'de geçerliliğini kontrol et
3. DB'de token'ı revoke et (`revoked=1`, `revoked_at` set)

---

## 11. Repository Katmanı

### 11.1. `person_repo.go` — `SQLitePersonRepo`

| Fonksiyon | Açıklama |
|---|---|
| `AddPerson(p Person)` | Yeni kullanıcı ekler |
| `GetAllPeople()` | Tüm kullanıcıları listeler |
| `GetPersonByID(id)` | ID'ye göre kullanıcı getirir |
| `GetPersonByEmail(email)` | E-postaya göre kullanıcı getirir (login için) |
| `EmailExists(email)` | E-postanın kayıtlı olup olmadığını kontrol eder |
| `DeletePerson(id)` | ID'ye göre kullanıcı siler |
| `UpdatePerson(p Person)` | Kullanıcı bilgilerini günceller |
| `UploadPhoto(file, header)` | Fotoğrafı `uploads/` klasörüne kaydeder, URL yolunu döner |
| `DeleteUploadedFile(urlPath)` | URL yoluyla belirtilen dosyayı diskten siler |

`UploadPhoto` özellikleri:
- Sadece `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp` uzantılarına izin verir
- Rastgele 16 byte hex dosya adı üretir (çakışma engellenir)
- `uploads/` klasörü yoksa otomatik oluşturur
- Dosya yolunu `/uploads/filename.ext` formatında döner

`DeleteUploadedFile` özellikleri:
- `/uploads/...` formatındaki URL yolunu `uploads/...` dosya sistemi yoluna çevirir
- Boş path veya bulunamayan dosya için sessizce devam eder (hata fırlatmaz)

### 11.2. `blog_repo.go` — `SQLiteBlogRepo`

| Fonksiyon | Açıklama |
|---|---|
| `CreateBlog(blog)` | Yeni blog oluşturur, `LastInsertId` döner |
| `GetAllBlogs()` | Tüm blog'ları tarihe göre ters sırayla getirir |
| `GetPublishedBlogs()` | Yalnızca `published=1` olan blog'ları getirir (ana sayfa için) |
| `GetBlogByID(id)` | Tek blog getirir |
| `GetBlogsByAuthor(authorID)` | Yazara ait blog'ları getirir |
| `UpdateBlog(blog)` | Blog günceller |
| `DeleteBlog(id)` | Blog siler |
| `UpdateBlogPublishStatus(id, published)` | Yalnızca yayın durumunu günceller |

`blog_repo.go` ayrıca `scanBlogRow` ve `parseTimeStr` yardımcı fonksiyonlarını içerir; SQLite'da metin olarak saklanan tarih alanlarını `time.Time`'a çevirir (RFC3339, `2006-01-02 15:04:05`, `2006-01-02T15:04:05Z` formatlarını destekler).

### 11.3. `auth_repo.go` — `SQLiteAuthRepo`

| Fonksiyon | Açıklama |
|---|---|
| `SaveRefreshToken(userID, token)` | Yeni refresh token'ı DB'ye kaydeder |
| `IsRefreshTokenValid(token)` | Token'ın var ve revoke edilmemiş olup olmadığını kontrol eder |
| `RevokeRefreshToken(token)` | Token'ı geçersiz kılar |

---

## 12. Paylaşılan Altyapı (`shared/`)

### 12.1. CORS Middleware (`cors.go`)

- Her yanıta `Access-Control-Allow-Origin: *` ekler
- İzin verilen metodlar: `GET, POST, PUT, DELETE, OPTIONS`
- Preflight (`OPTIONS`) isteklerine `204 No Content` ile anında yanıt verir

### 12.2. Rate Limiter (`rate_limiter.go`)

IP başına sabit pencere (fixed window) sayacı ile istek sınırlandırma:

- `NewRateLimiter(limit int, window time.Duration)` ile yapılandırılır
- `sync.Mutex` kullanılarak thread-safe, dış bağımlılık yok
- Süresi dolmuş kayıtlar arka planda periyodik olarak temizlenir
- Reverse proxy arkasında çalışıyorsa `X-Forwarded-For` header'ından IP alınır, aksi hâlde `RemoteAddr` kullanılır
- Limit aşıldığında `429 Too Many Requests` döner

Şu an uygulandığı endpoint'ler (dakikada 10 istek / IP):

| Endpoint | Açıklama |
|---|---|
| `POST /web-login` | Web form girişi |
| `POST /api/login` | REST API girişi |

### 12.3. Hata Yönetimi ve Uygulama Loglama (`errors.go`)

#### Hata Tipleri

`ErrorType` string sabitleri ile hataların kategorisi belirlenir:

| Sabit | Değer | HTTP Kodu |
|---|---|---|
| `ErrorTypeValidation` | `"validation"` | 400 |
| `ErrorTypeAuth` | `"auth"` | 401 |
| `ErrorTypeNotFound` | `"not_found"` | 404 |
| `ErrorTypePermission` | `"permission"` | 403 |
| `ErrorTypeDatabase` | `"database"` | 500 |
| `ErrorTypeInternal` | `"internal"` | 500 |

#### `CustomError` Yapısı

```go
type CustomError struct {
    Type    ErrorType              `json:"type"`
    Message string                 `json:"message"`
    Code    int                    `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

#### Constructor Fonksiyonları

| Fonksiyon | Tip | HTTP Kodu |
|---|---|---|
| `NewValidationError(message, details)` | validation | 400 |
| `NewAuthError(message)` | auth | 401 |
| `NewNotFoundError(message)` | not_found | 404 |
| `NewPermissionError(message)` | permission | 403 |
| `NewDatabaseError(message)` | database | 500 |
| `NewInternalError(message)` | internal | 500 |

#### HTTP Helper Fonksiyonları

| Fonksiyon | Açıklama |
|---|---|
| `WriteError(w, status, message, err)` | JSON hata yanıtı yazar, loga kaydeder |
| `HandleError(w, err, status, message) bool` | `err != nil` ise `WriteError` çağırır ve `true` döner |
| `HandleCustomError(w, *CustomError) bool` | `CustomError` tipinde hata işler, JSON yazar |
| `WriteSuccess(w, message, data)` | `{"success": true, "message": ..., "data": ...}` JSON yanıtı yazar |

#### Sabit Hata Mesajları

Tüm sabit hata metinleri merkezi olarak tanımlanmıştır:
- `ErrInvalidRequestBody`, `ErrEmailAlreadyExists`, `ErrPersonNotFound`
- `ErrAuthHeaderRequired`, `ErrBearerTokenRequired`, `ErrInvalidOrExpiredToken`
- `ErrAccessTokenGenerateFail`, `ErrRefreshTokenGenerateFail`, `ErrInvalidRefreshToken`
- `ErrInvalidEmailFormat`, `ErrInvalidPassword`, `ErrInvalidName`, vb.

#### Uygulama Loglama Fonksiyonları

| Fonksiyon | Açıklama |
|---|---|
| `LogError(level, message, details)` | Hata seviyesinde log yazar |
| `LogInfo(level, message, details)` | Bilgi seviyesinde log yazar |
| `LogAuth(action, userID, details)` | Auth olaylarını yapılandırılmış olarak loglar |

Her log girişi hem `log.Printf` (stdout) hem de `app.log` dosyasına JSON formatında yazılır.

### 12.4. HTTP İstek Loglama Middleware (`logging.go`)

`loggingResponseWriter` wrapper'ı ile her HTTP isteği için şunları loglar:

```
http request: method=POST path=/web-login status=302 duration=45ms bytes=0
```

### 12.5. Doğrulama Katmanı (`validation.go`)

`Validator` struct ile zincir (builder) halinde doğrulama yapılır:

```go
v := shared.NewValidator()
v.ValidateName(req.Name, "İsim").ValidateEmail(req.Email).ValidatePassword(req.Password)
if v.HasError() {
    shared.HandleCustomError(w, v.GetErrorAsCustomError())
}
```

#### Temel Validator Metodları

| Metod | Kural |
|---|---|
| `ValidateName(name, fieldName)` | Boş olamaz, 2-50 karakter |
| `ValidateEmail(email)` | Regex ile format kontrolü |
| `ValidatePassword(password)` | 6-50 karakter, en az 1 büyük harf, en az 1 rakam |
| `ValidateAge(age)` | 0-150 arası |
| `ValidatePhone(phone)` | İsteğe bağlı, format regex kontrolü (10-20 karakter) |
| `ValidateRole(role)` | `admin` veya `editor` olmalı |
| `ValidateRequired(value, fieldName)` | Boş string olamaz (TrimSpace ile) |
| `ValidateMaxLength(value, fieldName, max)` | Maksimum karakter uzunluğu |
| `ValidateMinLength(value, fieldName, min)` | Minimum karakter uzunluğu |
| `ValidateNumeric(value, fieldName)` | Sayısal değer kontrolü |

#### Bileşik Validator Metodları

| Metod | Açıklama |
|---|---|
| `ValidateCreatePersonRequest(req)` | Yeni kullanıcı oluşturma için tüm alanları doğrular |
| `ValidateUpdatePersonRequest(req)` | Kullanıcı güncelleme için tüm alanları doğrular (şifre isteğe bağlı) |
| `ValidateBlogRequest(title, content, summary)` | Başlık 3-200 karakter, içerik min 10 karakter, özet max 500 karakter |

#### Sonuç Metodları

| Metod | Açıklama |
|---|---|
| `HasError() bool` | Hata var mı? |
| `GetError() string` | Hataları `"; "` ile birleştirip döner |
| `GetErrorAsCustomError() *CustomError` | `CustomError` olarak döner (HTTP handler'larla entegre) |

Yardımcı fonksiyon: `ParseIntFromForm(s string) int` — form değerini int'e çevirir, hata durumunda 0 döner.

### 12.6. Template Veri Yönetimi (`web_helpers.go`)

`GetTemplateData(r)` her sayfa handler'ının başında çağrılır:
1. `auth_token` cookie'sini okur
2. Token'ı parse eder → `IsAuthenticated = true`
3. `GetPersonByID` ile kullanıcı adı ve rolü alır
4. `success_message` cookie'sini kontrol eder (form redirect sonrası mesaj gösterimi için)
5. `?registered=true` URL parametresini kontrol eder

```go
type TemplateData struct {
    Title           string
    IsAuthenticated bool
    UserName        string
    UserRole        string
    Users           []models.PersonResponse
    Blogs           []models.BlogResponse
    ErrorMessage    string
    SuccessMessage  string
}
```

---

## 13. Arayüz Katmanı (Templates + JS)

### 13.1. Şablon Sistemi

Tüm sayfalar `layout.html` temel şablonunu kullanır. Layout içerir:
- Bootstrap 5 (CDN)
- Navbar — role göre dinamik:
  - Giriş yapılmamışsa: Giriş, Kayıt Ol
  - `admin` rolündeyse: kullanıcı adı, Admin Panel, Swagger, pprof, Çıkış
  - `editor` rolündeyse: kullanıcı adı, Editor Panel, Çıkış
- Global alert alanı — hata ve başarı mesajları, 5 saniye sonra otomatik kapanır
- `/static/js/app.js`

### 13.2. Sayfalar

| Şablon | Açıklama |
|---|---|
| `home.html` | Yayınlanmış blog'ları kart görünümünde listeler, giriş yapmadan erişilebilir |
| `login.html` | E-posta ve şifre ile giriş formu |
| `register.html` | Kayıt formu: isim, soyisim, e-posta, yaş, telefon, rol, fotoğraf, şifre |
| `admin.html` | Kullanıcı listesi tablosu + kullanıcı ekleme/düzenleme modal'ları + Blog Yönetimi bağlantısı |
| `editor.html` | Editor'ün kendi blog'larının tablosu + blog ekleme/düzenleme modal'ları |
| `blog.html` | Admin/editor için blog yönetim tablosu + modal'lar |

### 13.3. JavaScript (`static/js/app.js`)

**`deleteUser(userId)`**
- `GET /web-delete?id=userId` AJAX isteği
- Başarılıysa `location.reload()` ile sayfayı yeniler

**`editUser(userId, name, surname, email, age, phone, role, photoPath)`**
- Düzenleme modal'ının alanlarını doldurur
- Mevcut fotoğrafı önizleme olarak gösterir
- Bootstrap Modal API ile modal'ı açar

### 13.4. Blog Edit — Inline Data Attributes Yaklaşımı

Blog düzenleme için ek API çağrısı yapılmaz. Veriler sunucu taraflı render sırasında `data-*` attribute'larına gömülür:

```html
<button data-blog-id="{{.ID}}"
        data-title="{{.Title}}"
        data-content="{{.Content}}"
        data-summary="{{.Summary}}"
        data-published="{{.Published}}"
        onclick="editBlog(this)">
```

`editBlog(btn)` fonksiyonu bu attribute'ları okuyarak modal'ı doldurur. Go'nun `html/template` paketi tüm değerleri otomatik HTML-escape eder.

---

## 14. Tipik Kullanım Akışları

### 14.1. Admin Girişi

```
/login  →  POST /web-login  →  auth_token cookie set  →  /admin
```

Admin panelinde:
- Tüm kullanıcıları listeler
- Modal ile yeni kullanıcı ekler
- Modal ile kullanıcı düzenler (ad, soyad, e-posta, şifre, fotoğraf, rol)
- AJAX ile kullanıcı siler
- `/blogs` üzerinden tüm blog'ları yönetir

### 14.2. Editor Girişi

```
/login  →  POST /web-login  →  auth_token cookie set  →  /editor
```

Editor panelinde:
- Yalnızca kendi blog'larını listeler
- Modal ile yeni blog ekler (`/blog/create`)
- Modal ile blog düzenler (`/blog/update`)
- Blog siler (`/blog/delete`)

### 14.3. REST API Akışı

```
POST /api/login  →  { accessToken, refreshToken }
                         ↓
GET /api/all  (Authorization: Bearer <accessToken>)
                         ↓
POST /api/refresh  →  yeni accessToken  (token süresi dolduğunda)
                         ↓
POST /api/logout  →  refresh token DB'de revoke edilir
```

### 14.4. Blog Yayınlama

```
Editor paneli  →  "Yeni Blog Ekle" modal  →  "Yayınla" checkbox işaretle
POST /blog/create  →  published=1  →  Ana sayfada kart olarak görünür
```

---

## 15. Güvenlik Notları

- **Şifre:** Asla düz metin tutulmaz, bcrypt ile hashlenir (default cost).
- **JWT Secret'ları:** Ortam değişkenleri üzerinden yapılandırılır. Production ortamında varsayılan değerler bırakılırsa uygulama panic verir.
- **HttpOnly Cookie:** Web UI'da auth cookie `HttpOnly: true` ile set edilir; tarayıcı JavaScript'i erişemez.
- **Email Benzersizliği:** Hem uygulama katmanında (`EmailExists`) hem DB `UNIQUE` constraint'i ile çift kayıt engellenir.
- **Fotoğraf Yükleme:** Yalnızca belirli uzantılara izin verilir, rastgele dosya adı kullanılır, path traversal kontrolü yapılır. Güncelleme işlemlerinde eski dosya diskten otomatik silinir.
- **Rate Limiting:** `/api/login` ve `/web-login` endpoint'leri için IP başına dakikada 10 istek sınırı uygulanır; aşıldığında `429 Too Many Requests` döner.
- **Rol Kontrolü:** Her korumalı route hem middleware'de hem handler içinde tekrar kontrol edilir.
- **Statik Dosya Güvenliği:** `StaticHandler`, path traversal saldırılarına karşı `filepath.Abs` ile yol doğrulaması yapar.
- **CORS:** Şu an `*` (herkese açık). Production'da belirli origin'lerle kısıtlanmalıdır.

---

## 16. Gelecekte Eklenecekler

### 16.1. Kullanıcı Deneyimi
- **Blog Arama ve Filtreleme** — Başlık, yazar, tarih aralığı ve yayın durumuna göre arama
- **Sayfalama (Pagination)** — Kullanıcı listesi ve blog listesi için sayfa sistemi
- **Kullanıcı Profil Sayfası** — Her kullanıcının kendi bilgilerini görüp düzenleyeceği sayfa
- **Admin Dashboard İstatistikleri** — Toplam kullanıcı, toplam blog, taslak/yayın sayısı gibi özet kartlar
- **Blog Kategorileri ve Etiketler** — Blog'ları gruplamak ve filtrelemek için
- **Blog Detay Sayfası** — Ana sayfadaki kart'a tıklandığında açılan tam içerik sayfası

### 16.2. Altyapı ve Güvenlik
- **Şifre Sıfırlama Akışı** — E-posta ile doğrulama ve şifre yenileme
- **E-posta Doğrulama** — Kayıt sonrası e-posta onayı
- **CORS Origin Kısıtlaması** — `*` yerine belirli origin listesi
- **Production Docker Kurulumu** — `Dockerfile` ve `docker-compose.yml`
- **Prometheus + Grafana** — pprof'un yanına görsel monitoring dashboard'u (Docker Compose ile)

### 16.3. API Geliştirmeleri
- **RESTful URL Yapısı** — `/api/get?id=` yerine `GET /api/people/{id}`, `/api/delete?id=` yerine `DELETE /api/people/{id}`
- **Blog REST API** — Blog'lar için tam REST API endpoint'leri (`/api/blogs`, `/api/blogs/{id}`)
- **Token'da Rol Bilgisi** — JWT claims'e rol alanı eklenerek DB sorgusu azaltılabilir

### 16.4. İçerik ve Yönetim
- **Yorum Sistemi** — Blog yazılarına okuyucu yorumları
- **Blog Versiyonlama** — Düzenleme geçmişini tutma
- **Toplu İşlemler** — Admin panelinde birden fazla kaydı tek seferde silme/yayınlama

---

## 17. Test Altyapısı

Proje, Go'nun yerleşik `testing` paketi ile yazılmış 126 birim testiyle donatılmıştır. Harici test kütüphanesi kullanılmaz.

### 17.1. Test Dosyaları

| Dosya | Paket | Kapsam |
|---|---|---|
| `shared/validation_test.go` | `shared` | Tüm validator metodları (email, şifre, isim, yaş, telefon, rol, blog, CreatePerson, UpdatePerson) |
| `shared/errors_test.go` | `shared` | CustomError constructor'ları, WriteError, HandleError, HandleCustomError, WriteSuccess |
| `shared/rate_limiter_test.go` | `shared` | Allow(), pencere sıfırlanması, Middleware (200/429), X-Forwarded-For, RemoteAddr |
| `service/auth_service_test.go` | `service` | Login (başarılı/hatalı), token üretimi/parse, refresh kaydetme/revoke — mock repo ile |
| `handlers/auth_handler_test.go` | `handlers_test` | LoginHandler, RefreshHandler, LogoutHandler, JwtAuthMiddleware — mock service ile |
| `handlers/registration_pipeline_test.go` | `handlers` | ensureEmailNotExistsStep, buildPersonStep, savePersonStep, runRegistrationPipeline |

### 17.2. Test Mimarisi

- **Birim testler:** Her katman kendi mock bağımlılıklarıyla bağımsız test edilir. Gerçek DB veya HTTP sunucusu kullanılmaz.
- **Mock'lar:** `service/auth_service_test.go` içinde `mockAuthRepo` ve `mockPersonRepo`; `handlers/auth_handler_test.go` içinde `mockAuthService` tanımlanmıştır. Bunlar ilgili interface'leri karşılayan sade struct'lardır.
- **White-box testler:** `registration_pipeline_test.go` dosyası `package handlers` ile tanımlanmıştır, bu sayede dışa kapalı (`unexported`) pipeline fonksiyonlarına doğrudan erişilir.
- **Black-box testler:** `auth_handler_test.go` dosyası `package handlers_test` ile tanımlanmıştır; yalnızca dışa açık API test edilir.

### 17.3. Testleri Çalıştırma Komutları

Tüm komutlar `go-auth-app/` dizininden çalıştırılır.

#### Tüm testleri çalıştır

```bash
go test ./...
```

#### Belirli bir paketi test et

```bash
go test ./shared/...
go test ./service/...
go test ./handlers/...
```

#### Ayrıntılı çıktı ile çalıştır (`-v`)

```bash
go test ./... -v
```

#### Tek bir test fonksiyonunu çalıştır

```bash
go test ./shared/... -run TestValidateEmail
go test ./service/... -run TestLogin_Basarili
go test ./handlers/... -run TestLoginHandler_Basarili
```

#### Test adına göre filtrele (pattern destekler)

```bash
go test ./... -run TestLogin
go test ./... -run TestRateLimiter
```

#### Test kapsamı (coverage) raporu

```bash
# Terminalde özet
go test ./... -cover

# HTML raporu oluştur ve tarayıcıda aç
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Timeout ile çalıştır (uzun testler için)

```bash
go test ./... -timeout 60s
```

#### Paralel test çalıştırma (varsayılan zaten paralel, sayı kısıtlamak için)

```bash
go test ./... -parallel 4
```

### 17.4. Beklenen Çıktı

```
ok  	go-kisi-api/shared    ~2.6s
ok  	go-kisi-api/service   ~2.4s
ok  	go-kisi-api/handlers  ~3.4s
```

Tüm testler geçtiğinde her satır `ok` ile başlar. Herhangi bir `FAIL` görülürse `-v` bayrağıyla ayrıntılı çıktıya bakılabilir.

---

*Bu doküman projenin güncel halini yansıtır. Yeni özellikler eklendikçe güncellenmesi önerilir.*
