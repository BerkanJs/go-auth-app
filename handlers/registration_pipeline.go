package handlers

import (
	"errors"

	"go-kisi-api/models"
	"go-kisi-api/repository"

	"golang.org/x/crypto/bcrypt"
)

type registrationContext struct {
	Req    models.CreatePersonRequest
	Person models.Person
}

var errEmailAlreadyExists = errors.New("email already exists")

// RegistrationHandler, Chain of Responsibility pattern için handler arayüzü.
// Her adım bir sonrakini çağırarak zinciri ilerletir.
type RegistrationHandler interface {
	SetNext(RegistrationHandler) RegistrationHandler
	Handle(*registrationContext, repository.PersonRepository) error
}

// BaseRegistrationHandler, next zincir referansını ve handleNext yardımcısını taşır.
type BaseRegistrationHandler struct {
	next RegistrationHandler
}

// SetNext, bir sonraki handler'ı ayarlar ve onu döner (akıcı zincirleme için).
func (b *BaseRegistrationHandler) SetNext(h RegistrationHandler) RegistrationHandler {
	b.next = h
	return h
}

// handleNext, zincirde bir sonraki handler varsa onu çalıştırır.
func (b *BaseRegistrationHandler) handleNext(ctx *registrationContext, repo repository.PersonRepository) error {
	if b.next != nil {
		return b.next.Handle(ctx, repo)
	}
	return nil
}

// EmailCheckHandler, email'in daha önce kayıtlı olup olmadığını kontrol eder.
type EmailCheckHandler struct {
	BaseRegistrationHandler
}

func (h *EmailCheckHandler) Handle(ctx *registrationContext, repo repository.PersonRepository) error {
	exists, err := repo.EmailExists(ctx.Req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errEmailAlreadyExists
	}
	return h.handleNext(ctx, repo)
}

// PersonBuildHandler, şifreyi hash'leyerek Person modelini oluşturur.
type PersonBuildHandler struct {
	BaseRegistrationHandler
}

func (h *PersonBuildHandler) Handle(ctx *registrationContext, repo repository.PersonRepository) error {
	person, err := buildPersonFromCreateRequest(ctx.Req)
	if err != nil {
		return err
	}
	ctx.Person = person
	return h.handleNext(ctx, repo)
}

// PersonSaveHandler, Person'ı veritabanına kaydeder.
type PersonSaveHandler struct {
	BaseRegistrationHandler
}

func (h *PersonSaveHandler) Handle(ctx *registrationContext, repo repository.PersonRepository) error {
	id, err := repo.AddPerson(ctx.Person)
	if err != nil {
		return err
	}
	ctx.Person.ID = int(id)
	return h.handleNext(ctx, repo)
}

// NewRegistrationChain, CoR zincirini kurar ve ilk handler'ı döner.
// Sıra: EmailCheck → PersonBuild → PersonSave
func NewRegistrationChain() RegistrationHandler {
	emailCheck := &EmailCheckHandler{}
	personBuild := &PersonBuildHandler{}
	personSave := &PersonSaveHandler{}

	emailCheck.SetNext(personBuild).SetNext(personSave)

	return emailCheck
}

// runRegistrationPipeline, her çağrıda yeni bir CoR zinciri oluşturarak çalıştırır.
func runRegistrationPipeline(ctx *registrationContext, repo repository.PersonRepository) error {
	return NewRegistrationChain().Handle(ctx, repo)
}

func buildPersonFromCreateRequest(req models.CreatePersonRequest) (models.Person, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Person{}, err
	}
	return models.Person{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		Age:          req.Age,
		Phone:        req.Phone,
		PhotoPath:    req.PhotoPath,
		Role:         req.Role,
		PasswordHash: string(hashed),
	}, nil
}
