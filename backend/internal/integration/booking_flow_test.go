// Package integration contains end-to-end tests that wire the full HTTP router
// (auth middleware, services, handlers) using only in-memory repositories so
// they are deterministic, require no external dependencies, and run in CI.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	httpRouter "box-office-go/backend/internal/http/router"
	"box-office-go/backend/internal/payment"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/repository/memory"
	"box-office-go/backend/internal/service"
)

// ---------------------------------------------------------------------------
// Minimal in-memory booking repository (no Postgres required)
// ---------------------------------------------------------------------------

type integrationBookingRepo struct {
	mu                   sync.Mutex
	holds                map[string]domain.BookingHold
	bookings             map[string]domain.BookingCheckoutResult
	confirmed            map[string]bool            // holdIDs already finalized
	heldSeats            map[string]map[string]bool // showtimeID → seat → held
	paymentByID          map[string]domain.PaymentTransaction
	paymentByIdempotency map[string]string // idempotencyKey -> txnID
}

func newIntegrationBookingRepo() *integrationBookingRepo {
	return &integrationBookingRepo{
		holds:                make(map[string]domain.BookingHold),
		bookings:             make(map[string]domain.BookingCheckoutResult),
		confirmed:            make(map[string]bool),
		heldSeats:            make(map[string]map[string]bool),
		paymentByID:          make(map[string]domain.PaymentTransaction),
		paymentByIdempotency: make(map[string]string),
	}
}

func (r *integrationBookingRepo) CleanupExpiredHolds(_ context.Context) error { return nil }

func (r *integrationBookingRepo) CreateHold(_ context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	seats, ok := r.heldSeats[input.ShowtimeID]
	if !ok {
		seats = make(map[string]bool)
		r.heldSeats[input.ShowtimeID] = seats
	}
	for _, s := range input.SeatNumbers {
		if seats[s] {
			return domain.BookingHold{}, repository.ErrSeatUnavailable
		}
	}
	for _, s := range input.SeatNumbers {
		seats[s] = true
	}

	hold := domain.BookingHold{
		HoldID:        holdID,
		UserID:        input.UserID,
		ShowtimeID:    input.ShowtimeID,
		SeatNumbers:   input.SeatNumbers,
		Status:        domain.BookingHoldStatusHeld,
		HoldExpiresAt: holdExpiresAt,
		TotalAmount:   float64(len(input.SeatNumbers)) * 12.50,
		CreatedAt:     time.Now().UTC(),
	}
	r.holds[holdID] = hold
	return hold, nil
}

func (r *integrationBookingRepo) CheckoutHold(_ context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.confirmed[holdID] {
		return domain.BookingCheckoutResult{}, repository.ErrHoldFinalized
	}
	hold, ok := r.holds[holdID]
	if !ok {
		return domain.BookingCheckoutResult{}, repository.ErrHoldNotFound
	}
	if hold.UserID != userID {
		return domain.BookingCheckoutResult{}, repository.ErrHoldNotFound
	}

	r.confirmed[holdID] = true
	now := time.Now().UTC()
	result := domain.BookingCheckoutResult{
		BookingID:   bookingID,
		HoldID:      holdID,
		UserID:      userID,
		ShowtimeID:  hold.ShowtimeID,
		SeatNumbers: hold.SeatNumbers,
		Status:      domain.BookingHoldStatusConfirmed,
		TotalAmount: hold.TotalAmount,
		ConfirmedAt: now,
	}
	r.bookings[bookingID] = result
	hold.Status = domain.BookingHoldStatusConfirmed
	r.holds[holdID] = hold
	return result, nil
}

func (r *integrationBookingRepo) ListByUserID(_ context.Context, userID string) ([]domain.UserBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var out []domain.UserBooking
	for _, b := range r.bookings {
		if b.UserID == userID {
			out = append(out, domain.UserBooking{
				BookingID:   b.BookingID,
				HoldID:      b.HoldID,
				UserID:      b.UserID,
				ShowtimeID:  b.ShowtimeID,
				SeatNumbers: b.SeatNumbers,
				Status:      b.Status,
			})
		}
	}
	return out, nil
}

