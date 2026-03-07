package handlers

import (
	"net/http"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"
)

// EditorPageHandler editor panelini gösterir
func EditorPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GetTemplateData(r)

	if !data.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data.Title = "Editor Panel"

	// Editörün kendi bloglarını getir
	claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var blogs []models.Blog
	if data.UserRole == "admin" {
		blogs, err = repository.GetAllBlogs()
	} else {
		blogs, err = repository.GetBlogsByAuthor(claims.UserID)
	}

	if err != nil {
		data.ErrorMessage = "Bloglar yüklenirken hata oluştu: " + err.Error()
		renderTemplate(w, "editor.html", data)
		return
	}

	data.Blogs = models.ToBlogResponseList(blogs)
	renderTemplate(w, "editor.html", data)
}
