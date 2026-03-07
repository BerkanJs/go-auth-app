package repository

import (
	"time"

	"go-kisi-api/db"
	"go-kisi-api/queries"
)

// SaveRefreshToken verilen kullanıcı için üretilen refresh token'ı saklar.
func SaveRefreshToken(userID int, token string) error {
	_, err := db.DB.Exec(
		queries.InsertRefreshToken,
		userID,
		token,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// IsRefreshTokenValid refresh token'ın veritabanında var ve revoke edilmemiş olduğunu kontrol eder.
func IsRefreshTokenValid(token string) (bool, error) {
	var revoked int
	err := db.DB.QueryRow(queries.SelectRefreshTokenRevoked, token).Scan(&revoked)
	if err != nil {
		// satır yoksa veya başka bir hata varsa, token geçersiz sayılır
		return false, nil
	}
	return revoked == 0, nil
}

// RevokeRefreshToken verilen refresh token'ı revoke eder.
func RevokeRefreshToken(token string) error {
	_, err := db.DB.Exec(
		queries.RevokeRefreshTokenQuery,
		time.Now().UTC().Format(time.RFC3339),
		token,
	)
	return err
}

