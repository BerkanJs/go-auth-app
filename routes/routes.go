package routes

import (
	"go-kisi-api/handlers"
	"go-kisi-api/shared"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes() {
	// Login endpoint'leri için rate limiter: dakikada 10 istek
	loginLimiter := shared.NewRateLimiter(10, time.Minute)

	// Web form işlemleri (spesifik route'lar önce)
	http.HandleFunc("/web-login", loginLimiter.Middleware(handlers.WebLoginHandler))
	http.HandleFunc("/web-register", handlers.WebRegisterHandler)
	http.HandleFunc("/web-logout", handlers.WebLogoutHandler)
	http.HandleFunc("/user/update", handlers.AdminMiddleware(handlers.UpdateUserHandler))
	http.HandleFunc("/web-delete", handlers.AdminMiddleware(handlers.WebDeletePersonHandler))

	// Blog işlemleri
	http.HandleFunc("/blog/create", handlers.EditorMiddleware(handlers.CreateBlogHandler))
	http.HandleFunc("/blog/update", handlers.EditorMiddleware(handlers.UpdateBlogHandler))
	http.HandleFunc("/blog/delete", handlers.EditorMiddleware(handlers.DeleteBlogHandler))

	// Statik dosyalar
	http.HandleFunc("/static/", handlers.StaticHandler)
	http.HandleFunc("/uploads/", handlers.StaticHandler)

	// API endpoint'leri
	http.HandleFunc("/api/login", loginLimiter.Middleware(handlers.LoginHandler))
	http.HandleFunc("/api/refresh", handlers.RefreshHandler)
	http.HandleFunc("/api/logout", handlers.JwtAuthMiddleware(handlers.LogoutHandler))
	http.HandleFunc("/api/health", handlers.HealthHandler)
	http.HandleFunc("/api/add", handlers.AddPersonHandler)
	http.HandleFunc("/api/all", handlers.JwtAuthMiddleware(handlers.GetAllPeopleHandler))
	http.HandleFunc("/api/get", handlers.JwtAuthMiddleware(handlers.GetPersonByIDHandler))
	http.HandleFunc("/api/delete", handlers.JwtAuthMiddleware(handlers.DeletePersonHandler))

	// Web arayüzü route'ları (genel route'lar sonra)
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginPageHandler)
	http.HandleFunc("/register", handlers.RegisterPageHandler)
	http.HandleFunc("/admin", handlers.AdminMiddleware(handlers.AdminPageHandler))
	http.HandleFunc("/editor", handlers.EditorMiddleware(handlers.EditorPageHandler))
	http.HandleFunc("/blogs", handlers.EditorMiddleware(handlers.BlogPageHandler))

	http.Handle("/swagger/", httpSwagger.WrapHandler)
}
