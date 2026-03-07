// @title Go Kisi API
// @version 1.0
// @description Kişi yönetimi ve kimlik doğrulama için basit API.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"go-kisi-api/db"
	"go-kisi-api/routes"
	"go-kisi-api/shared"
	"net/http"

	_ "go-kisi-api/docs"
)

func main() {
	db.Init()
	routes.RegisterRoutes()
	fmt.Println("Server çalışıyor :8080")

	// Global middleware zinciri: Logging -> CORS -> mux
	var handler http.Handler = http.DefaultServeMux
	handler = shared.CorsMiddleware(handler)
	handler = shared.LoggingMiddleware(handler)

	http.ListenAndServe(":8080", handler)
}
