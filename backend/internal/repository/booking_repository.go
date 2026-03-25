package repository

import (
	"context"
	"errors"
	"time"

	"box-office-go/backend/internal/domain"
)

var (
	ErrSeatUnavailable         = errors.New("one or more seats are not available")
	ErrInvalidSeatSelection    = errors.New("invalid seat selection")
	ErrHoldNotFound            = errors.New("booking hold not found")
	ErrHoldExpired             = errors.New("booking hold has expired")
	ErrHoldFinalized           = errors.New("booking hold already finalized")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
)

type BookingRepository interface {
	CleanupExpiredHolds(ctx context.Context) error
	CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.UserBooking, error)
	CancelBooking(ctx context.Context, bookingID string, userID string) error
}
