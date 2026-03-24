package repository

import (
	"context"
	"errors"
	"time"

	"box-office-go/backend/internal/domain"
)

var (
	ErrSeatUnavailable      = errors.New("one or more seats are not available")
	ErrInvalidSeatSelection = errors.New("invalid seat selection")
	ErrHoldNotFound         = errors.New("booking hold not found")
	ErrHoldExpired          = errors.New("booking hold has expired")
	ErrHoldFinalized        = errors.New("booking hold already finalized")
)

type BookingRepository interface {
	CleanupExpiredHolds(ctx context.Context) error
	CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
}
