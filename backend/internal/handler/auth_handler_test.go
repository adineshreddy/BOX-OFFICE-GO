package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"box-office-go/backend/internal/repository/memory"
	"box-office-go/backend/internal/service"
)

func TestAuthHandlerSignup_Success(t *testing.T) {
	authSvc := service.NewAuthService(memory.NewUserRepository())
	h := NewAuthHandler(authSvc)

	payload := []byte(`{"name":"Dinesh","phone":"+1234567890","email":"dinesh@example.com","password":"password123","confirmPassword":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	h.Signup(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerLogin_MethodNotAllowed(t *testing.T) {
	authSvc := service.NewAuthService(memory.NewUserRepository())
	h := NewAuthHandler(authSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
