# Sprint 2 Report

GitHub Link: https://github.com/adineshreddy/BOX-OFFICE-GO

## Sprint Overview
Sprint 2 focused on completing backend booking workflows and finalizing FE-consumable APIs for show selection, seat visibility, hold/checkout flow, and booking management.

## Sprint 2 Backend Scope (Completed)
- Show details + seat map + seat availability APIs.
- Booking hold + checkout workflow with hold expiry handling.
- Booking management APIs: user booking list and booking cancel.
- Contract hardening: ID-based query contract (`movieId`, `theaterId`, `showTime`) and `userId` query preference (with `user_id` compatibility).
- Stability updates: hold cleanup worker, non-destructive startup schema handling, and expanded unit tests.

## Sprint 2 Outcome Summary
- Backend APIs required by current frontend flows are implemented and wired.
- Hold lifecycle behavior is enforced (`HELD`, `EXPIRED`, `CONFIRMED`, `CANCELLED`) with automatic release after expiry.
- Seat pricing and showtime base-price behavior are normalized for current sprint assumptions.
- Backend test suite now covers core validation, handlers, and service branches.

## Backend Unit Tests (Implemented)

Current snapshot:
- Test files: 13
- Test functions: 66
- Production functions (backend/internal): 65

Test files:
- backend/internal/config/config_test.go
- backend/internal/handler/auth_handler_extra_test.go
- backend/internal/handler/auth_handler_test.go
- backend/internal/handler/booking_handler_test.go
- backend/internal/handler/movie_handler_extra_test.go
- backend/internal/handler/movie_handler_test.go
- backend/internal/http/middleware/cors_test.go
- backend/internal/http/response/response_test.go
- backend/internal/service/auth_service_test.go
- backend/internal/service/booking_service_test.go
- backend/internal/service/movie_service_test.go
- backend/internal/validation/login_validation_test.go
- backend/internal/validation/signup_validation_test.go

---

# Sprint 2 Backend API Documentation

## Scope
This section is a frontend-focused backend API guide for Sprint 2.
It includes:
- endpoint purpose
- request parameters and data types
- response contracts
- common errors
- curl examples
- integration notes important for FE implementation

Base URL:
- Local: `http://localhost:8080`
- API prefix: `/api/v1`

---

## Common Conventions

### Content Type
- Request body: `application/json`
- Response body: `application/json`

### Error Response Shape
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

### Date/Time Formats
- `showTime` query parameter must be RFC3339, for example:
  - `2026-03-21T15:00:00Z`
- Most timestamps in responses are ISO timestamps.

---

## Health

### API Name
Health Check

### API
`GET /health`

### What It Does
Checks whether backend server is running.

### Request Parameters
None.

### Response
```json
{
  "status": "ok"
}
```

### Curl
```bash
curl "http://localhost:8080/health"
```

---

## Authentication

### API Name
Signup

### API
`POST /api/v1/auth/signup`

### What It Does
Creates a user account.

### Request Body
| Field | Type | Required | Notes |
|---|---|---|---|
| name | string | yes | user display name |
| phone | string | yes | phone number |
| email | string | yes | unique email |
| password | string | yes | plain text input; backend hashes it |
| confirmPassword | string | yes | must match password |

### Success Response (201)
```json
{
  "message": "account created successfully",
  "user": {
    "id": "usr_...",
    "name": "Dinesh",
    "phone": "+919876543210",
    "email": "dinesh@example.com",
    "createdAt": "2026-03-21T...Z"
  }
}
```

### Curl
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/auth/signup \
  --header 'content-type: application/json' \
  --data '{
    "name": "Dinesh",
    "phone": "+919876543210",
    "email": "dinesh@example.com",
    "password": "password123",
    "confirmPassword": "password123"
  }'
```

---

### API Name
Login

### API
`POST /api/v1/auth/login`

### What It Does
Authenticates user credentials.

### Request Body
| Field | Type | Required |
|---|---|---|
| email | string | yes |
| password | string | yes |

### Success Response (200)
```json
{
  "message": "login successful",
  "user": {
    "id": "usr_...",
    "name": "Dinesh",
    "phone": "+919876543210",
    "email": "dinesh@example.com",
    "isAdmin": false,
    "isActive": true,
    "isVerified": false,
    "lastLoginAt": "2026-03-21T...Z",
    "createdAt": "2026-03-21T...Z",
    "updatedAt": "2026-03-21T...Z"
  }
}
```

### Curl
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/auth/login \
  --header 'content-type: application/json' \
  --data '{
    "email": "dinesh@example.com",
    "password": "password123"
  }'
```

