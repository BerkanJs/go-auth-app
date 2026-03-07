package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Hata tipleri
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeInternal   ErrorType = "internal"
)

// CustomError özel hata yapısı
type CustomError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *CustomError) Error() string {
	return e.Message
}

// Uygulama genelinde kullanılan hata mesajları.
const (
	ErrInvalidRequestBody       = "Geçersiz istek gövdesi"
	ErrEmailCheckFailed         = "Email kontrolü sırasında hata oluştu"
	ErrEmailAlreadyExists       = "Bu email ile kayıt zaten mevcut"
	ErrPersonNotFound           = "Kişi bulunamadı"
	ErrUnauthorized             = "Kullanıcı veya şifre hatalı"
	ErrAuthHeaderRequired       = "Authorization header gerekli"
	ErrBearerTokenRequired      = "Bearer token gerekli"
	ErrInvalidOrExpiredToken    = "Geçersiz veya süresi dolmuş token"
	ErrAccessTokenGenerateFail  = "Access token üretilemedi"
	ErrRefreshTokenGenerateFail = "Refresh token üretilemedi"
	ErrInvalidRefreshToken      = "Geçersiz refresh token"
)

// Validasyon hataları
const (
	ErrInvalidEmailFormat = "Geçersiz email formatı"
	ErrInvalidPassword    = "Şifre en az 6 karakter olmalıdır"
	ErrInvalidName        = "İsim boş olamaz"
	ErrInvalidSurname     = "Soyisim boş olamaz"
	ErrInvalidAge         = "Yaş 0 ile 150 arasında olmalıdır"
	ErrInvalidRole        = "Geçersiz rol"
)

// Hata oluşturma fonksiyonları
func NewValidationError(message string, details map[string]interface{}) *CustomError {
	return &CustomError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    http.StatusBadRequest,
		Details: details,
	}
}

func NewAuthError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeAuth,
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

func NewNotFoundError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Code:    http.StatusNotFound,
	}
}

func NewPermissionError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypePermission,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

func NewDatabaseError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

func NewInternalError(message string) *CustomError {
	return &CustomError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

// WriteError hem log yazar hem de HTTP hata yanıtını döner.
func WriteError(w http.ResponseWriter, status int, message string, err error) {
	LogError("HTTP_ERROR", message, map[string]interface{}{
		"status": status,
		"error":  err,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"message": message,
			"code":    status,
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// HandleError, sık kullanılan "if err != nil { WriteError; return }"
// pattern'ini tek satıra indirir. Eğer err != nil ise hatayı yazar ve true döner.
func HandleError(w http.ResponseWriter, err error, status int, message string) bool {
	if err == nil {
		return false
	}
	WriteError(w, status, message, err)
	return true
}

// CustomError handler
func HandleCustomError(w http.ResponseWriter, err *CustomError) bool {
	if err == nil {
		return false
	}

	LogError("CUSTOM_ERROR", err.Message, map[string]interface{}{
		"type":    err.Type,
		"code":    err.Code,
		"details": err.Details,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	response := map[string]interface{}{
		"success": false,
		"error":   err,
	}

	json.NewEncoder(w).Encode(response)
	return true
}

// Success response helper
func WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	LogInfo("SUCCESS", message, data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	json.NewEncoder(w).Encode(response)
}

// Logging fonksiyonları
func LogError(level, message string, details interface{}) {
	logMessage(level, message, details)
}

func LogInfo(level, message string, details interface{}) {
	logMessage(level, message, details)
}

func LogAuth(action, userID, details string) {
	LogInfo("AUTH", fmt.Sprintf("%s - User: %s", action, userID), map[string]interface{}{
		"action":  action,
		"user_id": userID,
		"details": details,
	})
}

func logMessage(level, message string, details interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := map[string]interface{}{
		"timestamp": timestamp,
		"level":     level,
		"message":   message,
		"details":   details,
	}

	// JSON formatında log
	logJSON, _ := json.Marshal(logEntry)
	log.Printf("%s", string(logJSON))

	// Ayrıca dosyaya yaz (isteğe bağlı)
	writeLogToFile(timestamp, level, message, details)
}

func writeLogToFile(timestamp, level, message string, details interface{}) {
	// Log dosyası oluştur/aç
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Log dosyası açılamadı: %v", err)
		return
	}
	defer logFile.Close()

	// Log formatı: [TIMESTAMP] LEVEL: MESSAGE - DETAILS
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	if details != nil {
		logLine += fmt.Sprintf(" - %+v", details)
	}
	logLine += "\n"

	logFile.WriteString(logLine)
}
