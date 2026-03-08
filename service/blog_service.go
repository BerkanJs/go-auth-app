package service

import (
	"errors"
	"time"

	"go-kisi-api/models"
	"go-kisi-api/repository"
)

var (
	ErrBlogNotFound     = errors.New("blog bulunamadı")
	ErrPermissionDenied = errors.New("bu işlem için yetkiniz yok")
)

type blogService struct {
	repo  repository.BlogRepository
	authz AuthorizationStrategy // Strategy: yetki kontrolü dışarıdan enjekte edilir
}

// NewBlogService, varsayılan olarak OwnerOrAdminStrategy kullanarak blogService oluşturur.
func NewBlogService(repo repository.BlogRepository) BlogService {
	return &blogService{repo: repo, authz: &OwnerOrAdminStrategy{}}
}

// NewBlogServiceWithStrategy, özel bir yetki stratejisiyle blogService oluşturur.
// Test veya farklı iş kuralları için kullanılır.
func NewBlogServiceWithStrategy(repo repository.BlogRepository, authz AuthorizationStrategy) BlogService {
	return &blogService{repo: repo, authz: authz}
}

func (s *blogService) GetBlogsForUser(userRole string, userID int) ([]models.Blog, error) {
	if userRole == "admin" {
		return s.repo.GetAllBlogs()
	}
	return s.repo.GetBlogsByAuthor(userID)
}

func (s *blogService) CreateBlog(title, content, summary, imagePath string, published bool, authorID int, authorName string) (int64, error) {
	blog := models.Blog{
		Title:      title,
		Content:    content,
		Summary:    summary,
		ImagePath:  imagePath,
		AuthorID:   authorID,
		AuthorName: authorName,
		Published:  published,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return s.repo.CreateBlog(blog)
}

func (s *blogService) UpdateBlog(blogID int, title, content, summary, imagePath string, published bool, userRole string, userID int) error {
	existing, err := s.repo.GetBlogByID(blogID)
	if err != nil {
		return ErrBlogNotFound
	}

	// Strategy Pattern: yetki kontrolü stratejiye devredilir, if/else zinciri kalktı
	if !s.authz.IsAuthorized(userRole, userID, existing.AuthorID) {
		return ErrPermissionDenied
	}

	if imagePath == "" {
		imagePath = existing.ImagePath
	} else {
		repository.DeleteUploadedFile(existing.ImagePath)
	}

	updated := models.Blog{
		ID:         blogID,
		Title:      title,
		Content:    content,
		Summary:    summary,
		ImagePath:  imagePath,
		AuthorID:   existing.AuthorID,
		AuthorName: existing.AuthorName,
		Published:  published,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  time.Now(),
	}
	return s.repo.UpdateBlog(updated)
}

func (s *blogService) DeleteBlog(blogID int, userRole string, userID int) error {
	blog, err := s.repo.GetBlogByID(blogID)
	if err != nil {
		return ErrBlogNotFound
	}

	// Strategy Pattern: UpdateBlog ile aynı strateji, tutarlı yetki kontrolü
	if !s.authz.IsAuthorized(userRole, userID, blog.AuthorID) {
		return ErrPermissionDenied
	}

	return s.repo.DeleteBlog(blogID)
}
