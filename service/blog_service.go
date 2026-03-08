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
	repo repository.BlogRepository
}

func NewBlogService(repo repository.BlogRepository) BlogService {
	return &blogService{repo: repo}
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

	if userRole != "admin" && existing.AuthorID != userID {
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

	if userRole != "admin" && blog.AuthorID != userID {
		return ErrPermissionDenied
	}

	return s.repo.DeleteBlog(blogID)
}
