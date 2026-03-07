package repository

import (
	"time"

	"go-kisi-api/db"
	"go-kisi-api/models"
	"go-kisi-api/queries"
)

// parseTimeStr SQLite'tan gelen string zamanı time.Time'a çevirir.
// Birden fazla format denenir; hiçbiri uymazsa zero time döner.
func parseTimeStr(s string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// scanBlogRows tek bir satırı blog struct'ına dönüştürür.
// *sql.Rows ve *sql.Row her ikisi de Scan metoduna sahiptir; interface ile soyutlandı.
func scanBlogRow(scanner interface {
	Scan(...interface{}) error
}) (models.Blog, error) {
	var blog models.Blog
	var createdAtStr, updatedAtStr string
	err := scanner.Scan(
		&blog.ID, &blog.Title, &blog.Content, &blog.Summary,
		&blog.ImagePath, &blog.AuthorID, &blog.AuthorName,
		&blog.Published, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		return models.Blog{}, err
	}
	blog.CreatedAt = parseTimeStr(createdAtStr)
	blog.UpdatedAt = parseTimeStr(updatedAtStr)
	return blog, nil
}

// CreateBlog yeni blog oluşturur
func CreateBlog(blog models.Blog) (int64, error) {
	result, err := db.DB.Exec(
		queries.InsertBlog,
		blog.Title, blog.Content, blog.Summary, blog.ImagePath,
		blog.AuthorID, blog.AuthorName, blog.Published,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAllBlogs tüm blogları getirir
func GetAllBlogs() ([]models.Blog, error) {
	rows, err := db.DB.Query(queries.SelectAllBlogs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		blog, err := scanBlogRow(rows)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	return blogs, nil
}

// GetPublishedBlogs yayınlanmış blogları getirir
func GetPublishedBlogs() ([]models.Blog, error) {
	rows, err := db.DB.Query(queries.SelectPublishedBlogs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		blog, err := scanBlogRow(rows)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	return blogs, nil
}

// GetBlogByID ID'ye göre blog getirir
func GetBlogByID(id int) (models.Blog, error) {
	row := db.DB.QueryRow(queries.SelectBlogByID, id)
	return scanBlogRow(row)
}

// GetBlogsByAuthor yazarın bloglarını getirir
func GetBlogsByAuthor(authorID int) ([]models.Blog, error) {
	rows, err := db.DB.Query(queries.SelectBlogsByAuthor, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		blog, err := scanBlogRow(rows)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	return blogs, nil
}

// UpdateBlog blog günceller
func UpdateBlog(blog models.Blog) error {
	_, err := db.DB.Exec(
		queries.UpdateBlogQuery,
		blog.Title, blog.Content, blog.Summary, blog.ImagePath,
		blog.Published, time.Now(), blog.ID,
	)
	return err
}

// DeleteBlog blog siler
func DeleteBlog(id int) error {
	_, err := db.DB.Exec(queries.DeleteBlogByID, id)
	return err
}

// UpdateBlogPublishStatus blog yayın durumunu günceller
func UpdateBlogPublishStatus(id int, published bool) error {
	_, err := db.DB.Exec(queries.UpdateBlogPublishStatusQuery, published, time.Now(), id)
	return err
}
