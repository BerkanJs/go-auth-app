package shared

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *CustomError
		wantType ErrorType
		wantCode int
	}{
		{"validation", NewValidationError("test", nil), ErrorTypeValidation, 400},
		{"auth", NewAuthError("test"), ErrorTypeAuth, 401},
		{"not_found", NewNotFoundError("test"), ErrorTypeNotFound, 404},
		{"permission", NewPermissionError("test"), ErrorTypePermission, 403},
		{"database", NewDatabaseError("test"), ErrorTypeDatabase, 500},
		{"internal", NewInternalError("test"), ErrorTypeInternal, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Type != tt.wantType {
				t.Errorf("Type=%q, beklenen=%q", tt.err.Type, tt.wantType)
			}
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code=%d, beklenen=%d", tt.err.Code, tt.wantCode)
			}
			if tt.err.Message != "test" {
				t.Errorf("Message=%q, beklenen 'test'", tt.err.Message)
			}
			if tt.err.Error() != "test" {
				t.Errorf("Error()=%q, beklenen 'test'", tt.err.Error())
			}
		})
	}
}

func TestValidationError_DetailsAktarilir(t *testing.T) {
	details := map[string]interface{}{"field": "email", "reason": "format hatalı"}
	err := NewValidationError("hata", details)

	if err.Details["field"] != "email" {
		t.Errorf("details['field'] bekleniyor 'email', alınan=%v", err.Details["field"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "test hata", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status=%d, beklenen=400", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type bekleniyor: application/json, alınan: %s", w.Header().Get("Content-Type"))
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON parse hatası: %v", err)
	}
	if body["success"] != false {
		t.Errorf("success=false beklendi, alınan=%v", body["success"])
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("body['error'] map beklendi")
	}
	if errObj["message"] != "test hata" {
		t.Errorf("message='test hata' beklendi, alınan=%v", errObj["message"])
	}
}

func TestWriteError_HataNesnesiyileBirlikte(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusInternalServerError, "sunucu hatası", errors.New("iç hata"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status=500 beklendi, alınan=%d", w.Code)
	}
}

func TestHandleError(t *testing.T) {
	t.Run("nil hata - false döner, yanıt yazılmaz", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleError(w, nil, http.StatusInternalServerError, "hata")
		if handled {
			t.Error("false beklendi")
		}
		if w.Code != 200 {
			t.Errorf("yanıt yazılmamalıydı, status=%d", w.Code)
		}
	})

	t.Run("hata varsa true döner ve yanıt yazar", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleError(w, errors.New("bir hata"), http.StatusInternalServerError, "hata mesajı")
		if !handled {
			t.Error("true beklendi")
		}
		if w.Code != http.StatusInternalServerError {
			t.Errorf("status=500 beklendi, alınan=%d", w.Code)
		}
	})
}

func TestHandleCustomError(t *testing.T) {
	t.Run("nil - false döner, yanıt yazılmaz", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleCustomError(w, nil)
		if handled {
			t.Error("false beklendi")
		}
		if w.Code != 200 {
			t.Errorf("yanıt yazılmamalıydı, status=%d", w.Code)
		}
	})

	t.Run("permission hatası - 403 döner", func(t *testing.T) {
		w := httptest.NewRecorder()
		handled := HandleCustomError(w, NewPermissionError("yasak"))
		if !handled {
			t.Error("true beklendi")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("status=403 beklendi, alınan=%d", w.Code)
		}
	})

	t.Run("not found - 404 döner", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleCustomError(w, NewNotFoundError("bulunamadı"))
		if w.Code != http.StatusNotFound {
			t.Errorf("status=404 beklendi, alınan=%d", w.Code)
		}
	})

	t.Run("yanıt JSON formatında", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleCustomError(w, NewAuthError("yetkisiz"))

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("JSON parse hatası: %v", err)
		}
		if body["success"] != false {
			t.Errorf("success=false beklendi")
		}
	})
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccess(w, "işlem başarılı", map[string]string{"id": "42"})

	if w.Code != http.StatusOK {
		t.Errorf("status=200 beklendi, alınan=%d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type=application/json beklendi")
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON parse hatası: %v", err)
	}
	if body["success"] != true {
		t.Errorf("success=true beklendi, alınan=%v", body["success"])
	}
	if body["message"] != "işlem başarılı" {
		t.Errorf("message='işlem başarılı' beklendi, alınan=%q", body["message"])
	}
}
