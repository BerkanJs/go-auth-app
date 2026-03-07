## Go Kisi API - Postman Örnekleri

Taban adres: `http://localhost:8080`

> Not: JWT gerektiren endpoint’ler için önce `/login` ile `accessToken` al, sonra `Authorization` header’ına `Bearer <accessToken>` ekle.

---

### 1. Kayıt / Kişi Ekle

- **Method**: `POST`  
- **URL**: `http://localhost:8080/add`  
- **Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <ACCESS_TOKEN>`
- **Body (raw / JSON)**:

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

---

### 2. Login

- **Method**: `POST`  
- **URL**: `http://localhost:8080/login`  
- **Headers**:
  - `Content-Type: application/json`
- **Body (raw / JSON)**:

```json
{
  "email": "berkan@example.com",
  "password": "Sifre123!"
}
```

**Response (örnek)**:

```json
{
  "accessToken": "<ACCESS_TOKEN>",
  "refreshToken": "<REFRESH_TOKEN>"
}
```

Bu `accessToken` ve `refreshToken` değerlerini diğer isteklerde kullanacaksın.

---

### 3. Access Token Yenile (Refresh)

- **Method**: `POST`  
- **URL**: `http://localhost:8080/refresh`  
- **Headers**:
  - `Content-Type: application/json`
- **Body (raw / JSON)**:

```json
{
  "refreshToken": "<REFRESH_TOKEN>"
}
```

**Response (örnek)**:

```json
{
  "accessToken": "<YENI_ACCESS_TOKEN>",
  "refreshToken": "<AYNI_VEYA_YENI_REFRESH_TOKEN>"
}
```

Yeni `accessToken` ile korumalı endpoint’leri çağırabilirsin.

---

### 4. Logout

- **Method**: `POST`  
- **URL**: `http://localhost:8080/logout`  
- **Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <ACCESS_TOKEN>`
- **Body (raw / JSON)**:

```json
{
  "refreshToken": "<REFRESH_TOKEN>"
}
```

Bu istekten sonra ilgili `refreshToken` revoke edilir ve `/refresh` ile kullanılamaz hale gelir.

---

### 5. Tüm Kişileri Listele

- **Method**: `GET`  
- **URL**: `http://localhost:8080/all`  
- **Headers**:
  - `Authorization: Bearer <ACCESS_TOKEN>`

Body yoktur.

---

### 6. ID’ye Göre Kişi Getir

- **Method**: `GET`  
- **URL**: `http://localhost:8080/get?id=1`  
- **Headers**:
  - `Authorization: Bearer <ACCESS_TOKEN>`

Body yoktur, `id` query parametresi ile gönderilir.

---

### 7. Kişi Sil

- **Method**: `GET`  
- **URL**: `http://localhost:8080/delete?id=1`  
- **Headers**:
  - `Authorization: Bearer <ACCESS_TOKEN>`

Body yoktur, `id` query parametresi ile gönderilir.  
İleride bu endpoint `DELETE /people/{id}` şeklinde RESTful hale getirilebilir.

