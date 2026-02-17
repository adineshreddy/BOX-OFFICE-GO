package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/response"
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

	loggedInUser, validationErrors, err := h.authService.Login(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to login", nil)
		return
	}

	if len(validationErrors) > 0 {
		response.Error(w, http.StatusBadRequest, "validation failed", validationErrors)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"user": map[string]any{
			"id":          loggedInUser.ID,
			"name":        loggedInUser.Name,
			"phone":       loggedInUser.Phone,
			"email":       loggedInUser.Email,
			"isAdmin":     loggedInUser.IsAdmin,
			"isActive":    loggedInUser.IsActive,
			"isVerified":  loggedInUser.IsVerified,
			"lastLoginAt": loggedInUser.LastLoginAt,
			"createdAt":   loggedInUser.CreatedAt,
			"updatedAt":   loggedInUser.UpdatedAt,
		},
	})
}
