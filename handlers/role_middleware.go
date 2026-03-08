package handlers

import (
	"net/http"

	"go-kisi-api/repository"
	"go-kisi-api/shared"
)

// RoleMiddleware belirli rollere erişim kontrolü yapar.
func RoleMiddleware(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			claims, err := shared.ParseAccessToken(cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			person, err := repository.GetPersonByID(claims.UserID)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			hasPermission := false
			for _, role := range allowedRoles {
				if person.Role == role {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, "Yetkiniz yok", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// AdminMiddleware sadece admin rolüne erişim verir.
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return RoleMiddleware("admin")(next)
}

// EditorMiddleware editor ve admin rollerine erişim verir.
func EditorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return RoleMiddleware("editor", "admin")(next)
}
