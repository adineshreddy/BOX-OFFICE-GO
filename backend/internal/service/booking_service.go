package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"

	"github.com/jung-kurt/gofpdf/v2"
)

const defaultHoldDuration = 7 * time.Minute

type BookingService struct {
	bookingRepository repository.BookingRepository
}

func NewBookingService(bookingRepository repository.BookingRepository) *BookingService {
	return &BookingService{bookingRepository: bookingRepository}
}

func (s *BookingService) ReleaseExpiredHolds(ctx context.Context) error {
	if err := s.bookingRepository.CleanupExpiredHolds(ctx); err != nil {
		return fmt.Errorf("cleanup expired holds: %w", err)
	}

	return nil
}

func (s *BookingService) CreateBookingHold(ctx context.Context, input domain.CreateBookingHoldInput) (domain.BookingHold, error) {
	normalizedInput, err := normalizeHoldInput(input)
	if err != nil {
		return domain.BookingHold{}, err
	}

	if err := s.bookingRepository.CleanupExpiredHolds(ctx); err != nil {
		return domain.BookingHold{}, fmt.Errorf("cleanup expired holds: %w", err)
	}

	holdID := fmt.Sprintf("hold_%d", time.Now().UnixNano())
	holdExpiresAt := time.Now().UTC().Add(defaultHoldDuration)

	hold, err := s.bookingRepository.CreateHold(ctx, normalizedInput, holdID, holdExpiresAt)
	if err != nil {
		return domain.BookingHold{}, err
	}

	return hold, nil
}

func (s *BookingService) CheckoutBookingHold(ctx context.Context, input domain.ConfirmBookingInput) (domain.BookingCheckoutResult, error) {
	holdID := strings.TrimSpace(input.HoldID)
	if holdID == "" {
		return domain.BookingCheckoutResult{}, fmt.Errorf("holdId is required")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return domain.BookingCheckoutResult{}, fmt.Errorf("userId is required")
	}

	if err := s.bookingRepository.CleanupExpiredHolds(ctx); err != nil {
		return domain.BookingCheckoutResult{}, fmt.Errorf("cleanup expired holds: %w", err)
	}

	bookingID := fmt.Sprintf("bok_%d", time.Now().UnixNano())
	return s.bookingRepository.CheckoutHold(ctx, holdID, userID, bookingID)
}

func normalizeHoldInput(input domain.CreateBookingHoldInput) (domain.CreateBookingHoldInput, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return domain.CreateBookingHoldInput{}, fmt.Errorf("userId is required")
	}

	showtimeID := strings.TrimSpace(input.ShowtimeID)
	if showtimeID == "" {
		return domain.CreateBookingHoldInput{}, fmt.Errorf("showtimeId is required")
	}

	if len(input.SeatNumbers) == 0 {
		return domain.CreateBookingHoldInput{}, fmt.Errorf("seatNumbers must contain at least one seat")
	}

	seen := make(map[string]struct{})
	seatNumbers := make([]string, 0, len(input.SeatNumbers))
	for _, seatNumber := range input.SeatNumbers {
		normalizedSeat := strings.ToUpper(strings.TrimSpace(seatNumber))
		if normalizedSeat == "" {
			return domain.CreateBookingHoldInput{}, fmt.Errorf("seatNumbers cannot contain empty values")
		}
		if _, exists := seen[normalizedSeat]; exists {
			return domain.CreateBookingHoldInput{}, fmt.Errorf("duplicate seat number: %s", normalizedSeat)
		}
		seen[normalizedSeat] = struct{}{}
		seatNumbers = append(seatNumbers, normalizedSeat)
	}

	return domain.CreateBookingHoldInput{
		UserID:      userID,
		ShowtimeID:  showtimeID,
		SeatNumbers: seatNumbers,
	}, nil
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID string) ([]domain.UserBooking, error) {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return nil, fmt.Errorf("userId is required")
	}

	return s.bookingRepository.ListByUserID(ctx, trimmed)
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID string, userID string) error {
	bid := strings.TrimSpace(bookingID)
	if bid == "" {
		return fmt.Errorf("bookingId is required")
	}

	uid := strings.TrimSpace(userID)
	if uid == "" {
		return fmt.Errorf("userId is required")
	}

	return s.bookingRepository.CancelBooking(ctx, bid, uid)
}

