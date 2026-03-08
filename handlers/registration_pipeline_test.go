package handlers

import (
	"context"
	"errors"
	"testing"

	"go-kisi-api/models"

	"golang.org/x/crypto/bcrypt"
)

// --- Mock PersonRepository ---

type mockPersonRepo struct {
	emailExists    bool
	emailExistsErr error
	addPersonErr   error
	addPersonID    int64
}

func (m *mockPersonRepo) EmailExists(_ context.Context, email string) (bool, error) {
	return m.emailExists, m.emailExistsErr
}
func (m *mockPersonRepo) AddPerson(_ context.Context, p models.Person) (int64, error) {
	return m.addPersonID, m.addPersonErr
}
func (m *mockPersonRepo) GetAllPeople(_ context.Context) ([]models.Person, error) {
	return nil, nil
}
func (m *mockPersonRepo) GetPersonByID(_ context.Context, id int) (models.Person, error) {
	return models.Person{}, nil
}
func (m *mockPersonRepo) GetPersonByEmail(_ context.Context, e string) (models.Person, error) {
	return models.Person{}, nil
}
func (m *mockPersonRepo) DeletePerson(_ context.Context, id int) error   { return nil }
func (m *mockPersonRepo) UpdatePerson(_ context.Context, p models.Person) error { return nil }

// --- EmailCheckHandler testleri ---

func TestEmailCheckHandler_EmailYok_Gecer(t *testing.T) {
	ctx := &registrationContext{Req: models.CreatePersonRequest{Email: "yeni@example.com"}}
	repo := &mockPersonRepo{emailExists: false}

	h := &EmailCheckHandler{}
	if err := h.Handle(context.Background(), ctx, repo); err != nil {
		t.Errorf("hata beklenmiyordu: %v", err)
	}
}

func TestEmailCheckHandler_EmailMevcut_Hata(t *testing.T) {
	ctx := &registrationContext{Req: models.CreatePersonRequest{Email: "mevcut@example.com"}}
	repo := &mockPersonRepo{emailExists: true}

	h := &EmailCheckHandler{}
	err := h.Handle(context.Background(), ctx, repo)
	if !errors.Is(err, errEmailAlreadyExists) {
		t.Errorf("errEmailAlreadyExists beklendi, alınan=%v", err)
	}
}

func TestEmailCheckHandler_RepoHatasi(t *testing.T) {
	ctx := &registrationContext{Req: models.CreatePersonRequest{Email: "test@example.com"}}
	repo := &mockPersonRepo{emailExistsErr: errors.New("db bağlantı hatası")}

	h := &EmailCheckHandler{}
	if err := h.Handle(context.Background(), ctx, repo); err == nil {
		t.Error("hata beklendi")
	}
}

// --- PersonBuildHandler testleri ---

func TestPersonBuildHandler_HashUretilir(t *testing.T) {
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Name:     "Ali",
			Surname:  "Veli",
			Email:    "ali@example.com",
			Password: "Test123",
			Role:     "editor",
		},
	}

	h := &PersonBuildHandler{}
	if err := h.Handle(context.Background(), ctx, &mockPersonRepo{}); err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}

	if ctx.Person.PasswordHash == "" {
		t.Error("PasswordHash dolu olmalı")
	}
	if ctx.Person.PasswordHash == "Test123" {
		t.Error("şifre düz metin olarak saklanmamalı")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ctx.Person.PasswordHash), []byte("Test123")); err != nil {
		t.Errorf("hash doğrulama başarısız: %v", err)
	}
}

func TestPersonBuildHandler_BilgilerAktarilir(t *testing.T) {
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Name:    "Ayşe",
			Surname: "Kaya",
			Email:   "ayse@example.com",
			Age:     28,
			Role:    "admin",
		},
	}

	h := &PersonBuildHandler{}
	h.Handle(context.Background(), ctx, &mockPersonRepo{})

	if ctx.Person.Name != "Ayşe" {
		t.Errorf("Name='Ayşe' beklendi, alınan=%q", ctx.Person.Name)
	}
	if ctx.Person.Email != "ayse@example.com" {
		t.Errorf("Email bekleniyor, alınan=%q", ctx.Person.Email)
	}
	if ctx.Person.Role != "admin" {
		t.Errorf("Role='admin' beklendi, alınan=%q", ctx.Person.Role)
	}
}

// --- PersonSaveHandler testleri ---

func TestPersonSaveHandler_Basarili_IDSetEdilir(t *testing.T) {
	ctx := &registrationContext{Person: models.Person{Name: "Ali"}}
	repo := &mockPersonRepo{addPersonID: 42}

	h := &PersonSaveHandler{}
	if err := h.Handle(context.Background(), ctx, repo); err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}
	if ctx.Person.ID != 42 {
		t.Errorf("ID=42 beklendi, alınan=%d", ctx.Person.ID)
	}
}

func TestPersonSaveHandler_RepoHatasi(t *testing.T) {
	ctx := &registrationContext{Person: models.Person{}}
	repo := &mockPersonRepo{addPersonErr: errors.New("db hatası")}

	h := &PersonSaveHandler{}
	if err := h.Handle(context.Background(), ctx, repo); err == nil {
		t.Error("hata beklendi")
	}
}

// --- Chain of Responsibility: SetNext ve zincirleme testleri ---

