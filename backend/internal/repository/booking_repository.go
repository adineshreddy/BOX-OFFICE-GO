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
	ErrHoldAlreadyReleased     = errors.New("booking hold is already released")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
	ErrCancellationWindowClosed = errors.New("booking can only be cancelled at least 1 hour before showtime")
	ErrBookingNotOwned         = errors.New("booking does not belong to the requesting user")
	ErrDuplicatePayment        = errors.New("duplicate payment for this idempotency key")
	ErrPaymentNotFound         = errors.New("payment transaction not found")
	ErrShowtimeStarted         = errors.New("showtime has already started")
)

type BookingRepository interface {
	CleanupExpiredHolds(ctx context.Context) error
	CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.UserBooking, error)
	CancelBooking(ctx context.Context, bookingID string, userID string) error
	GetBookingForTicket(ctx context.Context, bookingID string) (domain.TicketData, error)

	// Payment transaction methods
	CreatePaymentTransaction(ctx context.Context, txn domain.PaymentTransaction) error
	GetPaymentByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.PaymentTransaction, error)
	UpdatePaymentStatus(ctx context.Context, txnID string, status string, gatewayTxnID string, failureReason string) error
	GetHoldDetails(ctx context.Context, holdID string, userID string) (domain.BookingHold, error)
	ReleaseHold(ctx context.Context, holdID string, userID string) error
}
