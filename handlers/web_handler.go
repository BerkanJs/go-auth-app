package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"
)

// WebHandler web arayüzü sayfalarını yönetir.
type WebHandler struct {
	authSvc    service.AuthService
	personRepo repository.PersonRepository
	blogRepo   repository.BlogRepository
	personSvc  service.PersonService
}

func NewWebHandler(authSvc service.AuthService, personRepo repository.PersonRepository, blogRepo repository.BlogRepository, personSvc service.PersonService) *WebHandler {
	return &WebHandler{authSvc: authSvc, personRepo: personRepo, blogRepo: blogRepo, personSvc: personSvc}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[1:]
	if !isPathSafe(filePath) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filePath)
}

func isPathSafe(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	staticAbs, _ := filepath.Abs("static")
	uploadsAbs, _ := filepath.Abs("uploads")
	return hasPrefix(absPath, staticAbs) || hasPrefix(absPath, uploadsAbs)
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (h *WebHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Ana Sayfa"
	blogs, err := h.blogRepo.GetPublishedBlogs()
	if err != nil {
		shared.LogError("HOME_LOAD_ERROR", "Failed to load published blogs", map[string]interface{}{"error": err.Error()})
		data.ErrorMessage = "Blog'lar yüklenirken bir hata oluştu."
		renderTemplate(w, "home.html", data)
		return
	}
	data.Blogs = models.ToBlogResponseList(blogs)
	renderTemplate(w, "home.html", data)
}

func (h *WebHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Giriş Yap"
	if data.IsAuthenticated {
		if data.UserRole == "editor" {
			http.Redirect(w, r, "/editor", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		}
		return
	}
	renderTemplate(w, "login.html", data)
}

func (h *WebHandler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Kayıt Ol"
	if data.IsAuthenticated {
		if data.UserRole == "editor" {
			http.Redirect(w, r, "/editor", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		}
		return
	}
	renderTemplate(w, "register.html", data)
}

func (h *WebHandler) AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	data.Title = "Admin Panel"
	people, err := h.personRepo.GetAllPeople()
	if err != nil {
		shared.LogError("ADMIN_LOAD_ERROR", "Failed to load users", map[string]interface{}{"error": err.Error()})
		data.ErrorMessage = "Kullanıcılar yüklenirken bir hata oluştu."
		renderTemplate(w, "admin.html", data)
		return
	}
	data.Users = models.ToPersonResponseList(people)
	renderTemplate(w, "admin.html", data)
}

func (h *WebHandler) WebLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	person, err := h.authSvc.Login(email, password)
	if err != nil {
		shared.LogAuth("LOGIN_FAILED", email, "Invalid credentials")
		data := shared.GetTemplateData(r)
		data.Title = "Giriş Yap"
		data.ErrorMessage = "Email veya şifre hatalı"
		renderTemplate(w, "login.html", data)
		return
	}
	accessToken, err := h.authSvc.GenerateAccessToken(person.ID)
	if err != nil {
		shared.LogError("LOGIN_TOKEN_ERROR", "Access token generation failed", map[string]interface{}{"user_id": person.ID, "error": err.Error()})
		data := shared.GetTemplateData(r)
		data.Title = "Giriş Yap"
		data.ErrorMessage = "Giriş yapılırken bir hata oluştu."
		renderTemplate(w, "login.html", data)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: accessToken, Path: "/", MaxAge: 3600, HttpOnly: true})
	shared.LogAuth("LOGIN_SUCCESS", email, "Logged in via web")
	redirectURL := "/admin"
	if person.Role == "editor" {
		redirectURL = "/editor"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *WebHandler) WebRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		shared.LogError("REGISTER_PARSE_ERROR", "Multipart form parse failed", map[string]interface{}{"error": err.Error()})
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	req := models.CreatePersonRequest{
		Name:     r.FormValue("name"),
		Surname:  r.FormValue("surname"),
		Email:    r.FormValue("email"),
		Age:      shared.ParseIntFromForm(r.FormValue("age")),
		Phone:    r.FormValue("phone"),
		Role:     r.FormValue("role"),
		Password: r.FormValue("password"),
	}
	validator := shared.NewValidator()
	validator.ValidateCreatePersonRequest(req)
	if validator.HasError() {
		data := shared.GetTemplateData(r)
		data.Title = "Kayıt Ol"
		data.ErrorMessage = validator.GetError()
		renderTemplate(w, "register.html", data)
		return
	}
	file, header, err := r.FormFile("photo")
	if err == nil {
		photoPath, err := repository.UploadPhoto(file, header)
		if err != nil {
			shared.LogError("REGISTER_PHOTO_ERROR", "Photo upload failed", map[string]interface{}{"error": err.Error(), "email": req.Email})
			data := shared.GetTemplateData(r)
			data.Title = "Kayıt Ol"
			data.ErrorMessage = "Fotoğraf yüklenemedi."
			renderTemplate(w, "register.html", data)
			return
		}
		req.PhotoPath = photoPath
	}
	ctx := &registrationContext{Req: req}
	if err := runRegistrationPipeline(ctx, h.personRepo); err != nil {
		shared.LogAuth("REGISTER_FAILED", req.Email, err.Error())
		data := shared.GetTemplateData(r)
		data.Title = "Kayıt Ol"
		if errors.Is(err, errEmailAlreadyExists) {
			data.ErrorMessage = "Bu email zaten kayıtlı."
		} else {
			data.ErrorMessage = "Kayıt olurken bir hata oluştu."
		}
		renderTemplate(w, "register.html", data)
		return
	}
	shared.LogAuth("REGISTER_SUCCESS", req.Email, "User registered successfully")
	data := shared.GetTemplateData(r)
	if data.IsAuthenticated {
		http.SetCookie(w, &http.Cookie{Name: "success_message", Value: "Kullanıcı başarıyla eklendi!", Path: "/", MaxAge: 5, HttpOnly: false})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
	}
}

func WebLogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *WebHandler) WebDeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID parametresi gerekli", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Geçersiz ID", http.StatusBadRequest)
		return
	}
	if err := h.personRepo.DeletePerson(id); err != nil {
		shared.LogError("WEB_DELETE_ERROR", "Failed to delete user", map[string]interface{}{"user_id": id, "error": err.Error()})
		http.Error(w, "Kullanıcı silinemedi", http.StatusInternalServerError)
		return
	}
	shared.LogAuth("USER_DELETED", fmt.Sprintf("ID: %d", id), "User deleted via web")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Kullanıcı başarıyla silindi"}`))
}

func (h *WebHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
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
	var newPhotoPath string
	file, header, err := r.FormFile("photo")
	if err == nil {
		newPhotoPath, err = repository.UploadPhoto(file, header)
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
	if err := h.personSvc.UpdatePerson(req); err != nil {
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
	http.SetCookie(w, &http.Cookie{Name: "success_message", Value: "Kullanıcı başarıyla güncellendi!", Path: "/", MaxAge: 5, HttpOnly: false})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func renderTemplate(w http.ResponseWriter, templateName string, data shared.TemplateData) {
	templates, err := template.ParseFiles("templates/layout.html", "templates/"+templateName)
	if err != nil {
		shared.LogError("TEMPLATE_PARSE_ERROR", "Template parse failed", map[string]interface{}{"template": templateName, "error": err.Error()})
		http.Error(w, "Sayfa yüklenemedi", http.StatusInternalServerError)
		return
	}
	if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
		shared.LogError("TEMPLATE_EXEC_ERROR", "Template execute failed", map[string]interface{}{"template": templateName, "error": err.Error()})
	}
}
