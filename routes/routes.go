package routes

import (
	"net/http"
	"time"

	"go-kisi-api/handlers"
	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"

	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes() {
	// Repo'ları oluştur
	authRepo := repository.NewAuthRepo()
	blogRepo := repository.NewBlogRepo()
	personRepo := repository.NewPersonRepo()

	// Service'leri oluştur
	authSvc := service.NewAuthService(authRepo, personRepo)
	blogSvc := service.NewBlogService(blogRepo)
	personSvc := service.NewPersonService(personRepo)

	// Handler'ları oluştur
	authH := handlers.NewAuthHandler(authSvc)
	blogH := handlers.NewBlogHandler(blogSvc)
	personH := handlers.NewPersonHandler(personRepo)
	webH := handlers.NewWebHandler(authSvc, personRepo, blogRepo, personSvc)

	loginLimiter := shared.NewRateLimiter(10, time.Minute)

	// Web form işlemleri
	http.HandleFunc("/web-login", loginLimiter.Middleware(webH.WebLoginHandler))
	http.HandleFunc("/web-register", webH.WebRegisterHandler)
	http.HandleFunc("/web-logout", handlers.WebLogoutHandler)
	http.HandleFunc("/user/update", handlers.AdminMiddleware(webH.UpdateUserHandler))
	http.HandleFunc("/web-delete", handlers.AdminMiddleware(webH.WebDeletePersonHandler))

	// Blog işlemleri
	http.HandleFunc("/blog/create", handlers.EditorMiddleware(blogH.CreateBlogHandler))
	http.HandleFunc("/blog/update", handlers.EditorMiddleware(blogH.UpdateBlogHandler))
	http.HandleFunc("/blog/delete", handlers.EditorMiddleware(blogH.DeleteBlogHandler))

	// Statik dosyalar
	http.HandleFunc("/static/", handlers.StaticHandler)
	http.HandleFunc("/uploads/", handlers.StaticHandler)

	// API endpoint'leri
	http.HandleFunc("/api/login", loginLimiter.Middleware(authH.LoginHandler))
	http.HandleFunc("/api/refresh", authH.RefreshHandler)
	http.HandleFunc("/api/logout", handlers.JwtAuthMiddleware(authH.LogoutHandler))
	http.HandleFunc("/api/health", handlers.HealthHandler)
	http.HandleFunc("/api/add", personH.AddPersonHandler)
	http.HandleFunc("/api/all", handlers.JwtAuthMiddleware(personH.GetAllPeopleHandler))
	http.HandleFunc("/api/get", handlers.JwtAuthMiddleware(personH.GetPersonByIDHandler))
	http.HandleFunc("/api/delete", handlers.JwtAuthMiddleware(personH.DeletePersonHandler))

	// Web arayüzü route'ları
	http.HandleFunc("/", webH.HomeHandler)
	http.HandleFunc("/login", webH.LoginPageHandler)
	http.HandleFunc("/register", webH.RegisterPageHandler)
	http.HandleFunc("/admin", handlers.AdminMiddleware(webH.AdminPageHandler))
	http.HandleFunc("/editor", handlers.EditorMiddleware(blogH.EditorPageHandler))
	http.HandleFunc("/blogs", handlers.EditorMiddleware(blogH.BlogPageHandler))

	http.Handle("/swagger/", httpSwagger.WrapHandler)
}
