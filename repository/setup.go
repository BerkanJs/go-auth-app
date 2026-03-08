package repository

import "database/sql"

// SetDB, paket düzeyindeki varsayılan repository örneklerini başlatır.
// Bu fonksiyon main() içinde db.Init()'ten hemen sonra çağrılmalıdır.
// Amaç: shared/web_helpers.go gibi doğrudan paket fonksiyonlarını kullanan
// yerler için geriye dönük uyumluluğu korumak.
func SetDB(database *sql.DB) {
	defaultPersonRepo = &SQLitePersonRepo{db: database}
	defaultBlogRepo = &SQLiteBlogRepo{db: database}
	defaultAuthRepo = &SQLiteAuthRepo{db: database}
}
