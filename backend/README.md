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

Invalid password/email: HTTP `400` with `invalid email or password`.
