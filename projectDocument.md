## Go Kisi API - Proje Dokümantasyonu

Bu doküman, `Go Kisi API` projesinin mimarisini, katmanlarını ve akışlarını ayrıntılı şekilde açıklar.

---

## 1. Genel Bakış

- **Amaç**: Basit bir kişi yönetimi ve kimlik doğrulama (JWT + refresh token) servisi sunmak.
- **Teknolojiler**:
  - Programlama dili: **Go**
  - Web sunucusu: **net/http**
  - Veritabanı: **SQLite** (`people.db`)
  - ORM yok, **doğrudan SQL**
  - Dokümantasyon: **Swagger** (`swaggo/swag`, `swaggo/http-swagger`)
  - Kimlik doğrulama: **JWT** (`github.com/golang-jwt/jwt/v5`)
  - Şifre güvenliği: **bcrypt** (`golang.org/x/crypto/bcrypt`)

Uygulama temelde **katmanlı** bir yapıya sahiptir:

- **main**: Uygulamanın giriş noktası ve global middleware zinciri (CORS + logging).
- **routes**: HTTP endpoint’lerinin kayıt edildiği yer.
- **handlers**: HTTP isteklerinin iş mantığının yazıldığı katman.
  - `person_handler.go`: Kişi CRUD uçları.
  - `auth_handler.go`: Login, refresh, logout ve JWT işlemleri.
  - `registration_pipeline.go`: Kayıt (sign up) akışını yöneten hafif pipeline / Chain of Responsibility.
  - `health_handler.go`: Basit health check endpoint’i.
- **repository**: Veritabanı erişimi (SQL sorguları).
  - `person_repo.go`: Kişi tablosu ile ilgili sorgular.
  - `auth_repo.go`: Refresh token tablosu ile ilgili sorgular.
- **queries**: Tüm ham SQL sorgularının tutulduğu merkezi paket.
- **db**: Veritabanı bağlantısı ve tablo başlangıç kurulumu.
- **models**: Entity’ler ve DTO’lar.
- **shared**: Ortak altyapı bileşenleri (error handling, CORS, logging).
- **docs**: Swagger tarafından üretilen otomatik dokümantasyon dosyaları.

Klasör yapısı (özet):

- `main.go`
- `routes/routes.go`
- `handlers/person_handler.go`
- `handlers/auth_handler.go`
- `handlers/registration_pipeline.go`
- `handlers/health_handler.go`
- `repository/person_repo.go`
- `repository/auth_repo.go`
- `queries/queries.go`
- `models/person.go`
- `db/db.go`
- `shared/errors.go`
- `shared/cors.go`
- `shared/logging.go`
- `docs/` (swagger tarafından üretilir)
- `people.db` (SQLite veritabanı dosyası)
- `projectDocument.md` (bu doküman)
- `postmanExamples.md` (Postman istek örnekleri)

---

## 2. main katmanı (`main.go`)

- Uygulamanın giriş noktasıdır.
- Görevleri:
  - Veritabanı bağlantısını başlatmak: `db.Init()`
  - Route’ları kaydetmek: `routes.RegisterRoutes()`
  - HTTP sunucusunu ayağa kaldırmak: `http.ListenAndServe(":8080", handler)`
  - Global middleware zincirini kurmak:
    - `shared.CorsMiddleware` → CORS header’ları ve preflight cevapları.
    - `shared.LoggingMiddleware` → Tüm istekler için method, path, status, süre ve byte sayısını loglar.
  - Swagger meta verilerini sağlamak:
    - `@title`, `@version`, `@description`, `@host`, `@BasePath`
    - Güvenlik şeması tanımı:
      - `@securityDefinitions.apikey BearerAuth`
      - `@in header`, `@name Authorization`
  - Swagger dokümantasyonu için `docs` paketini import eder: `_ "go-kisi-api/docs"`

---

## 3. Route katmanı (`routes/routes.go`)

Bu katman, HTTP endpoint’lerinin URL’lere bağlandığı yerdir.

- **Swagger UI route’u**:
  - `http.Handle("/swagger/", httpSwagger.WrapHandler)`
  - Tarayıcı üzerinden Swagger UI:
    - `http://localhost:8080/swagger/index.html`

- **Auth endpoint’leri (public / yarı public)**:
  - `POST /login` → `handlers.LoginHandler`
  - `POST /refresh` → `handlers.RefreshHandler`
  - `POST /logout` → `handlers.JwtAuthMiddleware(handlers.LogoutHandler)` (JWT gerektirir)

