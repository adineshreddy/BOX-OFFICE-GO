package domain

import "time"

const (
	BookingHoldStatusHeld      = "HELD"
	BookingHoldStatusExpired   = "EXPIRED"
	BookingHoldStatusConfirmed = "CONFIRMED"
)

type CreateBookingHoldInput struct {
	UserID      string   `json:"userId"`
	ShowtimeID  string   `json:"showtimeId"`
	SeatNumbers []string `json:"seatNumbers"`
}

type BookingHold struct {
	HoldID        string    `json:"holdId"`
	UserID        string    `json:"userId"`
	ShowtimeID    string    `json:"showtimeId"`
	SeatNumbers   []string  `json:"seatNumbers"`
	Status        string    `json:"status"`
	HoldExpiresAt time.Time `json:"holdExpiresAt"`
	TotalAmount   float64   `json:"totalAmount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ConfirmBookingInput struct {
	HoldID string `json:"holdId"`
	UserID string `json:"userId"`
}

type BookingCheckoutResult struct {
	BookingID   string    `json:"bookingId"`
	HoldID      string    `json:"holdId"`
	UserID      string    `json:"userId"`
	ShowtimeID  string    `json:"showtimeId"`
	SeatNumbers []string  `json:"seatNumbers"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"totalAmount"`
	ConfirmedAt time.Time `json:"confirmedAt"`
}
