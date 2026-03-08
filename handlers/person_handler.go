package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"
)

// PersonHandler API kişi endpoint'lerini yönetir.
type PersonHandler struct {
	personSvc service.PersonService
}

func NewPersonHandler(personSvc service.PersonService) *PersonHandler {
	return &PersonHandler{personSvc: personSvc}
}

// AddPersonHandler godoc
// @Summary Yeni kişi ekle / kayıt ol
// @Tags people
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} models.PersonResponse
// @Failure 400 {string} string
// @Router /add [post]
func (h *PersonHandler) AddPersonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePersonRequest

	contentType := r.Header.Get("Content-Type")
	if contentType != "" && len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
			return
		}
		req.Name = r.FormValue("name")
		req.Surname = r.FormValue("surname")
		req.Email = r.FormValue("email")
		req.Age = parseIntFromForm(r.FormValue("age"))
		req.Phone = r.FormValue("phone")
		req.Password = r.FormValue("password")

		file, header, err := r.FormFile("photo")
		if err == nil {
			photoPath, err := repository.UploadPhoto(file, header)
			if err != nil {
				shared.HandleError(w, err, http.StatusBadRequest, "Fotoğraf yüklenemedi")
				return
			}
			req.PhotoPath = photoPath
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
			return
		}
	}

	person, err := h.personSvc.CreatePerson(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			shared.WriteError(w, http.StatusBadRequest, shared.ErrEmailAlreadyExists, nil)
		default:
			shared.WriteError(w, http.StatusInternalServerError, shared.ErrEmailCheckFailed, err)
		}
		return
	}

	shared.WriteSuccess(w, "Kullanıcı oluşturuldu", models.ToPersonResponse(person))
}

// GetAllPeopleHandler godoc
// @Summary Tüm kişileri getir
// @Tags people
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.PersonResponse
// @Router /all [get]
func (h *PersonHandler) GetAllPeopleHandler(w http.ResponseWriter, r *http.Request) {
	people, err := h.personSvc.GetAllPeople(r.Context())
	if shared.HandleError(w, err, http.StatusInternalServerError, "Kullanıcılar getirilemedi") {
		return
	}
	shared.WriteSuccess(w, "Kullanıcılar getirildi", models.ToPersonResponseList(people))
}

// GetPersonByIDHandler godoc
// @Summary ID'ye göre kişi getir
// @Tags people
// @Security BearerAuth
// @Produce json
// @Param id query int true "Kişi ID"
// @Success 200 {object} models.PersonResponse
// @Router /get [get]
func (h *PersonHandler) GetPersonByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	p, err := h.personSvc.GetPersonByID(r.Context(), id)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, shared.ErrPersonNotFound, nil)
		return
	}
	shared.WriteSuccess(w, "Kullanıcı getirildi", models.ToPersonResponse(p))
}

// DeletePersonHandler godoc
// @Summary Kişi sil
// @Tags people
// @Security BearerAuth
// @Param id query int true "Kişi ID"
// @Router /delete [get]
func (h *PersonHandler) DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if shared.HandleError(w, h.personSvc.DeletePerson(r.Context(), id), http.StatusInternalServerError, "Kullanıcı silinemedi") {
		return
	}
	shared.WriteSuccess(w, "Kullanıcı başarıyla silindi", nil)
}

// UpdatePersonHandler godoc
// @Summary Kişi güncelle
// @Tags people
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData int true "Kullanıcı ID"
// @Param name formData string false "Ad"
// @Param surname formData string false "Soyad"
// @Param email formData string false "Email"
// @Param age formData int false "Yaş"
// @Param phone formData string false "Telefon"
// @Param role formData string false "Rol"
// @Param password formData string false "Yeni Şifre"
// @Param photo formData file false "Fotoğraf"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Router /api/update [post]
func (h *PersonHandler) UpdatePersonHandler(w http.ResponseWriter, r *http.Request) {
	req, err := parsePersonUpdateForm(r)
	if err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.personSvc.UpdatePerson(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, service.ErrPersonNotFound):
			shared.WriteError(w, http.StatusNotFound, shared.ErrPersonNotFound, nil)
		case errors.Is(err, service.ErrEmailTaken):
			shared.WriteError(w, http.StatusBadRequest, shared.ErrEmailAlreadyExists, nil)
		case errors.Is(err, service.ErrPasswordHash):
			shared.WriteError(w, http.StatusInternalServerError, "Şifre işlenirken hata oluştu", err)
		default:
			shared.WriteError(w, http.StatusInternalServerError, "Kullanıcı güncellenemedi", err)
		}
		return
	}

	shared.WriteSuccess(w, "Kullanıcı başarıyla güncellendi", nil)
}

// parsePersonUpdateForm multipart formu okur, fotoğrafı yükler ve UpdatePersonRequest döner.
func parsePersonUpdateForm(r *http.Request) (service.UpdatePersonRequest, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return service.UpdatePersonRequest{}, fmt.Errorf("%s", shared.ErrInvalidRequestBody)
	}

	userID := parseIntFromForm(r.FormValue("user_id"))
	if userID == 0 {
		return service.UpdatePersonRequest{}, fmt.Errorf("geçersiz kullanıcı ID")
	}

	var newPhotoPath string
	file, header, err := r.FormFile("photo")
	if err == nil {
		newPhotoPath, err = repository.UploadPhoto(file, header)
		if err != nil {
			return service.UpdatePersonRequest{}, fmt.Errorf("fotoğraf yüklenemedi")
		}
	}

	return service.UpdatePersonRequest{
		UserID:       userID,
		Name:         r.FormValue("name"),
		Surname:      r.FormValue("surname"),
		Email:        r.FormValue("email"),
		Age:          parseIntFromForm(r.FormValue("age")),
		Phone:        r.FormValue("phone"),
		Role:         r.FormValue("role"),
		NewPassword:  r.FormValue("password"),
		NewPhotoPath: newPhotoPath,
	}, nil
}

func parseIntFromForm(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