func (r *integrationBookingRepo) CancelBooking(_ context.Context, bookingID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.bookings[bookingID]
	if !ok || b.UserID != userID {
		return repository.ErrBookingNotFound
	}
	if b.Status == "CANCELLED" {
		return repository.ErrBookingAlreadyCancelled
	}
	b.Status = "CANCELLED"
	r.bookings[bookingID] = b
	return nil
}

func (r *integrationBookingRepo) GetBookingForTicket(_ context.Context, bookingID string) (domain.TicketData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.bookings[bookingID]
	if !ok {
		return domain.TicketData{}, repository.ErrBookingNotFound
	}

	return domain.TicketData{
		BookingID:   b.BookingID,
		UserID:      b.UserID,
		MovieTitle:  "Integration Test Movie",
		TheaterName: "Integration Theater",
		City:        "New York",
		ScreenName:  "Screen 1",
		ShowTime:    time.Now().UTC().Add(2 * time.Hour),
		Language:    "EN",
		Format:      "2D",
		SeatNumbers: b.SeatNumbers,
		TotalAmount: b.TotalAmount,
		Status:      b.Status,
		ConfirmedAt: b.ConfirmedAt,
	}, nil
}

func (r *integrationBookingRepo) CreatePaymentTransaction(_ context.Context, txn domain.PaymentTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.paymentByIdempotency[txn.IdempotencyKey]; exists {
		return repository.ErrDuplicatePayment
	}
	r.paymentByID[txn.ID] = txn
	r.paymentByIdempotency[txn.IdempotencyKey] = txn.ID
	return nil
}

func (r *integrationBookingRepo) GetPaymentByIdempotencyKey(_ context.Context, idempotencyKey string) (*domain.PaymentTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txnID, ok := r.paymentByIdempotency[idempotencyKey]
	if !ok {
		return nil, nil
	}
	txn := r.paymentByID[txnID]
	copied := txn
	return &copied, nil
}

func (r *integrationBookingRepo) UpdatePaymentStatus(_ context.Context, txnID string, status string, gatewayTxnID string, failureReason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn, ok := r.paymentByID[txnID]
	if !ok {
		return repository.ErrPaymentNotFound
	}
	txn.Status = status
	txn.GatewayTxnID = gatewayTxnID
	txn.FailureReason = failureReason
	txn.UpdatedAt = time.Now().UTC()
	r.paymentByID[txnID] = txn
	return nil
}

func (r *integrationBookingRepo) GetHoldDetails(_ context.Context, holdID string, userID string) (domain.BookingHold, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hold, ok := r.holds[holdID]
	if !ok || hold.UserID != userID {
		return domain.BookingHold{}, repository.ErrHoldNotFound
	}
	return hold, nil
}

func (r *integrationBookingRepo) ReleaseHold(_ context.Context, holdID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	hold, ok := r.holds[holdID]
	if !ok || hold.UserID != userID {
		return repository.ErrHoldNotFound
	}
	if hold.Status == domain.BookingHoldStatusConfirmed {
		return repository.ErrHoldFinalized
	}
	if hold.Status == domain.BookingHoldStatusExpired {
		return repository.ErrHoldAlreadyReleased
	}

	held := r.heldSeats[hold.ShowtimeID]
	for _, seat := range hold.SeatNumbers {
		delete(held, seat)
	}
	hold.Status = domain.BookingHoldStatusExpired
	r.holds[holdID] = hold
	return nil
}

// ---------------------------------------------------------------------------
// Minimal no-op movie repository (booking tests don't hit movie endpoints)
// ---------------------------------------------------------------------------

type noopMovieRepo struct{}

func (noopMovieRepo) ListActive(_ context.Context, _ domain.MovieListQuery) (domain.MovieListResponse, error) {
	return domain.MovieListResponse{}, nil
}
func (noopMovieRepo) GetByID(_ context.Context, _ string) (domain.Movie, error) {
	return domain.Movie{}, repository.ErrMovieNotFound
}
func (noopMovieRepo) ListMovieShowtimeRecords(_ context.Context, _ string, _ *time.Time) ([]domain.MovieShowtimeRecord, error) {
	return nil, nil
}
func (noopMovieRepo) GetShowDetailsBySelection(_ context.Context, _, _ string, _ time.Time) (domain.ShowDetails, error) {
	return domain.ShowDetails{}, nil
}
func (noopMovieRepo) GetSeatMapBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatMapResponse, error) {
	return domain.SeatMapResponse{}, nil
}
func (noopMovieRepo) GetSeatAvailabilityBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatAvailabilityResponse, error) {
	return domain.SeatAvailabilityResponse{}, nil
}

