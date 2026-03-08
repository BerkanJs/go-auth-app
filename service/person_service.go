package service

import (
	"errors"

	"go-kisi-api/models"
	"go-kisi-api/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPersonNotFound = errors.New("kullanıcı bulunamadı")
	ErrEmailTaken     = errors.New("bu email zaten kayıtlı")
	ErrPasswordHash   = errors.New("şifre işlenirken hata oluştu")
	ErrPhotoUpload    = errors.New("fotoğraf yüklenemedi")
)

// UpdatePersonRequest güncelleme için gerekli alanları taşır.
type UpdatePersonRequest struct {
	UserID       int
	Name         string
	Surname      string
	Email        string
	Age          int
	Phone        string
	Role         string
	NewPassword  string
	NewPhotoPath string
}

type personService struct {
	repo repository.PersonRepository
}

func NewPersonService(repo repository.PersonRepository) PersonService {
	return &personService{repo: repo}
}

func (s *personService) UpdatePerson(req UpdatePersonRequest) error {
	existing, err := s.repo.GetPersonByID(req.UserID)
	if err != nil {
		return ErrPersonNotFound
	}

	if req.Email != existing.Email {
		exists, err := s.repo.EmailExists(req.Email)
		if err != nil {
			return err
		}
		if exists {
			return ErrEmailTaken
		}
	}

	photoPath := existing.PhotoPath
	if req.NewPhotoPath != "" {
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
	return s.repo.UpdatePerson(updated)
}