- **Kişi endpoint’leri (JWT ile korumalı)**:
  - `POST /add` → `handlers.JwtAuthMiddleware(handlers.AddPersonHandler)`
  - `GET /all` → `handlers.JwtAuthMiddleware(handlers.GetAllPeopleHandler)`
  - `GET /get` → `handlers.JwtAuthMiddleware(handlers.GetPersonByIDHandler)`
  - `GET /delete` → `handlers.JwtAuthMiddleware(handlers.DeletePersonHandler)`

- **Monitoring / health endpoint’i**:
  - `GET /health` → `handlers.HealthHandler` (herkese açık, sadece `200 OK` ve `"OK"` döner).

Bu sayede, kişi ile ilgili tüm uçlar için **Authorization: Bearer <access_token>** zorunlu hale gelirken, `login` ve `refresh` uçları herkese açık, `logout` ise hem JWT hem geçerli bir refresh token ister.

---

## 4. Model katmanı (`models/person.go`)

### 4.1. Entity: `Person`

Veritabanındaki `people` tablosunun Go karşılığıdır:

- **Alanlar**:
  - `ID int` — birincil anahtar.
  - `Name string`
  - `Surname string`
  - `Email string`
  - `Age int`
  - `Phone string`
  - `PasswordHash string` — **yalnızca veritabanında tutulur**, dışarıya JSON olarak çıkmaz (`json:"-"`).

Bu struct doğrudan veritabanı ile çalışmak için kullanılır, dış API’ye direkt açılmaz.

### 4.2. DTO: `CreatePersonRequest`

Yeni bir kullanıcı / kişi oluşturmak için kullanılan istek modeli:

- `Name string`
- `Surname string`
- `Email string`
- `Age int`
- `Phone string`
- `Password string`

Özellikler:

- Dışarıdan düz şifre (`Password`) alınır.
- Handler katmanında bu şifre **bcrypt** ile hash’lenerek `Person.PasswordHash`’e dönüştürülür.
- ID ve PasswordHash bu DTO’da **yoktur**, bunlar backend tarafında yönetilir.

### 4.3. DTO: `PersonResponse`

Kullanıcıya döndüğümüz kişi modeli:

- `ID int`
- `Name string`
- `Surname string`
- `Email string`
- `Age int`
- `Phone string`

Özellikler:

- Şifre veya hash gibi hassas bilgiler içermez.
- Tüm kişi uçları (`/add`, `/all`, `/get`) bu modeli döner.

### 4.4. DTO: `LoginRequest`

Login için kullanılan model:

- `Email string`
- `Password string`

### 4.5. DTO: `TokenResponse`

Login ve refresh sonrası dönen token modeli:

- `AccessToken string`
- `RefreshToken string`

### 4.6. DTO: `RefreshTokenRequest`

Yeni access token almak için kullanılan model:

- `RefreshToken string`

---

## 5. Veritabanı katmanı (`db/db.go`)

Bu katman, SQLite bağlantısını yönetir ve tabloyu oluşturur.

- `DB *sql.DB` — global veritabanı bağlantısı.
- `Init()` fonksiyonu:
  - `sql.Open("sqlite3", "people.db")` ile veritabanını açar/oluşturur.
  - `CREATE TABLE IF NOT EXISTS people (...)` ile tabloyu oluşturur.

### 5.1. `people` tablosu yapısı

Tablo şeması:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `name TEXT NOT NULL`
- `surname TEXT`
- `email TEXT NOT NULL UNIQUE`
- `age INTEGER`
- `phone TEXT`
- `password_hash TEXT NOT NULL`

Özellikler:

- `email` kolonu **UNIQUE**; aynı email ile ikinci kayıt yapılamaz.
- `password_hash` düz şifre değil, **hash** değeri taşır.

### 5.2. `refresh_tokens` tablosu yapısı

Refresh token’ların durumunu (geçerli / revoke) takip etmek için kullanılır:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `user_id INTEGER NOT NULL`
- `token TEXT NOT NULL UNIQUE`
- `revoked INTEGER NOT NULL DEFAULT 0`
- `created_at TEXT NOT NULL`
- `revoked_at TEXT`

Özellikler:

