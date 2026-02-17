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

func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (domain.User, map[string]string, error) {
	validationErrors := validation.ValidateLoginInput(input)
	if len(validationErrors) > 0 {
		return domain.User{}, validationErrors, nil
	}

	user, err := s.userRepository.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return domain.User{}, map[string]string{"credentials": "invalid email or password"}, nil
		}
		return domain.User{}, nil, fmt.Errorf("get user by email: %w", err)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); compareErr != nil {
		return domain.User{}, map[string]string{"credentials": "invalid email or password"}, nil
	}

	if !user.IsActive {
		return domain.User{}, map[string]string{"account": "account is inactive"}, nil
	}

	now := time.Now().UTC()
	if updateErr := s.userRepository.UpdateLastLogin(ctx, user.ID, now); updateErr != nil {
		return domain.User{}, nil, fmt.Errorf("update last login: %w", updateErr)
	}

	user.LastLoginAt = &now
	user.UpdatedAt = now

	return user, nil, nil
}
