package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/payment"
	"box-office-go/backend/internal/repository"
)

// ── repo stub ────────────────────────────────────────────────────────

type bookingRepoStub struct {
	cleanupExpiredFn          func(ctx context.Context) error
	createHoldFn              func(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error)
	checkoutHoldFn            func(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error)
	listByUserIDFn            func(ctx context.Context, userID string) ([]domain.UserBooking, error)
	cancelBookingFn           func(ctx context.Context, bookingID string, userID string) error
	getBookingForTicketFn     func(ctx context.Context, bookingID string) (domain.TicketData, error)
	createPaymentTxnFn        func(ctx context.Context, txn domain.PaymentTransaction) error
	getPaymentByIdempotencyFn func(ctx context.Context, key string) (*domain.PaymentTransaction, error)
	updatePaymentStatusFn     func(ctx context.Context, txnID string, status string, gatewayTxnID string, failureReason string) error
	getHoldDetailsFn          func(ctx context.Context, holdID string, userID string) (domain.BookingHold, error)
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
	if s.checkoutHoldFn == nil {
		return domain.BookingCheckoutResult{BookingID: bookingID, HoldID: holdID, UserID: userID}, nil
	}
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
func (s *bookingRepoStub) CreatePaymentTransaction(ctx context.Context, txn domain.PaymentTransaction) error {
	if s.createPaymentTxnFn == nil {
		return nil
	}
	return s.createPaymentTxnFn(ctx, txn)
}
func (s *bookingRepoStub) GetPaymentByIdempotencyKey(ctx context.Context, key string) (*domain.PaymentTransaction, error) {
	if s.getPaymentByIdempotencyFn == nil {
		return nil, nil
	}
	return s.getPaymentByIdempotencyFn(ctx, key)
}
func (s *bookingRepoStub) UpdatePaymentStatus(ctx context.Context, txnID string, status string, gatewayTxnID string, failureReason string) error {
	if s.updatePaymentStatusFn == nil {
		return nil
	}
	return s.updatePaymentStatusFn(ctx, txnID, status, gatewayTxnID, failureReason)
}
func (s *bookingRepoStub) GetHoldDetails(ctx context.Context, holdID string, userID string) (domain.BookingHold, error) {
	if s.getHoldDetailsFn == nil {
		return domain.BookingHold{
			HoldID:        holdID,
			UserID:        userID,
			ShowtimeID:    "st_1",
			SeatNumbers:   []string{"A01"},
			Status:        domain.BookingHoldStatusHeld,
			HoldExpiresAt: time.Now().UTC().Add(5 * time.Minute),
			TotalAmount:   12.50,
			CreatedAt:     time.Now().UTC(),
		}, nil
	}
	return s.getHoldDetailsFn(ctx, holdID, userID)
}

// ── gateway stub ─────────────────────────────────────────────────────

type gatewayStub struct {
	chargeFn func(ctx context.Context, req payment.ChargeRequest) (payment.ChargeResponse, error)
}

func (g *gatewayStub) Charge(ctx context.Context, req payment.ChargeRequest) (payment.ChargeResponse, error) {
	if g.chargeFn == nil {
		return payment.ChargeResponse{GatewayTxnID: "gw_ok", Status: "SUCCESS"}, nil
	}
	return g.chargeFn(ctx, req)
}

// helper to build a service with default mock gateway
func newTestService(repo *bookingRepoStub) *BookingService {
	return NewBookingService(repo, &gatewayStub{})
}

func newTestServiceWithGateway(repo *bookingRepoStub, gw payment.Gateway) *BookingService {
	return NewBookingService(repo, gw)
}

// ── existing tests (updated for new constructor) ─────────────────────

func TestNewBookingService(t *testing.T) {
	svc := newTestService(&bookingRepoStub{})
	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}
}