// GetTicketPDF fetches the booking, enforces ownership and confirmed status, then renders a PDF.
func (s *BookingService) GetTicketPDF(ctx context.Context, bookingID string, userID string) ([]byte, string, error) {
	bid := strings.TrimSpace(bookingID)
	if bid == "" {
		return nil, "", fmt.Errorf("bookingId is required")
	}

	uid := strings.TrimSpace(userID)
	if uid == "" {
		return nil, "", fmt.Errorf("userId is required")
	}

	ticket, err := s.bookingRepository.GetBookingForTicket(ctx, bid)
	if err != nil {
		return nil, "", err
	}

	if ticket.UserID != uid {
		return nil, "", repository.ErrBookingNotOwned
	}

	if ticket.Status != "CONFIRMED" {
		return nil, "", fmt.Errorf("ticket download is only available for confirmed bookings (current status: %s)", ticket.Status)
	}

	pdfBytes, err := renderTicketPDF(ticket)
	if err != nil {
		return nil, "", fmt.Errorf("render ticket pdf: %w", err)
	}

	filename := fmt.Sprintf("ticket_%s.pdf", bid)
	return pdfBytes, filename, nil
}

// renderTicketPDF builds a single-page PDF ticket with booking essentials.
func renderTicketPDF(t domain.TicketData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// --- Header band ---
	pdf.SetFillColor(30, 27, 46) // dark purple
	pdf.Rect(0, 0, 210, 40, "F")
	pdf.SetTextColor(167, 139, 250) // light purple
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetY(10)
	pdf.CellFormat(180, 12, "BOX OFFICE GO", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(200, 200, 200)
	pdf.CellFormat(180, 8, "E-Ticket", "", 1, "C", false, 0, "")

	// Reset text color
	pdf.SetTextColor(40, 40, 40)

	// --- Booking ID ---
	pdf.SetY(48)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(40, 8, "Booking ID:", "", 0, "", false, 0, "")
	pdf.SetFont("Courier", "", 11)
	pdf.CellFormat(0, 8, t.BookingID, "", 1, "", false, 0, "")

	// --- Divider ---
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY()+2, 195, pdf.GetY()+2)
	pdf.SetY(pdf.GetY() + 6)

	// --- Movie ---
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, t.MovieTitle, "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 6, fmt.Sprintf("%s  |  %s", t.Language, t.Format), "", 1, "", false, 0, "")
	pdf.SetTextColor(40, 40, 40)
	pdf.Ln(4)

	// --- Details grid ---
	labelW := 40.0
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelW, 7, "Theater:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, fmt.Sprintf("%s, %s", t.TheaterName, t.City), "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelW, 7, "Screen:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, t.ScreenName, "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelW, 7, "Show Time:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, t.ShowTime.Format("Mon, 02 Jan 2006 at 03:04 PM"), "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelW, 7, "Seats:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, strings.Join(t.SeatNumbers, ", "), "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(labelW, 7, "Total:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 7, fmt.Sprintf("$%.2f", t.TotalAmount), "", 1, "", false, 0, "")

	pdf.Ln(4)

	// --- Confirmed at ---
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 6, fmt.Sprintf("Confirmed at: %s", t.ConfirmedAt.Format(time.RFC3339)), "", 1, "", false, 0, "")

	// --- Footer ---
	pdf.SetY(270)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(160, 160, 160)
	pdf.CellFormat(0, 5, "Present this ticket at the theater entrance. Enjoy the movie!", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
