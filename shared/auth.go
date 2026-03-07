package shared

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	config           = GetConfig()
	jwtAccessSecret  = []byte(config.JWTAccessSecret)
	jwtRefreshSecret = []byte(config.JWTRefreshSecret)
	accessTokenTTL   = time.Duration(config.AccessTokenTTL) * time.Second
	refreshTokenTTL  = time.Duration(config.RefreshTokenTTL) * time.Second
)

type jwtClaims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// ParseAccessToken access token'ı parse eder
func ParseAccessToken(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtAccessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
