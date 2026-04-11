# Sprint 3 Report

GitHub Link: https://github.com/adineshreddy/BOX-OFFICE-GO

## Sprint Overview
Sprint 3 (backend scope in this report) focused on security hardening and session-based authentication for booking flows.

This report covers **2 completed backend stories**:
- BE-05: Token-based Auth Sessions + Logout
- BE-06: Protect Booking APIs with Auth Middleware (No client `userId` trust)

The other Sprint 3 backend stories (Ticket Download API and Payment-Aware Checkout) are not included here and will be delivered separately.

---

## Sprint 3 Backend Scope (Completed in This Report)

### 1) Token-based Auth Sessions + Logout
- Login now returns a signed bearer token (`accessToken`) and expiry metadata.
- Backend persists auth sessions in `auth_sessions` table.
- New logout API revokes current session/token.
- Revoked or invalid tokens are rejected for protected APIs.

### 2) Protected Booking APIs (Identity from Token)
- Booking APIs now require `Authorization: Bearer <accessToken>`.
- Backend derives user identity from token/session context.
- Backend no longer trusts client-supplied `userId` for booking actions.
- Booking list/cancel operate on the authenticated user only.

---

## Sprint 3 Outcome Summary (Backend)
- Introduced backend-managed auth sessions and token lifecycle (issue → validate → revoke).
- Added auth middleware and applied it to booking routes.
- Preserved existing booking business logic while tightening authorization model.
- Updated tests across service/handler/middleware layers.

---

# Frontend-Focused API Contract (Sprint 3 Changes)

## Base URL
- Local: `http://localhost:8080`
- API prefix: `/api/v1`

## Common Error Shape
All handled API errors follow:

```json
{
  "message": "human-readable message",
  "fields": {
    "fieldName": "validation detail"
  }
}
```

`fields` is optional.

---

## What Changed From Sprint 2

### A) Login response now includes token fields
`POST /api/v1/auth/login` now returns:
- `accessToken` (JWT-like bearer token)
- `tokenType` (`Bearer`)
- `expiresAt` (ISO timestamp)
- existing user object

### B) New API added
- `POST /api/v1/auth/logout` (requires bearer token)

### C) Booking endpoints are now protected
These endpoints require `Authorization` header:
- `POST /api/v1/bookings/holds`
- `POST /api/v1/bookings/checkout`
- `GET /api/v1/bookings`
- `DELETE /api/v1/bookings?bookingId=...`

### D) Client `userId` is no longer trusted for booking authorization
- FE should **not rely on sending `userId`** for auth decisions.
- For backward compatibility, `userId` can still appear in payload shape, but backend overwrites it using token identity.

---

## API Details + Examples

## 1) Login

### API
`POST /api/v1/auth/login`

### Request body
```json
{
  "email": "dinesh@example.com",
  "password": "password123"
}
```

### Success response (`200`)
```json
{
  "message": "login successful",
  "accessToken": "<token>",
  "tokenType": "Bearer",
  "expiresAt": "2026-04-12T18:45:00Z",
  "user": {
    "id": "usr_123",
    "name": "Dinesh",
    "phone": "+1234567890",
    "email": "dinesh@example.com",
    "isAdmin": false,
    "isActive": true,
    "isVerified": false,
    "lastLoginAt": "2026-04-11T18:45:00Z",
    "createdAt": "2026-03-01T11:00:00Z",
    "updatedAt": "2026-04-11T18:45:00Z"
  }
}
```

### Curl
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email":"dinesh@example.com",
    "password":"password123"
  }'
```

---

## 2) Logout (New)

### API
`POST /api/v1/auth/logout`

### Headers
- `Authorization: Bearer <accessToken>`

### Success response (`200`)
```json
{
  "message": "logout successful"
}
```

### Unauthorized (`401`) example
```json
{
  "message": "invalid or expired auth token"
}
```

### Curl
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <accessToken>"
```

---

## 3) Create Booking Hold (Protected)

### API
`POST /api/v1/bookings/holds`

### Headers
- `Authorization: Bearer <accessToken>`
- `Content-Type: application/json`

### Request body
```json
{
  "showtimeId": "st_123",
  "seatNumbers": ["A01", "A02"]
}
```

### Success response (`201`)
```json
{
  "message": "booking hold created",
  "hold": {
    "holdId": "hold_123",
    "userId": "usr_123",
    "showtimeId": "st_123",
    "seatNumbers": ["A01", "A02"],
    "status": "HELD",
    "holdExpiresAt": "2026-04-11T19:01:00Z",
    "totalAmount": 24,
    "createdAt": "2026-04-11T18:54:00Z"
  }
}
```

---

## 4) Checkout Booking Hold (Protected)

### API
`POST /api/v1/bookings/checkout`

### Headers
- `Authorization: Bearer <accessToken>`
- `Content-Type: application/json`

### Request body
```json
{
  "holdId": "hold_123"
}
```

### Success response (`200`)
```json
{
  "message": "checkout successful",
  "booking": {
    "bookingId": "bok_123",
    "holdId": "hold_123",
    "userId": "usr_123",
    "showtimeId": "st_123",
    "seatNumbers": ["A01", "A02"],
    "status": "CONFIRMED",
    "totalAmount": 24,
    "confirmedAt": "2026-04-11T18:56:00Z"
  }
}
```

---

## 5) Get My Bookings (Protected)

### API
`GET /api/v1/bookings`

### Headers
- `Authorization: Bearer <accessToken>`

### Success response (`200`)
```json
{
  "bookings": [
    {
      "bookingId": "bok_123",
      "movieTitle": "Starlight Horizon",
      "theaterName": "AMC Downtown 12",
      "seatNumbers": ["A01", "A02"]
    }
  ]
}
```

---

## 6) Cancel Booking (Protected)

### API
`DELETE /api/v1/bookings?bookingId=<bookingId>`

### Headers
- `Authorization: Bearer <accessToken>`

### Success response (`200`)
```json
{
  "message": "booking cancelled successfully"
}
```

### Curl
```bash
curl --request DELETE \
  --url "http://localhost:8080/api/v1/bookings?bookingId=bok_123" \
  --header "Authorization: Bearer <accessToken>"
```

---

## Frontend Integration Notes (Actionable)
- After login, store `accessToken` and attach it as `Authorization: Bearer <accessToken>` for booking/logout requests.
- Remove dependency on sending `userId` from FE for booking auth (backend derives user from token).
- On `401`, redirect user to login and clear stale local session/token.
- Public browse APIs remain unchanged and do not require token:
  - `/health`
  - `/api/v1/movies`
  - `/api/v1/movies/{movieId}`
  - `/api/v1/movies/{movieId}/theaters`
  - `/api/v1/shows/details`
  - `/api/v1/shows/seat-map`
  - `/api/v1/shows/seat-map/availability`

---

## Backend Notes
- Added migration: `backend/db/migrations/007_create_auth_sessions.sql`
- Added config support:
  - `JWT_SECRET` (default: `dev-secret-change-me`)
  - `AUTH_TOKEN_TTL_MINUTES` (default: `1440`)
- Auth + booking middleware/service/handler tests are passing.
