package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
)

type bookingRepoStub struct {
	cleanupExpiredFn func(ctx context.Context) error
	createHoldFn     func(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	checkoutHoldFn   func(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	listByUserIDFn   func(ctx context.Context, userID string) ([]domain.UserBooking, error)
	cancelBookingFn  func(ctx context.Context, bookingID string, userID string) error
}

func (s *bookingRepoStub) CleanupExpiredHolds(ctx context.Context) error {
	if s.cleanupExpiredFn == nil {
		return nil
	}
	return s.cleanupExpiredFn(ctx)
}
func (s *bookingRepoStub) CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error) {
	return s.createHoldFn(ctx, input, holdID, holdExpiresAt)
}
func (s *bookingRepoStub) CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
	return s.checkoutHoldFn(ctx, holdID, userID, bookingID)
}
func (s *bookingRepoStub) ListByUserID(ctx context.Context, userID string) ([]domain.UserBooking, error) {
	return s.listByUserIDFn(ctx, userID)
}
func (s *bookingRepoStub) CancelBooking(ctx context.Context, bookingID string, userID string) error {
	return s.cancelBookingFn(ctx, bookingID, userID)
}

func TestNewBookingService(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{})
	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}
}

func TestBookingServiceReleaseExpiredHolds_Error(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		cleanupExpiredFn: func(_ context.Context) error { return errors.New("db") },
	})
	if err := svc.ReleaseExpiredHolds(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBookingServiceCreateBookingHold_InvalidInput(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{})
	if _, err := svc.CreateBookingHold(context.Background(), domain.CreateBookingHoldInput{}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBookingServiceCreateBookingHold_CleanupError(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		cleanupExpiredFn: func(_ context.Context) error { return errors.New("cleanup failed") },
	})

	_, err := svc.CreateBookingHold(context.Background(), domain.CreateBookingHoldInput{
		UserID:      "usr_1",
		ShowtimeID:  "st_1",
		SeatNumbers: []string{"a01"},
	})
	if err == nil {
		t.Fatal("expected cleanup error")
	}
}

func TestBookingServiceCreateBookingHold_Success(t *testing.T) {
	repo := &bookingRepoStub{
		createHoldFn: func(_ context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error) {
			if holdID == "" || holdExpiresAt.IsZero() {
				t.Fatal("expected generated hold metadata")
			}
			if input.SeatNumbers[0] != "A01" {
				t.Fatalf("expected normalized seat, got %v", input.SeatNumbers)
			}
			return domain.BookingHold{HoldID: holdID}, nil
		},
	}
	svc := NewBookingService(repo)

	hold, err := svc.CreateBookingHold(context.Background(), domain.CreateBookingHoldInput{
		UserID:      "usr_1",
		ShowtimeID:  "st_1",
		SeatNumbers: []string{" a01 "},
	})
	if err != nil || hold.HoldID == "" {
		t.Fatalf("unexpected result hold=%+v err=%v", hold, err)
	}
}

func TestBookingServiceCheckoutBookingHold_Validation(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{})
	if _, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{UserID: "usr_1"}); err == nil {
		t.Fatal("expected holdId validation error")
	}
	if _, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{HoldID: "hold_1"}); err == nil {
		t.Fatal("expected userId validation error")
	}
}

func TestBookingServiceCheckoutBookingHold_Success(t *testing.T) {
	repo := &bookingRepoStub{
		checkoutHoldFn: func(_ context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
			if bookingID == "" {
				t.Fatal("expected generated bookingID")
			}
			return domain.BookingCheckoutResult{BookingID: bookingID, HoldID: holdID, UserID: userID}, nil
		},
	}
	svc := NewBookingService(repo)

	result, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{HoldID: "hold_1", UserID: "usr_1"})
	if err != nil || result.BookingID == "" {
		t.Fatalf("unexpected result=%+v err=%v", result, err)
	}
}

func TestBookingServiceGetUserBookings(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		listByUserIDFn: func(_ context.Context, userID string) ([]domain.UserBooking, error) {
			return []domain.UserBooking{{UserID: userID}}, nil
		},
	})

	bookings, err := svc.GetUserBookings(context.Background(), " usr_1 ")
	if err != nil || len(bookings) != 1 || bookings[0].UserID != "usr_1" {
		t.Fatalf("unexpected bookings=%+v err=%v", bookings, err)
	}
}

func TestBookingServiceCancelBooking(t *testing.T) {
	called := false
	svc := NewBookingService(&bookingRepoStub{
		cancelBookingFn: func(_ context.Context, bookingID string, userID string) error {
			called = true
			if bookingID != "bok_1" || userID != "usr_1" {
				t.Fatalf("unexpected params bookingID=%s userID=%s", bookingID, userID)
			}
			return nil
		},
	})

	if err := svc.CancelBooking(context.Background(), " bok_1 ", " usr_1 "); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected repo cancel call")
	}
}
