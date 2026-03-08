package handlers

import (
	"net/http"

	"go-kisi-api/shared"
)

// PageRenderer, sayfa render işlemi için Template Method pattern arayüzü.
// Her sayfa türü bu arayüzü implemente ederek sadece kendine özgü kısımları tanımlar.
type PageRenderer interface {
	RequiresAuth() bool
	Title() string
	TemplateName() string
	LoadData(data *shared.TemplateData, userID int) error
}

// RenderPage, Template Method: ortak iskelet burada tanımlıdır.
// Auth kontrolü, token parse, başlık atama ve template render
// değişmeyen adımlardır. Yalnızca LoadData her sayfa için farklıdır.
func RenderPage(w http.ResponseWriter, r *http.Request, renderer PageRenderer) {
	data := shared.GetTemplateData(r)

	var userID int
	if renderer.RequiresAuth() {
		if !data.IsAuthenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		claims, err := shared.ParseAccessToken(getTokenFromCookie(r))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID = claims.UserID
	}

	data.Title = renderer.Title()

	if err := renderer.LoadData(&data, userID); err != nil {
		shared.LogError("PAGE_LOAD_ERROR", "Failed to load page data", map[string]interface{}{
			"template": renderer.TemplateName(),
			"error":    err.Error(),
		})
		data.ErrorMessage = "Veriler yüklenirken bir hata oluştu."
	}

	renderTemplate(w, renderer.TemplateName(), data)
}
