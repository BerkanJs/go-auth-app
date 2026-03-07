package handlers

import (
	"net/http"
	"strconv"
	"time"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"
)

// BlogPageHandler blog sayfasını gösterir
func BlogPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)

	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data.Title = "Blog Yönetimi"

	// Kullanıcının rolüne göre blog'ları getir
	var blogs []models.Blog
	var err error

	if data.UserRole == "admin" {
		blogs, err = repository.GetAllBlogs()
	} else {
		// Editor ise sadece kendi blog'larını görebilir
		claims, _ := shared.ParseAccessToken(getTokenFromCookie(r))
		blogs, err = repository.GetBlogsByAuthor(claims.UserID)
	}

	if err != nil {
		data.ErrorMessage = "Blog'lar yüklenirken hata oluştu: " + err.Error()
		shared.LogError("BLOG_LOAD_ERROR", "Failed to load blogs", map[string]interface{}{
			"error":     err.Error(),
			"user_role": data.UserRole,
		})
		renderTemplate(w, "blog.html", data)
		return
	}

	data.Blogs = models.ToBlogResponseList(blogs)
	renderTemplate(w, "blog.html", data)
}

// CreateBlogHandler blog oluşturur
func CreateBlogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/blogs", http.StatusSeeOther)
		return
	}

	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Form verilerini işle
	claims, _ := ParseAccessToken(getTokenFromCookie(r))

	req := models.CreateBlogRequest{
		Title:     r.FormValue("title"),
		Content:   r.FormValue("content"),
		Summary:   r.FormValue("summary"),
		Published: r.FormValue("published") == "on",
	}

	// Fotoğraf yükle
	file, header, err := r.FormFile("image")
	if err == nil {
		imagePath, err := repository.UploadPhoto(file, header)
		if err != nil {
			data.ErrorMessage = "Fotoğraf yüklenemedi: " + err.Error()
			renderTemplate(w, "blog.html", data)
			return
		}
		req.ImagePath = imagePath
	}

	// Blog oluştur
	blog := models.Blog{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		ImagePath:  req.ImagePath,
		AuthorID:   claims.UserID,
		AuthorName: data.UserName,
		Published:  req.Published,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err = repository.CreateBlog(blog)
	if err != nil {
		data.ErrorMessage = "Blog oluşturulurken hata oluştu: " + err.Error()
		renderTemplate(w, "blog.html", data)
		return
	}

	// Başarılı mesajı ayarla ve yönlendir
	http.SetCookie(w, &http.Cookie{
		Name:     "success_message",
		Value:    "Blog başarıyla oluşturuldu!",
		Path:     "/",
		MaxAge:   5,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

// UpdateBlogHandler blog günceller
func UpdateBlogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/blogs", http.StatusSeeOther)
		return
	}

	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Blog ID'sini al
	blogID, err := strconv.Atoi(r.FormValue("blog_id"))
	if err != nil {
		data.ErrorMessage = "Geçersiz blog ID"
		renderTemplate(w, "blog.html", data)
		return
	}

	// Blog'u kontrol et (yetki kontrolü)
	blog, err := repository.GetBlogByID(blogID)
	if err != nil {
		data.ErrorMessage = "Blog bulunamadı"
		renderTemplate(w, "blog.html", data)
		return
	}

	claims, _ := ParseAccessToken(getTokenFromCookie(r))

	// Yetki kontrolü: admin herkesin blog'unu düzenleyebilir, editor sadece kendi blog'unu
	if data.UserRole != "admin" && blog.AuthorID != claims.UserID {
		data.ErrorMessage = "Bu blog'u düzenleme yetkiniz yok"
		renderTemplate(w, "blog.html", data)
		return
	}

	// Form verilerini işle
	req := models.UpdateBlogRequest{
		ID:        blogID,
		Title:     r.FormValue("title"),
		Content:   r.FormValue("content"),
		Summary:   r.FormValue("summary"),
		Published: r.FormValue("published") == "on",
	}

	// Fotoğraf yükle
	file, header, err := r.FormFile("image")
	if err == nil {
		imagePath, err := repository.UploadPhoto(file, header)
		if err != nil {
			data.ErrorMessage = "Fotoğraf yüklenemedi: " + err.Error()
			renderTemplate(w, "blog.html", data)
			return
		}
		req.ImagePath = imagePath
	} else {
		req.ImagePath = blog.ImagePath // Mevcut fotoğrafı koru
	}

	// Blog güncelle
	updatedBlog := models.Blog{
		ID:         req.ID,
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		ImagePath:  req.ImagePath,
		AuthorID:   blog.AuthorID,
		AuthorName: blog.AuthorName,
		Published:  req.Published,
		CreatedAt:  blog.CreatedAt,
		UpdatedAt:  time.Now(),
	}

	err = repository.UpdateBlog(updatedBlog)
	if err != nil {
		data.ErrorMessage = "Blog güncellenirken hata oluştu: " + err.Error()
		renderTemplate(w, "blog.html", data)
		return
	}

	// Başarılı mesajı ayarla ve yönlendir
	http.SetCookie(w, &http.Cookie{
		Name:     "success_message",
		Value:    "Blog başarıyla güncellendi!",
		Path:     "/",
		MaxAge:   5,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

// DeleteBlogHandler blog siler
func DeleteBlogHandler(w http.ResponseWriter, r *http.Request) {
	blogID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Geçersiz blog ID", http.StatusBadRequest)
		return
	}

	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Blog'u kontrol et
	blog, err := repository.GetBlogByID(blogID)
	if err != nil {
		data.ErrorMessage = "Blog bulunamadı"
		renderTemplate(w, "blog.html", data)
		return
	}

	claims, _ := ParseAccessToken(getTokenFromCookie(r))

	// Yetki kontrolü
	if data.UserRole != "admin" && blog.AuthorID != claims.UserID {
		data.ErrorMessage = "Bu blog'u silme yetkiniz yok"
		renderTemplate(w, "blog.html", data)
		return
	}

	// Blog sil
	err = repository.DeleteBlog(blogID)
	if err != nil {
		data.ErrorMessage = "Blog silinirken hata oluştu: " + err.Error()
		renderTemplate(w, "blog.html", data)
		return
	}

	// Başarılı mesajı ayarla ve yönlendir
	http.SetCookie(w, &http.Cookie{
		Name:     "success_message",
		Value:    "Blog başarıyla silindi!",
		Path:     "/",
		MaxAge:   5,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

// getTokenFromCookie cookie'den token'ı alır
func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
