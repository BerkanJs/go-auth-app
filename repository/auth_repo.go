package repository

import (
	"database/sql"
	"time"

	"go-kisi-api/queries"
)

// SQLiteAuthRepo, AuthRepository'yi SQLite üzerinde implement eder.
// db alanı constructor üzerinden enjekte edilir; global db.DB bağımlılığı yoktur.
type SQLiteAuthRepo struct {
	db *sql.DB
}

// NewAuthRepo, bağımlılık enjeksiyonuyla bir AuthRepository oluşturur.
func NewAuthRepo(database *sql.DB) AuthRepository {
	return &SQLiteAuthRepo{db: database}
}

func (r *SQLiteAuthRepo) SaveRefreshToken(userID int, token string) error {
	_, err := r.db.Exec(
		queries.InsertRefreshToken,
		userID,
		token,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteAuthRepo) IsRefreshTokenValid(token string) (bool, error) {
	var revoked int
	err := r.db.QueryRow(queries.SelectRefreshTokenRevoked, token).Scan(&revoked)
	if err != nil {
		return false, nil
	}
	return revoked == 0, nil
}

func (r *SQLiteAuthRepo) RevokeRefreshToken(token string) error {
	_, err := r.db.Exec(
		queries.RevokeRefreshTokenQuery,
		time.Now().UTC().Format(time.RFC3339),
		token,
	)
	return err
}

// defaultAuthRepo, paket düzeyinde wrapper fonksiyonlar için kullanılır.
// SetDB() çağrısıyla başlatılır; yeni kod için doğrudan AuthRepository arayüzünü tercih edin.
var defaultAuthRepo AuthRepository

// Paket düzeyinde wrapper fonksiyonlar — geriye dönük uyumluluk için korunur.
func SaveRefreshToken(userID int, token string) error { return defaultAuthRepo.SaveRefreshToken(userID, token) }
func IsRefreshTokenValid(token string) (bool, error)  { return defaultAuthRepo.IsRefreshTokenValid(token) }
func RevokeRefreshToken(token string) error           { return defaultAuthRepo.RevokeRefreshToken(token) }