func TestSetNext_ZincirDogru_Baglanir(t *testing.T) {
	// SetNext, kendisine verilen handler'ı döner — zincirleme için
	emailCheck := &EmailCheckHandler{}
	personBuild := &PersonBuildHandler{}

	returned := emailCheck.SetNext(personBuild)
	if returned != personBuild {
		t.Error("SetNext, eklenen handler'ı dönmeli")
	}
}

func TestEmailCheckHandler_BasariliOlunca_NextCagirilir(t *testing.T) {
	// Email geçerse PersonBuildHandler devreye girmeli
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Email:    "yeni@example.com",
			Password: "Test123",
			Role:     "editor",
		},
	}
	repo := &mockPersonRepo{emailExists: false, addPersonID: 99}

	emailCheck := &EmailCheckHandler{}
	personBuild := &PersonBuildHandler{}
	personSave := &PersonSaveHandler{}
	emailCheck.SetNext(personBuild).SetNext(personSave)

	if err := emailCheck.Handle(context.Background(), ctx, repo); err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}
	// Tüm zincir çalıştıysa hem hash hem ID set edilmeli
	if ctx.Person.PasswordHash == "" {
		t.Error("PersonBuildHandler çalışmadı: PasswordHash boş")
	}
	if ctx.Person.ID != 99 {
		t.Errorf("PersonSaveHandler çalışmadı: ID=99 beklendi, alınan=%d", ctx.Person.ID)
	}
}

func TestEmailCheckHandler_HataOlunca_ZincirDurur(t *testing.T) {
	// Email mevcutsa zincir durmalı, sonraki handler'lar çalışmamalı
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Email:    "mevcut@example.com",
			Password: "Test123",
		},
	}
	repo := &mockPersonRepo{emailExists: true}

	chain := NewRegistrationChain()
	err := chain.Handle(context.Background(), ctx, repo)

	if !errors.Is(err, errEmailAlreadyExists) {
		t.Errorf("errEmailAlreadyExists beklendi, alınan=%v", err)
	}
	// Zincir durdu: PersonSaveHandler çalışmadı, ID 0 kalmalı
	if ctx.Person.ID != 0 {
		t.Errorf("zincir erken durduğunda ID=0 olmalı, alınan=%d", ctx.Person.ID)
	}
}

// --- runRegistrationPipeline testleri ---

func TestRunRegistrationPipeline_TamAkis_Basarili(t *testing.T) {
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Name:     "Ali",
			Surname:  "Veli",
			Email:    "ali@example.com",
			Password: "Test123",
			Role:     "editor",
		},
	}
	repo := &mockPersonRepo{emailExists: false, addPersonID: 1}

	if err := runRegistrationPipeline(context.Background(), ctx, repo); err != nil {
		t.Errorf("hata beklenmiyordu: %v", err)
	}
	if ctx.Person.ID != 1 {
		t.Errorf("ID=1 beklendi, alınan=%d", ctx.Person.ID)
	}
	if ctx.Person.PasswordHash == "" {
		t.Error("PasswordHash dolu olmalı")
	}
}

func TestRunRegistrationPipeline_EmailMevcut_Durur(t *testing.T) {
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Email:    "mevcut@example.com",
			Password: "Test123",
		},
	}
	repo := &mockPersonRepo{emailExists: true}

	err := runRegistrationPipeline(context.Background(), ctx, repo)
	if !errors.Is(err, errEmailAlreadyExists) {
		t.Errorf("errEmailAlreadyExists beklendi, alınan=%v", err)
	}
	if ctx.Person.ID != 0 {
		t.Errorf("pipeline erken durduğunda ID=0 olmalı, alınan=%d", ctx.Person.ID)
	}
}

func TestRunRegistrationPipeline_KayitHatasi(t *testing.T) {
	ctx := &registrationContext{
		Req: models.CreatePersonRequest{
			Name:     "Ali",
			Surname:  "Veli",
			Email:    "ali@example.com",
			Password: "Test123",
			Role:     "editor",
		},
	}
	repo := &mockPersonRepo{
		emailExists:  false,
		addPersonErr: errors.New("kayıt hatası"),
	}

	if err := runRegistrationPipeline(context.Background(), ctx, repo); err == nil {
		t.Error("kayıt hatası durumunda hata beklendi")
	}
}

// --- buildPersonFromCreateRequest testleri ---

func TestBuildPersonFromCreateRequest_TumAlanlar(t *testing.T) {
	req := models.CreatePersonRequest{
		Name:      "Test",
		Surname:   "User",
		Email:     "test@example.com",
		Age:       30,
		Phone:     "05001234567",
		PhotoPath: "/uploads/photo.jpg",
		Role:      "editor",
		Password:  "Test123",
	}

	person, err := buildPersonFromCreateRequest(req)
	if err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}

	if person.Name != req.Name {
		t.Errorf("Name eşleşmiyor")
	}
	if person.Email != req.Email {
		t.Errorf("Email eşleşmiyor")
	}
	if person.PhotoPath != req.PhotoPath {
		t.Errorf("PhotoPath eşleşmiyor")
	}
	if person.PasswordHash == req.Password {
		t.Error("şifre hashlenmedi")
	}
	// ID sıfır olmalı — DB'den atanacak
	if person.ID != 0 {
		t.Errorf("ID=0 beklendi, alınan=%d", person.ID)
	}
}
