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

// GetBlogsForUser role göre blogları getirir.
// Admin tüm blogları, editor yalnızca kendi bloglarını görür.
func GetBlogsForUser(userRole string, userID int) ([]models.Blog, error) {
	if userRole == "admin" {
		return repository.GetAllBlogs()
	}
	return repository.GetBlogsByAuthor(userID)
}

// CreateBlog yeni blog oluşturur ve ID'yi döner.
func CreateBlog(title, content, summary, imagePath string, published bool, authorID int, authorName string) (int64, error) {
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
	return repository.CreateBlog(blog)
}

// UpdateBlog blogu günceller; yetki kontrolü de yapar.
// imagePath boş gelirse mevcut görsel korunur.
func UpdateBlog(blogID int, title, content, summary, imagePath string, published bool, userRole string, userID int) error {
	existing, err := repository.GetBlogByID(blogID)
	if err != nil {
		return ErrBlogNotFound
	}

	if userRole != "admin" && existing.AuthorID != userID {
		return ErrPermissionDenied
	}

	if imagePath == "" {
		imagePath = existing.ImagePath
	} else {
		// Yeni görsel yüklendiğinde eski görseli diskten sil
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
	return repository.UpdateBlog(updated)
}

// DeleteBlog blogu siler; yetki kontrolü de yapar.
func DeleteBlog(blogID int, userRole string, userID int) error {
	blog, err := repository.GetBlogByID(blogID)
	if err != nil {
		return ErrBlogNotFound
	}

	if userRole != "admin" && blog.AuthorID != userID {
		return ErrPermissionDenied
	}

	return repository.DeleteBlog(blogID)
}
