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

// UserBooking represents a confirmed booking with movie/theater details for the "Get My Bookings" API.
type UserBooking struct {
	BookingID   string    `json:"bookingId"`
	HoldID      string    `json:"holdId"`
	UserID      string    `json:"userId"`
	ShowtimeID  string    `json:"showtimeId"`
	SeatNumbers []string  `json:"seatNumbers"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"totalAmount"`
	ConfirmedAt time.Time `json:"confirmedAt"`
	MovieTitle  string    `json:"movieTitle"`
	TheaterName string    `json:"theaterName"`
	City        string    `json:"city"`
	ScreenName  string    `json:"screenName"`
	ShowTime    time.Time `json:"showTime"`
	Language    string    `json:"language"`
	Format      string    `json:"format"`
}

// TicketData holds the fully-joined booking information needed to render a PDF ticket.
type TicketData struct {
	BookingID   string
	UserID      string
	MovieTitle  string
	TheaterName string
	City        string
	ScreenName  string
	ShowTime    time.Time
	Language    string
	Format      string
	SeatNumbers []string
	TotalAmount float64
	Status      string
	ConfirmedAt time.Time
}
