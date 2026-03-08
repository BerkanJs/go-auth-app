package service

import (
	"context"
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

func (s *personService) CreatePerson(ctx context.Context, req models.CreatePersonRequest) (models.Person, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return models.Person{}, err
	}
	if exists {
		return models.Person{}, ErrEmailTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Person{}, ErrPasswordHash
	}

	p := models.Person{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		Age:          req.Age,
		Phone:        req.Phone,
		PhotoPath:    req.PhotoPath,
		Role:         req.Role,
		PasswordHash: string(hashed),
	}

	id, err := s.repo.AddPerson(ctx, p)
	if err != nil {
		return models.Person{}, err
	}
	p.ID = int(id)
	return p, nil
}

func (s *personService) GetAllPeople(ctx context.Context) ([]models.Person, error) {
	return s.repo.GetAllPeople(ctx)
}

func (s *personService) GetPersonByID(ctx context.Context, id int) (models.Person, error) {
	p, err := s.repo.GetPersonByID(ctx, id)
	if err != nil {
		return models.Person{}, ErrPersonNotFound
	}
	return p, nil
}

func (s *personService) DeletePerson(ctx context.Context, id int) error {
	return s.repo.DeletePerson(ctx, id)
}

func (s *personService) UpdatePerson(ctx context.Context, req UpdatePersonRequest) error {
	existing, err := s.repo.GetPersonByID(ctx, req.UserID)
	if err != nil {
		return ErrPersonNotFound
	}

	if req.Email != existing.Email {
		exists, err := s.repo.EmailExists(ctx, req.Email)
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
	return s.repo.UpdatePerson(ctx, updated)
}
