package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"box-office-go/backend/internal/repository/memory"
	"box-office-go/backend/internal/service"
)

func TestAuthHandlerSignup_MethodNotAllowed(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/signup", nil)
	rec := httptest.NewRecorder()

	h.Signup(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestAuthHandlerSignup_InvalidJSONSyntax(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	h.Signup(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerSignup_ValidationFailure(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBufferString(`{"name":"","phone":"x","email":"bad","password":"1","confirmPassword":"2"}`))
	rec := httptest.NewRecorder()

	h.Signup(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerLogin_InvalidPayload(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":123}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerLogin_ValidationFailure(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"bad","password":""}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerLogin_InvalidCredentials(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"dinesh@example.com","password":"wrongpass"}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
