package domain

import "time"

type User struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	IsAdmin      bool       `json:"isAdmin"`
	IsActive     bool       `json:"isActive"`
	IsVerified   bool       `json:"isVerified"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type SignupInput struct {
	Name            string `json:"name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
