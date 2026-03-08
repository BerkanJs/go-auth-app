package repository

import (
	"time"

	"go-kisi-api/db"
	"go-kisi-api/queries"
)

// SQLiteAuthRepo, AuthRepository'yi SQLite üzerinde implement eder.
type SQLiteAuthRepo struct{}

func NewAuthRepo() AuthRepository {
	return &SQLiteAuthRepo{}
}

func (r *SQLiteAuthRepo) SaveRefreshToken(userID int, token string) error {
	_, err := db.DB.Exec(
		queries.InsertRefreshToken,
		userID,
		token,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteAuthRepo) IsRefreshTokenValid(token string) (bool, error) {
	var revoked int
	err := db.DB.QueryRow(queries.SelectRefreshTokenRevoked, token).Scan(&revoked)
	if err != nil {
		return false, nil
	}
	return revoked == 0, nil
}

func (r *SQLiteAuthRepo) RevokeRefreshToken(token string) error {
	_, err := db.DB.Exec(
		queries.RevokeRefreshTokenQuery,
		time.Now().UTC().Format(time.RFC3339),
		token,
	)
	return err
}

// defaultAuthRepo, geriye dönük uyumluluk için kullanılan paket düzeyindeki örnek.
var defaultAuthRepo AuthRepository = &SQLiteAuthRepo{}

// Paket düzeyinde wrapper fonksiyonlar — shared ve diğer paketler bunları kullanmaya devam eder.
func SaveRefreshToken(userID int, token string) error {
	return defaultAuthRepo.SaveRefreshToken(userID, token)
}

func IsRefreshTokenValid(token string) (bool, error) {
	return defaultAuthRepo.IsRefreshTokenValid(token)
}

func RevokeRefreshToken(token string) error {
	return defaultAuthRepo.RevokeRefreshToken(token)
}
