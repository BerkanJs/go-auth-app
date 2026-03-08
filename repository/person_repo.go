package repository

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"go-kisi-api/db"
	"go-kisi-api/models"
	"go-kisi-api/queries"
)

// SQLitePersonRepo, PersonRepository'yi SQLite üzerinde implement eder.
type SQLitePersonRepo struct{}

func NewPersonRepo() PersonRepository {
	return &SQLitePersonRepo{}
}

func (r *SQLitePersonRepo) AddPerson(p models.Person) (int64, error) {
	result, err := db.DB.Exec(
		queries.InsertPerson,
		p.Name, p.Surname, p.Email, p.Age, p.Phone, p.PhotoPath, p.Role, p.PasswordHash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLitePersonRepo) GetAllPeople() ([]models.Person, error) {
	rows, err := db.DB.Query(queries.SelectAllPeople)
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

func (r *SQLitePersonRepo) GetPersonByID(id int) (models.Person, error) {
	var p models.Person
	row := db.DB.QueryRow(queries.SelectPersonByID, id)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PhotoPath, &p.Role, &p.PasswordHash)
	return p, err
}

func (r *SQLitePersonRepo) GetPersonByEmail(email string) (models.Person, error) {
	var p models.Person
	row := db.DB.QueryRow(queries.SelectPersonByEmail, email)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PhotoPath, &p.Role, &p.PasswordHash)
	return p, err
}

func (r *SQLitePersonRepo) EmailExists(email string) (bool, error) {
	var id int
	err := db.DB.QueryRow(queries.SelectPersonIDByEmail, email).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *SQLitePersonRepo) DeletePerson(id int) error {
	_, err := db.DB.Exec(queries.DeletePersonByID, id)
	return err
}

func (r *SQLitePersonRepo) UpdatePerson(p models.Person) error {
	_, err := db.DB.Exec(`
		UPDATE people
		SET name = ?, surname = ?, email = ?, age = ?, phone = ?, photo_path = ?, role = ?, password_hash = ?
		WHERE id = ?
	`, p.Name, p.Surname, p.Email, p.Age, p.Phone, p.PhotoPath, p.Role, p.PasswordHash, p.ID)
	return err
}

// defaultPersonRepo, geriye dönük uyumluluk için kullanılan paket düzeyindeki örnek.
var defaultPersonRepo PersonRepository = &SQLitePersonRepo{}

// Paket düzeyinde wrapper fonksiyonlar — shared ve diğer paketler bunları kullanmaya devam eder.
func AddPerson(p models.Person) (int64, error)              { return defaultPersonRepo.AddPerson(p) }
func GetAllPeople() ([]models.Person, error)               { return defaultPersonRepo.GetAllPeople() }
func GetPersonByID(id int) (models.Person, error)          { return defaultPersonRepo.GetPersonByID(id) }
func GetPersonByEmail(email string) (models.Person, error) { return defaultPersonRepo.GetPersonByEmail(email) }
func EmailExists(email string) (bool, error)               { return defaultPersonRepo.EmailExists(email) }
func DeletePerson(id int) error                            { return defaultPersonRepo.DeletePerson(id) }
func UpdatePerson(p models.Person) error                   { return defaultPersonRepo.UpdatePerson(p) }

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
