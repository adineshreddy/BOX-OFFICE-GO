package validation

import (
	"net/mail"
	"strings"

	"box-office-go/backend/internal/domain"
)

func ValidateLoginInput(input domain.LoginInput) map[string]string {
	errors := make(map[string]string)

	email := strings.TrimSpace(input.Email)
	if email == "" {
		errors["email"] = "email is required"
	} else if _, err := mail.ParseAddress(email); err != nil {
		errors["email"] = "email format is invalid"
	}

	if strings.TrimSpace(input.Password) == "" {
		errors["password"] = "password is required"
	}

	return errors
}
