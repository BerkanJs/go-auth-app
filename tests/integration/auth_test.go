package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// registerPayload, POST /api/add için JSON gövdesi.
type registerPayload struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// loginPayload, POST /api/login için JSON gövdesi.
type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse, /api/login başarılı yanıtının tam yapısı.
// shared.WriteSuccess: {"success": true, "message": "...", "data": {"accessToken": "...", "refreshToken": "..."}}
type loginResponse struct {
	Data struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
}

// doJSON, JSON body ile HTTP isteği gönderir ve yanıtı döner.
func doJSON(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("JSON encode hatası: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("istek oluşturulamadı: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("istek gönderilemedi: %v", err)
	}
	return resp
}

// --- Kayıt testleri ---

func TestRegister_GecerliKullanici_200Donerr(t *testing.T) {
	cleanTables(t)

	resp := doJSON(t, http.MethodPost, testServer.URL+"/api/add", registerPayload{
		Name:     "Ali",
		Surname:  "Veli",
		Email:    "ali@example.com",
		Age:      25,
		Phone:    "05001234567",
		Password: "Test1234",
		Role:     "editor",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("200 beklendi, alınan: %d", resp.StatusCode)
	}
}

func TestRegister_AyniEmail_400Doner(t *testing.T) {
	cleanTables(t)

	payload := registerPayload{
		Name:     "Ali",
		Surname:  "Veli",
		Email:    "tekrar@example.com",
		Password: "Test1234",
		Role:     "editor",
	}
	// İlk kayıt
	doJSON(t, http.MethodPost, testServer.URL+"/api/add", payload).Body.Close()
	// Aynı email ile ikinci kayıt
	resp := doJSON(t, http.MethodPost, testServer.URL+"/api/add", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("400 beklendi, alınan: %d", resp.StatusCode)
	}
}

// --- Login testleri ---

func TestLogin_DogruKimlik_TokenDoner(t *testing.T) {
	cleanTables(t)

	// Önce kullanıcı kaydet
	doJSON(t, http.MethodPost, testServer.URL+"/api/add", registerPayload{
		Name:     "Ali",
		Surname:  "Veli",
		Email:    "ali@example.com",
		Password: "Test1234",
		Role:     "editor",
	}).Body.Close()

	// Login
	resp := doJSON(t, http.MethodPost, testServer.URL+"/api/login", loginPayload{
		Email:    "ali@example.com",
		Password: "Test1234",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("200 beklendi, alınan: %d", resp.StatusCode)
	}

	var tokens loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		t.Fatalf("token yanıtı parse edilemedi: %v", err)
	}
	if tokens.Data.AccessToken == "" {
		t.Error("access_token boş olmamalı")
	}
	if tokens.Data.RefreshToken == "" {
		t.Error("refresh_token boş olmamalı")
	}
}

func TestLogin_YanlisŞifre_401Doner(t *testing.T) {
	cleanTables(t)

	doJSON(t, http.MethodPost, testServer.URL+"/api/add", registerPayload{
		Name:     "Ali",
		Email:    "ali@example.com",
		Password: "DogruSifre",
		Role:     "editor",
	}).Body.Close()

	resp := doJSON(t, http.MethodPost, testServer.URL+"/api/login", loginPayload{
		Email:    "ali@example.com",
		Password: "YanlisSifre",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("401 beklendi, alınan: %d", resp.StatusCode)
	}
}

func TestLogin_YokEmail_401Doner(t *testing.T) {
	cleanTables(t)

	resp := doJSON(t, http.MethodPost, testServer.URL+"/api/login", loginPayload{
		Email:    "yok@example.com",
		Password: "Test1234",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("401 beklendi, alınan: %d", resp.StatusCode)
	}
}

// --- Korumalı endpoint testleri ---

func TestGetAll_TokenSiz_401Doner(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/api/all")
	if err != nil {
		t.Fatalf("istek gönderilemedi: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("401 beklendi, alınan: %d", resp.StatusCode)
	}
}

func TestGetAll_GecerliToken_200Doner(t *testing.T) {
	cleanTables(t)

	// Kayıt + Login ile token al
	doJSON(t, http.MethodPost, testServer.URL+"/api/add", registerPayload{
		Name:     "Ali",
		Email:    "ali@example.com",
		Password: "Test1234",
		Role:     "editor",
	}).Body.Close()

	loginResp := doJSON(t, http.MethodPost, testServer.URL+"/api/login", loginPayload{
		Email:    "ali@example.com",
		Password: "Test1234",
	})
	var tokens loginResponse
	json.NewDecoder(loginResp.Body).Decode(&tokens)
	loginResp.Body.Close()

	// Token ile korumalı endpoint'e git
	req, _ := http.NewRequest(http.MethodGet, testServer.URL+"/api/all", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.Data.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("istek gönderilemedi: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("200 beklendi, alınan: %d", resp.StatusCode)
	}
}
