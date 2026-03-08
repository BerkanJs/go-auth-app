package repository

import "go-kisi-api/models"

// AuthRepository token yönetimi için DB operasyonlarını soyutlar.
type AuthRepository interface {
	SaveRefreshToken(userID int, token string) error
	IsRefreshTokenValid(token string) (bool, error)
	RevokeRefreshToken(token string) error
}

// BlogRepository blog CRUD operasyonlarını soyutlar.
type BlogRepository interface {
	CreateBlog(blog models.Blog) (int64, error)
	GetAllBlogs() ([]models.Blog, error)
	GetPublishedBlogs() ([]models.Blog, error)
	GetBlogByID(id int) (models.Blog, error)
	GetBlogsByAuthor(authorID int) ([]models.Blog, error)
	UpdateBlog(blog models.Blog) error
	DeleteBlog(id int) error
	UpdateBlogPublishStatus(id int, published bool) error
}

// PersonRepository kişi CRUD operasyonlarını soyutlar.
type PersonRepository interface {
	AddPerson(p models.Person) (int64, error)
	GetAllPeople() ([]models.Person, error)
	GetPersonByID(id int) (models.Person, error)
	GetPersonByEmail(email string) (models.Person, error)
	EmailExists(email string) (bool, error)
	DeletePerson(id int) error
	UpdatePerson(p models.Person) error
}
