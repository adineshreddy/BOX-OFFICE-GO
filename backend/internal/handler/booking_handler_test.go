package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type bookingRepoHandlerStub struct {
	cleanupExpiredFn func(ctx context.Context) error
	createHoldFn     func(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	checkoutHoldFn   func(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	listByUserIDFn   func(ctx context.Context, userID string) ([]domain.UserBooking, error)
	cancelBookingFn  func(ctx context.Context, bookingID string, userID string) error
}

func (s *bookingRepoHandlerStub) CleanupExpiredHolds(ctx context.Context) error {
	if s.cleanupExpiredFn == nil {
		return nil
	}
	return s.cleanupExpiredFn(ctx)
}

func (s *bookingRepoHandlerStub) CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error) {
	return s.createHoldFn(ctx, input, holdID, holdExpiresAt)
}

func (s *bookingRepoHandlerStub) CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
	return s.checkoutHoldFn(ctx, holdID, userID, bookingID)
}

func (s *bookingRepoHandlerStub) ListByUserID(ctx context.Context, userID string) ([]domain.UserBooking, error) {
	return s.listByUserIDFn(ctx, userID)
}

func (s *bookingRepoHandlerStub) CancelBooking(ctx context.Context, bookingID string, userID string) error {
	return s.cancelBookingFn(ctx, bookingID, userID)
}

func newBookingHandlerForTest(repo repository.BookingRepository) *BookingHandler {
	return NewBookingHandler(service.NewBookingService(repo))
}

func TestBookingHandlerCreateBookingHold_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		createHoldFn: func(_ context.Context, _ domain.CreateBookingHoldInput, holdID string, _ time.Time) (domain.BookingHold, error) {
			return domain.BookingHold{HoldID: holdID}, nil
		},
	})

	body := map[string]any{"userId": "usr_1", "showtimeId": "st_1", "seatNumbers": []string{"A01"}}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/holds", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	h.CreateBookingHold(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandlerCreateBookingHold_SeatUnavailable(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		createHoldFn: func(_ context.Context, _ domain.CreateBookingHoldInput, _ string, _ time.Time) (domain.BookingHold, error) {
			return domain.BookingHold{}, repository.ErrSeatUnavailable
		},
	})

	payload := []byte(`{"userId":"usr_1","showtimeId":"st_1","seatNumbers":["A01"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/holds", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	h.CreateBookingHold(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestBookingHandlerCheckoutBookingHold_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		checkoutHoldFn: func(_ context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
			return domain.BookingCheckoutResult{HoldID: holdID, UserID: userID, BookingID: bookingID}, nil
		},
	})

	payload := []byte(`{"holdId":"hold_1","userId":"usr_1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/checkout", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	h.CheckoutBookingHold(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandlerGetUserBookings_RequiresUserID(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	rec := httptest.NewRecorder()

	h.GetUserBookings(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBookingHandlerGetUserBookings_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		listByUserIDFn: func(_ context.Context, userID string) ([]domain.UserBooking, error) {
			return []domain.UserBooking{{UserID: userID}}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings?userId=usr_1", nil)
	rec := httptest.NewRecorder()

	h.GetUserBookings(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBookingHandlerCancelBooking_NotFound(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		cancelBookingFn: func(_ context.Context, _ string, _ string) error {
			return repository.ErrBookingNotFound
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1&userId=usr_1", nil)
	rec := httptest.NewRecorder()

	h.CancelBooking(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestBookingHandlerCancelBooking_InternalError(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		cancelBookingFn: func(_ context.Context, _ string, _ string) error {
			return errors.New("db error")
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1&userId=usr_1", nil)
	rec := httptest.NewRecorder()

	h.CancelBooking(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestBookingHandlerCancelBooking_QueryBookingID_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		cancelBookingFn: func(_ context.Context, bookingID string, userID string) error {
			if bookingID != "bok_1" || userID != "usr_1" {
				t.Fatalf("unexpected params bookingID=%s userID=%s", bookingID, userID)
			}
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1&userId=usr_1", nil)
	rec := httptest.NewRecorder()

	h.CancelBooking(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
