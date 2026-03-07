package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"

	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"
)

// UpdateUserHandler kullanıcı günceller (sadece admin)
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if data.UserRole != "admin" {
		data.ErrorMessage = "Kullanıcı güncelleme yetkiniz yok."
		renderTemplate(w, "admin.html", data)
		return
	}

	userID := parseIntFromForm(r.FormValue("user_id"))
	if userID == 0 {
		data.ErrorMessage = "Geçersiz kullanıcı ID."
		renderTemplate(w, "admin.html", data)
		return
	}

	// Yeni fotoğraf varsa yükle
	var newPhotoPath string
	file, header, err := r.FormFile("photo")
	if err == nil {
		newPhotoPath, err = uploadIfProvided(file, header)
		if err != nil {
			shared.LogError("USER_UPDATE_PHOTO_ERROR", "Photo upload failed", map[string]interface{}{"error": err.Error(), "user_id": userID})
			data.ErrorMessage = "Fotoğraf yüklenemedi."
			renderTemplate(w, "admin.html", data)
			return
		}
	}

	req := service.UpdatePersonRequest{
		UserID:       userID,
		Name:         r.FormValue("name"),
		Surname:      r.FormValue("surname"),
		Email:        r.FormValue("email"),
		Age:          parseIntFromForm(r.FormValue("age")),
		Phone:        r.FormValue("phone"),
		Role:         r.FormValue("role"),
		NewPassword:  r.FormValue("password"),
		NewPhotoPath: newPhotoPath,
	}

	if err := service.UpdatePerson(req); err != nil {
		shared.LogError("USER_UPDATE_ERROR", "Failed to update user", map[string]interface{}{"error": err.Error(), "user_id": userID})
		switch {
		case errors.Is(err, service.ErrPersonNotFound):
			data.ErrorMessage = "Kullanıcı bulunamadı."
		case errors.Is(err, service.ErrEmailTaken):
			data.ErrorMessage = "Bu email zaten kayıtlı."
		case errors.Is(err, service.ErrPasswordHash):
			data.ErrorMessage = "Şifre işlenirken bir hata oluştu."
		default:
			data.ErrorMessage = "Kullanıcı güncellenirken bir hata oluştu."
		}
		renderTemplate(w, "admin.html", data)
		return
	}

	shared.LogInfo("USER_UPDATED", "User updated successfully", map[string]interface{}{"user_id": userID})

	http.SetCookie(w, &http.Cookie{
		Name:     "success_message",
		Value:    "Kullanıcı başarıyla güncellendi!",
		Path:     "/",
		MaxAge:   5,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// uploadIfProvided fotoğraf yükleme işlemini yönetir
func uploadIfProvided(file multipart.File, header *multipart.FileHeader) (string, error) {
	return repository.UploadPhoto(file, header)
}
