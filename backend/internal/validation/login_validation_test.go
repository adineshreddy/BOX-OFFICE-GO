package validation

import (
	"testing"

	"box-office-go/backend/internal/domain"
)

func TestValidateLoginInput_Success(t *testing.T) {
	input := domain.LoginInput{
		Email:    "dinesh@example.com",
		Password: "password123",
	}

	errs := ValidateLoginInput(input)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateLoginInput_InvalidFields(t *testing.T) {
	input := domain.LoginInput{
		Email:    "bad-email",
		Password: " ",
	}

	errs := ValidateLoginInput(input)
	if len(errs) != 2 {
		t.Fatalf("expected 2 validation errors, got %d (%v)", len(errs), errs)
	}

	if errs["email"] == "" || errs["password"] == "" {
		t.Fatalf("expected email/password errors, got %v", errs)
	}
}
