# Sprint 3 Report

GitHub Link: https://github.com/adineshreddy/BOX-OFFICE-GO

## Sprint Overview
Sprint 3 focused on backend security and post-seat-selection completion flows.

All 4 backend stories are now implemented:
- BE-05: Token-based Auth Sessions + Logout
- BE-06: Protect Booking APIs with Auth Middleware (No client `userId` trust)
- BE-07: Ticket Download API (`/bookings/{id}/ticket`)
- BE-08: Payment-aware Checkout Flow

---

## Sprint 3 Backend Scope (Completed)

### 1) Auth Sessions + Logout (BE-05)
- Login now returns `accessToken`, `tokenType`, and `expiresAt`.
- Backend stores token sessions in `auth_sessions`.
- Logout endpoint revokes active session tokens.

### 2) Protected Booking APIs (BE-06)
- Booking APIs now require `Authorization: Bearer <token>`.
- Backend resolves user identity from token/session context.
- Client-supplied `userId` is no longer trusted for authorization.

### 3) Ticket Download API (BE-07)
- Added ticket download endpoint returning PDF.
- Enforces booking ownership and confirmed-booking status before download.

### 4) Payment-aware Checkout (BE-08)
- Checkout now requires `paymentMethod` and `idempotencyKey`.
- Payment transaction lifecycle is persisted (`PENDING`, `SUCCESS`, `FAILED`).
- Idempotency check prevents duplicate charges for the same idempotency key.

---

## New/Changed Backend APIs

## Newly Added APIs
- `POST /api/v1/auth/logout`
- `DELETE /api/v1/bookings/holds/{holdId}`
- `GET /api/v1/bookings/{bookingId}/ticket`

## Changed Contract APIs
- `POST /api/v1/auth/login` now returns token fields.
- `POST /api/v1/bookings/checkout` now requires payment fields.
- Booking APIs now require auth token headers.

---

## API Contract Details + Examples

Base URL:
- `http://localhost:8080`
- prefix: `/api/v1`

Common error shape:
```json
{
  "message": "human-readable message",
  "fields": {
    "fieldName": "validation detail"
  }
}
```

### 1) Login (updated)
`POST /api/v1/auth/login`

Request:
```json
{
  "email": "dinesh@example.com",
  "password": "password123"
}
```

Success (`200`):
```json
{
  "message": "login successful",
  "accessToken": "<token>",
  "tokenType": "Bearer",
  "expiresAt": "2026-04-13T20:30:00Z",
  "user": {
    "id": "usr_123",
    "name": "Dinesh",
    "email": "dinesh@example.com",
    "isAdmin": false
  }
}
```

### 2) Logout (new)
`POST /api/v1/auth/logout`

Header:
- `Authorization: Bearer <accessToken>`

Success (`200`):
```json
{
  "message": "logout successful"
}
```

### 3) Create Hold (protected)
`POST /api/v1/bookings/holds`

Headers:
- `Authorization: Bearer <accessToken>`
- `Content-Type: application/json`

Request:
```json
{
  "showtimeId": "st_123",
  "seatNumbers": ["A01", "A02"]
}
```

Success (`201`):
```json
{
  "message": "booking hold created",
  "hold": {
    "holdId": "hold_123",
    "userId": "usr_123",
    "status": "HELD"
  }
}
```

### 4) Release Hold (new, protected)
`DELETE /api/v1/bookings/holds/{holdId}`

Header:
- `Authorization: Bearer <accessToken>`

Success (`200`):
```json
{
  "message": "booking hold released successfully"
}
```

### 5) Checkout (payment-aware, protected)
`POST /api/v1/bookings/checkout`

Headers:
- `Authorization: Bearer <accessToken>`
- `Content-Type: application/json`

Request:
```json
{
  "holdId": "hold_123",
  "paymentMethod": "card",
  "idempotencyKey": "checkout_usr123_hold123_attempt1"
}
```

Success (`200`):
```json
{
  "message": "checkout successful",
  "booking": {
    "bookingId": "bok_123",
    "holdId": "hold_123",
    "status": "CONFIRMED",
    "transactionId": "txn_123"
  }
}
```

Payment decline (`402`):
```json
{
  "message": "payment declined"
}
```

Gateway timeout/unavailable (`502`):
```json
{
  "message": "payment gateway unavailable, please retry"
}
```

### 6) Download Ticket (new, protected)
`GET /api/v1/bookings/{bookingId}/ticket`

Header:
- `Authorization: Bearer <accessToken>`

