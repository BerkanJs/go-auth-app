package handlers

import (
	"errors"

	"go-kisi-api/models"
	"go-kisi-api/repository"
)

// registrationContext kayıt akışı boyunca taşınan veriyi tutar.
type registrationContext struct {
	Req    models.CreatePersonRequest
	Person models.Person
}

// registrationStep zincirdeki her adımın imzası.
type registrationStep func(*registrationContext) error

// Domain seviyesinde özel hatalar.
var (
	errEmailAlreadyExists = errors.New("email already exists")
)

// runRegistrationPipeline kayıt akışını adım adım çalıştırır.
func runRegistrationPipeline(ctx *registrationContext) error {
	steps := []registrationStep{
		ensureEmailNotExistsStep,
		buildPersonStep,
		savePersonStep,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ensureEmailNotExistsStep email'in benzersiz olmasını garanti eder.
func ensureEmailNotExistsStep(ctx *registrationContext) error {
	exists, err := repository.EmailExists(ctx.Req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errEmailAlreadyExists
	}
	return nil
}

// buildPersonStep, CreatePersonRequest'ten Person entity'si üretir.
func buildPersonStep(ctx *registrationContext) error {
	person, err := buildPersonFromCreateRequest(ctx.Req)
	if err != nil {
		return err
	}
	ctx.Person = person
	return nil
}

// savePersonStep, kişiyi veritabanına kaydeder ve ID'yi context'e yazar.
func savePersonStep(ctx *registrationContext) error {
	id, err := repository.AddPerson(ctx.Person)
	if err != nil {
		return err
	}
	ctx.Person.ID = int(id)
	return nil
}