// ---------------------------------------------------------------------------
// Test server factory
// ---------------------------------------------------------------------------

type testEnv struct {
	server      *httptest.Server
	bookingRepo *integrationBookingRepo
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	userRepo := memory.NewUserRepository()
	bookingRepo := newIntegrationBookingRepo()
	authSvc := service.NewAuthService(userRepo)
	movieSvc := service.NewMovieService(noopMovieRepo{})
	bookingSvc := service.NewBookingService(bookingRepo, payment.NewMockGateway())

	srv := httptest.NewServer(httpRouter.New(authSvc, movieSvc, bookingSvc))
	t.Cleanup(srv.Close)

	return &testEnv{server: srv, bookingRepo: bookingRepo}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func doJSON(t *testing.T, method, url string, body any, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func mustDecode(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func mustStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected HTTP %d, got %d — body: %s", want, resp.StatusCode, body)
	}
}

// signupAndLogin creates a fresh user and returns its Bearer token.
func signupAndLogin(t *testing.T, base string, suffix string) string {
	t.Helper()
	email := fmt.Sprintf("user_%s@example.com", suffix)
	creds := map[string]string{
		"name":            "Test User",
		"phone":           "+1234567890",
		"email":           email,
		"password":        "Password123!",
		"confirmPassword": "Password123!",
	}

	resp := doJSON(t, http.MethodPost, base+"/api/v1/auth/signup", creds, "")
	mustStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	resp = doJSON(t, http.MethodPost, base+"/api/v1/auth/login",
		map[string]string{"email": email, "password": "Password123!"}, "")
	mustStatus(t, resp, http.StatusOK)

	var loginBody struct {
		AccessToken string `json:"accessToken"`
	}
	mustDecode(t, resp, &loginBody)
	if loginBody.AccessToken == "" {
		t.Fatal("login returned empty accessToken")
	}
	return loginBody.AccessToken
}

func checkoutPayload(holdID string, idempotencyKey string) map[string]string {
	return map[string]string{
		"holdId":         holdID,
		"paymentMethod":  "card",
		"cardNumber":     "4111111111111111",
		"cardExpiry":     "12/29",
		"cardCvv":        "123",
		"idempotencyKey": idempotencyKey,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestBookingFlow_HappyPath exercises the complete booking lifecycle:
// signup → login → create hold → checkout → list bookings.
func TestBookingFlow_HappyPath(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "happy")

	// Create hold
	holdPayload := map[string]any{
		"showtimeId":  "st_happy_1",
		"seatNumbers": []string{"A01", "A02"},
	}
	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds", holdPayload, token)
	mustStatus(t, resp, http.StatusCreated)

	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)
	if holdBody.Hold.HoldID == "" {
		t.Fatal("expected non-empty holdId in response")
	}
	holdID := holdBody.Hold.HoldID

	// Checkout the hold
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout", checkoutPayload(holdID, "happy_checkout_1"), token)
	mustStatus(t, resp, http.StatusOK)

	var checkoutBody struct {
		Booking struct {
			BookingID string `json:"bookingId"`
			HoldID    string `json:"holdId"`
			Status    string `json:"status"`
		} `json:"booking"`
	}
	mustDecode(t, resp, &checkoutBody)
	if checkoutBody.Booking.BookingID == "" {
		t.Fatal("expected non-empty bookingId in checkout response")
	}
	if checkoutBody.Booking.HoldID != holdID {
		t.Fatalf("checkout holdId mismatch: want %s got %s", holdID, checkoutBody.Booking.HoldID)
	}

	// List bookings — should contain the confirmed booking
	resp = doJSON(t, http.MethodGet, base+"/api/v1/bookings", nil, token)
	mustStatus(t, resp, http.StatusOK)

	var listBody struct {
		Bookings []struct {
			BookingID string `json:"bookingId"`
		} `json:"bookings"`
	}
	mustDecode(t, resp, &listBody)
	if len(listBody.Bookings) == 0 {
		t.Fatal("expected at least one booking in list response")
	}
	found := false
	for _, b := range listBody.Bookings {
		if b.BookingID == checkoutBody.Booking.BookingID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("booking %s not found in user booking list", checkoutBody.Booking.BookingID)
	}
}

// TestBookingFlow_NoAuthToken verifies that protected endpoints return 401
// when no Authorization header is provided.
func TestBookingFlow_NoAuthToken(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	endpoints := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/api/v1/bookings/holds", map[string]any{"showtimeId": "st_1", "seatNumbers": []string{"A01"}}},
		{http.MethodPost, "/api/v1/bookings/checkout", checkoutPayload("hold_1", "no_auth_checkout")},
		{http.MethodGet, "/api/v1/bookings", nil},
	}

	for _, ep := range endpoints {
		resp := doJSON(t, ep.method, base+ep.path, ep.body, "" /* no token */)
		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Errorf("%s %s without token: expected 401, got %d — body: %s",
				ep.method, ep.path, resp.StatusCode, body)
		} else {
			resp.Body.Close()
		}
	}
}

