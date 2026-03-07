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
// @Description JSON body ile yeni bir kişi ekler, email benzersiz olmalıdır
// @Tags people
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param person body models.CreatePersonRequest true "Kişi"
// @Success 200 {object} models.PersonResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /add [post]
func AddPersonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
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
}
