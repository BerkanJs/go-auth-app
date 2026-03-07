package service

import (
	"errors"

	"go-kisi-api/models"
	"go-kisi-api/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPersonNotFound   = errors.New("kullanıcı bulunamadı")
	ErrEmailTaken       = errors.New("bu email zaten kayıtlı")
	ErrPasswordHash     = errors.New("şifre işlenirken hata oluştu")
	ErrPhotoUpload      = errors.New("fotoğraf yüklenemedi")
)

// UpdatePersonRequest güncelleme için gerekli alanları taşır.
type UpdatePersonRequest struct {
	UserID      int
	Name        string
	Surname     string
	Email       string
	Age         int
	Phone       string
	Role        string
	NewPassword string    // boşsa mevcut şifre korunur
	NewPhotoPath string   // boşsa mevcut fotoğraf korunur
}

// UpdatePerson kullanıcı bilgilerini günceller.
// Email değişiyorsa benzersizlik kontrolü yapar.
// Şifre ve fotoğraf değişmemişse mevcut değerler korunur.
func UpdatePerson(req UpdatePersonRequest) error {
	existing, err := repository.GetPersonByID(req.UserID)
	if err != nil {
		return ErrPersonNotFound
	}

	if req.Email != existing.Email {
		exists, err := repository.EmailExists(req.Email)
		if err != nil {
			return err
		}
		if exists {
			return ErrEmailTaken
		}
	}

	photoPath := existing.PhotoPath
	if req.NewPhotoPath != "" {
		// Yeni fotoğraf yüklendiğinde eski fotoğrafı diskten sil
		repository.DeleteUploadedFile(existing.PhotoPath)
		photoPath = req.NewPhotoPath
	}

	passwordHash := existing.PasswordHash
	if req.NewPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return ErrPasswordHash
		}
		passwordHash = string(hashed)
	}

	updated := models.Person{
		ID:           req.UserID,
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		Age:          req.Age,
		Phone:        req.Phone,
		Role:         req.Role,
		PhotoPath:    photoPath,
		PasswordHash: passwordHash,
	}

	return repository.UpdatePerson(updated)
}
