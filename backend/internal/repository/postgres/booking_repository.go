package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
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
	var showStart time.Time
	if err := tx.QueryRowContext(
		ctx,
		`SELECT base_price, start_time FROM showtimes WHERE id = $1 AND is_active = TRUE`,
		input.ShowtimeID,
	).Scan(&basePrice, &showStart); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BookingHold{}, repository.ErrShowtimeNotFound
		}
		return domain.BookingHold{}, err
	}

	if !showStart.After(time.Now().UTC()) {
		return domain.BookingHold{}, repository.ErrShowtimeStarted
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

func (r *BookingRepository) ListByUserID(ctx context.Context, userID string) ([]domain.UserBooking, error) {
	query := `
	SELECT
		b.id,
		b.hold_id,
		b.user_id,
		b.showtime_id,
		b.status,
		b.total_amount,
		b.confirmed_at,
		m.title AS movie_title,
		t.name AS theater_name,
		t.city,
		s.screen_name,
		s.start_time,
		s.language,
		s.format
	FROM bookings b
	JOIN showtimes s ON s.id = b.showtime_id
	JOIN movies m ON m.id = s.movie_id
	JOIN theaters t ON t.id = s.theater_id
	WHERE b.user_id = $1
	ORDER BY b.confirmed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domain.UserBooking, 0)
	for rows.Next() {
		var ub domain.UserBooking

		if scanErr := rows.Scan(
			&ub.BookingID,
			&ub.HoldID,
			&ub.UserID,
			&ub.ShowtimeID,
			&ub.Status,
			&ub.TotalAmount,
			&ub.ConfirmedAt,
			&ub.MovieTitle,
			&ub.TheaterName,
			&ub.City,
			&ub.ScreenName,
			&ub.ShowTime,
			&ub.Language,
			&ub.Format,
		); scanErr != nil {
			return nil, scanErr
		}

		// Fetch seat numbers from booking_hold_seats
		seatRows, seatErr := r.db.QueryContext(ctx,
			`SELECT seat_number FROM booking_hold_seats WHERE hold_id = $1 ORDER BY seat_number ASC`,
			ub.HoldID,
		)
		if seatErr != nil {
			return nil, seatErr
		}

		seats := make([]string, 0)
		for seatRows.Next() {
			var seat string
			if err := seatRows.Scan(&seat); err != nil {
				seatRows.Close()
				return nil, err
			}
			seats = append(seats, seat)
		}
		seatRows.Close()
		if err := seatRows.Err(); err != nil {
			return nil, err
		}

		ub.SeatNumbers = seats
		bookings = append(bookings, ub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID string, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var holdID string
	var showStart time.Time

	err = tx.QueryRowContext(ctx,
		`SELECT b.status, b.hold_id, s.start_time
		 FROM bookings b
		 JOIN showtimes s ON s.id = b.showtime_id
		 WHERE b.id = $1 AND b.user_id = $2
		 FOR UPDATE OF b`,
		bookingID, userID,
	).Scan(&status, &holdID, &showStart)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrBookingNotFound
		}
		return err
	}

	if status == "CANCELLED" {
		return repository.ErrBookingAlreadyCancelled
	}

	now := time.Now().UTC()
	if !showStart.After(now) {
		return repository.ErrShowtimeStarted
	}

	// Cancel the booking
	if _, err := tx.ExecContext(ctx,
		`UPDATE bookings SET status = 'CANCELLED', updated_at = $2 WHERE id = $1`,
		bookingID, now,
	); err != nil {
		return err
	}

	// Cancel the associated hold
	if _, err := tx.ExecContext(ctx,
		`UPDATE booking_holds SET status = 'CANCELLED', updated_at = $2 WHERE id = $1`,
		holdID, now,
	); err != nil {
		return err
	}

	// Release seats: mark them as available and not held
	if _, err := tx.ExecContext(ctx,
		`UPDATE seat_inventory si
		 SET is_available = TRUE, is_held = FALSE, updated_at = $2
		 FROM booking_hold_seats bhs
		 WHERE bhs.hold_id = $1
		   AND si.showtime_id = bhs.showtime_id
		   AND si.seat_number = bhs.seat_number`,
		holdID, now,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BookingRepository) GetBookingForTicket(ctx context.Context, bookingID string) (domain.TicketData, error) {
	query := `
	SELECT
		b.id,
		b.user_id,
		b.status,
		b.total_amount,
		b.confirmed_at,
		b.hold_id,
		m.title AS movie_title,
		t.name  AS theater_name,
		t.city,
		s.screen_name,
		s.start_time,
		s.language,
		s.format
	FROM bookings b
	JOIN showtimes s ON s.id = b.showtime_id
	JOIN movies m    ON m.id = s.movie_id
	JOIN theaters t  ON t.id = s.theater_id
	WHERE b.id = $1
	`

	var td domain.TicketData
	var holdID string

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&td.BookingID,
		&td.UserID,
		&td.Status,
		&td.TotalAmount,
		&td.ConfirmedAt,
		&holdID,
		&td.MovieTitle,
		&td.TheaterName,
		&td.City,
		&td.ScreenName,
		&td.ShowTime,
		&td.Language,
		&td.Format,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TicketData{}, repository.ErrBookingNotFound
		}
		return domain.TicketData{}, err
	}

	seatRows, err := r.db.QueryContext(ctx,
		`SELECT seat_number FROM booking_hold_seats WHERE hold_id = $1 ORDER BY seat_number ASC`,
		holdID,
	)
	if err != nil {
		return domain.TicketData{}, err
	}
	defer seatRows.Close()

	seats := make([]string, 0)
	for seatRows.Next() {
		var seat string
		if scanErr := seatRows.Scan(&seat); scanErr != nil {
			return domain.TicketData{}, scanErr
		}
		seats = append(seats, seat)
	}
	if err := seatRows.Err(); err != nil {
		return domain.TicketData{}, err
	}

	td.SeatNumbers = seats
	return td, nil
}

// ── Payment transaction methods ──────────────────────────────────────

func (r *BookingRepository) CreatePaymentTransaction(ctx context.Context, txn domain.PaymentTransaction) error {
	query := `
	INSERT INTO payment_transactions (id, hold_id, user_id, amount, payment_method, gateway_txn_id, status, failure_reason, idempotency_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		txn.ID,
		txn.HoldID,
		txn.UserID,
		txn.Amount,
		txn.PaymentMethod,
		txn.GatewayTxnID,
		txn.Status,
		txn.FailureReason,
		txn.IdempotencyKey,
		txn.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrDuplicatePayment
		}
		return err
	}
	return err
}

