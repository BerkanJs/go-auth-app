package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"
)

// BlogHandler blog CRUD ve görüntüleme endpoint'lerini yönetir.
type BlogHandler struct {
	blogSvc service.BlogService
}

func NewBlogHandler(blogSvc service.BlogService) *BlogHandler {
	return &BlogHandler{blogSvc: blogSvc}
}

// blogPageRenderer, Blog Yönetimi sayfası için Template Method implementasyonu.
type blogPageRenderer struct {
	blogSvc service.BlogService
}

func (p *blogPageRenderer) RequiresAuth() bool  { return true }
func (p *blogPageRenderer) Title() string        { return "Blog Yönetimi" }
func (p *blogPageRenderer) TemplateName() string { return "blog.html" }
func (p *blogPageRenderer) LoadData(ctx context.Context, data *shared.TemplateData, userID int) error {
	blogs, err := p.blogSvc.GetBlogsForUser(ctx, data.UserRole, userID)
	if err != nil {
		shared.LogError("BLOG_PAGE_ERROR", "Failed to load blogs", map[string]interface{}{"error": err.Error()})
		return err
	}
	data.Blogs = models.ToBlogResponseList(blogs)
	return nil
}

func (h *BlogHandler) BlogPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderPage(w, r, &blogPageRenderer{blogSvc: h.blogSvc})
}

// editorPageRenderer, Editor Panel sayfası için Template Method implementasyonu.
type editorPageRenderer struct {
	blogSvc service.BlogService
}

func (p *editorPageRenderer) RequiresAuth() bool  { return true }
func (p *editorPageRenderer) Title() string        { return "Editor Panel" }
func (p *editorPageRenderer) TemplateName() string { return "editor.html" }
func (p *editorPageRenderer) LoadData(ctx context.Context, data *shared.TemplateData, userID int) error {
	blogs, err := p.blogSvc.GetBlogsForUser(ctx, data.UserRole, userID)
	if err != nil {
		shared.LogError("EDITOR_PAGE_ERROR", "Failed to load blogs", map[string]interface{}{"error": err.Error()})
		return err
	}
	data.Blogs = models.ToBlogResponseList(blogs)
	return nil
}

func (h *BlogHandler) EditorPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderPage(w, r, &editorPageRenderer{blogSvc: h.blogSvc})
}

func (h *BlogHandler) CreateBlogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/blogs", http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	summary := r.FormValue("summary")
	published := r.FormValue("published") == "on"

	var imagePath string
	file, header, err := r.FormFile("image")
	if err == nil {
		imagePath, err = repository.UploadPhoto(file, header)
		if err != nil {
			shared.LogError("BLOG_PHOTO_ERROR", "Photo upload failed", map[string]interface{}{"error": err.Error()})
			data.ErrorMessage = "Görsel yüklenemedi."
			renderTemplate(w, "blog.html", data)
			return
		}
	}

	if _, err := h.blogSvc.CreateBlog(ctx, title, content, summary, imagePath, published, claims.UserID, data.UserName); err != nil {
		shared.LogError("BLOG_CREATE_ERROR", "Failed to create blog", map[string]interface{}{"error": err.Error()})
		data.ErrorMessage = "Blog oluşturulurken bir hata oluştu."
		renderTemplate(w, "blog.html", data)
		return
	}

	shared.LogInfo("BLOG_CREATED", "Blog created successfully", map[string]interface{}{"author_id": claims.UserID})
	http.SetCookie(w, &http.Cookie{Name: "success_message", Value: "Blog başarıyla oluşturuldu!", Path: "/", MaxAge: 5, HttpOnly: false})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

func (h *BlogHandler) UpdateBlogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/blogs", http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	blogID, err := strconv.Atoi(r.FormValue("blog_id"))
	if err != nil {
		data.ErrorMessage = "Geçersiz blog ID."
		renderTemplate(w, "blog.html", data)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	summary := r.FormValue("summary")
	published := r.FormValue("published") == "on"

	var imagePath string
	file, header, err := r.FormFile("image")
	if err == nil {
		imagePath, err = repository.UploadPhoto(file, header)
		if err != nil {
			shared.LogError("BLOG_PHOTO_ERROR", "Photo upload failed", map[string]interface{}{"error": err.Error()})
			data.ErrorMessage = "Görsel yüklenemedi."
			renderTemplate(w, "blog.html", data)
			return
		}
	}

	err = h.blogSvc.UpdateBlog(ctx, blogID, title, content, summary, imagePath, published, data.UserRole, claims.UserID)
	if err != nil {
		shared.LogError("BLOG_UPDATE_ERROR", "Failed to update blog", map[string]interface{}{"blog_id": blogID, "error": err.Error()})
		switch {
		case errors.Is(err, service.ErrBlogNotFound):
			data.ErrorMessage = "Blog bulunamadı."
		case errors.Is(err, service.ErrPermissionDenied):
			data.ErrorMessage = "Bu blogu düzenleme yetkiniz yok."
		default:
			data.ErrorMessage = "Blog güncellenirken bir hata oluştu."
		}
		renderTemplate(w, "blog.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "success_message", Value: "Blog başarıyla güncellendi!", Path: "/", MaxAge: 5, HttpOnly: false})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

func (h *BlogHandler) DeleteBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := shared.GetTemplateData(r)
	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	blogID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		data.ErrorMessage = "Geçersiz blog ID."
		renderTemplate(w, "blog.html", data)
		return
	}

	claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = h.blogSvc.DeleteBlog(ctx, blogID, data.UserRole, claims.UserID)
	if err != nil {
		shared.LogError("BLOG_DELETE_ERROR", "Failed to delete blog", map[string]interface{}{"blog_id": blogID, "error": err.Error()})
		switch {
		case errors.Is(err, service.ErrBlogNotFound):
			data.ErrorMessage = "Blog bulunamadı."
		case errors.Is(err, service.ErrPermissionDenied):
			data.ErrorMessage = "Bu blogu silme yetkiniz yok."
		default:
			data.ErrorMessage = "Blog silinirken bir hata oluştu."
		}
		renderTemplate(w, "blog.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "success_message", Value: "Blog başarıyla silindi!", Path: "/", MaxAge: 5, HttpOnly: false})
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
