package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-kisi-api/models"
	"go-kisi-api/service"
	"go-kisi-api/shared"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// AuthHandler API kimlik doğrulama endpoint'lerini yönetir.
type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
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
		claims, err := shared.ParseAccessToken(parts[1])
		if shared.HandleError(w, err, http.StatusUnauthorized, shared.ErrInvalidOrExpiredToken) {
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

// LoginHandler godoc
// @Summary Kullanıcı girişi
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Giriş bilgileri"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {string} string
// @Router /login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleCustomError(w, shared.NewValidationError("Geçersiz istek formatı", nil))
		return
	}

	validator := shared.NewValidator()
	validator.ValidateRequired(req.Email, "Email")
	validator.ValidateRequired(req.Password, "Şifre")
	if validator.HasError() {
		shared.HandleCustomError(w, validator.GetErrorAsCustomError())
		return
	}

	ctx := r.Context()
	person, err := h.authSvc.Login(ctx, req.Email, req.Password)
	if err != nil {
		shared.LogAuth("LOGIN_FAILED", req.Email, "Invalid credentials")
		shared.HandleCustomError(w, shared.NewAuthError("Kullanıcı veya şifre hatalı"))
		return
	}

	accessToken, err := h.authSvc.GenerateAccessToken(person.ID)
	if err != nil {
		shared.LogError("TOKEN_GENERATE", "Access token generation failed", map[string]interface{}{"user_id": person.ID, "error": err})
		shared.HandleCustomError(w, shared.NewInternalError("Access token üretilemedi"))
		return
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken(ctx, person.ID)
	if err != nil {
		shared.LogError("TOKEN_GENERATE", "Refresh token generation failed", map[string]interface{}{"user_id": person.ID, "error": err})
		shared.HandleCustomError(w, shared.NewInternalError("Refresh token üretilemedi"))
		return
	}

	shared.LogAuth("LOGIN_SUCCESS", req.Email, "User logged in successfully")
	shared.WriteSuccess(w, "Giriş başarılı", models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshHandler godoc
// @Summary Access token yenile
// @Tags auth
// @Accept json
// @Produce json
// @Param token body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse
// @Failure 401 {string} string
// @Router /refresh [post]
func (h *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
	}

	valid, err := h.authSvc.IsRefreshTokenValid(r.Context(), req.RefreshToken)
	if err != nil || !valid {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrInvalidRefreshToken, err)
		return
	}

	userID, err := h.authSvc.ParseRefreshToken(req.RefreshToken)
	if shared.HandleError(w, err, http.StatusUnauthorized, shared.ErrInvalidRefreshToken) {
		return
	}

	accessToken, err := h.authSvc.GenerateAccessToken(userID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, shared.ErrAccessTokenGenerateFail, err)
		return
	}

	shared.WriteSuccess(w, "Token yenilendi", models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	})
}

// LogoutHandler godoc
// @Summary Çıkış yap
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param token body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {string} string
// @Failure 401 {string} string
// @Router /logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.HandleError(w, err, http.StatusBadRequest, shared.ErrInvalidRequestBody)
		return
	}

	valid, err := h.authSvc.IsRefreshTokenValid(r.Context(), req.RefreshToken)
	if err != nil || !valid {
		shared.WriteError(w, http.StatusUnauthorized, shared.ErrInvalidRefreshToken, err)
		return
	}

	if err := h.authSvc.RevokeRefreshToken(r.Context(), req.RefreshToken); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "Çıkış yapılırken bir hata oluştu", err)
		return
	}

	shared.WriteSuccess(w, "Çıkış başarılı", nil)
}