---

## Movies and Show Discovery

### API Name
List Movies

### API
`GET /api/v1/movies`

### What It Does
Returns active movies with optional filters.

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| title | string | no | partial match |
| genre | string | no | partial match |

### Success Response (200)
```json
{
  "movies": [
    {
      "id": "mov_001",
      "title": "Starlight Horizon",
      "description": "...",
      "genre": "Sci-Fi",
      "language": "English",
      "durationMinutes": 128,
      "releaseDate": "2024-06-14T00:00:00Z",
      "rating": 8.4,
      "posterUrl": "https://...",
      "isActive": true,
      "createdAt": "2026-...",
      "updatedAt": "2026-..."
    }
  ]
}
```

### Curl
```bash
curl "http://localhost:8080/api/v1/movies"
```

With filters:
```bash
curl "http://localhost:8080/api/v1/movies?title=Star&genre=Sci"
```

---

### API Name
Get Movie by ID

### API
`GET /api/v1/movies/{movieId}`

### What It Does
Returns details of a single movie by `movieId`.

### Path Parameters
| Param | Type | Required |
|---|---|---|
| movieId | string | yes |

### Success Response (200)
```json
{
  "id": "mov_001",
  "title": "Starlight Horizon",
  "description": "...",
  "genre": "Sci-Fi",
  "language": "English",
  "durationMinutes": 128,
  "releaseDate": "2024-06-14T00:00:00Z",
  "rating": 8.4,
  "posterUrl": "https://...",
  "isActive": true,
  "createdAt": "2026-...",
  "updatedAt": "2026-..."
}
```

### Curl
```bash
curl "http://localhost:8080/api/v1/movies/mov_001"
```

---

### API Name
List Theaters and Showtimes for a Movie

### API
`GET /api/v1/movies/{movieId}/theaters`

### What It Does
Returns all theaters and upcoming showtimes for a selected movie.

### Path Parameters
| Param | Type | Required |
|---|---|---|
| movieId | string | yes |

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| date | string | no | format `YYYY-MM-DD` |

### Success Response (200)
```json
{
  "movieId": "mov_001",
  "movieTitle": "Starlight Horizon",
  "durationMinutes": 128,
  "theaters": [
    {
      "theaterId": "th_001",
      "theaterName": "AMC Downtown 12",
      "city": "New York",
      "addressLine1": "123 W 42nd St, Manhattan",
      "timezone": "America/New_York",
      "showtimes": [
        {
          "showtimeId": "st_...",
          "screenName": "Screen 1",
          "startTime": "2026-03-21T15:00:00Z",
          "endTime": "2026-03-21T17:08:00Z",
          "language": "English",
          "format": "2D",
          "basePrice": 12
        }
      ]
    }
  ]
}
```

### Curl
```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters"
```

With date filter:
```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters?date=2026-03-21"
```

---

## Show Details and Seat APIs (Sprint 2 BE-02)

### API Name
Get Show Details (selected movie/theater/showtime)

### API
`GET /api/v1/shows/details`

### What It Does
Returns detailed info for one selected show, including seat counts.

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| movieId | string | yes | example `mov_001` |
| theaterId | string | yes | example `th_001` |
| showTime | string | yes | RFC3339, exact show start time |

### Success Response (200)
```json
{
  "showtimeId": "st_...",
  "movieId": "mov_001",
  "movieTitle": "Starlight Horizon",
  "theaterId": "th_001",
  "theaterName": "AMC Downtown 12",
  "city": "New York",
  "addressLine1": "123 W 42nd St, Manhattan",
  "screenName": "Screen 1",
  "startTime": "2026-03-21T15:00:00Z",
  "language": "English",
  "format": "2D",
  "basePrice": 12,
  "durationMinutes": 128,
  "availableSeats": 120,
  "totalSeats": 120,
  "unavailableSeats": 0
}
```

### Curl
```bash
curl -G "http://localhost:8080/api/v1/shows/details" \
  --data-urlencode "movieId=mov_001" \
  --data-urlencode "theaterId=th_001" \
  --data-urlencode "showTime=2026-03-21T15:00:00Z"
```

