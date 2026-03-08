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

// PersonHandler API kişi endpoint'lerini yönetir.
type PersonHandler struct {
	personRepo repository.PersonRepository
}

func NewPersonHandler(personRepo repository.PersonRepository) *PersonHandler {
	return &PersonHandler{personRepo: personRepo}
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
	var photoPath string

	contentType := r.Header.Get("Content-Type")
	if contentType != "" && len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
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
			photoPath, err = repository.UploadPhoto(file, header)
			if err != nil {
				shared.HandleError(w, err, http.StatusBadRequest, "Fotoğraf yüklenemedi")
				return
			}
		}
		req.PhotoPath = photoPath
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
			return
		}
	}

	regCtx := &registrationContext{Req: req}
	if err := runRegistrationPipeline(r.Context(), regCtx, h.personRepo); err != nil {
		switch {
		case errors.Is(err, errEmailAlreadyExists):
			shared.WriteError(w, http.StatusBadRequest, shared.ErrEmailAlreadyExists, nil)
		default:
			shared.WriteError(w, http.StatusInternalServerError, shared.ErrEmailCheckFailed, err)
		}
		return
	}

	json.NewEncoder(w).Encode(models.ToPersonResponse(regCtx.Person))
}

// GetAllPeopleHandler godoc
// @Summary Tüm kişileri getir
// @Tags people
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.PersonResponse
// @Router /all [get]
func (h *PersonHandler) GetAllPeopleHandler(w http.ResponseWriter, r *http.Request) {
	people, err := h.personRepo.GetAllPeople(r.Context())
	if shared.HandleError(w, err, http.StatusInternalServerError, "Kullanıcılar getirilemedi") {
		return
	}
	json.NewEncoder(w).Encode(models.ToPersonResponseList(people))
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
	p, err := h.personRepo.GetPersonByID(r.Context(), id)
	if shared.HandleError(w, err, http.StatusNotFound, shared.ErrPersonNotFound) {
		return
	}
	json.NewEncoder(w).Encode(models.ToPersonResponse(p))
}

// DeletePersonHandler godoc
// @Summary Kişi sil
// @Tags people
// @Security BearerAuth
// @Param id query int true "Kişi ID"
// @Router /delete [get]
func (h *PersonHandler) DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	err := h.personRepo.DeletePerson(r.Context(), id)
	if shared.HandleError(w, err, http.StatusInternalServerError, "Kullanıcı silinemedi") {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Kullanıcı başarıyla silindi"}`))
}

func parseIntFromForm(s string) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 0
}
