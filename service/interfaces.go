package service

import "go-kisi-api/models"

// AuthService kimlik doğrulama ve token yönetimi iş mantığını soyutlar.
type AuthService interface {
	Login(email, password string) (models.Person, error)
	GenerateAccessToken(userID int) (string, error)
	GenerateRefreshToken(userID int) (string, error)
	IsRefreshTokenValid(token string) (bool, error)
	ParseRefreshToken(token string) (int, error)
	RevokeRefreshToken(token string) error
}

// BlogService blog CRUD iş mantığını soyutlar.
type BlogService interface {
	GetBlogsForUser(userRole string, userID int) ([]models.Blog, error)
	CreateBlog(title, content, summary, imagePath string, published bool, authorID int, authorName string) (int64, error)
	UpdateBlog(blogID int, title, content, summary, imagePath string, published bool, userRole string, userID int) error
	DeleteBlog(blogID int, userRole string, userID int) error
}

// PersonService kişi güncelleme iş mantığını soyutlar.
type PersonService interface {
	UpdatePerson(req UpdatePersonRequest) error
}