---

### API Name
Get Seat Map

### API
`GET /api/v1/shows/seat-map`

### What It Does
Returns row-wise seat map with per-seat availability and pricing multiplier.

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| movieId | string | yes |
| theaterId | string | yes |
| showTime | string | yes | RFC3339 |

### Success Response (200)
```json
{
  "showtimeId": "st_...",
  "movieTitle": "Starlight Horizon",
  "theaterName": "AMC Downtown 12",
  "screenName": "Screen 1",
  "showTime": "2026-03-21T15:00:00Z",
  "totalSeats": 120,
  "availableSeats": 120,
  "unavailableSeats": 0,
  "rows": [
    {
      "rowLabel": "A",
      "seats": [
        {
          "seatNumber": "A01",
          "rowLabel": "A",
          "seatIndex": 1,
          "seatType": "premium",
          "priceFactor": 1.4,
          "isAvailable": true,
          "isHeld": false
        }
      ]
    }
  ]
}
```

### Curl
```bash
curl -G "http://localhost:8080/api/v1/shows/seat-map" \
  --data-urlencode "movieId=mov_001" \
  --data-urlencode "theaterId=th_001" \
  --data-urlencode "showTime=2026-03-21T15:00:00Z"
```

---

### API Name
Refresh Seat Availability

### API
`GET /api/v1/shows/seat-map/availability`

### What It Does
Returns lightweight seat counters for polling-based availability refresh.

### Query Parameters
| Param | Type | Required |
|---|---|---|
| movieId | string | yes |
| theaterId | string | yes |
| showTime | string | yes |

### Success Response (200)
```json
{
  "showtimeId": "st_...",
  "movieTitle": "Starlight Horizon",
  "theaterName": "AMC Downtown 12",
  "showTime": "2026-03-21T15:00:00Z",
  "totalSeats": 120,
  "availableSeats": 118,
  "unavailableSeats": 2,
  "lastRefreshedAt": "2026-03-21T16:10:00Z"
}
```

### Curl
```bash
curl -G "http://localhost:8080/api/v1/shows/seat-map/availability" \
  --data-urlencode "movieId=mov_001" \
  --data-urlencode "theaterId=th_001" \
  --data-urlencode "showTime=2026-03-21T15:00:00Z"
```

---

## Booking Hold and Checkout APIs (Sprint 2 BE-04)

### API Name
Create Booking Hold

### API
`POST /api/v1/bookings/holds`

### What It Does
Temporarily reserves selected seats for checkout.

### Request Body
| Field | Type | Required | Notes |
|---|---|---|---|
| userId | string | yes | currently provided by FE from auth user payload |
| showtimeId | string | yes | from show listing/details APIs |
| seatNumbers | string[] | yes | one or more seat ids, example `A01` |

### Success Response (201)
```json
{
  "message": "booking hold created",
  "hold": {
    "holdId": "hold_...",
    "userId": "usr_...",
    "showtimeId": "st_...",
    "seatNumbers": ["A01", "A02"],
    "status": "HELD",
    "holdExpiresAt": "2026-03-21T16:17:00Z",
    "totalAmount": 33.6,
    "createdAt": "2026-03-21T16:10:00Z"
  }
}
```

### Curl
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/bookings/holds \
  --header 'content-type: application/json' \
  --data '{
    "userId": "usr_123456789",
    "showtimeId": "st_xxxxxxxxxxxxxxxxxxxx",
    "seatNumbers": ["A01", "A02"]
  }'
```

---

### API Name
Confirm Checkout

### API
`POST /api/v1/bookings/checkout`

### What It Does
Finalizes a hold as a confirmed booking.

### Request Body
| Field | Type | Required |
|---|---|---|
| holdId | string | yes |
| userId | string | yes |

### Success Response (200)
```json
{
  "message": "checkout successful",
  "booking": {
    "bookingId": "bok_...",
    "holdId": "hold_...",
    "userId": "usr_...",
    "showtimeId": "st_...",
    "seatNumbers": ["A01", "A02"],
    "status": "CONFIRMED",
    "totalAmount": 33.6,
    "confirmedAt": "2026-03-21T16:12:00Z"
  }
}
```

### Curl
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/bookings/checkout \
  --header 'content-type: application/json' \
  --data '{
    "holdId": "hold_123456789",
    "userId": "usr_123456789"
  }'
```

