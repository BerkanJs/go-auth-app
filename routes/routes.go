package routes

import "go-kisi-api/db"

// RegisterRoutes, AppBuilder kullanarak tüm bağımlılıkları ve route'ları kurar.
// Yapılandırma değiştirmek için NewAppBuilder(...).WithLoginRateLimit(...).Build() kullanılabilir.
func RegisterRoutes() {
	NewAppBuilder(db.DB).Build()
}
