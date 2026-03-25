package validation

import (
	"testing"

	"box-office-go/backend/internal/domain"
)

func TestValidateSignupInput_Success(t *testing.T) {
	input := domain.SignupInput{
		Name:            "Dinesh",
		Phone:           "+1234567890",
		Email:           "dinesh@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	errs := ValidateSignupInput(input)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateSignupInput_InvalidFields(t *testing.T) {
	input := domain.SignupInput{
		Name:            "  ",
		Phone:           "abc",
		Email:           "not-an-email",
		Password:        "short",
		ConfirmPassword: "different",
	}

	errs := ValidateSignupInput(input)
	if len(errs) != 5 {
		t.Fatalf("expected 5 validation errors, got %d (%v)", len(errs), errs)
	}

	if errs["name"] == "" || errs["phone"] == "" || errs["email"] == "" || errs["password"] == "" || errs["confirmPassword"] == "" {
		t.Fatalf("expected all field errors, got %v", errs)
	}
}