- Her üretilen refresh token veritabanına kaydedilir.
- Logout işlemi (`/logout`), ilgili token için `revoked = 1` ve `revoked_at` set eder.
- `/refresh` endpoint’i önce token’ın bu tabloda **revoked olmadığına** bakar, sonra JWT doğrulaması yapar.

> Not: Şema değiştiği için eski `people.db` dosyası bu tablo yapısı ile uyuşmayabilir. Gerekirse dosya silinip uygulama yeniden başlatılarak tablolar yeni şema ile oluşturulabilir.

---

## 6. Repository katmanı (`repository/...`)

Repository katmanı, SQL sorgularını kapsüller ve handler katmanına daha soyut bir arayüz sunar. SQL metinleri `queries/queries.go` içinde merkezi olarak tutulur; repository fonksiyonları bu sabitleri kullanır.

### 6.1. `person_repo.go`

#### 6.1.1. `AddPerson(p models.Person) (int64, error)`

- Görev: Yeni bir kişi kaydı eklemek.
- SQL (özetle): `INSERT INTO people(name, surname, email, age, phone, password_hash) VALUES (?, ?, ?, ?, ?, ?)`
- Dönüş:
  - Başarılıysa eklenen satırın ID’si (`LastInsertId`).

#### 6.1.2. `GetAllPeople() ([]models.Person, error)`

- Görev: Tüm kişileri listelemek.
- SQL (özetle): `SELECT id, name, surname, email, age, phone, password_hash FROM people`
- Dönüş:
  - `[]Person` listesi.

#### 6.1.3. `GetPersonByID(id int) (models.Person, error)`

- Görev: ID’ye göre tek bir kişi döndürmek.

#### 6.1.4. `GetPersonByEmail(email string) (models.Person, error)`

- Görev: Email’e göre kullanıcıyı bulmak (login için).

#### 6.1.5. `EmailExists(email string) (bool, error)`

- Görev: Belirli bir email’in daha önce kullanılmış olup olmadığını kontrol etmek.
- Dönüş:
  - Kayıt yoksa: `(false, nil)`
  - Varsa: `(true, nil)`

#### 6.1.6. `DeletePerson(id int) error`

- Görev: ID’ye göre kişiyi silmek.

### 6.2. `auth_repo.go`

#### 6.2.1. `SaveRefreshToken(userID int, token string) error`

- Görev: Üretilen refresh token’ı `refresh_tokens` tablosuna kaydetmek.

#### 6.2.2. `IsRefreshTokenValid(token string) (bool, error)`

- Görev: Verilen refresh token’ın tabloda var ve `revoked = 0` olup olmadığını kontrol etmek.
- Dönüş:
  - Geçersiz veya bulunamadı: `(false, nil)` (veya hata).
  - Geçerli: `(true, nil)`.

#### 6.2.3. `RevokeRefreshToken(token string) error`

- Görev: Verilen refresh token’ı revoke etmek (`revoked = 1`, `revoked_at` set).

---

## 7. Handler katmanı – Kişi işlemleri (`handlers/person_handler.go`)

Bu dosya, kişi ile ilgili HTTP endpoint’lerinin iş mantığını içerir. Handler’lar sadece **HTTP ile ilgili detayları** ve **DTO dönüşümlerini** yönetir; veritabanı işlerini repository’e bırakır.

Tüm bu handler’lar, route katmanında **JWT middleware** ile korunmuştur (`JwtAuthMiddleware`).

### 7.1. `AddPersonHandler` – `POST /add`

- Amaç: Yeni kişi oluşturmak / kullanıcı kaydı.
- İstek gövdesi: `CreatePersonRequest` (JSON).
- İş akışı:
  1. Body’den DTO okunur.
  2. `EmailExists` ile email’in daha önce kullanılıp kullanılmadığı kontrol edilir.
     - Kullanılmışsa: `400` ve `"Bu email ile kayıt zaten mevcut"`.
  3. `buildPersonFromCreateRequest` fonksiyonu ile:
     - Şifre bcrypt ile hash’lenir.
     - `Person` entity’si oluşturulur.
  4. `AddPerson` ile veritabanına kaydedilir.
  5. Sonuç olarak `PersonResponse` döndürülür.

Swagger:
- `@Tags people`
- `@Security BearerAuth`
- `@Accept json`, `@Produce json`
- `@Param person body models.CreatePersonRequest true`
- `@Success 200 {object} models.PersonResponse`