Success (`200`):
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename="ticket_<bookingId>.pdf"`

Not owner (`403`):
```json
{
  "message": "you do not have access to this booking"
}
```

Not confirmed (`409`):
```json
{
  "message": "ticket download is only available for confirmed bookings (current status: ...)"
}
```

---

---

## Backend Verification Snapshot
- `go test ./...` passes.
- `go vet ./...` passes.
- Current backend/internal ratio snapshot:
  - Production functions: `95`
  - Test functions: `105`
  - Result: test:function ratio is at least `1:1`.

---

## Backend Schema + Config Notes
- Added auth session migration:
  - `backend/db/migrations/007_create_auth_sessions.sql`
- Added payment transaction migration:
  - `backend/db/migrations/008_create_payment_transactions.sql`
- Auth config support:
  - `JWT_SECRET` (default: `dev-secret-change-me`)
  - `AUTH_TOKEN_TTL_MINUTES` (default: `1440`)

---

## Sprint 2 carryover and Sprint 3 completion

Sprint 2 left the booking journey through seat hold in place; Sprint 3 completes **post-hold** flows with **auth**, **payment-aware checkout**, **hold release**, and **ticket download**, plus tests for those paths.

---

## Frontend unit tests (Sprint 2 baseline + Sprint 3 additions)

**Location:** `frontend/src/**/*.spec.ts`

**Snapshot (verified locally):** 7 test files, **13** tests passing.

| File | Notes |
|------|--------|
| `frontend/src/app/app.spec.ts` | App shell / routing smoke |
| `frontend/src/app/services/auth.service.spec.ts` | Auth service behavior |
| `frontend/src/app/services/movie.service.spec.ts` | Movie API client |
| `frontend/src/app/pages/movie-detail/movie-detail.component.spec.ts` | Movie detail |
| `frontend/src/app/pages/movie-seats/movie-seats.component.spec.ts` | Seat selection |
| `frontend/src/app/pages/payment/payment.component.spec.ts` | **Sprint 3:** payment / checkout UI |
| `frontend/src/app/pages/booking-success/booking-success.component.spec.ts` | **Sprint 3:** post-checkout success |

**How to run (shows all unit test results in the terminal):**

```bash
cd frontend
npm test
```

Use `npm test -- --watch=false` if your environment supports a non-watch single run (Angular 21 + Vitest may still open the test runner depending on CLI defaults; for the narrated demo, a full green run of `npm test` is acceptable).

---

## Backend unit tests (Sprint 2 baseline + Sprint 3 additions)

**How to run (entire suite):**

```bash
cd backend
go test ./...
```

Verbose (good for screen recording):

```bash
go test -v ./...
```

**Test files (15 files, `*_test.go`):**

| Package / path |
|----------------|
| `backend/internal/config/config_test.go` |
| `backend/internal/handler/auth_handler_test.go` |
| `backend/internal/handler/auth_handler_extra_test.go` |
| `backend/internal/handler/booking_handler_test.go` |
| `backend/internal/handler/movie_handler_test.go` |
| `backend/internal/handler/movie_handler_extra_test.go` |
| `backend/internal/http/middleware/auth_test.go` |
| `backend/internal/http/middleware/cors_test.go` |
| `backend/internal/http/response/response_test.go` |
| `backend/internal/payment/gateway_test.go` |
| `backend/internal/service/auth_service_test.go` |
| `backend/internal/service/booking_service_test.go` |
| `backend/internal/service/movie_service_test.go` |
| `backend/internal/validation/login_validation_test.go` |
| `backend/internal/validation/signup_validation_test.go` |

Sprint 3-related coverage includes middleware auth, payment gateway behavior, checkout idempotency/decline/timeout paths, ticket PDF service checks, and extended booking/auth handler tests.

---

## Backend API documentation (updated for Sprint 3)

Use these together for a full picture:

1. **`Sprint2.md`** (in repo root next to this file) — complete **baseline** REST reference for movies, shows, seat map, holds (pre–Sprint 3 contract where `userId` was passed on holds), checkout (pre-payment), and booking list/cancel. Still useful for **status tables**, **curl** examples, and **FE integration notes**; **auth and booking hold/checkout contracts changed in Sprint 3** (see below).

2. **`backend/README.md`** — quick start, env vars (`JWT_SECRET`, `AUTH_TOKEN_TTL_MINUTES`), signup/login curl (login response includes **access token** fields), movies/theaters, and **authenticated** cancel booking example.

3. **This document (`Sprint3.md`)** — **Sprint 3 deltas:** `POST /api/v1/auth/logout`, Bearer protection on booking routes, hold body without trusted client `userId`, `DELETE /api/v1/bookings/holds/{holdId}`, payment-aware `POST /api/v1/bookings/checkout`, and `GET /api/v1/bookings/{bookingId}/ticket` (PDF).

---

## Submission checklist (course requirements)

- [ ] **GitHub:** repo link on the submission page (update if the fork/org changed).
- [ ] **Video:** narrated; **each teammate narrates a segment**; demo **new Sprint 3 functionality**; show **all unit tests** — run **frontend** (`npm test` in `frontend`) and **backend** (`go test ./...` or `go test -v ./...` in `backend`) with results visible.
- [ ] **Commits:** everyone needs **individual commits** on the branch you submit; coordinate with your TA if contribution is blocked.
- [ ] **This file:** keep `Sprint3.md` accurate (work completed, test lists, doc pointers above).
