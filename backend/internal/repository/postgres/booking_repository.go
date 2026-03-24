package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CleanupExpiredHolds(ctx context.Context) error {
	query := `
	WITH expired AS (
		UPDATE booking_holds
		SET status = 'EXPIRED', updated_at = NOW()
		WHERE status = 'HELD' AND hold_expires_at < NOW()
		RETURNING id
	)
	UPDATE seat_inventory si
	SET is_held = FALSE, updated_at = NOW()
	FROM booking_hold_seats bhs
	JOIN expired e ON e.id = bhs.hold_id
	WHERE si.showtime_id = bhs.showtime_id
	  AND si.seat_number = bhs.seat_number
	  AND si.is_available = TRUE;
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *BookingRepository) CreateHold(ctx context.Context, input domain.CreateBookingHoldInput, holdID string, holdExpiresAt time.Time) (domain.BookingHold, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.BookingHold{}, err
	}
	defer tx.Rollback()

	var basePrice float64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT base_price FROM showtimes WHERE id = $1 AND is_active = TRUE`,
		input.ShowtimeID,
	).Scan(&basePrice); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BookingHold{}, repository.ErrShowtimeNotFound
		}
		return domain.BookingHold{}, err
	}

	seatState := make(map[string]struct {
		isAvailable bool
		isHeld      bool
	})

	rows, err := tx.QueryContext(
		ctx,
		`SELECT seat_number, is_available, is_held FROM seat_inventory WHERE showtime_id = $1 AND seat_number = ANY($2) FOR UPDATE`,
		input.ShowtimeID,
		input.SeatNumbers,
	)
	if err != nil {
		return domain.BookingHold{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var seatNumber string
		var isAvailable bool
		var isHeld bool
		if scanErr := rows.Scan(&seatNumber, &isAvailable, &isHeld); scanErr != nil {
			return domain.BookingHold{}, scanErr
		}
		seatState[seatNumber] = struct {
			isAvailable bool
			isHeld      bool
		}{
			isAvailable: isAvailable,
			isHeld:      isHeld,
		}
	}
	if err := rows.Err(); err != nil {
		return domain.BookingHold{}, err
	}

	if len(seatState) != len(input.SeatNumbers) {
		return domain.BookingHold{}, repository.ErrInvalidSeatSelection
	}

	for _, seatNumber := range input.SeatNumbers {
		state := seatState[seatNumber]
		if !state.isAvailable || state.isHeld {
			return domain.BookingHold{}, repository.ErrSeatUnavailable
		}
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO booking_holds (id, user_id, showtime_id, status, hold_expires_at, total_amount, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 0, $6, $6)`,
		holdID,
		input.UserID,
		input.ShowtimeID,
		domain.BookingHoldStatusHeld,
		holdExpiresAt,
		now,
	)
	if err != nil {
		return domain.BookingHold{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO booking_hold_seats (hold_id, showtime_id, seat_number, price_at_hold, created_at)
		 SELECT $1, $2, si.seat_number, ROUND(($3::numeric * si.price_multiplier)::numeric, 2), NOW()
		 FROM seat_inventory si
		 WHERE si.showtime_id = $2 AND si.seat_number = ANY($4)`,
		holdID,
		input.ShowtimeID,
		basePrice,
		input.SeatNumbers,
	)
	if err != nil {
		return domain.BookingHold{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE seat_inventory
		 SET is_held = TRUE, updated_at = NOW()
		 WHERE showtime_id = $1 AND seat_number = ANY($2)`,
		input.ShowtimeID,
		input.SeatNumbers,
	)
	if err != nil {
		return domain.BookingHold{}, err
	}

	var totalAmount float64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(price_at_hold), 0) FROM booking_hold_seats WHERE hold_id = $1`, holdID).Scan(&totalAmount); err != nil {
		return domain.BookingHold{}, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE booking_holds SET total_amount = $2, updated_at = NOW() WHERE id = $1`, holdID, totalAmount)
	if err != nil {
		return domain.BookingHold{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.BookingHold{}, err
	}

	return domain.BookingHold{
		HoldID:        holdID,
		UserID:        input.UserID,
		ShowtimeID:    input.ShowtimeID,
		SeatNumbers:   input.SeatNumbers,
		Status:        domain.BookingHoldStatusHeld,
		HoldExpiresAt: holdExpiresAt,
		TotalAmount:   totalAmount,
		CreatedAt:     now,
	}, nil
}

func (r *BookingRepository) CheckoutHold(ctx context.Context, holdID string, userID string, bookingID string) (domain.BookingCheckoutResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.BookingCheckoutResult{}, err
	}
	defer tx.Rollback()

	var holdStatus string
	var showtimeID string
	var holdExpiresAt time.Time
	var totalAmount float64

	err = tx.QueryRowContext(
		ctx,
		`SELECT status, showtime_id, hold_expires_at, total_amount
		 FROM booking_holds
		 WHERE id = $1 AND user_id = $2
		 FOR UPDATE`,
		holdID,
		userID,
	).Scan(&holdStatus, &showtimeID, &holdExpiresAt, &totalAmount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BookingCheckoutResult{}, repository.ErrHoldNotFound
		}
		return domain.BookingCheckoutResult{}, err
	}

	if holdStatus == domain.BookingHoldStatusConfirmed {
		return domain.BookingCheckoutResult{}, repository.ErrHoldFinalized
	}

	now := time.Now().UTC()
	if holdStatus != domain.BookingHoldStatusHeld || holdExpiresAt.Before(now) {
		_, _ = tx.ExecContext(ctx, `UPDATE booking_holds SET status = 'EXPIRED', updated_at = NOW() WHERE id = $1`, holdID)
		_, _ = tx.ExecContext(
			ctx,
			`UPDATE seat_inventory si
			 SET is_held = FALSE, updated_at = NOW()
			 FROM booking_hold_seats bhs
			 WHERE bhs.hold_id = $1
			   AND si.showtime_id = bhs.showtime_id
			   AND si.seat_number = bhs.seat_number
			   AND si.is_available = TRUE`,
			holdID,
		)
		if commitErr := tx.Commit(); commitErr != nil {
			return domain.BookingCheckoutResult{}, commitErr
		}
		return domain.BookingCheckoutResult{}, repository.ErrHoldExpired
	}

	seatRows, err := tx.QueryContext(ctx, `SELECT seat_number FROM booking_hold_seats WHERE hold_id = $1 ORDER BY seat_number ASC`, holdID)
	if err != nil {
		return domain.BookingCheckoutResult{}, err
	}
	defer seatRows.Close()

	seatNumbers := make([]string, 0)
	for seatRows.Next() {
		var seatNumber string
		if scanErr := seatRows.Scan(&seatNumber); scanErr != nil {
			return domain.BookingCheckoutResult{}, scanErr
		}
		seatNumbers = append(seatNumbers, seatNumber)
	}
	if err := seatRows.Err(); err != nil {
		return domain.BookingCheckoutResult{}, err
	}

	if len(seatNumbers) == 0 {
		return domain.BookingCheckoutResult{}, repository.ErrInvalidSeatSelection
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE seat_inventory
		 SET is_available = FALSE,
		     is_held = FALSE,
		     updated_at = NOW()
		 WHERE showtime_id = $1 AND seat_number = ANY($2)`,
		showtimeID,
		seatNumbers,
	)
	if err != nil {
		return domain.BookingCheckoutResult{}, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE booking_holds SET status = 'CONFIRMED', updated_at = NOW() WHERE id = $1`, holdID)
	if err != nil {
		return domain.BookingCheckoutResult{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO bookings (id, hold_id, user_id, showtime_id, status, total_amount, confirmed_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'CONFIRMED', $5, $6, $6, $6)`,
		bookingID,
		holdID,
		userID,
		showtimeID,
		totalAmount,
		now,
	)
	if err != nil {
		return domain.BookingCheckoutResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.BookingCheckoutResult{}, err
	}

	sort.Strings(seatNumbers)
	return domain.BookingCheckoutResult{
		BookingID:   bookingID,
		HoldID:      holdID,
		UserID:      userID,
		ShowtimeID:  showtimeID,
		SeatNumbers: seatNumbers,
		Status:      domain.BookingHoldStatusConfirmed,
		TotalAmount: totalAmount,
		ConfirmedAt: now,
	}, nil
}

var _ repository.BookingRepository = (*BookingRepository)(nil)
