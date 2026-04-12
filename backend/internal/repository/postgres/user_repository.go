package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

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

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string, loggedAt time.Time) error {
	query := `
	UPDATE users
	SET last_login_at = $2, updated_at = $2
	WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, loggedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) CreateSession(ctx context.Context, session domain.AuthSession) error {
	query := `
	INSERT INTO auth_sessions (
		id,
		user_id,
		token_id,
		expires_at,
		revoked_at,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.TokenID,
		session.ExpiresAt,
		session.RevokedAt,
		session.CreatedAt,
	)
	return err
}

func (r *UserRepository) GetSessionByTokenID(ctx context.Context, tokenID string) (domain.AuthSession, error) {
	query := `
	SELECT
		id,
		user_id,
		token_id,
		expires_at,
		revoked_at,
		created_at
	FROM auth_sessions
	WHERE token_id = $1
	`

	var session domain.AuthSession
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(tokenID)).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenID,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AuthSession{}, repository.ErrSessionNotFound
		}
		return domain.AuthSession{}, err
	}

	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return session, nil
}

func (r *UserRepository) RevokeSessionByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error {
	query := `
	UPDATE auth_sessions
	SET revoked_at = COALESCE(revoked_at, $2)
	WHERE token_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, strings.TrimSpace(tokenID), revokedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrSessionNotFound
	}

	return nil
}
