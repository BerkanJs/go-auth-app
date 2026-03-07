package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"

	"golang.org/x/crypto/bcrypt"
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

// StaticHandler statik dosyaları sunar
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[1:] // /static/... -> static/...

	// Güvenlik kontrolü - sadece static klasöründen dosyalara izin ver
	if !isPathSafe(filePath) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

// isPathSafe dosya yolunun güvenli olup olmadığını kontrol eder
func isPathSafe(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	staticAbs, err := filepath.Abs("static")
	if err != nil {
		return false
	}

	uploadsAbs, err := filepath.Abs("uploads")
	if err != nil {
		return false
	}

	// Yol static veya uploads klasörü içinde mi?
	return hasPrefix(absPath, staticAbs) || hasPrefix(absPath, uploadsAbs)
}

// hasPrefix string'in prefix ile başlayıp başlamadığını kontrol eder
func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

// HomeHandler ana sayfayı gösterir
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Ana Sayfa"

	// Yayınlanmış blog'ları getir
	blogs, err := repository.GetPublishedBlogs()
	if err != nil {
		data.ErrorMessage = "Blog'lar yüklenirken hata oluştu: " + err.Error()
		renderTemplate(w, "home.html", data)
		return
	}

	data.Blogs = models.ToBlogResponseList(blogs)
	renderTemplate(w, "home.html", data)
}

// LoginPageHandler login sayfasını gösterir
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Giriş Yap"

	// Zaten giriş yapmışsa admin paneline yönlendir
	if data.IsAuthenticated {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "login.html", data)
}

// RegisterPageHandler register sayfasını gösterir
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)
	data.Title = "Kayıt Ol"

	// Zaten giriş yapmışsa admin paneline yönlendir
	if data.IsAuthenticated {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "register.html", data)
}

// AdminPageHandler admin panelini gösterir
func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)

	shared.LogInfo("DEBUG", "AdminPageHandler called", map[string]interface{}{
		"is_authenticated": data.IsAuthenticated,
		"user_name":        data.UserName,
		"user_role":        data.UserRole,
	})

	if !data.IsAuthenticated {
		shared.LogInfo("DEBUG", "User not authenticated, redirecting to login", nil)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data.Title = "Admin Panel"

	// Tüm kullanıcıları getir
	people, err := repository.GetAllPeople()
	if err != nil {
		shared.LogError("DEBUG", "Failed to get all people", map[string]interface{}{
			"error": err.Error(),
		})
		data.ErrorMessage = "Kullanıcılar yüklenirken hata oluştu: " + err.Error()
		renderTemplate(w, "admin.html", data)
		return
	}

	data.Users = models.ToPersonResponseList(people)

	shared.LogInfo("DEBUG", "Rendering admin template", map[string]interface{}{
		"user_count": len(data.Users),
		"user_role":  data.UserRole,
	})

	renderTemplate(w, "admin.html", data)
}

