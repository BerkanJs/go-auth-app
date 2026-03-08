package shared

import (
	"strings"
	"testing"

	"go-kisi-api/models"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{"geçerli", "test@example.com", false, ""},
		{"noktalı geçerli", "test.user+tag@sub.example.com", false, ""},
		{"boş", "", true, "Email boş olamaz"},
		{"@ yok", "testexample.com", true, "Geçersiz email formatı"},
		{"domain yok", "test@", true, "Geçersiz email formatı"},
		{"tld yok", "test@example", true, "Geçersiz email formatı"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateEmail(tt.email)
			if tt.wantErr {
				if !v.HasError() {
					t.Errorf("hata beklendi ama alınmadı")
				}
				if tt.errMsg != "" && !strings.Contains(v.GetError(), tt.errMsg) {
					t.Errorf("beklenen hata mesajı %q, alınan %q", tt.errMsg, v.GetError())
				}
			} else {
				if v.HasError() {
					t.Errorf("hata beklenmiyordu ama alındı: %s", v.GetError())
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{"geçerli", "Abc123", false},
		{"boş", "", true},
		{"çok kısa - 3 karakter", "Ab1", true},
		{"büyük harf yok", "abc123", true},
		{"rakam yok", "Abcdef", true},
		{"çok uzun - 51 karakter", strings.Repeat("A", 46) + "bcde1", true},
		{"tam 6 karakter sınır", "Abc12!", false},
		{"tam 50 karakter sınır", strings.Repeat("A", 44) + "bcde1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidatePassword(tt.pass)
			if tt.wantErr != v.HasError() {
				t.Errorf("wantErr=%v, HasError=%v, errors=%q", tt.wantErr, v.HasError(), v.GetError())
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"geçerli", "Ali", false},
		{"boş", "", true},
		{"tek karakter", "A", true},
		{"tam 2 karakter", "Al", false},
		{"tam 50 karakter", strings.Repeat("a", 50), false},
		{"51 karakter - sınır aşımı", strings.Repeat("a", 51), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateName(tt.input, "İsim")
			if tt.wantErr != v.HasError() {
				t.Errorf("wantErr=%v, HasError=%v, errors=%q", tt.wantErr, v.HasError(), v.GetError())
			}
		})
	}
}

func TestValidateAge(t *testing.T) {
	tests := []struct {
		age     int
		wantErr bool
	}{
		{0, false},
		{25, false},
		{150, false},
		{-1, true},
		{151, true},
	}

	for _, tt := range tests {
		v := NewValidator()
		v.ValidateAge(tt.age)
		if tt.wantErr != v.HasError() {
			t.Errorf("yaş=%d: wantErr=%v, HasError=%v", tt.age, tt.wantErr, v.HasError())
		}
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"boş - isteğe bağlı", "", false},
		{"geçerli Türkiye", "05321234567", false},
		{"geçerli uluslararası", "+905321234567", false},
		{"çok kısa", "123", true},
		{"harf içeriyor", "abc1234567", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidatePhone(tt.phone)
			if tt.wantErr != v.HasError() {
				t.Errorf("wantErr=%v, HasError=%v, errors=%q", tt.wantErr, v.HasError(), v.GetError())
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		role    string
		wantErr bool
	}{
		{"admin", false},
		{"editor", false},
		{"", true},
		{"superadmin", true},
		{"user", true},
	}

	for _, tt := range tests {
		v := NewValidator()
		v.ValidateRole(tt.role)
		if tt.wantErr != v.HasError() {
			t.Errorf("rol=%q: wantErr=%v, HasError=%v", tt.role, tt.wantErr, v.HasError())
		}
	}
}

func TestValidateBlogRequest(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		summary string
		wantErr bool
	}{
		{"geçerli - özetsiz", "Blog Başlığı", "Bu blog içeriği yeterince uzundur.", "", false},
		{"geçerli - özetli", "Blog Başlığı", "Bu blog içeriği yeterince uzundur.", "Kısa özet", false},
		{"başlık boş", "", "Bu blog içeriği yeterince uzundur.", "", true},
		{"başlık çok kısa - 2 karakter", "ab", "Bu blog içeriği yeterince uzundur.", "", true},
		{"başlık çok uzun - 201 karakter", strings.Repeat("a", 201), "Bu blog içeriği yeterince uzundur.", "", true},
		{"içerik boş", "Geçerli Başlık", "", "", true},
		{"içerik çok kısa - 4 karakter", "Geçerli Başlık", "kısa", "", true},
		{"özet tam 500 karakter", "Geçerli Başlık", "Bu blog içeriği yeterince uzundur.", strings.Repeat("a", 500), false},
		{"özet 501 karakter - sınır aşımı", "Geçerli Başlık", "Bu blog içeriği yeterince uzundur.", strings.Repeat("a", 501), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateBlogRequest(tt.title, tt.content, tt.summary)
			if tt.wantErr != v.HasError() {
				t.Errorf("wantErr=%v, HasError=%v, errors=%q", tt.wantErr, v.HasError(), v.GetError())
			}
		})
	}
}

func TestValidateCreatePersonRequest(t *testing.T) {
	validReq := models.CreatePersonRequest{
		Name:     "Ali",
		Surname:  "Veli",
		Email:    "ali@example.com",
		Password: "Test123",
		Age:      25,
		Role:     "editor",
	}

	t.Run("geçerli istek", func(t *testing.T) {
		v := NewValidator()
		v.ValidateCreatePersonRequest(validReq)
		if v.HasError() {
			t.Errorf("hata beklenmiyordu: %s", v.GetError())
		}
	})

	t.Run("eksik isim", func(t *testing.T) {
		req := validReq
		req.Name = ""
		v := NewValidator()
		v.ValidateCreatePersonRequest(req)
		if !v.HasError() {
			t.Error("hata beklendi")
		}
	})

	t.Run("geçersiz email", func(t *testing.T) {
		req := validReq
		req.Email = "gecersiz-email"
		v := NewValidator()
		v.ValidateCreatePersonRequest(req)
		if !v.HasError() {
			t.Error("hata beklendi")
		}
	})

	t.Run("geçersiz rol", func(t *testing.T) {
		req := validReq
		req.Role = "superadmin"
		v := NewValidator()
		v.ValidateCreatePersonRequest(req)
		if !v.HasError() {
			t.Error("hata beklendi")
		}
	})
}

func TestValidateUpdatePersonRequest_SifreIstegeBagli(t *testing.T) {
	req := models.CreatePersonRequest{
		Name:    "Ali",
		Surname: "Veli",
		Email:   "ali@example.com",
		Age:     25,
		Role:    "editor",
		// Şifre boş - güncelleme için isteğe bağlı
	}

	v := NewValidator()
	v.ValidateUpdatePersonRequest(req)
	if v.HasError() {
		t.Errorf("şifresiz güncelleme geçerli olmalı, hata: %s", v.GetError())
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"değer", false},
		{"", true},
		{"   ", true}, // sadece boşluk
	}

	for _, tt := range tests {
		v := NewValidator()
		v.ValidateRequired(tt.input, "Alan")
		if tt.wantErr != v.HasError() {
			t.Errorf("input=%q: wantErr=%v, HasError=%v", tt.input, tt.wantErr, v.HasError())
		}
	}
}

