package handlers

import (
	"net/http"

	"go-kisi-api/models"
	"go-kisi-api/service"
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

	claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	blogs, err := service.GetBlogsForUser(data.UserRole, claims.UserID)
	if err != nil {
		shared.LogError("EDITOR_PAGE_ERROR", "Failed to load blogs", map[string]interface{}{"error": err.Error(), "user_id": claims.UserID})
		data.ErrorMessage = "Blog'lar yüklenirken bir hata oluştu."
		renderTemplate(w, "editor.html", data)
		return
	}

	data.Blogs = models.ToBlogResponseList(blogs)
	renderTemplate(w, "editor.html", data)
}
