package validation

import (
	"net/mail"
	"regexp"
	"strings"

	"box-office-go/backend/internal/domain"
)

var phoneRegex = regexp.MustCompile(`^[0-9+]{8,15}$`)

func ValidateSignupInput(input domain.SignupInput) map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}

	phone := strings.TrimSpace(input.Phone)
	if phone == "" {
		errors["phone"] = "phone is required"
	} else if !phoneRegex.MatchString(phone) {
		errors["phone"] = "phone must be 8-15 chars and only digits/+"
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
