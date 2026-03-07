package shared

import (
	"log"
	"net/http"
)

// Uygulama genelinde kullanılan hata mesajları.
const (
	ErrInvalidRequestBody      = "Geçersiz istek gövdesi"
	ErrEmailCheckFailed        = "Email kontrolü sırasında hata oluştu"
	ErrEmailAlreadyExists      = "Bu email ile kayıt zaten mevcut"
	ErrPersonNotFound          = "Kişi bulunamadı"
	ErrUnauthorized            = "Kullanıcı veya şifre hatalı"
	ErrAuthHeaderRequired      = "Authorization header gerekli"
	ErrBearerTokenRequired     = "Bearer token gerekli"
	ErrInvalidOrExpiredToken   = "Geçersiz veya süresi dolmuş token"
	ErrAccessTokenGenerateFail = "Access token üretilemedi"
	ErrRefreshTokenGenerateFail = "Refresh token üretilemedi"
	ErrInvalidRefreshToken     = "Geçersiz refresh token"
)

// WriteError hem log yazar hem de HTTP hata yanıtını döner.
// err parametresi opsiyoneldir; yoksa nil verilebilir.
func WriteError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Printf("http error: status=%d, message=%s, error=%v", status, message, err)
	} else {
		log.Printf("http error: status=%d, message=%s", status, message)
	}
	http.Error(w, message, status)
}

// HandleError, sık kullanılan "if err != nil { WriteError; return }"
// pattern'ini tek satıra indirir. Eğer err != nil ise hatayı yazar ve true döner.
// Kullanım:
//   if shared.HandleError(w, err, http.StatusInternalServerError, shared.ErrSomething) { return }
func HandleError(w http.ResponseWriter, err error, status int, message string) bool {
	if err == nil {
		return false
	}
	WriteError(w, status, message, err)
	return true
}

