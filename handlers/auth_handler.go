package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go-kisi-api/models"
	"go-kisi-api/repository"
	"go-kisi-api/shared"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT ile ilgili sabitler.
var (
	jwtAccessSecret  = []byte("super-secret-access-key")
	jwtRefreshSecret = []byte("super-secret-refresh-key")
	accessTokenTTL   = 15 * time.Minute
	refreshTokenTTL  = 7 * 24 * time.Hour
)

type jwtClaims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

type contextKey string

const userIDContextKey contextKey = "userID"

// buildPersonFromCreateRequest CreatePersonRequest -> Person dönüşümü ve şifre hashleme yapar.
func buildPersonFromCreateRequest(req models.CreatePersonRequest) (models.Person, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Person{}, err
	}

	return models.Person{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		Age:          req.Age,
		Phone:        req.Phone,
		PasswordHash: string(hashed),
	}, nil
}

func generateAccessToken(userID int) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtAccessSecret)
}

func generateRefreshToken(userID int) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtRefreshSecret)
	if err != nil {
		return "", err
	}

	// üretilen refresh token'ı veritabanında sakla
	if err := repository.SaveRefreshToken(userID, tokenStr); err != nil {
		return "", err
	}

	return tokenStr, nil
}

func parseAccessToken(tokenStr string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
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

func parseRefreshToken(tokenStr string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtRefreshSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// JwtAuthMiddleware korumalı endpoint'ler için Authorization: Bearer <token> kontrolü yapar.
func JwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			shared.WriteError(w, http.StatusUnauthorized, shared.ErrAuthHeaderRequired, nil)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			shared.WriteError(w, http.StatusUnauthorized, shared.ErrBearerTokenRequired, nil)
			return
		}

		claims, err := parseAccessToken(parts[1])
		if shared.HandleError(w, err, http.StatusUnauthorized, shared.ErrInvalidOrExpiredToken) {
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

// LoginHandler godoc
// @Summary Kullanıcı girişi
// @Description Email ve şifre ile giriş yapar, access ve refresh token döner
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Giriş bilgileri"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {string} string
// @Router /login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
	}

	person, err := repository.GetPersonByEmail(req.Email)
	if shared.HandleError(w, err, http.StatusUnauthorized, shared.ErrUnauthorized) {
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(person.PasswordHash), []byte(req.Password)); err != nil {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrUnauthorized, err)
		return
	}

	accessToken, err := generateAccessToken(person.ID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrAccessTokenGenerateFail, err)
		return
	}
	refreshToken, err := generateRefreshToken(person.ID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrRefreshTokenGenerateFail, err)
		return
	}

	resp := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	json.NewEncoder(w).Encode(resp)
}

// RefreshHandler godoc
// @Summary Access token yenile
// @Description Geçerli bir refresh token ile yeni access token döner
// @Tags auth
// @Accept json
// @Produce json
// @Param token body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {string} string
// @Router /refresh [post]
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
	}

	// Önce refresh token'ın veritabanında geçerli olup olmadığını kontrol et
	valid, err := repository.IsRefreshTokenValid(req.RefreshToken)
	if err != nil || !valid {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrInvalidRefreshToken, err)
		return
	}

	claims, err := parseRefreshToken(req.RefreshToken)
	if shared.HandleError(w, err, http.StatusUnauthorized, shared.ErrInvalidRefreshToken) {
		return
	}

	accessToken, err := generateAccessToken(claims.UserID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrAccessTokenGenerateFail, err)
		return
	}

	// İsteğe bağlı: refresh token'ı da yenileyebilirsin; burada aynısını geri dönüyoruz.
	resp := models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}
	json.NewEncoder(w).Encode(resp)
}

// LogoutHandler godoc
// @Summary Çıkış yap
// @Description Refresh token'ı revoke ederek kullanıcının çıkış yapmasını sağlar
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param token body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {string} string
// @Failure 401 {string} string
// @Router /logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
	}

	// Refresh token geçerli mi kontrol et
	valid, err := repository.IsRefreshTokenValid(req.RefreshToken)
	if err != nil || !valid {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrInvalidRefreshToken, err)
		return
	}

	// Token'ı revoke et
	if err := repository.RevokeRefreshToken(req.RefreshToken); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrInvalidRefreshToken, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Logout başarılı"))
}


