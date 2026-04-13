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
