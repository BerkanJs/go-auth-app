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

type registrationStep func(*registrationContext, repository.PersonRepository) error

var errEmailAlreadyExists = errors.New("email already exists")

func runRegistrationPipeline(ctx *registrationContext, repo repository.PersonRepository) error {
	steps := []registrationStep{
		ensureEmailNotExistsStep,
		buildPersonStep,
		savePersonStep,
	}
	for _, step := range steps {
		if err := step(ctx, repo); err != nil {
			return err
		}
	}
	return nil
}

func ensureEmailNotExistsStep(ctx *registrationContext, repo repository.PersonRepository) error {
	exists, err := repo.EmailExists(ctx.Req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errEmailAlreadyExists
	}
	return nil
}

func buildPersonStep(ctx *registrationContext, repo repository.PersonRepository) error {
	person, err := buildPersonFromCreateRequest(ctx.Req)
	if err != nil {
		return err
	}
	ctx.Person = person
	return nil
}

func savePersonStep(ctx *registrationContext, repo repository.PersonRepository) error {
	id, err := repo.AddPerson(ctx.Person)
	if err != nil {
		return err
	}
	ctx.Person.ID = int(id)
	return nil
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
