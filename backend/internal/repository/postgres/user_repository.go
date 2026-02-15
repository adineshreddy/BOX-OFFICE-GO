package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query := `
	INSERT INTO users (
		id,
		name,
		phone,
		email,
		password_hash,
		is_admin,
		is_active,
		email_verified,
		last_login_at,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Phone,
		strings.ToLower(strings.TrimSpace(user.Email)),
		user.PasswordHash,
		user.IsAdmin,
		user.IsActive,
		user.IsVerified,
		user.LastLoginAt,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, repository.ErrEmailExists
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
	SELECT
		id,
		name,
		phone,
		email,
		password_hash,
		is_admin,
		is_active,
		email_verified,
		last_login_at,
		created_at,
		updated_at
	FROM users
	WHERE email = $1
	`

	var user domain.User
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.IsActive,
		&user.IsVerified,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}
