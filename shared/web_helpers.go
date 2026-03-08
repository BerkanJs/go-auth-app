package shared

import (
	"net/http"

	"go-kisi-api/models"
	"go-kisi-api/repository"
)

// TemplateData web sayfaları için veri yapısı
type TemplateData struct {
	Title           string
	IsAuthenticated bool
	UserName        string
	UserRole        string
	Users           []models.PersonResponse
	Blogs           []models.BlogResponse
	ErrorMessage    string
	SuccessMessage  string
}

// GetTemplateData template verisi hazırlar
func GetTemplateData(r *http.Request) TemplateData {
	data := TemplateData{}

	// Cookie'den token'ı oku
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		LogInfo("DEBUG", "Cookie found", map[string]interface{}{
			"cookie_value": cookie.Value[:20] + "...", // Sadece ilk 20 karakter
		})

		claims, err := ParseAccessToken(cookie.Value)
		if err == nil {
			data.IsAuthenticated = true
			LogInfo("DEBUG", "Token parsed successfully", map[string]interface{}{
				"user_id": claims.UserID,
			})

			// Kullanıcı bilgisini al
			person, err := repository.GetPersonByID(r.Context(), claims.UserID)
			if err == nil {
				data.UserName = person.Name + " " + person.Surname
				data.UserRole = person.Role
				LogInfo("DEBUG", "User data loaded", map[string]interface{}{
					"user_name": data.UserName,
					"user_role": data.UserRole,
				})
			} else {
				LogError("DEBUG", "Failed to get user by ID", map[string]interface{}{
					"user_id": claims.UserID,
					"error":   err.Error(),
				})
			}
		} else {
			LogError("DEBUG", "Failed to parse token", map[string]interface{}{
				"error": err.Error(),
			})
		}
	} else {
		LogInfo("DEBUG", "No auth cookie found", nil)
	}

	// URL parametrelerini kontrol et
	if r.URL.Query().Get("registered") == "true" {
		data.SuccessMessage = "Kayıt başarılı! Giriş yapabilirsiniz."
	}

	// Success message cookie'sini kontrol et
	if successCookie, err := r.Cookie("success_message"); err == nil {
		data.SuccessMessage = successCookie.Value
	}

	LogInfo("DEBUG", "Template data prepared", map[string]interface{}{
		"is_authenticated": data.IsAuthenticated,
		"user_name":        data.UserName,
		"user_role":        data.UserRole,
	})

	return data
}
