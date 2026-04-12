package repository

import (
	"context"
	"errors"
	"time"

	"box-office-go/backend/internal/domain"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already registered")
	ErrSessionNotFound = errors.New("auth session not found")
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateLastLogin(ctx context.Context, userID string, loggedAt time.Time) error
	CreateSession(ctx context.Context, session domain.AuthSession) error
	GetSessionByTokenID(ctx context.Context, tokenID string) (domain.AuthSession, error)
	RevokeSessionByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error
}
