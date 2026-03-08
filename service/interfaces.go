package service

import (
	"context"

	"go-kisi-api/models"
)

// AuthService kimlik doğrulama ve token yönetimi iş mantığını soyutlar.
// GenerateAccessToken ve ParseRefreshToken DB'ye dokunmadığı için ctx almaz.
type AuthService interface {
	Login(ctx context.Context, email, password string) (models.Person, error)
	GenerateAccessToken(userID int) (string, error)
	GenerateRefreshToken(ctx context.Context, userID int) (string, error)
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
	ParseRefreshToken(token string) (int, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

// BlogService blog CRUD iş mantığını soyutlar.
type BlogService interface {
	GetBlogsForUser(ctx context.Context, userRole string, userID int) ([]models.Blog, error)
	CreateBlog(ctx context.Context, title, content, summary, imagePath string, published bool, authorID int, authorName string) (int64, error)
	UpdateBlog(ctx context.Context, blogID int, title, content, summary, imagePath string, published bool, userRole string, userID int) error
	DeleteBlog(ctx context.Context, blogID int, userRole string, userID int) error
}

// PersonService kişi güncelleme iş mantığını soyutlar.
type PersonService interface {
	UpdatePerson(ctx context.Context, req UpdatePersonRequest) error
}
