package repository

import (
	"context"
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

func (r *SQLiteAuthRepo) SaveRefreshToken(ctx context.Context, userID int, token string) error {
	_, err := r.db.ExecContext(ctx,
		queries.InsertRefreshToken,
		userID,
		token,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteAuthRepo) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	var revoked int
	err := r.db.QueryRowContext(ctx, queries.SelectRefreshTokenRevoked, token).Scan(&revoked)
	if err != nil {
		return false, nil
	}
	return revoked == 0, nil
}

func (r *SQLiteAuthRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx,
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
func SaveRefreshToken(ctx context.Context, userID int, token string) error {
	return defaultAuthRepo.SaveRefreshToken(ctx, userID, token)
}
func IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	return defaultAuthRepo.IsRefreshTokenValid(ctx, token)
}
func RevokeRefreshToken(ctx context.Context, token string) error {
	return defaultAuthRepo.RevokeRefreshToken(ctx, token)
}
