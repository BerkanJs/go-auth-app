package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"go-kisi-api/models"
	"go-kisi-api/queries"
)

// SQLitePersonRepo, PersonRepository'yi SQLite üzerinde implement eder.
// db alanı constructor üzerinden enjekte edilir; global db.DB bağımlılığı yoktur.
type SQLitePersonRepo struct {
	db *sql.DB
}

// NewPersonRepo, bağımlılık enjeksiyonuyla bir PersonRepository oluşturur.
func NewPersonRepo(database *sql.DB) PersonRepository {
	return &SQLitePersonRepo{db: database}
}

func (r *SQLitePersonRepo) AddPerson(ctx context.Context, p models.Person) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		queries.InsertPerson,
		p.Name, p.Surname, p.Email, p.Age, p.Phone, p.PhotoPath, p.Role, p.PasswordHash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLitePersonRepo) GetAllPeople(ctx context.Context) ([]models.Person, error) {
	rows, err := r.db.QueryContext(ctx, queries.SelectAllPeople)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var people []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PhotoPath, &p.Role, &p.PasswordHash); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func (r *SQLitePersonRepo) GetPersonByID(ctx context.Context, id int) (models.Person, error) {
	var p models.Person
	row := r.db.QueryRowContext(ctx, queries.SelectPersonByID, id)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PhotoPath, &p.Role, &p.PasswordHash)
	return p, err
}

func (r *SQLitePersonRepo) GetPersonByEmail(ctx context.Context, email string) (models.Person, error) {
	var p models.Person
	row := r.db.QueryRowContext(ctx, queries.SelectPersonByEmail, email)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PhotoPath, &p.Role, &p.PasswordHash)
	return p, err
}

func (r *SQLitePersonRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var id int
	err := r.db.QueryRowContext(ctx, queries.SelectPersonIDByEmail, email).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *SQLitePersonRepo) DeletePerson(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, queries.DeletePersonByID, id)
	return err
}

func (r *SQLitePersonRepo) UpdatePerson(ctx context.Context, p models.Person) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE people
		SET name = ?, surname = ?, email = ?, age = ?, phone = ?, photo_path = ?, role = ?, password_hash = ?
		WHERE id = ?
	`, p.Name, p.Surname, p.Email, p.Age, p.Phone, p.PhotoPath, p.Role, p.PasswordHash, p.ID)
	return err
}

// defaultPersonRepo, web_helpers.go gibi doğrudan repo erişimi gereken yerler için kullanılır.
// SetDB() çağrısıyla başlatılır; sıfır değerde kullanımı panic'e yol açar.
var defaultPersonRepo PersonRepository

// Paket düzeyinde wrapper fonksiyonlar — shared/web_helpers.go bunları kullanır.
// Yeni kod için doğrudan PersonRepository arayüzünü tercih edin.
func AddPerson(ctx context.Context, p models.Person) (int64, error) {
	return defaultPersonRepo.AddPerson(ctx, p)
}
func GetAllPeople(ctx context.Context) ([]models.Person, error) {
	return defaultPersonRepo.GetAllPeople(ctx)
}
func GetPersonByID(ctx context.Context, id int) (models.Person, error) {
	return defaultPersonRepo.GetPersonByID(ctx, id)
}
func GetPersonByEmail(ctx context.Context, email string) (models.Person, error) {
	return defaultPersonRepo.GetPersonByEmail(ctx, email)
}
func EmailExists(ctx context.Context, email string) (bool, error) {
	return defaultPersonRepo.EmailExists(ctx, email)
}
func DeletePerson(ctx context.Context, id int) error { return defaultPersonRepo.DeletePerson(ctx, id) }
func UpdatePerson(ctx context.Context, p models.Person) error {
	return defaultPersonRepo.UpdatePerson(ctx, p)
}

// DeleteUploadedFile ve UploadPhoto dosya sistemi operasyonlarıdır; DB ile ilgisi yoktur.
// PersonRepository interface'ine dahil edilmezler, serbest fonksiyon olarak kalırlar.

func DeleteUploadedFile(urlPath string) {
	if urlPath == "" {
		return
	}
	fsPath := strings.TrimPrefix(urlPath, "/")
	os.Remove(fsPath)
}

func UploadPhoto(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		return "", fmt.Errorf("dosya uzantısı bulunamadı")
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[strings.ToLower(ext)] {
		return "", fmt.Errorf("sadece resim dosyalarına izin veriliyor")
	}

	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	fileName := fmt.Sprintf("%x%s", randomBytes, ext)

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	filePath := filepath.Join(uploadDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/" + filePath, nil
}
