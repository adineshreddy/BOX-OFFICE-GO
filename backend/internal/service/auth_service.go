package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/validation"
)

type AuthService struct {
	userRepository repository.UserRepository
}

func NewAuthService(userRepository repository.UserRepository) *AuthService {
	return &AuthService{userRepository: userRepository}
}

func (s *AuthService) Signup(ctx context.Context, input domain.SignupInput) (domain.User, map[string]string, error) {
	validationErrors := validation.ValidateSignupInput(input)
	if len(validationErrors) > 0 {
		return domain.User{}, validationErrors, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()

	user := domain.User{
		ID:           fmt.Sprintf("usr_%d", time.Now().UnixNano()),
		Name:         strings.TrimSpace(input.Name),
		Phone:        strings.TrimSpace(input.Phone),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(hashedPassword),
		IsAdmin:      false,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdUser, createErr := s.userRepository.Create(ctx, user)
	if createErr != nil {
		if createErr == repository.ErrEmailExists {
			return domain.User{}, map[string]string{"email": "email already exists"}, nil
		}
		return domain.User{}, nil, fmt.Errorf("create user: %w", createErr)
	}

	return createdUser, nil, nil
}