// WebLoginHandler web üzerinden giriş yapar
func WebLoginHandler(w http.ResponseWriter, r *http.Request) {
	shared.LogInfo("DEBUG", "WebLoginHandler CALLED - ROUTING FIXED!", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	fmt.Printf("DEBUG: WebLoginHandler called - Method: %s, Path: %s\n", r.Method, r.URL.Path)

	if r.Method != "POST" {
		fmt.Printf("DEBUG: Method not POST, redirecting to login\n")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fmt.Printf("DEBUG: Method is POST, processing login\n")
	email := r.FormValue("email")
	password := r.FormValue("password")

	fmt.Printf("DEBUG: Email: %s, Password length: %d\n", email, len(password))

	person, err := repository.GetPersonByEmail(email)
	if err != nil {
		fmt.Printf("DEBUG: User not found: %v\n", err)
		data := shared.GetTemplateData(r)
		data.Title = "Giriş Yap"
		data.ErrorMessage = "Email veya şifre hatalı"
		renderTemplate(w, "login.html", data)
		return
	}

	fmt.Printf("DEBUG: User found: %s %s (Role: %s)\n", person.Name, person.Surname, person.Role)
	fmt.Printf("DEBUG: Stored hash: %s\n", person.PasswordHash)
	fmt.Printf("DEBUG: Input password: %s\n", password)

	if err := bcrypt.CompareHashAndPassword([]byte(person.PasswordHash), []byte(password)); err != nil {
		fmt.Printf("DEBUG: Password mismatch: %v\n", err)
		data := shared.GetTemplateData(r)
		data.Title = "Giriş Yap"
		data.ErrorMessage = "Email veya şifre hatalı"
		renderTemplate(w, "login.html", data)
		return
	}

	fmt.Printf("DEBUG: Password correct, generating token\n")
	// Token oluştur ve cookie'ye kaydet
	accessToken, err := GenerateAccessToken(person.ID)
	if err != nil {
		fmt.Printf("DEBUG: Token generation failed: %v\n", err)
		data := shared.GetTemplateData(r)
		data.Title = "Giriş Yap"
		data.ErrorMessage = "Giriş yapılırken hata oluştu"
		renderTemplate(w, "login.html", data)
		return
	}

	fmt.Printf("DEBUG: Token generated, setting cookie\n")
	// Cookie'ye token'ı kaydet
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   3600, // 1 saat
		HttpOnly: true,
	})

	// Role göre yönlendir
	redirectURL := "/admin"
	if person.Role == "editor" {
		redirectURL = "/editor"
	}
	fmt.Printf("DEBUG: Cookie set, redirecting to %s\n", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// WebRegisterHandler web üzerinden kayıt yapar
func WebRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Form'u parse et (multipart/form-data için ParseMultipartForm kullanılmalı)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		fmt.Printf("DEBUG: Form parse error: %v\n", err)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Form verilerini işle
	req := models.CreatePersonRequest{
		Name:      r.FormValue("name"),
		Surname:   r.FormValue("surname"),
		Email:     r.FormValue("email"),
		Age:       shared.ParseIntFromForm(r.FormValue("age")),
		Phone:     r.FormValue("phone"),
		PhotoPath: "", // Bu sonradan doldurulacak
		Role:      r.FormValue("role"),
		Password:  r.FormValue("password"),
	}

	fmt.Printf("DEBUG: WebRegister - Name: %s, Email: %s, Password: %s, Role: %s\n",
		req.Name, req.Email, req.Password, req.Role)

	// Validasyon
	validator := shared.NewValidator()
	validator.ValidateCreatePersonRequest(req)

	if validator.HasError() {
		data := shared.GetTemplateData(r)
		data.Title = "Kayıt Ol"
		data.ErrorMessage = validator.GetError()
		renderTemplate(w, "register.html", data)
		return
	}

	// Fotoğraf yükle
	file, header, err := r.FormFile("photo")
	if err == nil {
		photoPath, err := repository.UploadPhoto(file, header)
		if err != nil {
			shared.LogError("UPLOAD_ERROR", "Photo upload failed", map[string]interface{}{
				"error": err.Error(),
				"email": req.Email,
			})
			data := shared.GetTemplateData(r)
			data.Title = "Kayıt Ol"
			data.ErrorMessage = "Fotoğraf yüklenemedi: " + err.Error()
			renderTemplate(w, "register.html", data)
			return
		}
		req.PhotoPath = photoPath
	}

	// Kayıt işlemini yap
	ctx := &registrationContext{Req: req}
	if err := runRegistrationPipeline(ctx); err != nil {
		shared.LogAuth("REGISTER_FAILED", req.Email, err.Error())
		data := shared.GetTemplateData(r)
		data.Title = "Kayıt Ol"
		if err == errEmailAlreadyExists {
			data.ErrorMessage = "Bu email zaten kayıtlı"
		} else {
			data.ErrorMessage = "Kayıt olurken hata oluştu: " + err.Error()
		}
		renderTemplate(w, "register.html", data)
		return
	}

	// Başarılı kayıt log'u
	shared.LogAuth("REGISTER_SUCCESS", req.Email, "User registered successfully")

	// Başarılı kayıt sonrası admin paneline yönlendir (eğer giriş yapmışsa)
	data := shared.GetTemplateData(r)
	if data.IsAuthenticated {
		// Admin paneline yönlendir ve başarı mesajını session'a kaydet
		http.SetCookie(w, &http.Cookie{
			Name:     "success_message",
			Value:    "Kullanıcı başarıyla eklendi!",
			Path:     "/",
			MaxAge:   5,
			HttpOnly: false,
		})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		// Login sayfasına yönlendir
		http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
	}
}

// WebLogoutHandler web üzerinden çıkış yapar
func WebLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Cookie'yi sil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// WebDeletePersonHandler web üzerinden kullanıcı siler
func WebDeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı sil
	err = repository.DeletePerson(id)
	if err != nil {
		shared.LogError("WEB_DELETE_ERROR", "Failed to delete user", map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		})
		http.Error(w, "Kullanıcı silinemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Başarılı silme log'u
	shared.LogAuth("USER_DELETED", fmt.Sprintf("ID: %d", id), "User deleted successfully via web")

	// JSON response döndür
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Kullanıcı başarıyla silindi"}`))
}

// renderTemplate template'i render eder
func renderTemplate(w http.ResponseWriter, templateName string, data shared.TemplateData) {
	shared.LogInfo("DEBUG", "Rendering template", map[string]interface{}{
		"template_name": templateName,
		"title":         data.Title,
		"user_role":     data.UserRole,
	})

	// Template'leri parse et
	templates, err := template.ParseFiles(
		"templates/layout.html",
		"templates/"+templateName,
	)
	if err != nil {
		shared.LogError("DEBUG", "Template parse error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Template'i execute et
	err = templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		shared.LogError("DEBUG", "Template execute error", map[string]interface{}{
			"error": err.Error(),
			"data":  data,
		})
		http.Error(w, "Template execute error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.LogInfo("DEBUG", "Template rendered successfully", map[string]interface{}{
		"template_name": templateName,
	})
}
