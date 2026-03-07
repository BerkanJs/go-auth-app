package models

import "time"

// Rol tipleri
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
)

// Blog veritabanındaki blog modelimizdir.
type Blog struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	ImagePath   string    `json:"imagePath"`
	AuthorID    int       `json:"authorId"`
	AuthorName  string    `json:"authorName"`
	Published   bool      `json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateBlogRequest yeni blog oluşturmak için kullanılan DTO.
type CreateBlogRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Summary   string `json:"summary"`
	ImagePath string `json:"imagePath"`
	Published bool   `json:"published"`
}

// UpdateBlogRequest blog güncellemek için kullanılan DTO.
type UpdateBlogRequest struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Summary   string `json:"summary"`
	ImagePath string `json:"imagePath"`
	Published bool   `json:"published"`
}

// BlogResponse dışarıya döndüğümüz DTO.
type BlogResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	ImagePath   string    `json:"imagePath"`
	AuthorID    int       `json:"authorId"`
	AuthorName  string    `json:"authorName"`
	Published   bool      `json:"published"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ToBlogResponse bir Blog entity'sini dışarıya döneceğimiz DTO'ya çevirir.
func ToBlogResponse(blog Blog) BlogResponse {
	return BlogResponse{
		ID:          blog.ID,
		Title:       blog.Title,
		Content:     blog.Content,
		Summary:     blog.Summary,
		ImagePath:   blog.ImagePath,
		AuthorID:    blog.AuthorID,
		AuthorName:  blog.AuthorName,
		Published:   blog.Published,
		CreatedAt:   blog.CreatedAt,
		UpdatedAt:   blog.UpdatedAt,
	}
}

// ToBlogResponseList Blog slice'ını BlogResponse slice'ına dönüştürür.
func ToBlogResponseList(blogs []Blog) []BlogResponse {
	responses := make([]BlogResponse, 0, len(blogs))
	for _, blog := range blogs {
		responses = append(responses, ToBlogResponse(blog))
	}
	return responses
}
