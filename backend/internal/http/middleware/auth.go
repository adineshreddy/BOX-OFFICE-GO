package middleware

import (
	"context"
	"net/http"
	"strings"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/response"
)

type authContextKey string

const authIdentityContextKey authContextKey = "auth_identity"

type Authenticator interface {
	AuthenticateToken(ctx context.Context, token string) (domain.AuthIdentity, error)
}

type Auth struct {
	authenticator Authenticator
}

func NewAuth(authenticator Authenticator) *Auth {
	return &Auth{authenticator: authenticator}
}

func (a *Auth) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if header == "" {
			response.Error(w, http.StatusUnauthorized, "authorization token is required", nil)
			return
		}

		token, ok := extractBearerToken(header)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header", nil)
			return
		}

		identity, err := a.authenticator.AuthenticateToken(r.Context(), token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid or expired auth token", nil)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithAuthIdentity(r.Context(), identity)))
	})
}

func WithAuthIdentity(ctx context.Context, identity domain.AuthIdentity) context.Context {
	return context.WithValue(ctx, authIdentityContextKey, identity)
}

func AuthIdentityFromContext(ctx context.Context) (domain.AuthIdentity, bool) {
	value := ctx.Value(authIdentityContextKey)
	if value == nil {
		return domain.AuthIdentity{}, false
	}

	identity, ok := value.(domain.AuthIdentity)
	return identity, ok
}

func extractBearerToken(header string) (string, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
