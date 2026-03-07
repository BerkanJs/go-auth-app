package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"
)

// AddPersonHandler godoc
// @Summary Yeni kişi ekle / kayıt ol
// @Description JSON body veya multipart form ile yeni bir kişi ekler, email benzersiz olmalıdır
// @Tags people
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Param person body models.CreatePersonRequest false "Kişi (JSON)"
// @Param name formData string false "İsim (Form)"
// @Param surname formData string false "Soyisim (Form)"
// @Param email formData string false "Email (Form)"
// @Param age formData int false "Yaş (Form)"
// @Param phone formData string false "Telefon (Form)"
// @Param password formData string false "Şifre (Form)"
// @Param photo formData file false "Proje Fotoğrafı"
// @Success 200 {object} models.PersonResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /add [post]
func AddPersonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePersonRequest
	var photoPath string

	// Content-Type kontrolü
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Multipart form handling
		err := r.ParseMultipartForm(10 << 20) // 10MB max
		if err != nil {
			shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
			return
		}

		// Form verilerini al
		req.Name = r.FormValue("name")
		req.Surname = r.FormValue("surname")
		req.Email = r.FormValue("email")
		req.Age = parseIntFromForm(r.FormValue("age"))
		req.Phone = r.FormValue("phone")
		req.Password = r.FormValue("password")

		// Fotoğraf yükle
		file, header, err := r.FormFile("photo")
		if err == nil {
			photoPath, err = repository.UploadPhoto(file, header)
			if err != nil {
				shared.HandleError(w, err, http.StatusBadRequest, "Fotoğraf yüklenemedi")
				return
			}
		}
		req.PhotoPath = photoPath
	} else {
		// JSON handling
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
			return
		}
	}

	ctx := &registrationContext{Req: req}
	if err := runRegistrationPipeline(ctx); err != nil {
		switch {
		case errors.Is(err, errEmailAlreadyExists):
			shared.WriteError(w, http.StatusBadRequest, shared.ErrEmailAlreadyExists, nil)
		default:
			shared.WriteError(w, http.StatusInternalServerError, shared.ErrEmailCheckFailed, err)
		}
		return
	}

	response := models.ToPersonResponse(ctx.Person)
	json.NewEncoder(w).Encode(response)
}

// parseIntFromForm string'dan int'e çevirme helper
func parseIntFromForm(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}

// GetAllPeopleHandler godoc
// @Summary Tüm kişileri getir
// @Description Veritabanındaki tüm kişileri döndürür
// @Tags people
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.PersonResponse
// @Failure 500 {string} string
// @Router /all [get]
func GetAllPeopleHandler(w http.ResponseWriter, r *http.Request) {
	people, err := repository.GetAllPeople()
	if shared.HandleError(w, err, http.StatusInternalServerError, err.Error()) {
		return
	}

	// Person -> PersonResponse dönüşümü helper üzerinden
	responses := models.ToPersonResponseList(people)

	json.NewEncoder(w).Encode(responses)
}

// GetPersonByIDHandler godoc
// @Summary ID'ye göre kişi getir
// @Description Verilen ID'ye göre kişiyi döndürür
// @Tags people
// @Security BearerAuth
// @Produce json
// @Param id query int true "Kişi ID"
// @Success 200 {object} models.PersonResponse
// @Failure 404 {string} string
// @Router /get [get]
func GetPersonByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	p, err := repository.GetPersonByID(id)
	if shared.HandleError(w, err, http.StatusNotFound, shared.ErrPersonNotFound) {
		return
	}
	response := models.ToPersonResponse(p)

	json.NewEncoder(w).Encode(response)
}

// DeletePersonHandler godoc
// @Summary Kişi sil
// @Description ID'ye göre kişiyi siler
// @Tags people
// @Security BearerAuth
// @Param id query int true "Kişi ID"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /delete [get]
func DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	err := repository.DeletePerson(id)
	if shared.HandleError(w, err, http.StatusInternalServerError, err.Error()) {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Kullanıcı başarıyla silindi"}`))
}
