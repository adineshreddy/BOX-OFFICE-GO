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
	"box-office-go/backend/internal/http/middleware"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type bookingRepoHandlerStub struct {
	cleanupExpiredFn      func(ctx context.Context) error
	createHoldFn          func(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	checkoutHoldFn        func(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	listByUserIDFn        func(ctx context.Context, userID string) ([]domain.UserBooking, error)
	cancelBookingFn       func(ctx context.Context, bookingID string, userID string) error
	getBookingForTicketFn func(ctx context.Context, bookingID string) (domain.TicketData, error)
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

func (s *bookingRepoHandlerStub) GetBookingForTicket(ctx context.Context, bookingID string) (domain.TicketData, error) {
	if s.getBookingForTicketFn == nil {
		return domain.TicketData{}, nil
	}
	return s.getBookingForTicketFn(ctx, bookingID)
}

func newBookingHandlerForTest(repo repository.BookingRepository) *BookingHandler {
	return NewBookingHandler(service.NewBookingService(repo))
}

func withAuth(req *http.Request, userID string) *http.Request {
	ctx := middleware.WithAuthIdentity(req.Context(), domain.AuthIdentity{UserID: userID, TokenID: "tok_test"})
	return req.WithContext(ctx)
}

func TestBookingHandlerCreateBookingHold_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		createHoldFn: func(_ context.Context, _ domain.CreateBookingHoldInput, holdID string, _ time.Time) (domain.BookingHold, error) {
			return domain.BookingHold{HoldID: holdID}, nil
		},
	})

	body := map[string]any{"showtimeId": "st_1", "seatNumbers": []string{"A01"}}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/holds", bytes.NewReader(payload))
	req = withAuth(req, "usr_1")
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

	payload := []byte(`{"showtimeId":"st_1","seatNumbers":["A01"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/holds", bytes.NewReader(payload))
	req = withAuth(req, "usr_1")
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

	payload := []byte(`{"holdId":"hold_1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/checkout", bytes.NewReader(payload))
	req = withAuth(req, "usr_1")
	rec := httptest.NewRecorder()

	h.CheckoutBookingHold(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandlerGetUserBookings_RequiresAuth(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	rec := httptest.NewRecorder()

	h.GetUserBookings(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBookingHandlerGetUserBookings_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		listByUserIDFn: func(_ context.Context, userID string) ([]domain.UserBooking, error) {
			return []domain.UserBooking{{UserID: userID}}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings", nil)
	req = withAuth(req, "usr_1")
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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1", nil)
	req = withAuth(req, "usr_1")
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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1", nil)
	req = withAuth(req, "usr_1")
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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bookings?bookingId=bok_1", nil)
	req = withAuth(req, "usr_1")
	rec := httptest.NewRecorder()

	h.CancelBooking(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- DownloadTicket tests ---

// ticketRouteRequest creates a request routed through a mux so PathValue works.
func ticketRouteRequest(t *testing.T, h *BookingHandler, bookingID string, userID string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/bookings/{bookingId}/ticket", h.DownloadTicket)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/"+bookingID+"/ticket", nil)
	if userID != "" {
		req = withAuth(req, userID)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestDownloadTicket_RequiresAuth(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{})
	rec := ticketRouteRequest(t, h, "bok_1", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDownloadTicket_NotFound(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{}, repository.ErrBookingNotFound
		},
	})
	rec := ticketRouteRequest(t, h, "bok_999", "usr_1")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDownloadTicket_NotOwned(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{UserID: "usr_other", Status: "CONFIRMED"}, nil
		},
	})
	rec := ticketRouteRequest(t, h, "bok_1", "usr_1")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDownloadTicket_NonConfirmed(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
		getBookingForTicketFn: func(_ context.Context, _ string) (domain.TicketData, error) {
			return domain.TicketData{UserID: "usr_1", Status: "CANCELLED"}, nil
		},
	})
	rec := ticketRouteRequest(t, h, "bok_1", "usr_1")
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDownloadTicket_Success(t *testing.T) {
	h := newBookingHandlerForTest(&bookingRepoHandlerStub{
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
	rec := ticketRouteRequest(t, h, "bok_1", "usr_1")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/pdf" {
		t.Fatalf("expected Content-Type application/pdf, got %s", ct)
	}
	cd := rec.Header().Get("Content-Disposition")
	if cd == "" {
		t.Fatal("expected Content-Disposition header")
	}
	body := rec.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty response body")
	}
	if body[:4] != "%PDF" {
		t.Fatal("expected PDF magic header in response")
	}
}
