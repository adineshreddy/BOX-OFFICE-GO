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

func TestValidateSignupInput_FormattedPhoneAccepted(t *testing.T) {
	input := domain.SignupInput{
		Name:            "Dinesh",
		Phone:           "+1 (555) 123-4567",
		Email:           "dinesh@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	errs := ValidateSignupInput(input)
	if len(errs) != 0 {
		t.Fatalf("expected formatted phone to normalize cleanly, got %v", errs)
	}
}

func TestNormalizePhone(t *testing.T) {
	if got := NormalizePhone("+1 (555) 123-4567"); got != "15551234567" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizePhone("  +91 98765 43210 "); got != "919876543210" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateSignupInput_InvalidFields(t *testing.T) {
	input := domain.SignupInput{
		Name:            "a",
		Phone:           "1234567",
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
