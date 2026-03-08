package service

import (
	"context"
	"errors"
	"time"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("geçersiz email veya şifre")

type authClaims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

type authService struct {
	authRepo      repository.AuthRepository
	personRepo    repository.PersonRepository
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewAuthService(ar repository.AuthRepository, pr repository.PersonRepository) AuthService {
	cfg := shared.GetConfig()
	return &authService{
		authRepo:      ar,
		personRepo:    pr,
		accessSecret:  []byte(cfg.JWTAccessSecret),
		refreshSecret: []byte(cfg.JWTRefreshSecret),
		accessTTL:     time.Duration(cfg.AccessTokenTTL) * time.Second,
		refreshTTL:    time.Duration(cfg.RefreshTokenTTL) * time.Second,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (models.Person, error) {
	person, err := s.personRepo.GetPersonByEmail(ctx, email)
	if err != nil {
		return models.Person{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(person.PasswordHash), []byte(password)); err != nil {
		return models.Person{}, ErrInvalidCredentials
	}
	return person, nil
}

func (s *authService) GenerateAccessToken(userID int) (string, error) {
	claims := authClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

func (s *authService) GenerateRefreshToken(ctx context.Context, userID int) (string, error) {
	claims := authClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.refreshSecret)
	if err != nil {
		return "", err
	}
	if err := s.authRepo.SaveRefreshToken(ctx, userID, tokenStr); err != nil {
		return "", err
	}
	return tokenStr, nil
}

func (s *authService) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	return s.authRepo.IsRefreshTokenValid(ctx, token)
}

func (s *authService) ParseRefreshToken(tokenStr string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.refreshSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*authClaims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, jwt.ErrTokenInvalidClaims
}

func (s *authService) RevokeRefreshToken(ctx context.Context, token string) error {
	return s.authRepo.RevokeRefreshToken(ctx, token)
}
