# Sprint 2 Backend API Documentation

## Scope
This document is a frontend-focused backend API guide for Sprint 2.
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
| GET /api/v1/movies/{movieId}/theaters | yes | no | yes | yes | yes | yes | yes |
| GET /api/v1/shows/details | yes | no | yes | yes | yes | no | yes |
| GET /api/v1/shows/seat-map | yes | no | yes | yes | yes | no | yes |
| GET /api/v1/shows/seat-map/availability | yes | no | yes | yes | yes | no | yes |
| POST /api/v1/bookings/holds | no | yes | yes | yes | yes | yes | yes |
| POST /api/v1/bookings/checkout | yes | no | yes | yes | yes | yes | yes |

---

## FE Integration Notes

1. Current backend login returns `user` object only, not JWT token.
2. API prefix is `/api/v1`.
3. `showTime` must be exact RFC3339 timestamp matching selected show start time.
4. `basePrice` is currently normalized to `12.00` for showtimes.
5. Seat map pricing per seat can differ by `priceFactor`; total hold amount is computed backend-side.
