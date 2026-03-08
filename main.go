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
	"go-kisi-api/repository"
	"go-kisi-api/routes"
	"go-kisi-api/shared"
	"net/http"

	_ "go-kisi-api/docs"
	_ "net/http/pprof"
)

func main() {
	db.Init()
	// Repository Pattern: paket düzeyindeki varsayılan repo'ları başlat
	// (shared/web_helpers.go gibi wrapper fonksiyon kullanan yerler için)
	repository.SetDB(db.DB)
	fmt.Println("Server çalışıyor :8081")

	// Global middleware zinciri: Logging -> CORS -> mux
	var handler http.Handler = http.DefaultServeMux
	handler = shared.CorsMiddleware(handler)
	handler = shared.LoggingMiddleware(handler)

	// Route'ları middleware'den SONRA kaydet
	routes.RegisterRoutes()

	http.ListenAndServe(":8081", handler)
}
