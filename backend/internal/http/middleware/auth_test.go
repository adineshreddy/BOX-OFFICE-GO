package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
)

type authStub struct {
	authenticateFn func(ctx context.Context, token string) (domain.AuthIdentity, error)
}

func (s *authStub) AuthenticateToken(ctx context.Context, token string) (domain.AuthIdentity, error) {
	return s.authenticateFn(ctx, token)
}

func TestAuthRequire_MissingHeader(t *testing.T) {
	mw := NewAuth(&authStub{
		authenticateFn: func(_ context.Context, _ string) (domain.AuthIdentity, error) {
			t.Fatal("authenticate should not be called")
			return domain.AuthIdentity{}, nil
		},
	})

	handler := mw.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRequire_InvalidScheme(t *testing.T) {
	mw := NewAuth(&authStub{
		authenticateFn: func(_ context.Context, _ string) (domain.AuthIdentity, error) {
			t.Fatal("authenticate should not be called")
			return domain.AuthIdentity{}, nil
		},
	})

	handler := mw.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRequire_UnauthorizedToken(t *testing.T) {
	mw := NewAuth(&authStub{
		authenticateFn: func(_ context.Context, _ string) (domain.AuthIdentity, error) {
			return domain.AuthIdentity{}, context.DeadlineExceeded
		},
	})

	handler := mw.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRequire_SetsContextIdentity(t *testing.T) {
	expectedIdentity := domain.AuthIdentity{
		UserID:    "usr_1",
		TokenID:   "tok_1",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	mw := NewAuth(&authStub{
		authenticateFn: func(_ context.Context, token string) (domain.AuthIdentity, error) {
			if token != "token123" {
				t.Fatalf("unexpected token: %s", token)
			}
			return expectedIdentity, nil
		},
	})

	handler := mw.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := AuthIdentityFromContext(r.Context())
		if !ok {
			t.Fatal("expected auth identity in context")
		}
		if identity.UserID != expectedIdentity.UserID || identity.TokenID != expectedIdentity.TokenID {
			t.Fatalf("unexpected identity: %+v", identity)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
