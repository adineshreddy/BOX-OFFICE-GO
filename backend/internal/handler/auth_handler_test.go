package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/middleware"
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

func TestAuthHandlerLogin_ReturnsAccessToken(t *testing.T) {
	authSvc := service.NewAuthService(memory.NewUserRepository())
	h := NewAuthHandler(authSvc)

	signupPayload := []byte(`{"name":"Dinesh","phone":"+1234567890","email":"dinesh@example.com","password":"password123","confirmPassword":"password123"}`)
	signupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupPayload))
	signupRec := httptest.NewRecorder()
	h.Signup(signupRec, signupReq)
	if signupRec.Code != http.StatusCreated {
		t.Fatalf("expected signup 201, got %d", signupRec.Code)
	}

	loginPayload := []byte(`{"email":"dinesh@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var responseBody struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if responseBody.AccessToken == "" {
		t.Fatal("expected access token in login response")
	}
	if responseBody.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %s", responseBody.TokenType)
	}
}

func TestAuthHandlerLogout_Success(t *testing.T) {
	authSvc := service.NewAuthService(memory.NewUserRepository())
	h := NewAuthHandler(authSvc)

	signupPayload := []byte(`{"name":"Dinesh","phone":"+1234567890","email":"dinesh@example.com","password":"password123","confirmPassword":"password123"}`)
	signupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupPayload))
	signupRec := httptest.NewRecorder()
	h.Signup(signupRec, signupReq)
	if signupRec.Code != http.StatusCreated {
		t.Fatalf("expected signup 201, got %d", signupRec.Code)
	}

	loginPayload := []byte(`{"email":"dinesh@example.com","password":"password123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
	loginRec := httptest.NewRecorder()
	h.Login(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginRec.Code, loginRec.Body.String())
	}

	var loginResponse struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	identity, err := authSvc.AuthenticateToken(context.Background(), loginResponse.AccessToken)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq = logoutReq.WithContext(middleware.WithAuthIdentity(logoutReq.Context(), domain.AuthIdentity{
		UserID:  identity.UserID,
		TokenID: identity.TokenID,
	}))
	logoutRec := httptest.NewRecorder()

	h.Logout(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", logoutRec.Code, logoutRec.Body.String())
	}
}
