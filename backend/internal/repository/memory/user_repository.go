package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type UserRepository struct {
	mu      sync.RWMutex
	users   map[string]domain.User
	usersBy map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:   make(map[string]domain.User),
		usersBy: make(map[string]string),
	}
}

func (r *UserRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if _, exists := r.usersBy[normalizedEmail]; exists {
		return domain.User{}, repository.ErrEmailExists
	}

	r.users[user.ID] = user
	r.usersBy[normalizedEmail] = user.ID

	return user, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	userID, exists := r.usersBy[normalizedEmail]
	if !exists {
		return domain.User{}, repository.ErrUserNotFound
	}

	user, found := r.users[userID]
	if !found {
		return domain.User{}, repository.ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) UpdateLastLogin(_ context.Context, userID string, loggedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, found := r.users[userID]
	if !found {
		return repository.ErrUserNotFound
	}

	user.LastLoginAt = &loggedAt
	user.UpdatedAt = loggedAt
	r.users[userID] = user

	return nil
}
