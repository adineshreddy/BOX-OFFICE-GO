package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type authUserRepoStub struct {
	createFn                 func(ctx context.Context, user domain.User) (domain.User, error)
	getByEmailFn             func(ctx context.Context, email string) (domain.User, error)
	updateLastLoginFn        func(ctx context.Context, userID string, loggedAt time.Time) error
	createSessionFn          func(ctx context.Context, session domain.AuthSession) error
	getSessionByTokenIDFn    func(ctx context.Context, tokenID string) (domain.AuthSession, error)
	revokeSessionByTokenIDFn func(ctx context.Context, tokenID string, revokedAt time.Time) error
}

func (s *authUserRepoStub) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if s.createFn == nil {
		return user, nil
	}
	return s.createFn(ctx, user)
}

func (s *authUserRepoStub) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if s.getByEmailFn == nil {
		return domain.User{}, repository.ErrUserNotFound
	}
	return s.getByEmailFn(ctx, email)
}

func (s *authUserRepoStub) UpdateLastLogin(ctx context.Context, userID string, loggedAt time.Time) error {
	if s.updateLastLoginFn == nil {
		return nil
	}
	return s.updateLastLoginFn(ctx, userID, loggedAt)
}

func (s *authUserRepoStub) CreateSession(ctx context.Context, session domain.AuthSession) error {
	if s.createSessionFn == nil {
		return nil
	}
	return s.createSessionFn(ctx, session)
}

func (s *authUserRepoStub) GetSessionByTokenID(ctx context.Context, tokenID string) (domain.AuthSession, error) {
	if s.getSessionByTokenIDFn == nil {
		return domain.AuthSession{}, repository.ErrSessionNotFound
	}
	return s.getSessionByTokenIDFn(ctx, tokenID)
}

func (s *authUserRepoStub) RevokeSessionByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error {
	if s.revokeSessionByTokenIDFn == nil {
		return nil
	}
	return s.revokeSessionByTokenIDFn(ctx, tokenID, revokedAt)
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{})
	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}
}

func TestAuthServiceSignup_ValidationFailure(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{})

	_, validationErrs, err := svc.Signup(context.Background(), domain.SignupInput{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(validationErrs) == 0 {
		t.Fatal("expected validation errors, got none")
	}
}

func TestAuthServiceSignup_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{
		createFn: func(_ context.Context, _ domain.User) (domain.User, error) {
			return domain.User{}, repository.ErrEmailExists
		},
	})

	_, validationErrs, err := svc.Signup(context.Background(), domain.SignupInput{
		Name:            "Dinesh",
		Phone:           "+1234567890",
		Email:           "dinesh@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if validationErrs["email"] != "email already exists" {
		t.Fatalf("expected duplicate email validation, got %v", validationErrs)
	}
}

func TestAuthServiceSignup_Success(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{
		createFn: func(_ context.Context, user domain.User) (domain.User, error) {
			if user.PasswordHash == "" {
				t.Fatal("expected password hash to be set")
			}
			return user, nil
		},
	})

	user, validationErrs, err := svc.Signup(context.Background(), domain.SignupInput{
		Name:            "Dinesh",
		Phone:           "+1234567890",
		Email:           "dinesh@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(validationErrs) != 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrs)
	}
	if user.ID == "" || user.Email != "dinesh@example.com" {
		t.Fatalf("unexpected user created: %+v", user)
	}
}

func TestAuthServiceLogin_UserNotFound(t *testing.T) {
	svc := NewAuthService(&authUserRepoStub{
		getByEmailFn: func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{}, repository.ErrUserNotFound
		},
	})

	_, validationErrs, err := svc.Login(context.Background(), domain.LoginInput{Email: "dinesh@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if validationErrs["credentials"] == "" {
		t.Fatalf("expected credentials validation error, got %v", validationErrs)
	}
}

func TestAuthServiceLogin_InactiveAccount(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	svc := NewAuthService(&authUserRepoStub{
		getByEmailFn: func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{ID: "usr_1", PasswordHash: string(hash), IsActive: false}, nil
		},
	})

	_, validationErrs, err := svc.Login(context.Background(), domain.LoginInput{Email: "dinesh@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if validationErrs["account"] == "" {
		t.Fatalf("expected inactive account error, got %v", validationErrs)
	}
}

func TestAuthServiceLogin_UpdateFailure(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	svc := NewAuthService(&authUserRepoStub{
		getByEmailFn: func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{ID: "usr_1", PasswordHash: string(hash), IsActive: true}, nil
		},
		updateLastLoginFn: func(_ context.Context, _ string, _ time.Time) error {
			return errors.New("db down")
		},
	})

	_, _, err := svc.Login(context.Background(), domain.LoginInput{Email: "dinesh@example.com", Password: "password123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthServiceLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	updated := false
	svc := NewAuthService(&authUserRepoStub{
		getByEmailFn: func(_ context.Context, _ string) (domain.User, error) {
			return domain.User{ID: "usr_1", PasswordHash: string(hash), IsActive: true}, nil
		},
		updateLastLoginFn: func(_ context.Context, _ string, _ time.Time) error {
			updated = true
			return nil
		},
	})

	loginResult, validationErrs, err := svc.Login(context.Background(), domain.LoginInput{Email: "dinesh@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(validationErrs) != 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrs)
	}
	if !updated || loginResult.User.LastLoginAt == nil {
		t.Fatalf("expected last login update, got user=%+v updated=%v", loginResult.User, updated)
	}
	if loginResult.AccessToken == "" {
		t.Fatal("expected access token in login result")
	}
	if loginResult.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %s", loginResult.TokenType)
	}
}

func TestAuthServiceAuthenticateToken_Success(t *testing.T) {
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)

	var createdSession domain.AuthSession
	svc := NewAuthServiceWithConfig(&authUserRepoStub{
		getByEmailFn: func(_ context.Context, _ string) (domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return domain.User{ID: "usr_1", PasswordHash: string(hash), IsActive: true}, nil
		},
		createSessionFn: func(_ context.Context, session domain.AuthSession) error {
			createdSession = session
			return nil
		},
		getSessionByTokenIDFn: func(_ context.Context, tokenID string) (domain.AuthSession, error) {
			if tokenID != createdSession.TokenID {
				t.Fatalf("unexpected token id: %s", tokenID)
			}
			return createdSession, nil
		},
	}, "test-secret", time.Hour)
	svc.nowFn = func() time.Time { return now }

	loginResult, validationErrs, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "dinesh@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(validationErrs) != 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrs)
	}

	identity, authErr := svc.AuthenticateToken(context.Background(), loginResult.AccessToken)
	if authErr != nil {
		t.Fatalf("expected nil auth error, got %v", authErr)
	}
	if identity.UserID != "usr_1" || identity.TokenID == "" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
}

func TestAuthServiceLogout_RevokesSession(t *testing.T) {
	revoked := false
	svc := NewAuthServiceWithConfig(&authUserRepoStub{
		revokeSessionByTokenIDFn: func(_ context.Context, tokenID string, _ time.Time) error {
			if tokenID != "tok_123" {
				t.Fatalf("unexpected token id: %s", tokenID)
			}
			revoked = true
			return nil
		},
	}, "secret", time.Hour)

	err := svc.Logout(context.Background(), "tok_123")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !revoked {
		t.Fatal("expected session to be revoked")
	}
}
