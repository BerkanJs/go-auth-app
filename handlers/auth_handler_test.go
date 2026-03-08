package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-kisi-api/handlers"
	"go-kisi-api/models"
)

// --- Mock AuthService ---

type mockAuthService struct {
	loginPerson   models.Person
	loginErr      error
	accessToken   string
	accessErr     error
	refreshToken  string
	refreshErr    error
	isValidResult bool
	isValidErr    error
	parseUserID   int
	parseErr      error
	revokeErr     error
}

func (m *mockAuthService) Login(email, password string) (models.Person, error) {
	return m.loginPerson, m.loginErr
}
func (m *mockAuthService) GenerateAccessToken(userID int) (string, error) {
	return m.accessToken, m.accessErr
}
func (m *mockAuthService) GenerateRefreshToken(userID int) (string, error) {
	return m.refreshToken, m.refreshErr
}
func (m *mockAuthService) IsRefreshTokenValid(token string) (bool, error) {
	return m.isValidResult, m.isValidErr
}
func (m *mockAuthService) ParseRefreshToken(token string) (int, error) {
	return m.parseUserID, m.parseErr
}
func (m *mockAuthService) RevokeRefreshToken(token string) error {
	return m.revokeErr
}

// --- Yardımcı ---

func postJSON(t *testing.T, handler http.HandlerFunc, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

// --- LoginHandler testleri ---

func TestLoginHandler_Basarili(t *testing.T) {
	svc := &mockAuthService{
		loginPerson:  models.Person{ID: 1, Email: "test@example.com"},
		accessToken:  "access-token",
		refreshToken: "refresh-token",
	}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LoginHandler, models.LoginRequest{
		Email:    "test@example.com",
		Password: "Test123",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status=200 beklendi, alınan=%d, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("success=true beklendi, alınan=%v", resp["success"])
	}
}

func TestLoginHandler_GecersizKimlik(t *testing.T) {
	svc := &mockAuthService{loginErr: errors.New("geçersiz kimlik")}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LoginHandler, models.LoginRequest{
		Email:    "test@example.com",
		Password: "YanlisŞifre",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=401 beklendi, alınan=%d", w.Code)
	}
}

func TestLoginHandler_BosAlanlar(t *testing.T) {
	svc := &mockAuthService{}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LoginHandler, models.LoginRequest{Email: "", Password: ""})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status=400 beklendi, alınan=%d", w.Code)
	}
}

func TestLoginHandler_GecersizJSON(t *testing.T) {
	svc := &mockAuthService{}
	h := handlers.NewAuthHandler(svc)

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString("bu-json-degil"))
	w := httptest.NewRecorder()
	h.LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status=400 beklendi, alınan=%d", w.Code)
	}
}

func TestLoginHandler_AccessTokenUretimHatasi(t *testing.T) {
	svc := &mockAuthService{
		loginPerson: models.Person{ID: 1},
		accessErr:   errors.New("token üretilemedi"),
	}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LoginHandler, models.LoginRequest{
		Email:    "test@example.com",
		Password: "Test123",
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status=500 beklendi, alınan=%d", w.Code)
	}
}

// --- RefreshHandler testleri ---

func TestRefreshHandler_Basarili(t *testing.T) {
	svc := &mockAuthService{
		isValidResult: true,
		parseUserID:   5,
		accessToken:   "yeni-access-token",
	}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.RefreshHandler, models.RefreshTokenRequest{RefreshToken: "gecerli-refresh"})

	if w.Code != http.StatusOK {
		t.Errorf("status=200 beklendi, alınan=%d, body=%s", w.Code, w.Body.String())
	}
}

func TestRefreshHandler_GecersizToken(t *testing.T) {
	svc := &mockAuthService{isValidResult: false}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.RefreshHandler, models.RefreshTokenRequest{RefreshToken: "gecersiz"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=401 beklendi, alınan=%d", w.Code)
	}
}

// --- LogoutHandler testleri ---

func TestLogoutHandler_Basarili(t *testing.T) {
	svc := &mockAuthService{isValidResult: true}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LogoutHandler, models.RefreshTokenRequest{RefreshToken: "gecerli-token"})

	if w.Code != http.StatusOK {
		t.Errorf("status=200 beklendi, alınan=%d", w.Code)
	}
}

func TestLogoutHandler_GecersizToken(t *testing.T) {
	svc := &mockAuthService{isValidResult: false}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LogoutHandler, models.RefreshTokenRequest{RefreshToken: "gecersiz-token"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=401 beklendi, alınan=%d", w.Code)
	}
}

func TestLogoutHandler_RevokeHatasi(t *testing.T) {
	svc := &mockAuthService{
		isValidResult: true,
		revokeErr:     errors.New("db hatası"),
	}
	h := handlers.NewAuthHandler(svc)

	w := postJSON(t, h.LogoutHandler, models.RefreshTokenRequest{RefreshToken: "gecerli-token"})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status=500 beklendi, alınan=%d", w.Code)
	}
}

// --- JwtAuthMiddleware testleri ---

func TestJwtAuthMiddleware_HeaderYok(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := handlers.JwtAuthMiddleware(next)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=401 beklendi, alınan=%d", w.Code)
	}
}

func TestJwtAuthMiddleware_GecersizFormat(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := handlers.JwtAuthMiddleware(next)

	tests := []struct {
		name   string
		header string
	}{
		{"Bearer yok", "sadece-token"},
		{"boşluk yok", "Bearertoken"},
		{"Basic kullanıldı", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()
			handler(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("status=401 beklendi, alınan=%d", w.Code)
			}
		})
	}
}

func TestJwtAuthMiddleware_GecersizToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := handlers.JwtAuthMiddleware(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer gecersiz.jwt.token")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=401 beklendi, alınan=%d", w.Code)
	}
}
