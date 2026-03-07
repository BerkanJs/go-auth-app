package models

// Person veritabanındaki tam kişi modelimizdir.
type Person struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Surname      string `json:"surname"`
	Email        string `json:"email"`
	Age          int    `json:"age"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"` // dışarıya asla gösterme
}

// CreatePersonRequest yeni kişi oluşturmak / kayıt olmak için kullanılan DTO.
// ID ve PasswordHash burada yok, sadece düz şifre gelir.
type CreatePersonRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// PersonResponse dışarıya döndüğümüz DTO.
type PersonResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Phone   string `json:"phone"`
}

// LoginRequest giriş için kullanılan DTO.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse access ve refresh token sonucunu taşır.
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenRequest yeni access token almak için kullanılan DTO.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ToPersonResponse bir Person entity'sini dışarıya döneceğimiz DTO'ya çevirir.
func ToPersonResponse(p Person) PersonResponse {
	return PersonResponse{
		ID:      p.ID,
		Name:    p.Name,
		Surname: p.Surname,
		Email:   p.Email,
		Age:     p.Age,
		Phone:   p.Phone,
	}
}

// ToPersonResponseList Person slice'ını PersonResponse slice'ına dönüştürür.
func ToPersonResponseList(people []Person) []PersonResponse {
	responses := make([]PersonResponse, 0, len(people))
	for _, p := range people {
		responses = append(responses, ToPersonResponse(p))
	}
	return responses
}
