# Sprint 4 Report

GitHub Link: https://github.com/adineshreddy/BOX-OFFICE-GO

## Sprint Overview
Sprint 4 focused on completing remaining Sprint 3 carryover work, tightening payment and booking UX, and improving production-readiness with stronger validation and tests.

This sprint includes major frontend enhancements plus backend policy alignment for cancellations.

---

## Sprint 4 Frontend Scope (Completed)

### 1) Payment page aligned to updated backend checkout contract
- Updated checkout request payload to include required fields:
  - `paymentMethod`
  - `cardNumber`
  - `cardExpiry`
  - `cardCvv`
  - `idempotencyKey`
- Added payment form fields in UI so users can provide required checkout details.

### 2) Method-specific payment UX (major feature)
- Payment form now changes based on selected method:
  - **Card**: card number, expiry, CVV
  - **UPI**: UPI ID
  - **Net banking**: bank name, account number, routing number
- Added method-aware validation and method-aware error messages before API submit.

### 3) Strong frontend field validation (major feature)
- Card validations:
  - Card number must be 13-19 digits
  - Expiry must be `MM/YY` and not in the past
  - CVV must be 3-4 digits
- UPI validation:
  - UPI ID format validated (`name@bank`)
- Net banking validations:
  - Bank name minimum length
  - Account number length and numeric format
  - Routing number must be 9 digits
- Added input-level constraints (`maxlength`, numeric input mode hints).

### 4) My Bookings upgrades (major feature set)
- Added **search** across movie, theater, city, seat number, and booking ID.
- Added **status filter** (`All`, `Confirmed`, `Cancelled`).
- Added **sorting controls**:
  - Showtime latest first
  - Showtime earliest first
  - Recently booked
- Added **grouped sections**:
  - Upcoming bookings
  - Past bookings
- Added richer empty states and filter reset behavior.

### 5) Cancellation rule UX alignment
- Updated frontend cancellation availability logic to enforce:
  - cancellation only when showtime is at least 1 hour away
- Updated My Bookings messaging to clearly explain the cancellation window.

---

## Sprint 4 Backend Scope (Completed)

### 1) Cancellation window policy enforcement
- Added backend rule: booking can be cancelled only if showtime is at least 1 hour in the future.
- Added explicit repository error for closed cancellation window.
- Added clear API response message for cancellation-window conflicts.

### 2) Seat release behavior on cancel (confirmed)
- Cancellation flow keeps seat rollback behavior:
  - `is_available = TRUE`
  - `is_held = FALSE`
- Seats are released back to inventory during successful cancellation.

---

## Key API Contract Notes (Sprint 4)

### Checkout request (frontend-aligned)
`POST /api/v1/bookings/checkout`

Request:
```json
{
  "holdId": "hold_123",
  "paymentMethod": "card",
  "cardNumber": "4111111111111111",
  "cardExpiry": "12/29",
  "cardCvv": "123",
  "idempotencyKey": "checkout_usr123_hold123_attempt1"
}
```

### Cancel booking policy update
`DELETE /api/v1/bookings?bookingId=<bookingId>`

Behavior:
- Allowed only when showtime is at least 1 hour away.
- Returns conflict message when cancellation window is closed.

---

## Testing and Verification Snapshot

### Frontend
- Unit tests passing after Sprint 4 updates:
  - payment component tests expanded for method-specific behavior and validation
- Frontend build passes (`npm run build`)

### Backend
- Backend tests pass (`go test ./...`)
- Integration tests updated and stabilized for checkout/payment idempotency behavior.

---

## Files/Areas Updated in Sprint 4

### Frontend
- `frontend/src/app/pages/payment/payment.component.ts`
- `frontend/src/app/pages/payment/payment.component.html`
- `frontend/src/app/pages/payment/payment.component.scss`
- `frontend/src/app/pages/payment/payment.component.spec.ts`
- `frontend/src/app/services/booking.service.ts`
- `frontend/src/app/pages/my-bookings/my-bookings.component.ts`
- `frontend/src/app/pages/my-bookings/my-bookings.component.html`
- `frontend/src/app/pages/my-bookings/my-bookings.component.scss`
- `frontend/cypress/e2e/payment-flow.cy.ts`
- `frontend/cypress/e2e/full-booking-flow.cy.ts`
- `frontend/README.md`

### Backend
- `backend/internal/repository/booking_repository.go`
- `backend/internal/repository/postgres/booking_repository.go`
- `backend/internal/handler/booking_handler.go`
- `backend/internal/integration/booking_flow_test.go`

---

## Sprint 4 Summary
- Completed major frontend carryover from Sprint 3 with a production-grade payment form and better booking management UX.
- Added three substantial user-facing improvements on My Bookings (search/filter/sort + grouped views).
- Enforced cancellation policy consistently across backend and frontend.
- Extended/verified test coverage to keep the sprint merge-ready.
