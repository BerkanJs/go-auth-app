package repository

import (
	"context"
	"database/sql"
	"time"

	"go-kisi-api/models"
	"go-kisi-api/queries"
)

// SQLiteBlogRepo, BlogRepository'yi SQLite üzerinde implement eder.
// db alanı constructor üzerinden enjekte edilir; global db.DB bağımlılığı yoktur.
type SQLiteBlogRepo struct {
	db *sql.DB
}

// NewBlogRepo, bağımlılık enjeksiyonuyla bir BlogRepository oluşturur.
func NewBlogRepo(database *sql.DB) BlogRepository {
	return &SQLiteBlogRepo{db: database}
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

func (r *SQLiteBlogRepo) CreateBlog(ctx context.Context, blog models.Blog) (int64, error) {
	result, err := r.db.ExecContext(ctx,
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

func (r *SQLiteBlogRepo) GetAllBlogs(ctx context.Context) ([]models.Blog, error) {
	rows, err := r.db.QueryContext(ctx, queries.SelectAllBlogs)
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

func (r *SQLiteBlogRepo) GetPublishedBlogs(ctx context.Context) ([]models.Blog, error) {
	rows, err := r.db.QueryContext(ctx, queries.SelectPublishedBlogs)
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

func (r *SQLiteBlogRepo) GetBlogByID(ctx context.Context, id int) (models.Blog, error) {
	row := r.db.QueryRowContext(ctx, queries.SelectBlogByID, id)
	return scanBlogRow(row)
}

func (r *SQLiteBlogRepo) GetBlogsByAuthor(ctx context.Context, authorID int) ([]models.Blog, error) {
	rows, err := r.db.QueryContext(ctx, queries.SelectBlogsByAuthor, authorID)
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

func (r *SQLiteBlogRepo) UpdateBlog(ctx context.Context, blog models.Blog) error {
	_, err := r.db.ExecContext(ctx,
		queries.UpdateBlogQuery,
		blog.Title, blog.Content, blog.Summary, blog.ImagePath,
		blog.Published, time.Now(), blog.ID,
	)
	return err
}

func (r *SQLiteBlogRepo) DeleteBlog(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, queries.DeleteBlogByID, id)
	return err
}

func (r *SQLiteBlogRepo) UpdateBlogPublishStatus(ctx context.Context, id int, published bool) error {
	_, err := r.db.ExecContext(ctx, queries.UpdateBlogPublishStatusQuery, published, time.Now(), id)
	return err
}

// defaultBlogRepo, paket düzeyinde wrapper fonksiyonlar için kullanılır.
// SetDB() çağrısıyla başlatılır; yeni kod için doğrudan BlogRepository arayüzünü tercih edin.
var defaultBlogRepo BlogRepository

// Paket düzeyinde wrapper fonksiyonlar — geriye dönük uyumluluk için korunur.
func CreateBlog(ctx context.Context, blog models.Blog) (int64, error) {
	return defaultBlogRepo.CreateBlog(ctx, blog)
}
func GetAllBlogs(ctx context.Context) ([]models.Blog, error) {
	return defaultBlogRepo.GetAllBlogs(ctx)
}
func GetPublishedBlogs(ctx context.Context) ([]models.Blog, error) {
	return defaultBlogRepo.GetPublishedBlogs(ctx)
}
func GetBlogByID(ctx context.Context, id int) (models.Blog, error) {
	return defaultBlogRepo.GetBlogByID(ctx, id)
}
func GetBlogsByAuthor(ctx context.Context, id int) ([]models.Blog, error) {
	return defaultBlogRepo.GetBlogsByAuthor(ctx, id)
}
func UpdateBlog(ctx context.Context, blog models.Blog) error {
	return defaultBlogRepo.UpdateBlog(ctx, blog)
}
func DeleteBlog(ctx context.Context, id int) error { return defaultBlogRepo.DeleteBlog(ctx, id) }
func UpdateBlogPublishStatus(ctx context.Context, id int, p bool) error {
	return defaultBlogRepo.UpdateBlogPublishStatus(ctx, id, p)
}
