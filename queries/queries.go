package queries

// Bu paket, projedeki tüm ham SQL sorgularını merkezileştirmek için kullanılır.

// People tablosu ile ilgili sorgular.
const (
	InsertPerson = `
INSERT INTO people(name, surname, email, age, phone, photo_path, role, password_hash)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	SelectAllPeople = `
SELECT id, name, surname, email, age, phone, photo_path, role, password_hash
FROM people`

	SelectPersonByID = `
SELECT id, name, surname, email, age, phone, photo_path, role, password_hash
FROM people
WHERE id = ?`

	SelectPersonByEmail = `
SELECT id, name, surname, email, age, phone, photo_path, role, password_hash
FROM people
WHERE email = ?`

	SelectPersonIDByEmail = `
SELECT id
FROM people
WHERE email = ?`

	DeletePersonByID = `
DELETE FROM people
WHERE id = ?`
)

// Blogs tablosu ile ilgili sorgular.
const (
	InsertBlog = `
INSERT INTO blogs(title, content, summary, image_path, author_id, author_name, published, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	SelectAllBlogs = `
SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
FROM blogs
ORDER BY created_at DESC`

	SelectPublishedBlogs = `
SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
FROM blogs
WHERE published = 1
ORDER BY created_at DESC`

	SelectBlogByID = `
SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
FROM blogs
WHERE id = ?`

	SelectBlogsByAuthor = `
SELECT id, title, content, summary, image_path, author_id, author_name, published, created_at, updated_at
FROM blogs
WHERE author_id = ?
ORDER BY created_at DESC`

	UpdateBlogQuery = `
UPDATE blogs
SET title = ?, content = ?, summary = ?, image_path = ?, published = ?, updated_at = ?
WHERE id = ?`

	UpdateBlogPublishStatusQuery = `
UPDATE blogs
SET published = ?, updated_at = ?
WHERE id = ?`

	DeleteBlogByID = `DELETE FROM blogs WHERE id = ?`
)

// Refresh token tablosu ile ilgili sorgular.
const (
	InsertRefreshToken = `
INSERT INTO refresh_tokens(user_id, token, revoked, created_at)
VALUES (?, ?, 0, ?)`

	SelectRefreshTokenRevoked = `
SELECT revoked
FROM refresh_tokens
WHERE token = ?`

	RevokeRefreshTokenQuery = `
UPDATE refresh_tokens
SET revoked = 1, revoked_at = ?
WHERE token = ? AND revoked = 0`
)
