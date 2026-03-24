package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
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
