package repository

import (
	"time"

	"go-kisi-api/db"
	"go-kisi-api/models"
	"go-kisi-api/queries"
)

// SQLiteBlogRepo, BlogRepository'yi SQLite üzerinde implement eder.
type SQLiteBlogRepo struct{}

func NewBlogRepo() BlogRepository {
	return &SQLiteBlogRepo{}
}

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

func (r *SQLiteBlogRepo) CreateBlog(blog models.Blog) (int64, error) {
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

func (r *SQLiteBlogRepo) GetAllBlogs() ([]models.Blog, error) {
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

func (r *SQLiteBlogRepo) GetPublishedBlogs() ([]models.Blog, error) {
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

func (r *SQLiteBlogRepo) GetBlogByID(id int) (models.Blog, error) {
	row := db.DB.QueryRow(queries.SelectBlogByID, id)
	return scanBlogRow(row)
}

func (r *SQLiteBlogRepo) GetBlogsByAuthor(authorID int) ([]models.Blog, error) {
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

func (r *SQLiteBlogRepo) UpdateBlog(blog models.Blog) error {
	_, err := db.DB.Exec(
		queries.UpdateBlogQuery,
		blog.Title, blog.Content, blog.Summary, blog.ImagePath,
		blog.Published, time.Now(), blog.ID,
	)
	return err
}

func (r *SQLiteBlogRepo) DeleteBlog(id int) error {
	_, err := db.DB.Exec(queries.DeleteBlogByID, id)
	return err
}

func (r *SQLiteBlogRepo) UpdateBlogPublishStatus(id int, published bool) error {
	_, err := db.DB.Exec(queries.UpdateBlogPublishStatusQuery, published, time.Now(), id)
	return err
}

// defaultBlogRepo, geriye dönük uyumluluk için kullanılan paket düzeyindeki örnek.
var defaultBlogRepo BlogRepository = &SQLiteBlogRepo{}

// Paket düzeyinde wrapper fonksiyonlar.
func CreateBlog(blog models.Blog) (int64, error)             { return defaultBlogRepo.CreateBlog(blog) }
func GetAllBlogs() ([]models.Blog, error)                    { return defaultBlogRepo.GetAllBlogs() }
func GetPublishedBlogs() ([]models.Blog, error)              { return defaultBlogRepo.GetPublishedBlogs() }
func GetBlogByID(id int) (models.Blog, error)                { return defaultBlogRepo.GetBlogByID(id) }
func GetBlogsByAuthor(id int) ([]models.Blog, error)         { return defaultBlogRepo.GetBlogsByAuthor(id) }
func UpdateBlog(blog models.Blog) error                      { return defaultBlogRepo.UpdateBlog(blog) }
func DeleteBlog(id int) error                                { return defaultBlogRepo.DeleteBlog(id) }
func UpdateBlogPublishStatus(id int, published bool) error   { return defaultBlogRepo.UpdateBlogPublishStatus(id, published) }