func TestValidateMaxMinLength(t *testing.T) {
	t.Run("max aşıldı", func(t *testing.T) {
		v := NewValidator()
		v.ValidateMaxLength("toolong", "Alan", 5)
		if !v.HasError() {
			t.Error("hata beklendi")
		}
	})

	t.Run("max aşılmadı", func(t *testing.T) {
		v := NewValidator()
		v.ValidateMaxLength("ok", "Alan", 5)
		if v.HasError() {
			t.Errorf("hata beklenmiyordu: %s", v.GetError())
		}
	})

	t.Run("min altında", func(t *testing.T) {
		v := NewValidator()
		v.ValidateMinLength("ab", "Alan", 5)
		if !v.HasError() {
			t.Error("hata beklendi")
		}
	})
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"42", false},
		{"0", false},
		{"-5", false},
		{"abc", true},
		{"3.14", true},
		{"", true},
	}

	for _, tt := range tests {
		v := NewValidator()
		v.ValidateNumeric(tt.input, "Alan")
		if tt.wantErr != v.HasError() {
			t.Errorf("input=%q: wantErr=%v, HasError=%v", tt.input, tt.wantErr, v.HasError())
		}
	}
}

func TestGetErrorAsCustomError(t *testing.T) {
	t.Run("hata yokken nil döner", func(t *testing.T) {
		v := NewValidator()
		if v.GetErrorAsCustomError() != nil {
			t.Error("nil beklendi")
		}
	})

	t.Run("hata varken CustomError döner", func(t *testing.T) {
		v := NewValidator()
		v.ValidateEmail("")
		ce := v.GetErrorAsCustomError()
		if ce == nil {
			t.Fatal("CustomError beklendi ama nil alındı")
		}
		if ce.Code != 400 {
			t.Errorf("code=400 beklendi, alınan=%d", ce.Code)
		}
		if ce.Type != ErrorTypeValidation {
			t.Errorf("type=validation beklendi, alınan=%q", ce.Type)
		}
		if _, ok := ce.Details["validation_errors"]; !ok {
			t.Error("details içinde validation_errors beklendi")
		}
	})
}

func TestParseIntFromForm(t *testing.T) {
	tests := []struct {
		input  string
		expect int
	}{
		{"42", 42},
		{"0", 0},
		{"-5", -5},
		{"", 0},
		{"abc", 0},
		{"3.14", 0},
	}

	for _, tt := range tests {
		got := ParseIntFromForm(tt.input)
		if got != tt.expect {
			t.Errorf("ParseIntFromForm(%q) = %d, beklenen %d", tt.input, got, tt.expect)
		}
	}
}

func TestValidatorChaining_CokluHata(t *testing.T) {
	v := NewValidator()
	v.ValidateEmail("").ValidatePassword("").ValidateRole("")
	errStr := v.GetError()
	if !strings.Contains(errStr, ";") {
		t.Errorf("birden fazla hata beklendi (';' içermeli), alınan: %q", errStr)
	}
}
