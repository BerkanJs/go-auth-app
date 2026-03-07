package handlers

import (
	"net/http"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"

	"golang.org/x/crypto/bcrypt"
)

// UpdateUserHandler kullanıcı günceller
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

	// Sadece admin kullanıcı güncelleyebilir
	if data.UserRole != "admin" {
		data.ErrorMessage = "Kullanıcı güncelleme yetkiniz yok"
		renderTemplate(w, "admin.html", data)
		return
	}

	// Kullanıcı ID'sini al
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		data.ErrorMessage = "Geçersiz kullanıcı ID"
		renderTemplate(w, "admin.html", data)
		return
	}

	// Mevcut kullanıcı bilgisini al
	person, err := repository.GetPersonByID(userID)
	if err != nil {
		data.ErrorMessage = "Kullanıcı bulunamadı"
		renderTemplate(w, "admin.html", data)
		return
	}

	// Form verilerini işle
	updatedPerson := models.Person{
		ID:        userID,
		Name:      r.FormValue("name"),
		Surname:   r.FormValue("surname"),
		Email:     r.FormValue("email"),
		Age:       parseIntFromForm(r.FormValue("age")),
		Phone:     r.FormValue("phone"),
		Role:      r.FormValue("role"),
		PhotoPath: person.PhotoPath, // Mevcut fotoğrafı koru
	}

	// Email değişiyse benzersiz mi kontrol et
	if updatedPerson.Email != person.Email {
		exists, err := repository.EmailExists(updatedPerson.Email)
		if err != nil {
			data.ErrorMessage = "Email kontrolü sırasında hata oluştu"
			renderTemplate(w, "admin.html", data)
			return
		}
		if exists {
			data.ErrorMessage = "Bu email zaten kayıtlı"
			renderTemplate(w, "admin.html", data)
			return
		}
	}

	// Fotoğraf yükle
	file, header, err := r.FormFile("photo")
	if err == nil {
		photoPath, err := repository.UploadPhoto(file, header)
		if err != nil {
			data.ErrorMessage = "Fotoğraf yüklenemedi: " + err.Error()
			renderTemplate(w, "admin.html", data)
			return
		}
		updatedPerson.PhotoPath = photoPath
	}

	// Şifre değişiyse hash'le
	newPassword := r.FormValue("password")
	if newPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			data.ErrorMessage = "Şifre işlenirken hata oluştu"
			renderTemplate(w, "admin.html", data)
			return
		}
		updatedPerson.PasswordHash = string(hashed)
	} else {
		updatedPerson.PasswordHash = person.PasswordHash // Mevcut şifreyi koru
	}

	// Kullanıcıyı güncelle (UpdatePerson fonksiyonunu repository'e eklememiz gerekiyor)
	err = repository.UpdatePerson(updatedPerson)
	if err != nil {
		data.ErrorMessage = "Kullanıcı güncellenirken hata oluştu: " + err.Error()
		renderTemplate(w, "admin.html", data)
		return
	}

	// Başarılı mesajı ayarla ve yönlendir
	http.SetCookie(w, &http.Cookie{
		Name:     "success_message",
		Value:    "Kullanıcı başarıyla güncellendi!",
		Path:     "/",
		MaxAge:   5,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
