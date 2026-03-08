package routes

import (
	"database/sql"
	"net/http"
	"time"

	"go-kisi-api/handlers"
	"go-kisi-api/repository"
	"go-kisi-api/service"
	"go-kisi-api/shared"

	httpSwagger "github.com/swaggo/http-swagger"
)

// AppBuilder, uygulama bağımlılık grafiğini ve route kayıtlarını inşa eder.
// Builder Pattern: nesne oluşturma adımları açık ve zincirleme ile yapılandırılabilir.
type AppBuilder struct {
	db          *sql.DB
	loginLimit  int
	loginWindow time.Duration
}

// NewAppBuilder, veritabanı bağlantısıyla bir AppBuilder oluşturur.
func NewAppBuilder(database *sql.DB) *AppBuilder {
	return &AppBuilder{
		db:          database,
		loginLimit:  10,
		loginWindow: time.Minute,
	}
}

// WithLoginRateLimit, login endpoint'i için rate limit ayarını değiştirir.
// Zincirleme kullanım için *AppBuilder döner.
func (b *AppBuilder) WithLoginRateLimit(limit int, window time.Duration) *AppBuilder {
	b.loginLimit = limit
	b.loginWindow = window
	return b
}

// --- İç yapılar: inşa edilen katmanları gruplar ---

type appRepos struct {
	auth   repository.AuthRepository
	blog   repository.BlogRepository
	person repository.PersonRepository
}

type appServices struct {
	auth   service.AuthService
	blog   service.BlogService
	person service.PersonService
}

type appHandlers struct {
	auth   *handlers.AuthHandler
	blog   *handlers.BlogHandler
	person *handlers.PersonHandler
	web    *handlers.WebHandler
	roles  *handlers.RoleChecker
}

// --- Builder adımları ---

func (b *AppBuilder) buildRepos() appRepos {
	return appRepos{
		auth:   repository.NewAuthRepo(b.db),
		blog:   repository.NewBlogRepo(b.db),
		person: repository.NewPersonRepo(b.db),
	}
}

func (b *AppBuilder) buildServices(repos appRepos) appServices {
	return appServices{
		auth:   service.NewAuthService(repos.auth, repos.person),
		blog:   service.NewBlogService(repos.blog),
		person: service.NewPersonService(repos.person),
	}
}

func (b *AppBuilder) buildHandlers(repos appRepos, svcs appServices) appHandlers {
	return appHandlers{
		auth:   handlers.NewAuthHandler(svcs.auth),
		blog:   handlers.NewBlogHandler(svcs.blog, repos.person),
		person: handlers.NewPersonHandler(svcs.person),
		web:    handlers.NewWebHandler(svcs.auth, repos.person, repos.blog, svcs.person),
		roles:  handlers.NewRoleChecker(repos.person),
	}
}

func (b *AppBuilder) registerRoutes(h appHandlers) {
	loginLimiter := shared.NewRateLimiter(b.loginLimit, b.loginWindow)

	// Web form işlemleri
	http.HandleFunc("/web-login", loginLimiter.Middleware(h.web.WebLoginHandler))
	http.HandleFunc("/web-register", h.web.WebRegisterHandler)
	http.HandleFunc("/web-logout", handlers.WebLogoutHandler)
	http.HandleFunc("/user/update", h.roles.AdminMiddleware(h.web.UpdateUserHandler))
	http.HandleFunc("/web-delete", h.roles.AdminMiddleware(h.web.WebDeletePersonHandler))

	// Blog işlemleri
	http.HandleFunc("/blog/create", h.roles.EditorMiddleware(h.blog.CreateBlogHandler))
	http.HandleFunc("/blog/update", h.roles.EditorMiddleware(h.blog.UpdateBlogHandler))
	http.HandleFunc("/blog/delete", h.roles.EditorMiddleware(h.blog.DeleteBlogHandler))

	// Statik dosyalar
	http.HandleFunc("/static/", handlers.StaticHandler)
	http.HandleFunc("/uploads/", handlers.StaticHandler)

	// API endpoint'leri
	http.HandleFunc("/api/login", loginLimiter.Middleware(h.auth.LoginHandler))
	http.HandleFunc("/api/refresh", h.auth.RefreshHandler)
	http.HandleFunc("/api/logout", handlers.JwtAuthMiddleware(h.auth.LogoutHandler))
	http.HandleFunc("/api/health", handlers.HealthHandler)
	http.HandleFunc("/api/add", h.person.AddPersonHandler)
	http.HandleFunc("/api/all", handlers.JwtAuthMiddleware(h.person.GetAllPeopleHandler))
	http.HandleFunc("/api/get", handlers.JwtAuthMiddleware(h.person.GetPersonByIDHandler))
	http.HandleFunc("/api/delete", handlers.JwtAuthMiddleware(h.person.DeletePersonHandler))
	http.HandleFunc("/api/update", handlers.JwtAuthMiddleware(h.person.UpdatePersonHandler))

	// Blog API endpoint'leri (React için)
	http.HandleFunc("/api/blogs", handlers.JwtAuthMiddleware(h.blog.ApiBlogListHandler))
	http.HandleFunc("/api/blogs/create", handlers.JwtAuthMiddleware(h.blog.ApiCreateBlogHandler))
	http.HandleFunc("/api/blogs/update", handlers.JwtAuthMiddleware(h.blog.ApiUpdateBlogHandler))
	http.HandleFunc("/api/blogs/delete", handlers.JwtAuthMiddleware(h.blog.ApiDeleteBlogHandler))

	// Web arayüzü route'ları
	http.HandleFunc("/", h.web.HomeHandler)
	http.HandleFunc("/login", h.web.LoginPageHandler)
	http.HandleFunc("/register", h.web.RegisterPageHandler)
	http.HandleFunc("/admin", h.roles.AdminMiddleware(h.web.AdminPageHandler))
	http.HandleFunc("/editor", h.roles.EditorMiddleware(h.blog.EditorPageHandler))
	http.HandleFunc("/blogs", h.roles.EditorMiddleware(h.blog.BlogPageHandler))

	http.Handle("/swagger/", httpSwagger.WrapHandler)
}

// Build, tüm adımları sırasıyla çalıştırır:
// Repo'lar → Service'ler → Handler'lar → Route Kaydı
func (b *AppBuilder) Build() {
	repos := b.buildRepos()
	svcs := b.buildServices(repos)
	h := b.buildHandlers(repos, svcs)
	b.registerRoutes(h)
}