### 7.2. `GetAllPeopleHandler` – `GET /all`

- Amaç: Tüm kişileri listelemek.
- Akış:
  1. `GetAllPeople` repository fonksiyonu çağrılır (`[]Person` döner).
  2. Her `Person`, `PersonResponse`’a dönüştürülür.
  3. JSON olarak `[]PersonResponse` döndürülür.

Swagger:
- `@Tags people`
- `@Security BearerAuth`
- `@Produce json`
- `@Success 200 {array} models.PersonResponse`

### 7.3. `GetPersonByIDHandler` – `GET /get?id=<id>`

- Amaç: Belirli bir ID’ye sahip kişiyi getirir.
- Akış:
  1. Query string’den `id` okunur.
  2. `GetPersonByID` repository fonksiyonu çağrılır.
  3. Kayıt yoksa `404` ve `"Kişi bulunamadı"`.
  4. Varsa `PersonResponse`’a dönüştürülür ve JSON döndürülür.

Swagger:
- `@Tags people`
- `@Security BearerAuth`
- `@Param id query int true`
- `@Success 200 {object} models.PersonResponse`

### 7.4. `DeletePersonHandler` – `GET /delete?id=<id>`

- Amaç: ID’ye göre kişiyi silmek.
- Akış:
  1. Query string’den `id` okunur.
  2. `DeletePerson(id)` çağrılır.
  3. Hata yoksa `200` döner.

> İleride iyileştirme: Bu endpoint **HTTP DELETE** metoduna ve daha RESTful bir URL yapısına (`/people/{id}`) taşınabilir.

Swagger:
- `@Tags people`
- `@Security BearerAuth`
- `@Param id query int true`

---

## 8. Handler katmanı – Kimlik doğrulama (`handlers/auth_handler.go`)

Bu dosyada:

- JWT üretimi ve doğrulaması,
- Bcrypt ile şifre hashleme ve karşılaştırma,
- Login ve refresh endpoint’leri,
- JWT tabanlı middleware

yer alır.

### 8.1. JWT ayarları

- Sabitler:
  - `jwtAccessSecret` — access token’ın imzalandığı secret.
  - `jwtRefreshSecret` — refresh token’ın imzalandığı secret.
  - `accessTokenTTL` — access token geçerlilik süresi (15 dakika).
  - `refreshTokenTTL` — refresh token geçerlilik süresi (7 gün).

> Güvenlik için bu secret değerleri ileride **environment variable** üzerinden yönetilmelidir.

### 8.2. Claims yapısı: `jwtClaims`

- `UserID int` alanı içerir.
- `jwt.RegisteredClaims` gömülü struct ile:
  - `ExpiresAt`
  - `IssuedAt`
  gibi alanlar tutulur.

### 8.3. Yardımcı fonksiyon: `buildPersonFromCreateRequest`

- Girdi: `CreatePersonRequest`
- Çıktı: `Person`
- İşler:
  - `req.Password` bcrypt ile hash’lenir.
  - Geriye `Person` döndürülür (`PasswordHash` alanı dolu).

### 8.4. Token üretimi

- `generateAccessToken(userID int) (string, error)`:
  - `jwtClaims` ile access token oluşturur.
  - Expire süresi: `accessTokenTTL`.

- `generateRefreshToken(userID int) (string, error)`:
  - Aynı şekilde refresh token oluşturur.
  - Expire süresi: `refreshTokenTTL`.

### 8.5. Token doğrulama

- `parseAccessToken(tokenStr string) (*jwtClaims, error)`:
  - Access token’ı parse eder ve secret ile doğrular.

- `parseRefreshToken(tokenStr string) (*jwtClaims, error)`:
  - Refresh token’ı parse eder ve ilgili secret ile doğrular.

### 8.6. JWT Middleware: `JwtAuthMiddleware`

- İmza: `func JwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc`
- Akış:
  1. `Authorization` header’ı okunur.
  2. `"Bearer <token>"` formatı kontrol edilir.
  3. `parseAccessToken` ile token doğrulanır.
  4. Claims içindeki `UserID` request context’ine yazılır.
  5. `next` handler çağrılır.

Hata durumları:

- Header yoksa veya format hatalıysa: `401 Unauthorized`.
- Token geçersiz veya süresi dolmuşsa: `401 Unauthorized`.

### 8.7. `LoginHandler` – `POST /login`

