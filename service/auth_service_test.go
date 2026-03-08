package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-kisi-api/models"

	"golang.org/x/crypto/bcrypt"
)

// --- Mock Repository'ler ---

type mockAuthRepo struct {
	tokens        map[string]bool
	saveErr       error
	isValidErr    error
	revokeErr     error
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{tokens: make(map[string]bool)}
}

func (m *mockAuthRepo) SaveRefreshToken(_ context.Context, userID int, token string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.tokens[token] = true
	return nil
}

func (m *mockAuthRepo) IsRefreshTokenValid(_ context.Context, token string) (bool, error) {
	if m.isValidErr != nil {
		return false, m.isValidErr
	}
	valid, ok := m.tokens[token]
	return ok && valid, nil
}

func (m *mockAuthRepo) RevokeRefreshToken(_ context.Context, token string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	m.tokens[token] = false
	return nil
}

type mockPersonRepo struct {
	person  models.Person
	findErr error
}

func (m *mockPersonRepo) GetPersonByEmail(_ context.Context, email string) (models.Person, error) {
	return m.person, m.findErr
}
func (m *mockPersonRepo) AddPerson(_ context.Context, p models.Person) (int64, error) {
	return 1, nil
}
func (m *mockPersonRepo) GetAllPeople(_ context.Context) ([]models.Person, error) { return nil, nil }
func (m *mockPersonRepo) GetPersonByID(_ context.Context, id int) (models.Person, error) {
	return models.Person{}, nil
}
func (m *mockPersonRepo) EmailExists(_ context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockPersonRepo) DeletePerson(_ context.Context, id int) error   { return nil }
func (m *mockPersonRepo) UpdatePerson(_ context.Context, p models.Person) error { return nil }

// --- Yardımcı ---

func newTestAuthService(ar *mockAuthRepo, pr *mockPersonRepo) *authService {
	return &authService{
		authRepo:      ar,
		personRepo:    pr,
		accessSecret:  []byte("test-access-secret"),
		refreshSecret: []byte("test-refresh-secret"),
		accessTTL:     900 * time.Second,
		refreshTTL:    3600 * time.Second,
	}
}

// --- Login testleri ---

func TestLogin_Basarili(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Test123"), bcrypt.MinCost)
	personRepo := &mockPersonRepo{
		person: models.Person{ID: 1, Email: "test@example.com", PasswordHash: string(hash)},
	}
	svc := newTestAuthService(newMockAuthRepo(), personRepo)

	person, err := svc.Login(context.Background(), "test@example.com", "Test123")
	if err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}
	if person.ID != 1 {
		t.Errorf("ID=1 beklendi, alınan=%d", person.ID)
	}
}

func TestLogin_YanlisŞifre(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("DogruSifre1"), bcrypt.MinCost)
	personRepo := &mockPersonRepo{
		person: models.Person{ID: 1, Email: "test@example.com", PasswordHash: string(hash)},
	}
	svc := newTestAuthService(newMockAuthRepo(), personRepo)

	_, err := svc.Login(context.Background(), "test@example.com", "YanlisŞifre1")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ErrInvalidCredentials beklendi, alınan=%v", err)
	}
}

func TestLogin_KullaniciBulunamadi(t *testing.T) {
	personRepo := &mockPersonRepo{findErr: errors.New("bulunamadı")}
	svc := newTestAuthService(newMockAuthRepo(), personRepo)

	_, err := svc.Login(context.Background(), "yok@example.com", "Test123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ErrInvalidCredentials beklendi, alınan=%v", err)
	}
}

// --- GenerateAccessToken testleri ---

func TestGenerateAccessToken_TokenUretir(t *testing.T) {
	svc := newTestAuthService(newMockAuthRepo(), &mockPersonRepo{})

	token, err := svc.GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("token üretilemedi: %v", err)
	}
	if token == "" {
		t.Error("token boş olamaz")
	}
}

func TestGenerateAccessToken_FarkliIDler_FarkliTokenlar(t *testing.T) {
	svc := newTestAuthService(newMockAuthRepo(), &mockPersonRepo{})

	token1, _ := svc.GenerateAccessToken(1)
	token2, _ := svc.GenerateAccessToken(2)

	if token1 == token2 {
		t.Error("farklı userID'ler farklı tokenlar üretmeli")
	}
}

