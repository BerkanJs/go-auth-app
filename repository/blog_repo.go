package repository

import (
	"fmt"
	"time"

	"go-kisi-api/db"
	"go-kisi-api/models"
)

// CreateBlog yeni blog oluşturur
func CreateBlog(blog models.Blog) (int64, error) {
	result, err := db.DB.Exec(
		`INSERT INTO blogs(title, content, summary, image_path, author_id, author_name, published, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		blog.Title, blog.Content, blog.Summary, blog.ImagePath, blog.AuthorID, blog.AuthorName, blog.Published, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAllBlogs tüm blogları getirir
func GetAllBlogs() ([]models.Blog, error) {
	rows, err := db.DB.Query(`
		SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
		FROM blogs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Summary, &blog.ImagePath, &blog.AuthorID, &blog.AuthorName, &blog.Published, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}

		// Debug: Gelen zaman verisini logla
		fmt.Printf("DEBUG: Raw time strings - CreatedAt: '%s', UpdatedAt: '%s'\n", createdAtStr, updatedAtStr)

		// String zaman formatını parse et - birden fazla formatı dene
		if createdAtStr != "" {
			// Önce RFC3339 formatını dene
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as RFC3339: %v\n", createdAt)
			} else if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as SQL format: %v\n", createdAt)
			} else if createdAt, err := time.Parse("2006-01-02T15:04:05Z", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as UTC format: %v\n", createdAt)
			} else {
				fmt.Printf("DEBUG: Failed to parse CreatedAt: '%s', error: %v\n", createdAtStr, err)
			}
		}
		if updatedAtStr != "" {
			// Önce RFC3339 formatını dene
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as RFC3339: %v\n", updatedAt)
			} else if updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as SQL format: %v\n", updatedAt)
			} else if updatedAt, err := time.Parse("2006-01-02T15:04:05Z", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as UTC format: %v\n", updatedAt)
			} else {
				fmt.Printf("DEBUG: Failed to parse UpdatedAt: '%s', error: %v\n", updatedAtStr, err)
			}
		}

		blogs = append(blogs, blog)
	}
	return blogs, nil
}

// GetPublishedBlogs yayınlanmış blogları getirir
func GetPublishedBlogs() ([]models.Blog, error) {
	fmt.Printf("DEBUG: GetPublishedBlogs called\n")
	rows, err := db.DB.Query(`
		SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
		FROM blogs
		WHERE published = 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Summary, &blog.ImagePath, &blog.AuthorID, &blog.AuthorName, &blog.Published, &createdAtStr, &updatedAtStr); err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			return nil, err
		}

		// Debug: Gelen zaman verisini logla
		fmt.Printf("DEBUG: Raw time strings - CreatedAt: '%s', UpdatedAt: '%s'\n", createdAtStr, updatedAtStr)

		// String zaman formatını parse et - birden fazla formatı dene
		if createdAtStr != "" {
			// Önce RFC3339 formatını dene
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as RFC3339: %v\n", createdAt)
			} else if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as SQL format: %v\n", createdAt)
			} else if createdAt, err := time.Parse("2006-01-02T15:04:05Z", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
				fmt.Printf("DEBUG: Parsed CreatedAt as UTC format: %v\n", createdAt)
			} else {
				fmt.Printf("DEBUG: Failed to parse CreatedAt: '%s', error: %v\n", createdAtStr, err)
			}
		}
		if updatedAtStr != "" {
			// Önce RFC3339 formatını dene
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as RFC3339: %v\n", updatedAt)
			} else if updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as SQL format: %v\n", updatedAt)
			} else if updatedAt, err := time.Parse("2006-01-02T15:04:05Z", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
				fmt.Printf("DEBUG: Parsed UpdatedAt as UTC format: %v\n", updatedAt)
			} else {
				fmt.Printf("DEBUG: Failed to parse UpdatedAt: '%s', error: %v\n", updatedAtStr, err)
			}
		}

		blogs = append(blogs, blog)
	}
	fmt.Printf("DEBUG: GetPublishedBlogs returning %d blogs\n", len(blogs))
	return blogs, nil
}

// GetBlogByID ID'ye göre blog getirir
func GetBlogByID(id int) (models.Blog, error) {
	var blog models.Blog
	var createdAtStr, updatedAtStr string
	row := db.DB.QueryRow(`
		SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
		FROM blogs
		WHERE id = ?
	`, id)
	err := row.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Summary, &blog.ImagePath, &blog.AuthorID, &blog.AuthorName, &blog.Published, &createdAtStr, &updatedAtStr)

	// String zaman formatını parse et
	if err == nil {
		if createdAtStr != "" {
			if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
			}
		}
		if updatedAtStr != "" {
			if updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
			}
		}
	}

	return blog, err
}

// GetBlogsByAuthor yazarın bloglarını getirir
func GetBlogsByAuthor(authorID int) ([]models.Blog, error) {
	rows, err := db.DB.Query(`
		SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
		FROM blogs
		WHERE author_id = ?
		ORDER BY created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Summary, &blog.ImagePath, &blog.AuthorID, &blog.AuthorName, &blog.Published, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}

		// String zaman formatını parse et - birden fazla formatı dene
		if createdAtStr != "" {
			// Önce RFC3339 formatını dene
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				blog.CreatedAt = createdAt
			} else if createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
			} else if createdAt, err := time.Parse("2006-01-02T15:04:05Z", createdAtStr); err == nil {
				blog.CreatedAt = createdAt
			}
		}
		if updatedAtStr != "" {
			// Önce RFC3339 formatını dene
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
			} else if updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
			} else if updatedAt, err := time.Parse("2006-01-02T15:04:05Z", updatedAtStr); err == nil {
				blog.UpdatedAt = updatedAt
			}
		}

		blogs = append(blogs, blog)
	}
	return blogs, nil
}

// UpdateBlog blog günceller
func UpdateBlog(blog models.Blog) error {
	_, err := db.DB.Exec(`
		UPDATE blogs
		SET title = ?, content = ?, summary = ?, image_path = ?, published = ?, updated_at = ?
		WHERE id = ?
	`, blog.Title, blog.Content, blog.Summary, blog.ImagePath, blog.Published, time.Now(), blog.ID)
	return err
}

// DeleteBlog blog siler
func DeleteBlog(id int) error {
	_, err := db.DB.Exec("DELETE FROM blogs WHERE id = ?", id)
	return err
}

// UpdateBlogPublishStatus blog yayın durumunu günceller
func UpdateBlogPublishStatus(id int, published bool) error {
	_, err := db.DB.Exec(`
		UPDATE blogs
		SET published = ?, updated_at = ?
		WHERE id = ?
	`, published, time.Now(), id)
	return err
}