- Amaç: Email ve şifre ile giriş yapmak.
- İstek gövdesi: `LoginRequest` (`email`, `password`).
- Akış:
  1. Body’den DTO okunur.
  2. `GetPersonByEmail(email)` ile kullanıcı bulunmaya çalışılır.
  3. Kullanıcı yoksa veya bcrypt ile şifre doğrulanamazsa:
     - `401` ve `"Kullanıcı veya şifre hatalı"`.
  4. Başarılıysa:
     - `generateAccessToken(user.ID)`
     - `generateRefreshToken(user.ID)`
     çağrılır.
  5. `TokenResponse` ile iki token JSON olarak döndürülür.

Swagger:
- `@Tags auth`
- `@Accept json`, `@Produce json`
- `@Param credentials body models.LoginRequest`
- `@Success 200 {object} models.TokenResponse`

### 8.8. `RefreshHandler` – `POST /refresh`

- Amaç: Geçerli bir `refreshToken` ile yeni bir `accessToken` üretmek.
- İstek gövdesi: `RefreshTokenRequest`.
- Akış:
  1. Body’den `refreshToken` okunur.
  2. `parseRefreshToken` ile token doğrulanır.
  3. Claims’ten `UserID` alınır.
  4. `generateAccessToken` ile yeni access token üretilir.
  5. Yanıt olarak:

```json
{
  "accessToken": "<yeni_access_token>",
  "refreshToken": "<gönderilen_refresh_token_veya_yenisini>"
}
```

Swagger:
- `@Tags auth`
- `@Accept json`, `@Produce json`
- `@Param token body models.RefreshTokenRequest`
- `@Success 200 {object} models.TokenResponse`

---

## 9. Swagger / OpenAPI dokümantasyonu (`docs` klasörü)

Swagger dokümanları `swag init -g main.go` komutu ile otomatik üretilir:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

Swagger UI’ye erişim:

- `http://localhost:8080/swagger/index.html`

UI üzerinden:

- Tüm endpoint’leri görebilir,
- DTO şemalarını inceleyebilir,
- `Authorize` butonu ile Bearer token girip korumalı endpoint’leri test edebilirsin.

---

## 10. Tipik Akış Senaryoları

### 10.1. Kayıt (Sign up) + Login + Protected endpoint

1. **Kayıt (sign up)**:
   - Endpoint: `POST /add`
   - Body:

```json
{
  "name": "Berkan",
  "surname": "Yılmaz",
  "email": "berkan@example.com",
  "age": 30,
  "phone": "0500 000 00 00",
  "password": "Sifre123!"
}
```

2. **Login**:
   - Endpoint: `POST /login`
   - Body:

```json
{
  "email": "berkan@example.com",
  "password": "Sifre123!"
}
```

   - Response:

```json
{
  "accessToken": "<jwt_access_token>",
  "refreshToken": "<jwt_refresh_token>"
}
```

3. **Korumalı endpoint çağırma**:
   - Örneğin: `GET /all`
   - Header:
     - `Authorization: Bearer <jwt_access_token>`

4. **Token süresi dolduğunda yenileme**:
   - Endpoint: `POST /refresh`
   - Body:

```json
{
  "refreshToken": "<jwt_refresh_token>"
}
```

   - Dönen yeni `accessToken` ile tekrar korumalı endpoint’lere erişilir.

---

## 11. İyileştirme Önerileri

- **Güvenlik**:
  - JWT secret’ları kod içine gömülü değil, ortam değişkenleri (`ENV`) üzerinden okunmalı.
  - `refreshToken` değerleri veritabanında saklanarak invalidation/blacklist mekanizması eklenebilir.

- **REST Tasarımı**:
  - `/delete` endpoint’i `DELETE /people/{id}` şeklinde yeniden tasarlanabilir.
  - `/get` endpoint’i `GET /people/{id}` biçimine taşınabilir.
  - `/all` endpoint’i `GET /people` olabilir.

- **Validasyon**:
  - Body alanlarına (email formatı, şifre uzunluğu vb.) daha ayrıntılı validasyon kuralları eklenebilir.

- **Logging & Monitoring**:
  - Önemli aksiyonlar (login, başarısız login, kayıt vb.) loglanabilir.

Bu doküman, projenin mevcut halini ve ana akışlarını ayrıntılı bir şekilde özetler. Yeni özellikler eklendikçe bu dosya güncellenerek mimari bütünlük korunabilir.