func (r *BookingRepository) ReleaseHold(ctx context.Context, holdID string, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var holdStatus string
	err = tx.QueryRowContext(
		ctx,
		`SELECT status FROM booking_holds WHERE id = $1 AND user_id = $2 FOR UPDATE`,
		holdID,
		userID,
	).Scan(&holdStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrHoldNotFound
		}
		return err
	}

	if holdStatus == domain.BookingHoldStatusConfirmed {
		return repository.ErrHoldFinalized
	}
	if holdStatus != domain.BookingHoldStatusHeld {
		return repository.ErrHoldAlreadyReleased
	}

	_, err = tx.ExecContext(ctx, `UPDATE booking_holds SET status = 'EXPIRED', updated_at = NOW() WHERE id = $1`, holdID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
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
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BookingRepository) GetPaymentByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.PaymentTransaction, error) {
	query := `
	SELECT id, hold_id, user_id, amount, payment_method, status, COALESCE(gateway_txn_id,''), COALESCE(failure_reason,''), idempotency_key, created_at, updated_at
	FROM payment_transactions
	WHERE idempotency_key = $1
	`
	var txn domain.PaymentTransaction
	err := r.db.QueryRowContext(ctx, query, idempotencyKey).Scan(
		&txn.ID,
		&txn.HoldID,
		&txn.UserID,
		&txn.Amount,
		&txn.PaymentMethod,
		&txn.Status,
		&txn.GatewayTxnID,
		&txn.FailureReason,
		&txn.IdempotencyKey,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // no previous payment — not an error
		}
		return nil, err
	}
	return &txn, nil
}

func (r *BookingRepository) UpdatePaymentStatus(ctx context.Context, txnID string, status string, gatewayTxnID string, failureReason string) error {
	query := `
	UPDATE payment_transactions
	SET status = $2, gateway_txn_id = $3, failure_reason = $4, updated_at = NOW()
	WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, txnID, status, gatewayTxnID, failureReason)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrPaymentNotFound
	}
	return nil
}

func (r *BookingRepository) GetHoldDetails(ctx context.Context, holdID string, userID string) (domain.BookingHold, error) {
	query := `
	SELECT id, user_id, showtime_id, status, hold_expires_at, total_amount, created_at
	FROM booking_holds
	WHERE id = $1 AND user_id = $2
	`
	var hold domain.BookingHold
	err := r.db.QueryRowContext(ctx, query, holdID, userID).Scan(
		&hold.HoldID,
		&hold.UserID,
		&hold.ShowtimeID,
		&hold.Status,
		&hold.HoldExpiresAt,
		&hold.TotalAmount,
		&hold.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BookingHold{}, repository.ErrHoldNotFound
		}
		return domain.BookingHold{}, err
	}

	seatRows, err := r.db.QueryContext(ctx,
		`SELECT seat_number FROM booking_hold_seats WHERE hold_id = $1 ORDER BY seat_number ASC`,
		holdID,
	)
	if err != nil {
		return domain.BookingHold{}, err
	}
	defer seatRows.Close()

	seats := make([]string, 0)
	for seatRows.Next() {
		var seat string
		if scanErr := seatRows.Scan(&seat); scanErr != nil {
			return domain.BookingHold{}, scanErr
		}
		seats = append(seats, seat)
	}
	if err := seatRows.Err(); err != nil {
		return domain.BookingHold{}, err
	}

	hold.SeatNumbers = seats
	return hold, nil
}

var _ repository.BookingRepository = (*BookingRepository)(nil)