func TestBookingServiceReleaseExpiredHolds_Error(t *testing.T) {
	svc := newTestService(&bookingRepoStub{
		cleanupExpiredFn: func(_ context.Context) error { return errors.New("db") },
	})
	if err := svc.ReleaseExpiredHolds(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBookingServiceCreateBookingHold_InvalidInput(t *testing.T) {
	svc := newTestService(&bookingRepoStub{})
	if _, err := svc.CreateBookingHold(context.Background(), domain.CreateBookingHoldInput{}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBookingServiceCreateBookingHold_CleanupError(t *testing.T) {
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(repo)

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
	svc := newTestService(&bookingRepoStub{})

	// missing holdId
	_, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_1",
	})
	if err == nil || !strings.Contains(err.Error(), "holdId is required") {
		t.Fatalf("expected holdId validation error, got %v", err)
	}

	// missing userId
	_, err = svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", PaymentMethod: "card", IdempotencyKey: "key_1",
	})
	if err == nil || !strings.Contains(err.Error(), "userId is required") {
		t.Fatalf("expected userId validation error, got %v", err)
	}

	// missing paymentMethod
	_, err = svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", IdempotencyKey: "key_1",
	})
	if err == nil || !strings.Contains(err.Error(), "paymentMethod is required") {
		t.Fatalf("expected paymentMethod validation error, got %v", err)
	}

	// missing idempotencyKey
	_, err = svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card",
	})
	if err == nil || !strings.Contains(err.Error(), "idempotencyKey is required") {
		t.Fatalf("expected idempotencyKey validation error, got %v", err)
	}
}

func TestBookingServiceCheckoutBookingHold_Success(t *testing.T) {
	repo := &bookingRepoStub{
		checkoutHoldFn: func(_ context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
			if bookingID == "" {
				t.Fatal("expected generated bookingID")
			}
			return domain.BookingCheckoutResult{BookingID: bookingID, HoldID: holdID, UserID: userID, TotalAmount: 12.50}, nil
		},
	}
	svc := newTestService(repo)

	result, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_1",
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if result.BookingID == "" {
		t.Fatal("expected bookingID")
	}
	if result.TransactionID == "" {
		t.Fatal("expected transactionID in result")
	}
}

func TestBookingServiceGetUserBookings(t *testing.T) {
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(&bookingRepoStub{})
	_, _, err := svc.GetTicketPDF(context.Background(), "  ", "usr_1")
	if err == nil || !strings.Contains(err.Error(), "bookingId is required") {
		t.Fatalf("expected bookingId validation error, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_EmptyUserID(t *testing.T) {
	svc := newTestService(&bookingRepoStub{})
	_, _, err := svc.GetTicketPDF(context.Background(), "bok_1", "  ")
	if err == nil || !strings.Contains(err.Error(), "userId is required") {
		t.Fatalf("expected userId validation error, got %v", err)
	}
}

func TestBookingServiceGetTicketPDF_BookingNotFound(t *testing.T) {
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(&bookingRepoStub{
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
	svc := newTestService(&bookingRepoStub{
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
	if !strings.HasPrefix(string(pdfBytes), "%PDF") {
		t.Fatal("expected PDF magic header")
	}
}

// ── Payment integration tests ────────────────────────────────────────

func TestCheckout_PaymentSuccess_ConfirmsBooking(t *testing.T) {
	var paymentCreated, paymentUpdated bool
	repo := &bookingRepoStub{
		createPaymentTxnFn: func(_ context.Context, txn domain.PaymentTransaction) error {
			paymentCreated = true
			if txn.Status != domain.PaymentStatusPending {
				t.Fatalf("expected PENDING status, got %s", txn.Status)
			}
			if txn.PaymentMethod != "card" {
				t.Fatalf("expected payment method card, got %s", txn.PaymentMethod)
			}
			return nil
		},
		updatePaymentStatusFn: func(_ context.Context, _ string, status string, gatewayTxnID string, _ string) error {
			paymentUpdated = true
			if status != domain.PaymentStatusSuccess {
				t.Fatalf("expected SUCCESS, got %s", status)
			}
			if gatewayTxnID == "" {
				t.Fatal("expected gateway txn ID")
			}
			return nil
		},
		checkoutHoldFn: func(_ context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
			return domain.BookingCheckoutResult{BookingID: bookingID, HoldID: holdID, UserID: userID, TotalAmount: 12.50}, nil
		},
	}
	svc := newTestService(repo) // default gateway returns SUCCESS

	result, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_success",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !paymentCreated {
		t.Fatal("expected payment transaction to be created")
	}
	if !paymentUpdated {
		t.Fatal("expected payment status to be updated to SUCCESS")
	}
	if result.TransactionID == "" {
		t.Fatal("expected transactionID in result")
	}
	if result.BookingID == "" {
		t.Fatal("expected bookingID in result")
	}
}

func TestCheckout_PaymentDeclined_LeavesHoldRetryable(t *testing.T) {
	var statusUpdatedTo string
	repo := &bookingRepoStub{
		updatePaymentStatusFn: func(_ context.Context, _ string, status string, _ string, failureReason string) error {
			statusUpdatedTo = status
			if failureReason == "" {
				t.Fatal("expected failure reason on decline")
			}
			return nil
		},
	}
	gw := &gatewayStub{
		chargeFn: func(_ context.Context, _ payment.ChargeRequest) (payment.ChargeResponse, error) {
			return payment.ChargeResponse{Status: "FAILED", Reason: "card declined by issuing bank"}, payment.ErrPaymentDeclined
		},
	}
	svc := newTestServiceWithGateway(repo, gw)

	_, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card_decline", IdempotencyKey: "key_decline",
	})
	if !errors.Is(err, payment.ErrPaymentDeclined) {
		t.Fatalf("expected ErrPaymentDeclined, got %v", err)
	}
	if statusUpdatedTo != domain.PaymentStatusFailed {
		t.Fatalf("expected payment status FAILED, got %s", statusUpdatedTo)
	}
}

func TestCheckout_GatewayTimeout_LeavesHoldRetryable(t *testing.T) {
	var statusUpdatedTo string
	repo := &bookingRepoStub{
		updatePaymentStatusFn: func(_ context.Context, _ string, status string, _ string, _ string) error {
			statusUpdatedTo = status
			return nil
		},
	}
	gw := &gatewayStub{
		chargeFn: func(_ context.Context, _ payment.ChargeRequest) (payment.ChargeResponse, error) {
			return payment.ChargeResponse{}, payment.ErrGatewayTimeout
		},
	}
	svc := newTestServiceWithGateway(repo, gw)

	_, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card_timeout", IdempotencyKey: "key_timeout",
	})
	if !errors.Is(err, payment.ErrGatewayTimeout) {
		t.Fatalf("expected ErrGatewayTimeout, got %v", err)
	}
	if statusUpdatedTo != domain.PaymentStatusFailed {
		t.Fatalf("expected payment status FAILED, got %s", statusUpdatedTo)
	}
}

func TestCheckout_Idempotency_ReturnsPreviousSuccess(t *testing.T) {
	repo := &bookingRepoStub{
		getPaymentByIdempotencyFn: func(_ context.Context, key string) (*domain.PaymentTransaction, error) {
			return &domain.PaymentTransaction{
				ID:             "txn_prev",
				HoldID:         "hold_1",
				UserID:         "usr_1",
				Amount:         12.50,
				PaymentMethod:  "card",
				Status:         domain.PaymentStatusSuccess,
				IdempotencyKey: key,
			}, nil
		},
	}
	svc := newTestService(repo)

	result, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_dup",
	})
	if err != nil {
		t.Fatalf("expected success for idempotent replay, got %v", err)
	}
	if result.TransactionID != "txn_prev" {
		t.Fatalf("expected previous transactionID txn_prev, got %s", result.TransactionID)
	}
	if result.Status != domain.BookingHoldStatusConfirmed {
		t.Fatalf("expected CONFIRMED status, got %s", result.Status)
	}
}

