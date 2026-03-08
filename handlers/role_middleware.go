package handlers

import (
	"net/http"

	"go-kisi-api/repository"
	"go-kisi-api/shared"
)

// RoleChecker, PersonRepository bağımlılığını enjekte ederek
// rol tabanlı yetki kontrolü yapar. Global paket fonksiyonları yerine
// struct metotları kullanılır — test edilebilirlik ve esneklik sağlar.
type RoleChecker struct {
	personRepo repository.PersonRepository
}

// NewRoleChecker, belirtilen repo ile bir RoleChecker oluşturur.
func NewRoleChecker(personRepo repository.PersonRepository) *RoleChecker {
	return &RoleChecker{personRepo: personRepo}
}

// Middleware, izin verilen rolleri kontrol eden bir HTTP middleware döner.
func (rc *RoleChecker) Middleware(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
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

			// Enjekte edilen repo kullanılır — global paket fonksiyonu değil
			person, err := rc.personRepo.GetPersonByID(claims.UserID)
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

// AdminMiddleware, yalnızca admin rolüne erişim verir.
func (rc *RoleChecker) AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return rc.Middleware("admin")(next)
}

// EditorMiddleware, editor ve admin rollerine erişim verir.
func (rc *RoleChecker) EditorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return rc.Middleware("editor", "admin")(next)
}