// TestBookingFlow_InvalidToken verifies that a bogus Bearer token returns 401.
func TestBookingFlow_InvalidToken(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_1", "seatNumbers": []string{"A01"}},
		"this-is-not-a-valid-token")
	mustStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

// TestBookingFlow_MalformedAuthHeader verifies that a non-Bearer auth scheme
// returns 401.
func TestBookingFlow_MalformedAuthHeader(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/bookings/holds",
		bytes.NewReader([]byte(`{"showtimeId":"st_1","seatNumbers":["A01"]}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // basic auth, not bearer

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed auth header, got %d", resp.StatusCode)
	}
}

// TestBookingFlow_AfterLogout_TokenRevoked verifies that a token cannot be
// used after the user logs out (session revocation).
func TestBookingFlow_AfterLogout_TokenRevoked(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "logout_test")

	// Logout — revokes the session
	resp := doJSON(t, http.MethodPost, base+"/api/v1/auth/logout", nil, token)
	mustStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Try to use the same token after logout — must be rejected
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_1", "seatNumbers": []string{"A01"}},
		token)
	mustStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

// TestBookingFlow_ReplayCheckout verifies that checking out the same hold a
// second time returns 409 (hold already finalized) — idempotency guard.
func TestBookingFlow_ReplayCheckout(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "replay")

	// Create hold
	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_replay_1", "seatNumbers": []string{"B01"}},
		token)
	mustStatus(t, resp, http.StatusCreated)

	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)
	holdID := holdBody.Hold.HoldID

	// First checkout — must succeed
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdID, "replay_checkout_1"), token)
	mustStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Second checkout (replay) — must be rejected with 409
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdID, "replay_checkout_2"), token)
	if resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected 409 on replay checkout, got %d — body: %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}

// TestBookingFlow_ReplayCheckout_SameIdempotencyKeyReturnsSameBooking verifies
// deterministic replay semantics for network retries.
func TestBookingFlow_ReplayCheckout_SameIdempotencyKeyReturnsSameBooking(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "replay_same_key")

	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_replay_same_1", "seatNumbers": []string{"B02"}},
		token)
	mustStatus(t, resp, http.StatusCreated)

	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)
	holdID := holdBody.Hold.HoldID

	key := "replay_same_key_1"
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdID, key), token)
	mustStatus(t, resp, http.StatusOK)

	var first struct {
		Booking struct {
			BookingID string `json:"bookingId"`
		} `json:"booking"`
	}
	mustDecode(t, resp, &first)
	if first.Booking.BookingID == "" {
		t.Fatal("expected bookingId on first checkout")
	}

	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdID, key), token)
	mustStatus(t, resp, http.StatusOK)

	var second struct {
		Booking struct {
			BookingID string `json:"bookingId"`
		} `json:"booking"`
	}
	mustDecode(t, resp, &second)
	if second.Booking.BookingID != first.Booking.BookingID {
		t.Fatalf("expected same bookingId on replay, want %s got %s", first.Booking.BookingID, second.Booking.BookingID)
	}
}

// TestBookingFlow_SeatConflict verifies that two concurrent hold requests for
// the same seat result in the second one being rejected with 409.
func TestBookingFlow_SeatConflict(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	tokenA := signupAndLogin(t, base, "conflict_a")
	tokenB := signupAndLogin(t, base, "conflict_b")

	holdPayload := map[string]any{
		"showtimeId":  "st_conflict_1",
		"seatNumbers": []string{"C01"},
	}

	// User A grabs the seat
	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds", holdPayload, tokenA)
	mustStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// User B tries the same seat — must be rejected
	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds", holdPayload, tokenB)
	if resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected 409 seat conflict for user B, got %d — body: %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}

// TestBookingFlow_GetBookings_IsolatedByUser verifies that each user only sees
// their own bookings, not another user's.
func TestBookingFlow_GetBookings_IsolatedByUser(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	tokenA := signupAndLogin(t, base, "isolation_a")
	tokenB := signupAndLogin(t, base, "isolation_b")

	// User A creates and checks out a booking
	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_iso_1", "seatNumbers": []string{"D01"}},
		tokenA)
	mustStatus(t, resp, http.StatusCreated)
	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)

	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdBody.Hold.HoldID, "isolation_checkout_1"), tokenA)
	mustStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// User B lists their bookings — must be empty (cannot see user A's booking)
	resp = doJSON(t, http.MethodGet, base+"/api/v1/bookings", nil, tokenB)
	mustStatus(t, resp, http.StatusOK)

	var listBody struct {
		Bookings []any `json:"bookings"`
	}
	mustDecode(t, resp, &listBody)
	if len(listBody.Bookings) != 0 {
		t.Fatalf("expected 0 bookings for user B, got %d", len(listBody.Bookings))
	}
}

// TestBookingFlow_CreateHold_MissingFields verifies that hold creation with
// missing required fields returns a clear 400 error (no auth required to learn
// this — but we still need auth to reach the handler).
func TestBookingFlow_CreateHold_MissingFields(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "missing_fields")

	cases := []struct {
		label   string
		payload any
	}{
		{"missing showtimeId", map[string]any{"seatNumbers": []string{"A01"}}},
		{"empty seatNumbers", map[string]any{"showtimeId": "st_1", "seatNumbers": []string{}}},
		{"duplicate seat", map[string]any{"showtimeId": "st_1", "seatNumbers": []string{"A01", "a01"}}},
	}

	for _, tc := range cases {
		resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds", tc.payload, token)
		if resp.StatusCode != http.StatusBadRequest {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Errorf("[%s] expected 400, got %d — body: %s", tc.label, resp.StatusCode, body)
		} else {
			resp.Body.Close()
		}
	}
}

// TestBookingFlow_Checkout_MissingHoldID verifies that checkout without a
// holdId returns 400.
func TestBookingFlow_Checkout_MissingHoldID(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "checkout_missing")

	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload("", "missing_hold_id"), token)
	mustStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

// TestBookingFlow_Signup_DuplicateEmail verifies that signing up with an
// already-registered email returns 400 with a validation error.
func TestBookingFlow_Signup_DuplicateEmail(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	creds := map[string]string{
		"name":            "First User",
		"phone":           "+1234567890",
		"email":           "dup@example.com",
		"password":        "Password123!",
		"confirmPassword": "Password123!",
	}

	// First signup — must succeed
	resp := doJSON(t, http.MethodPost, base+"/api/v1/auth/signup", creds, "")
	mustStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// Second signup with same email — must fail
	resp = doJSON(t, http.MethodPost, base+"/api/v1/auth/signup", creds, "")
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Fatalf("expected 400 or 409 for duplicate email signup, got %d — body: %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}

// TestBookingFlow_Login_WrongPassword verifies that login with a wrong
// password returns a non-200 response with an actionable error.
func TestBookingFlow_Login_WrongPassword(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	creds := map[string]string{
		"name":            "Pw User",
		"phone":           "+1234567890",
		"email":           "pw@example.com",
		"password":        "Password123!",
		"confirmPassword": "Password123!",
	}
	resp := doJSON(t, http.MethodPost, base+"/api/v1/auth/signup", creds, "")
	mustStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	resp = doJSON(t, http.MethodPost, base+"/api/v1/auth/login",
		map[string]string{"email": "pw@example.com", "password": "wrong_password"}, "")
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Fatal("expected non-200 for wrong password login, got 200")
	}
	resp.Body.Close()
}

// TestBookingFlow_CancelBooking verifies that a confirmed booking can be
// cancelled by its owner.
func TestBookingFlow_CancelBooking(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	token := signupAndLogin(t, base, "cancel")

	// Create + checkout a booking
	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_cancel_1", "seatNumbers": []string{"E01"}},
		token)
	mustStatus(t, resp, http.StatusCreated)
	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)

	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdBody.Hold.HoldID, "cancel_checkout_1"), token)
	mustStatus(t, resp, http.StatusOK)
	var checkoutBody struct {
		Booking struct {
			BookingID string `json:"bookingId"`
		} `json:"booking"`
	}
	mustDecode(t, resp, &checkoutBody)
	bookingID := checkoutBody.Booking.BookingID

	// Cancel the booking
	cancelURL := fmt.Sprintf("%s/api/v1/bookings?bookingId=%s", base, bookingID)
	req, _ := http.NewRequest(http.MethodDelete, cancelURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	cancelResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("cancel request: %v", err)
	}
	defer cancelResp.Body.Close()
	if cancelResp.StatusCode != http.StatusOK && cancelResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(cancelResp.Body)
		t.Fatalf("expected 200/204 on cancel, got %d — body: %s", cancelResp.StatusCode, body)
	}
}

// TestBookingFlow_TicketDownload verifies that the booking owner can download
// ticket PDF and other users cannot access that booking ticket.
func TestBookingFlow_TicketDownload(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	tokenA := signupAndLogin(t, base, "ticket_a")
	tokenB := signupAndLogin(t, base, "ticket_b")

	resp := doJSON(t, http.MethodPost, base+"/api/v1/bookings/holds",
		map[string]any{"showtimeId": "st_ticket_1", "seatNumbers": []string{"F01"}},
		tokenA)
	mustStatus(t, resp, http.StatusCreated)

	var holdBody struct {
		Hold struct {
			HoldID string `json:"holdId"`
		} `json:"hold"`
	}
	mustDecode(t, resp, &holdBody)

	resp = doJSON(t, http.MethodPost, base+"/api/v1/bookings/checkout",
		checkoutPayload(holdBody.Hold.HoldID, "ticket_checkout_1"), tokenA)
	mustStatus(t, resp, http.StatusOK)

	var checkoutBody struct {
		Booking struct {
			BookingID string `json:"bookingId"`
		} `json:"booking"`
	}
	mustDecode(t, resp, &checkoutBody)
	bookingID := checkoutBody.Booking.BookingID

	ticketURL := fmt.Sprintf("%s/api/v1/bookings/%s/ticket", base, bookingID)
	req, _ := http.NewRequest(http.MethodGet, ticketURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)

	ticketResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("ticket request: %v", err)
	}
	defer ticketResp.Body.Close()

	if ticketResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(ticketResp.Body)
		t.Fatalf("expected 200 ticket download for owner, got %d — body: %s", ticketResp.StatusCode, body)
	}
	if got := ticketResp.Header.Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("expected application/pdf content type, got %s", got)
	}

	pdfBytes, _ := io.ReadAll(ticketResp.Body)
	if len(pdfBytes) == 0 || !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected non-empty PDF bytes with %PDF header")
	}

	reqForbidden, _ := http.NewRequest(http.MethodGet, ticketURL, nil)
	reqForbidden.Header.Set("Authorization", "Bearer "+tokenB)
	forbiddenResp, err := http.DefaultClient.Do(reqForbidden)
	if err != nil {
		t.Fatalf("ticket forbidden request: %v", err)
	}
	defer forbiddenResp.Body.Close()
	if forbiddenResp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(forbiddenResp.Body)
		t.Fatalf("expected 403 for non-owner ticket request, got %d — body: %s", forbiddenResp.StatusCode, body)
	}
}

// TestBookingFlow_EmptyAuthHeader verifies that an empty Authorization header
// value (only whitespace) is treated as missing and returns 401.
func TestBookingFlow_EmptyAuthHeader(t *testing.T) {
	env := newTestEnv(t)
	base := env.server.URL

	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/bookings/holds",
		bytes.NewReader([]byte(`{"showtimeId":"st_1","seatNumbers":["A01"]}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", strings.Repeat(" ", 10)) // only whitespace

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for empty auth header, got %d", resp.StatusCode)
	}
}
