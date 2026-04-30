# Backend (Go)

## Prerequisites

- Go 1.22+
- Neon Postgres project (free tier is fine)

## Environment

Create a `.env` file in `backend` (or export variables in shell):

```bash
cp .env.example .env
```

Set `DATABASE_URL` from Neon dashboard (Connection Details → Pooled connection string).

Optional auth configuration:
- `JWT_SECRET` (default: `dev-secret-change-me`)
- `AUTH_TOKEN_TTL_MINUTES` (default: `1440`)

## Initialize Schema in Neon

Option 1 (already automated): app startup creates `users` table if missing.

Option 2 (manual): run SQL from `db/migrations/001_create_users.sql` in Neon SQL Editor.

## Run

```bash
set -a && source .env && set +a
go mod tidy
go run ./cmd/api
```

Server starts on `http://localhost:8080` by default.

## Run Unit Tests

Run all backend unit tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Run tests for a specific package:

```bash
go test ./internal/service
go test ./internal/handler
go test ./internal/validation
```

Run a specific test function:

```bash
go test ./internal/service -run TestBookingServiceCreateBookingHold_Success
```

## Signup API

- **Method**: `POST`
- **Path**: `/api/v1/auth/signup`

Request body:

```json
{
  "name": "Dinesh",
  "phone": "+919876543210",
  "email": "dinesh@example.com",
  "password": "password123",
  "confirmPassword": "password123"
}
```

### Notes

- Password is hashed using bcrypt.
- Email uniqueness is enforced.
- This implementation writes users directly to Neon Postgres.

## Test Signup (curl)

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Dinesh",
    "phone":"+919876543210",
    "email":"dinesh@example.com",
    "password":"password123",
    "confirmPassword":"password123"
  }'
```

Expected: HTTP `201` and created user payload.

Try same email again to verify duplicate check:

Expected: HTTP `400` with `email already exists` in response fields.

## Verify Row in Neon

Run this in Neon SQL Editor:

```sql
SELECT id, name, phone, email, created_at
FROM users
ORDER BY created_at DESC;
```

## Login API

- **Method**: `POST`
- **Path**: `/api/v1/auth/login`

Request body:

```json
{
  "email": "dinesh@example.com",
  "password": "password123"
}
```

## Test Login (curl)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email":"dinesh@example.com",
    "password":"password123"
  }'
```

Expected: HTTP `200` and user payload.
Response now also includes `accessToken`, `tokenType` (`Bearer`), and `expiresAt`.

Invalid password/email: HTTP `400` with `invalid email or password`.

## Movie Theaters by Selected Movie API

- **Method**: `GET`
- **Path**: `/api/v1/movies/{movieId}/theaters`
- **Optional Query**: `date=YYYY-MM-DD` (filters showtimes for that calendar day in the server’s local timezone)
  - When `date` is provided, it must be **today through today + 14 days** (inclusive). Past dates or dates beyond that window return **HTTP 400** with an error message.

Example:

```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters"
```

With date filter:

```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters?date=2026-02-18"
```

### Behavior

- Returns theater list for the selected movie with showtimes.
- Uses many-to-many relation through `showtimes` (`movie_id` + `theater_id`).
- Shows only active and upcoming showtimes.
- Validates schedule quality: gap between consecutive showtime start times in the same theater must be greater than movie runtime.

## Curl Commands for Non-Auth APIs

### 1) Health Check

```bash
curl "http://localhost:8080/health"
```

### 2) List Movies

```bash
curl "http://localhost:8080/api/v1/movies"
```

### 3) List Movies (filter by title)

```bash
curl "http://localhost:8080/api/v1/movies?title=Starlight"
```

### 4) List Movies (filter by genre)

```bash
curl "http://localhost:8080/api/v1/movies?genre=Drama"
```

### 5) List Theaters + Showtimes for Selected Movie

```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters"
```

```bash
curl "http://localhost:8080/api/v1/movies/mov_002/theaters"
```

```bash
curl "http://localhost:8080/api/v1/movies/mov_003/theaters"
```

### 6) List Theaters + Showtimes for Selected Movie on Specific Date

```bash
curl "http://localhost:8080/api/v1/movies/mov_001/theaters?date=2026-02-18"
```

### 7) Cancel Booking

```bash
curl --request DELETE \
  --url "http://localhost:8080/api/v1/bookings?bookingId=bok_123456789" \
  --header "Authorization: Bearer <access-token>"
```
