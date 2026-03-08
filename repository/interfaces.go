package repository

import (
	"context"

	"go-kisi-api/models"
)

// AuthRepository token yönetimi için DB operasyonlarını soyutlar.
type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, userID int, token string) error
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

// BlogRepository blog CRUD operasyonlarını soyutlar.
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

// PersonRepository kişi CRUD operasyonlarını soyutlar.
type PersonRepository interface {
	AddPerson(ctx context.Context, p models.Person) (int64, error)
	GetAllPeople(ctx context.Context) ([]models.Person, error)
	GetPersonByID(ctx context.Context, id int) (models.Person, error)
	GetPersonByEmail(ctx context.Context, email string) (models.Person, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	DeletePerson(ctx context.Context, id int) error
	UpdatePerson(ctx context.Context, p models.Person) error
}
