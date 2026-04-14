package validation

import (
	"net/mail"
	"strings"
	"unicode"

	"box-office-go/backend/internal/domain"
)

// NormalizePhone strips formatting; keeps digits only (E.164-style input with +, spaces, dashes, etc.).
func NormalizePhone(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func ValidateSignupInput(input domain.SignupInput) map[string]string {
	errors := make(map[string]string)

	name := strings.TrimSpace(input.Name)
	if name == "" {
		errors["name"] = "name is required"
	} else if len(name) < 2 {
		errors["name"] = "name must be at least 2 characters"
	}

	phone := NormalizePhone(input.Phone)
	if phone == "" {
		errors["phone"] = "phone is required"
	} else if len(phone) < 8 || len(phone) > 15 {
		errors["phone"] = "phone must be 8-15 digits (country code included)"
	}

	email := strings.TrimSpace(input.Email)
	if email == "" {
		errors["email"] = "email is required"
	} else if _, err := mail.ParseAddress(email); err != nil {
		errors["email"] = "email format is invalid"
	}

	if strings.TrimSpace(input.Password) == "" {
		errors["password"] = "password is required"
	} else if len(input.Password) < 8 {
		errors["password"] = "password must be at least 8 characters"
	}

	if input.ConfirmPassword == "" {
		errors["confirmPassword"] = "confirmPassword is required"
	} else if input.Password != input.ConfirmPassword {
		errors["confirmPassword"] = "password and confirmPassword do not match"
	}

	return errors
}
