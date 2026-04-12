package domain

import "time"

type AuthSession struct {
	ID        string
	UserID    string
	TokenID   string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthIdentity struct {
	UserID    string
	TokenID   string
	ExpiresAt time.Time
}

type LoginResult struct {
	User        User
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}