---

### API Name
Get User Bookings

### API
`GET /api/v1/bookings`

### What It Does
Returns all bookings for a given user.

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| userId | string | yes | primary param |
| user_id | string | no | backward-compatible alias |

### Success Response (200)
```json
{
  "bookings": [
    {
      "bookingId": "bok_...",
      "holdId": "hold_...",
      "userId": "usr_...",
      "showtimeId": "st_...",
      "seatNumbers": ["A01", "A02"],
      "status": "CONFIRMED",
      "totalAmount": 33.6,
      "confirmedAt": "2026-03-21T16:12:00Z",
      "movieTitle": "Starlight Horizon",
      "theaterName": "AMC Downtown 12",
      "city": "New York",
      "screenName": "Screen 1",
      "showTime": "2026-03-21T15:00:00Z",
      "language": "English",
      "format": "2D"
    }
  ]
}
```

### Curl
```bash
curl -G "http://localhost:8080/api/v1/bookings" \
  --data-urlencode "userId=usr_123456789"
```

---

### API Name
Cancel Booking

### API
`DELETE /api/v1/bookings?bookingId={bookingId}&userId={userId}`

### What It Does
Cancels a user booking and releases seats back to availability.

### Query Parameters
| Param | Type | Required | Notes |
|---|---|---|---|
| bookingId | string | yes | booking to cancel |
| userId | string | yes | primary param |
| user_id | string | no | backward-compatible alias |

### Success Response (200)
```json
{
  "message": "booking cancelled successfully"
}
```

### Curl
```bash
curl --request DELETE \
  --url "http://localhost:8080/api/v1/bookings?bookingId=bok_123456789&userId=usr_123456789"
```

---

## Booking Lifecycle Behavior (Important for FE)

### Hold Duration
- Hold duration is **7 minutes** from hold creation.

### Automatic Expiry
- Backend runs a background cleanup every 1 minute.
- Expired holds are marked expired.
- Their seats are released automatically.

### Expired Hold Checkout
- Checkout after expiry returns conflict error.
- FE should re-fetch seat map/availability and ask user to reselect seats.

---

## Recommended FE Flow

1. List movies.
2. Select movie and fetch theaters/showtimes.
3. Fetch seat map for chosen show.
4. On seat selection, call create hold.
5. Show countdown timer using `holdExpiresAt`.
6. Call checkout before timer expires.
7. On failure (`booking hold expired` or seat unavailable), refresh seat availability and restart selection.

---

## Status Code Quick Map

| Endpoint | 200 | 201 | 400 | 404 | 405 | 409 | 500 |
|---|---|---|---|---|---|---|---|
| GET /health | yes | no | no | no | no | no | no |
| POST /api/v1/auth/signup | no | yes | yes | no | yes | no | yes |
| POST /api/v1/auth/login | yes | no | yes | no | yes | no | yes |
| GET /api/v1/movies | yes | no | no | no | yes | no | yes |
| GET /api/v1/movies/{movieId} | yes | no | yes | yes | yes | no | yes |
| GET /api/v1/movies/{movieId}/theaters | yes | no | yes | yes | yes | yes | yes |
| GET /api/v1/shows/details | yes | no | yes | yes | yes | no | yes |
| GET /api/v1/shows/seat-map | yes | no | yes | yes | yes | no | yes |
| GET /api/v1/shows/seat-map/availability | yes | no | yes | yes | yes | no | yes |
| POST /api/v1/bookings/holds | no | yes | yes | yes | yes | yes | yes |
| POST /api/v1/bookings/checkout | yes | no | yes | yes | yes | yes | yes |
| GET /api/v1/bookings | yes | no | yes | no | yes | no | yes |
| DELETE /api/v1/bookings | yes | no | yes | yes | yes | yes | yes |

---

## FE Integration Notes

1. Current backend login returns `user` object only, not JWT token.
2. API prefix is `/api/v1`.
3. `showTime` must be exact RFC3339 timestamp matching selected show start time.
4. `basePrice` is currently normalized to `12.00` for showtimes.
5. Seat map pricing per seat can differ by `priceFactor`; total hold amount is computed backend-side.
6. Booking list/cancel endpoints now prefer `userId` query param; `user_id` is still accepted for backward compatibility.
