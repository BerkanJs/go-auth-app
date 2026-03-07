package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go-kisi-api/models"
)

// Validator validasyon işlemleri için yapı
type Validator struct {
	errors []string
}

// NewValidator yeni validator oluşturur
func NewValidator() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// ValidateEmail email formatını kontrol eder
func (v *Validator) ValidateEmail(email string) *Validator {
	if email == "" {
		v.errors = append(v.errors, "Email boş olamaz")
		return v
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.errors = append(v.errors, "Geçersiz email formatı")
	}
	return v
}

// ValidatePassword şifreyi kontrol eder
func (v *Validator) ValidatePassword(password string) *Validator {
	if password == "" {
		v.errors = append(v.errors, "Şifre boş olamaz")
		return v
	}

	if len(password) < 6 {
		v.errors = append(v.errors, "Şifre en az 6 karakter olmalıdır")
	}

	if len(password) > 50 {
		v.errors = append(v.errors, "Şifre en fazla 50 karakter olabilir")
	}

	// En az bir büyük harf kontrolü
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		v.errors = append(v.errors, "Şifre en az bir büyük harf içermelidir")
	}

	// En az bir rakam kontrolü
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		v.errors = append(v.errors, "Şifre en az bir rakam içermelidir")
	}

	return v
}

// ValidateName ismi kontrol eder
func (v *Validator) ValidateName(name, fieldName string) *Validator {
	if name == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s boş olamaz", fieldName))
		return v
	}

	if len(name) < 2 {
		v.errors = append(v.errors, fmt.Sprintf("%s en az 2 karakter olmalıdır", fieldName))
	}

	if len(name) > 50 {
		v.errors = append(v.errors, fmt.Sprintf("%s en fazla 50 karakter olabilir", fieldName))
	}

	return v
}

// ValidateAge yaşı kontrol eder
func (v *Validator) ValidateAge(age int) *Validator {
	if age < 0 || age > 150 {
		v.errors = append(v.errors, "Yaş 0 ile 150 arasında olmalıdır")
	}
	return v
}

// ValidatePhone telefon numarasını kontrol eder
func (v *Validator) ValidatePhone(phone string) *Validator {
	if phone == "" {
		return v // Telefon isteğe bağlı
	}

	// Basit telefon formatı kontrolü
	phoneRegex := regexp.MustCompile(`^[0-9+\-\s\(\)]{10,20}$`)
	if !phoneRegex.MatchString(phone) {
		v.errors = append(v.errors, "Geçersiz telefon formatı")
	}

	return v
}

// ValidateRole rolü kontrol eder
func (v *Validator) ValidateRole(role string) *Validator {
	validRoles := []string{"admin", "editor"}

	if role == "" {
		v.errors = append(v.errors, "Rol seçilmelidir")
		return v
	}

	isValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValid = true
			break
		}
	}

	if !isValid {
		v.errors = append(v.errors, "Geçersiz rol")
	}

	return v
}

// ValidateCreatePersonRequest kullanıcı oluşturma isteğini doğrular
func (v *Validator) ValidateCreatePersonRequest(req models.CreatePersonRequest) *Validator {
	v.ValidateName(req.Name, "İsim")
	v.ValidateName(req.Surname, "Soyisim")
	v.ValidateEmail(req.Email)
	v.ValidatePassword(req.Password)
	v.ValidateAge(req.Age)
	v.ValidatePhone(req.Phone)
	v.ValidateRole(req.Role)
	return v
}

// ValidateUpdatePersonRequest kullanıcı güncelleme isteğini doğrular
func (v *Validator) ValidateUpdatePersonRequest(req models.CreatePersonRequest) *Validator {
	v.ValidateName(req.Name, "İsim")
	v.ValidateName(req.Surname, "Soyisim")
	v.ValidateEmail(req.Email)
	v.ValidateAge(req.Age)
	v.ValidatePhone(req.Phone)
	v.ValidateRole(req.Role)

	// Şifre güncelleme isteğe bağlı
	if req.Password != "" {
		v.ValidatePassword(req.Password)
	}

	return v
}

// ValidateBlogRequest blog isteğini doğrular
func (v *Validator) ValidateBlogRequest(title, content, summary string) *Validator {
	if title == "" {
		v.errors = append(v.errors, "Blog başlığı boş olamaz")
	}

	if len(title) < 3 {
		v.errors = append(v.errors, "Blog başlığı en az 3 karakter olmalıdır")
	}

	if len(title) > 200 {
		v.errors = append(v.errors, "Blog başlığı en fazla 200 karakter olabilir")
	}

	if content == "" {
		v.errors = append(v.errors, "Blog içeriği boş olamaz")
	}

	if len(content) < 10 {
		v.errors = append(v.errors, "Blog içeriği en az 10 karakter olmalıdır")
	}

	if summary != "" && len(summary) > 500 {
		v.errors = append(v.errors, "Blog özeti en fazla 500 karakter olabilir")
	}

	return v
}

// HasError hata olup olmadığını kontrol eder
func (v *Validator) HasError() bool {
	return len(v.errors) > 0
}

// GetError hataları birleştirip döner
func (v *Validator) GetError() string {
	if len(v.errors) == 0 {
		return ""
	}
	return strings.Join(v.errors, "; ")
}

// GetErrorAsCustomError hataları CustomError olarak döner
func (v *Validator) GetErrorAsCustomError() *CustomError {
	if !v.HasError() {
		return nil
	}

	return NewValidationError(v.GetError(), map[string]interface{}{
		"validation_errors": v.errors,
	})
}

// ValidateRequired zorunlu alanları kontrol eder
func (v *Validator) ValidateRequired(value, fieldName string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s alanı zorunludur", fieldName))
	}
	return v
}

// ValidateMaxLength maksimum uzunluğu kontrol eder
func (v *Validator) ValidateMaxLength(value, fieldName string, maxLength int) *Validator {
	if len(value) > maxLength {
		v.errors = append(v.errors, fmt.Sprintf("%s en fazla %d karakter olabilir", fieldName, maxLength))
	}
	return v
}

// ValidateMinLength minimum uzunluğu kontrol eder
func (v *Validator) ValidateMinLength(value, fieldName string, minLength int) *Validator {
	if len(value) < minLength {
		v.errors = append(v.errors, fmt.Sprintf("%s en az %d karakter olmalıdır", fieldName, minLength))
	}
	return v
}

// ValidateNumeric sayısal değeri kontrol eder
func (v *Validator) ValidateNumeric(value, fieldName string) *Validator {
	if _, err := strconv.Atoi(value); err != nil {
		v.errors = append(v.errors, fmt.Sprintf("%s sayısal bir değer olmalıdır", fieldName))
	}
	return v
}

// ParseIntFromForm string'dan int'e çevirme helper
func ParseIntFromForm(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
