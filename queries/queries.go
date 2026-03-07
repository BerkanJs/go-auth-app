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
