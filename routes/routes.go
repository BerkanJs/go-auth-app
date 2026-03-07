package routes

import (
	"go-kisi-api/handlers"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes() {
	// Auth endpoints (public)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/refresh", handlers.RefreshHandler)
	http.HandleFunc("/logout", handlers.JwtAuthMiddleware(handlers.LogoutHandler))

	// Sağlık kontrolü (monitoring için temel endpoint)
	http.HandleFunc("/health", handlers.HealthHandler)

	// Kişi endpoint'leri (JWT ile korumalı)
	http.HandleFunc("/add", handlers.JwtAuthMiddleware(handlers.AddPersonHandler))
	http.HandleFunc("/all", handlers.JwtAuthMiddleware(handlers.GetAllPeopleHandler))
	http.HandleFunc("/get", handlers.JwtAuthMiddleware(handlers.GetPersonByIDHandler))
	http.HandleFunc("/delete", handlers.JwtAuthMiddleware(handlers.DeletePersonHandler))

	http.Handle("/swagger/", httpSwagger.WrapHandler)
}
