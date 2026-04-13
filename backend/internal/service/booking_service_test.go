package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type bookingRepoStub struct {
	cleanupExpiredFn      func(ctx context.Context) error
	createHoldFn          func(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	checkoutHoldFn        func(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	listByUserIDFn        func(ctx context.Context, userID string) ([]domain.UserBooking, error)
	cancelBookingFn       func(ctx context.Context, bookingID string, userID string) error
	getBookingForTicketFn func(ctx context.Context, bookingID string) (domain.TicketData, error)
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
func (s *bookingRepoStub) GetBookingForTicket(ctx context.Context, bookingID string) (domain.TicketData, error) {
	if s.getBookingForTicketFn == nil {
		return domain.TicketData{}, nil
	}
	return s.getBookingForTicketFn(ctx, bookingID)
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

func TestBookingServiceGetTicketPDF_EmptyBookingID(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{})
	_, _, err := svc.GetTicketPDF(context.Background(), "  ", "usr_1")
	if err == nil || !strings.Contains(err.Error(), "bookingId is required") {
		t.Fatalf("expected bookingId validation error, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_EmptyUserID(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{})
	_, _, err := svc.GetTicketPDF(context.Background(), "bok_1", "  ")
	if err == nil || !strings.Contains(err.Error(), "userId is required") {
		t.Fatalf("expected userId validation error, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_BookingNotFound(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{}, repository.ErrBookingNotFound
		},
	})
	_, _, err := svc.GetTicketPDF(context.Background(), "bok_1", "usr_1")
	if !errors.Is(err, repository.ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_NotOwned(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{UserID: "usr_other", Status: "CONFIRMED"}, nil
		},
	})
	_, _, err := svc.GetTicketPDF(context.Background(), "bok_1", "usr_1")
	if !errors.Is(err, repository.ErrBookingNotOwned) {
		t.Fatalf("expected ErrBookingNotOwned, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_NotConfirmed(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{UserID: "usr_1", Status: "CANCELLED"}, nil
		},
	})
	_, _, err := svc.GetTicketPDF(context.Background(), "bok_1", "usr_1")
	if err == nil || !strings.Contains(err.Error(), "only available for confirmed bookings") {
		t.Fatalf("expected non-confirmed error, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_Success(t *testing.T) {
	svc := NewBookingService(&bookingRepoStub{
		getBookingForTicketFn: func(_ context.Context, bookingID string) (domain.TicketData, error) {
			return domain.TicketData{
				BookingID:   bookingID,
				UserID:      "usr_1",
				MovieTitle:  "Test Movie",
				TheaterName: "Grand Theater",
				City:        "Gainesville",
				ScreenName:  "Screen 1",
				ShowTime:    time.Date(2026, 4, 15, 19, 0, 0, 0, time.UTC),
				Language:    "English",
				Format:      "2D",
				SeatNumbers: []string{"A01", "A02"},
				TotalAmount: 25.50,
				Status:      "CONFIRMED",
				ConfirmedAt: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC),
			}, nil
		},
	})

	pdfBytes, filename, err := svc.GetTicketPDF(context.Background(), "bok_1", "usr_1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if filename != "ticket_bok_1.pdf" {
		t.Fatalf("expected filename ticket_bok_1.pdf, got %s", filename)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	// PDF files start with %PDF
	if !strings.HasPrefix(string(pdfBytes), "%PDF") {
		t.Fatal("expected PDF magic header")
	}
}
