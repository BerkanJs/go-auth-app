// Package integration, gerçek HTTP katmanını uçtan uca test eder.
// Unit testlerin aksine burada gerçek server, gerçek route'lar ve
// in-memory SQLite veritabanı kullanılır.
package integration_test

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"

	"go-kisi-api/repository"
	"go-kisi-api/routes"

	_ "modernc.org/sqlite"
)

// testServer, tüm integration testlerinin kullandığı ortak HTTP test sunucusu.
var testServer *httptest.Server

// testDB, testler arasında temizlenebilen in-memory veritabanı.
var testDB *sql.DB

// TestMain, integration test suite başlamadan önce bir kez çalışır.
// Gerçek sunucunun tam yığınını (DB → Repo → Service → Handler → Route) kurar.
func TestMain(m *testing.M) {
	var err error

	// In-memory SQLite: her test çalışmasında temiz bir DB
	testDB, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		panic("test DB açılamadı: " + err.Error())
	}
	defer testDB.Close()

	if err := createSchema(testDB); err != nil {
		panic("schema oluşturulamadı: " + err.Error())
	}

	// shared/web_helpers.go gibi paket düzeyinde wrapper kullanan yerler için
	repository.SetDB(testDB)

	// AppBuilder: gerçek uygulamayla aynı bağımlılık grafiği, test DB ile
	routes.NewAppBuilder(testDB).Build()

	// Tüm testler için tek bir HTTP test sunucusu
	testServer = httptest.NewServer(nil) // http.DefaultServeMux kullanır
	defer testServer.Close()

	os.Exit(m.Run())
}

// createSchema, db/db.go'daki Init() ile aynı tabloları in-memory DB'de oluşturur.
func createSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS people (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			surname TEXT,
			email TEXT NOT NULL UNIQUE,
			age INTEGER,
			phone TEXT,
			photo_path TEXT,
			role TEXT NOT NULL DEFAULT 'editor',
			password_hash TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL UNIQUE,
			revoked INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			revoked_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS blogs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			summary TEXT,
			image_path TEXT,
			author_id INTEGER NOT NULL,
			author_name TEXT NOT NULL,
			published INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (author_id) REFERENCES people(id)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// cleanTables, testler arasında DB'yi temizler.
// Her testin bağımsız başlamasını sağlar.
func cleanTables(t *testing.T) {
	t.Helper()
	tables := []string{"refresh_tokens", "blogs", "people"}
	for _, tbl := range tables {
		if _, err := testDB.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("tablo temizlenemedi (%s): %v", tbl, err)
		}
	}
}
