package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/middleware"
	"box-office-go/backend/internal/http/response"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var input domain.SignupInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			response.Error(w, http.StatusBadRequest, "invalid JSON format", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "invalid request payload", nil)
		return
	}

	createdUser, validationErrors, err := h.authService.Signup(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create account", nil)
		return
	}

	if len(validationErrors) > 0 {
		response.Error(w, http.StatusBadRequest, "validation failed", validationErrors)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"message": "account created successfully",
		"user": map[string]any{
			"id":        createdUser.ID,
			"name":      createdUser.Name,
			"phone":     createdUser.Phone,
			"email":     createdUser.Email,
			"createdAt": createdUser.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var input domain.LoginInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			response.Error(w, http.StatusBadRequest, "invalid JSON format", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, "invalid request payload", nil)
		return
	}

	loginResult, validationErrors, err := h.authService.Login(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to login", nil)
		return
	}

	if len(validationErrors) > 0 {
		response.Error(w, http.StatusBadRequest, "validation failed", validationErrors)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message":     "login successful",
		"accessToken": loginResult.AccessToken,
		"tokenType":   loginResult.TokenType,
		"expiresAt":   loginResult.ExpiresAt,
		"user": map[string]any{
			"id":          loginResult.User.ID,
			"name":        loginResult.User.Name,
			"phone":       loginResult.User.Phone,
			"email":       loginResult.User.Email,
			"isAdmin":     loginResult.User.IsAdmin,
			"isActive":    loginResult.User.IsActive,
			"isVerified":  loginResult.User.IsVerified,
			"lastLoginAt": loginResult.User.LastLoginAt,
			"createdAt":   loginResult.User.CreatedAt,
			"updatedAt":   loginResult.User.UpdatedAt,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.TokenID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	if err := h.authService.Logout(r.Context(), identity.TokenID); err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) || errors.Is(err, service.ErrInvalidAuthToken) {
			response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to logout", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}
