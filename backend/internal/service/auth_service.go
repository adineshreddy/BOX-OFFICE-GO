package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/validation"
)

var (
	ErrInvalidAuthToken   = errors.New("invalid auth token")
	ErrAuthSessionExpired = errors.New("auth session expired")
	ErrAuthSessionRevoked = errors.New("auth session revoked")
)

const (
	defaultJWTSecret = "dev-secret-change-me"
	defaultTokenTTL  = 24 * time.Hour
	defaultTokenType = "Bearer"
)

type AuthService struct {
	userRepository repository.UserRepository
	jwtSecret      []byte
	tokenTTL       time.Duration
	nowFn          func() time.Time
}

func NewAuthService(userRepository repository.UserRepository) *AuthService {
	return NewAuthServiceWithConfig(userRepository, defaultJWTSecret, defaultTokenTTL)
}

func NewAuthServiceWithConfig(userRepository repository.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	secret := strings.TrimSpace(jwtSecret)
	if secret == "" {
		secret = defaultJWTSecret
	}

	if tokenTTL <= 0 {
		tokenTTL = defaultTokenTTL
	}

	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      []byte(secret),
		tokenTTL:       tokenTTL,
		nowFn:          time.Now,
	}
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

func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (domain.LoginResult, map[string]string, error) {
	validationErrors := validation.ValidateLoginInput(input)
	if len(validationErrors) > 0 {
		return domain.LoginResult{}, validationErrors, nil
	}

	user, err := s.userRepository.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return domain.LoginResult{}, map[string]string{"credentials": "invalid email or password"}, nil
		}
		return domain.LoginResult{}, nil, fmt.Errorf("get user by email: %w", err)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); compareErr != nil {
		return domain.LoginResult{}, map[string]string{"credentials": "invalid email or password"}, nil
	}

	if !user.IsActive {
		return domain.LoginResult{}, map[string]string{"account": "account is inactive"}, nil
	}

	now := s.nowFn().UTC()
	if updateErr := s.userRepository.UpdateLastLogin(ctx, user.ID, now); updateErr != nil {
		return domain.LoginResult{}, nil, fmt.Errorf("update last login: %w", updateErr)
	}

	tokenID := fmt.Sprintf("tok_%d", now.UnixNano())
	expiresAt := now.Add(s.tokenTTL)
	session := domain.AuthSession{
		ID:        fmt.Sprintf("ses_%d", now.UnixNano()+1),
		UserID:    user.ID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if sessionErr := s.userRepository.CreateSession(ctx, session); sessionErr != nil {
		return domain.LoginResult{}, nil, fmt.Errorf("create auth session: %w", sessionErr)
	}

	accessToken, tokenErr := s.signToken(authTokenClaims{
		Sub: user.ID,
		JTI: tokenID,
		Exp: expiresAt.Unix(),
		Iat: now.Unix(),
	})
	if tokenErr != nil {
		return domain.LoginResult{}, nil, fmt.Errorf("sign auth token: %w", tokenErr)
	}

	user.LastLoginAt = &now
	user.UpdatedAt = now

	return domain.LoginResult{
		User:        user,
		AccessToken: accessToken,
		TokenType:   defaultTokenType,
		ExpiresAt:   expiresAt,
	}, nil, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenID string) error {
	trimmed := strings.TrimSpace(tokenID)
	if trimmed == "" {
		return ErrInvalidAuthToken
	}

	if err := s.userRepository.RevokeSessionByTokenID(ctx, trimmed, s.nowFn().UTC()); err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}

	return nil
}

func (s *AuthService) AuthenticateToken(ctx context.Context, token string) (domain.AuthIdentity, error) {
	claims, err := s.parseAndVerifyToken(token)
	if err != nil {
		return domain.AuthIdentity{}, ErrInvalidAuthToken
	}

	now := s.nowFn().UTC()
	if claims.Exp <= now.Unix() {
		return domain.AuthIdentity{}, ErrAuthSessionExpired
	}

	session, sessionErr := s.userRepository.GetSessionByTokenID(ctx, claims.JTI)
	if sessionErr != nil {
		return domain.AuthIdentity{}, ErrInvalidAuthToken
	}
	if session.UserID != claims.Sub {
		return domain.AuthIdentity{}, ErrInvalidAuthToken
	}
	if session.RevokedAt != nil {
		return domain.AuthIdentity{}, ErrAuthSessionRevoked
	}
	if !session.ExpiresAt.After(now) {
		return domain.AuthIdentity{}, ErrAuthSessionExpired
	}

	return domain.AuthIdentity{
		UserID:    session.UserID,
		TokenID:   session.TokenID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

type authTokenClaims struct {
	Sub string `json:"sub"`
	JTI string `json:"jti"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func (s *AuthService) signToken(claims authTokenClaims) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsignedToken := headerPart + "." + payloadPart

	mac := hmac.New(sha256.New, s.jwtSecret)
	_, _ = mac.Write([]byte(unsignedToken))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return unsignedToken + "." + signaturePart, nil
}

func (s *AuthService) parseAndVerifyToken(token string) (authTokenClaims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return authTokenClaims{}, ErrInvalidAuthToken
	}

	unsignedToken := parts[0] + "." + parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return authTokenClaims{}, ErrInvalidAuthToken
	}

	mac := hmac.New(sha256.New, s.jwtSecret)
	_, _ = mac.Write([]byte(unsignedToken))
	expectedSignature := mac.Sum(nil)
	if !hmac.Equal(signatureBytes, expectedSignature) {
		return authTokenClaims{}, ErrInvalidAuthToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return authTokenClaims{}, ErrInvalidAuthToken
	}

	var claims authTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return authTokenClaims{}, ErrInvalidAuthToken
	}
	if strings.TrimSpace(claims.Sub) == "" || strings.TrimSpace(claims.JTI) == "" || claims.Exp <= 0 {
		return authTokenClaims{}, ErrInvalidAuthToken
	}

	return claims, nil
}