func TestCheckout_HoldExpired_ReturnsError(t *testing.T) {
	repo := &bookingRepoStub{
		getHoldDetailsFn: func(_ context.Context, _ string, _ string) (domain.BookingHold, error) {
			return domain.BookingHold{
				HoldID:        "hold_1",
				UserID:        "usr_1",
				Status:        domain.BookingHoldStatusHeld,
				HoldExpiresAt: time.Now().UTC().Add(-1 * time.Minute), // expired
				TotalAmount:   12.50,
			}, nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_exp",
	})
	if !errors.Is(err, repository.ErrHoldExpired) {
		t.Fatalf("expected ErrHoldExpired, got %v", err)
	}
}

func TestCheckout_HoldAlreadyConfirmed_ReturnsFinalized(t *testing.T) {
	repo := &bookingRepoStub{
		getHoldDetailsFn: func(_ context.Context, _ string, _ string) (domain.BookingHold, error) {
			return domain.BookingHold{
				HoldID:        "hold_1",
				UserID:        "usr_1",
				Status:        domain.BookingHoldStatusConfirmed,
				HoldExpiresAt: time.Now().UTC().Add(5 * time.Minute),
				TotalAmount:   12.50,
			}, nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.CheckoutBookingHold(context.Background(), domain.ConfirmBookingInput{
		HoldID: "hold_1", UserID: "usr_1", PaymentMethod: "card", IdempotencyKey: "key_fin",
	})
	if !errors.Is(err, repository.ErrHoldFinalized) {
		t.Fatalf("expected ErrHoldFinalized, got %v", err)
	}
}