// --- GenerateRefreshToken testleri ---

func TestGenerateRefreshToken_RepoyaKaydeder(t *testing.T) {
	authRepo := newMockAuthRepo()
	svc := newTestAuthService(authRepo, &mockPersonRepo{})

	token, err := svc.GenerateRefreshToken(context.Background(), 10)
	if err != nil {
		t.Fatalf("refresh token üretilemedi: %v", err)
	}

	valid, _ := authRepo.IsRefreshTokenValid(context.Background(), token)
	if !valid {
		t.Error("token repo'ya kaydedilmedi")
	}
}

func TestGenerateRefreshToken_RepoHatasi(t *testing.T) {
	authRepo := newMockAuthRepo()
	authRepo.saveErr = errors.New("db bağlantı hatası")
	svc := newTestAuthService(authRepo, &mockPersonRepo{})

	_, err := svc.GenerateRefreshToken(context.Background(), 10)
	if err == nil {
		t.Error("repo hatası durumunda hata beklendi")
	}
}

// --- ParseRefreshToken testleri ---

func TestParseRefreshToken_Basarili(t *testing.T) {
	svc := newTestAuthService(newMockAuthRepo(), &mockPersonRepo{})

	token, _ := svc.GenerateRefreshToken(context.Background(), 99)
	userID, err := svc.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("parse hatası: %v", err)
	}
	if userID != 99 {
		t.Errorf("userID=99 beklendi, alınan=%d", userID)
	}
}

func TestParseRefreshToken_GecersizToken(t *testing.T) {
	svc := newTestAuthService(newMockAuthRepo(), &mockPersonRepo{})

	_, err := svc.ParseRefreshToken("gecersiz.token.string")
	if err == nil {
		t.Error("geçersiz token için hata beklendi")
	}
}

func TestParseRefreshToken_YanlisSecret(t *testing.T) {
	// Token farklı secret ile üretilip farklı secret ile parse edilirse hata beklenir
	svc1 := newTestAuthService(newMockAuthRepo(), &mockPersonRepo{})
	svc2 := &authService{
		authRepo:      newMockAuthRepo(),
		personRepo:    &mockPersonRepo{},
		accessSecret:  []byte("farkli-access"),
		refreshSecret: []byte("farkli-refresh"),
		accessTTL:     900 * time.Second,
		refreshTTL:    3600 * time.Second,
	}

	token, _ := svc1.GenerateRefreshToken(context.Background(), 5)
	_, err := svc2.ParseRefreshToken(token)
	if err == nil {
		t.Error("farklı secret ile parse hatası beklendi")
	}
}

// --- IsRefreshTokenValid ve RevokeRefreshToken testleri ---

func TestIsRefreshTokenValid_GecerliToken(t *testing.T) {
	authRepo := newMockAuthRepo()
	svc := newTestAuthService(authRepo, &mockPersonRepo{})

	token, _ := svc.GenerateRefreshToken(context.Background(), 5)
	valid, err := svc.IsRefreshTokenValid(context.Background(), token)
	if err != nil {
		t.Fatalf("hata beklenmiyordu: %v", err)
	}
	if !valid {
		t.Error("token geçerli olmalı")
	}
}

func TestRevokeRefreshToken_TokenGecersizKilar(t *testing.T) {
	authRepo := newMockAuthRepo()
	svc := newTestAuthService(authRepo, &mockPersonRepo{})

	token, _ := svc.GenerateRefreshToken(context.Background(), 5)

	if err := svc.RevokeRefreshToken(context.Background(), token); err != nil {
		t.Fatalf("revoke hatası: %v", err)
	}

	valid, err := svc.IsRefreshTokenValid(context.Background(), token)
	if err != nil {
		t.Fatalf("kontrol hatası: %v", err)
	}
	if valid {
		t.Error("revoke sonrası token geçersiz olmalı")
	}
}

func TestRevokeRefreshToken_RepoHatasi(t *testing.T) {
	authRepo := newMockAuthRepo()
	authRepo.revokeErr = errors.New("db hatası")
	svc := newTestAuthService(authRepo, &mockPersonRepo{})

	authRepo.tokens["test-token"] = true
	if err := svc.RevokeRefreshToken(context.Background(), "test-token"); err == nil {
		t.Error("repo hatası durumunda hata beklendi")
	}
}
