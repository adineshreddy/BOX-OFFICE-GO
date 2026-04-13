package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/middleware"
	"box-office-go/backend/internal/http/response"
	"box-office-go/backend/internal/payment"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) CreateBookingHold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var input domain.CreateBookingHoldInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	input.UserID = identity.UserID

	hold, err := h.bookingService.CreateBookingHold(r.Context(), input)
	if err != nil {
		handleBookingError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"message": "booking hold created",
		"hold":    hold,
	})
}

func (h *BookingHandler) CheckoutBookingHold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var input domain.ConfirmBookingInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	input.UserID = identity.UserID

	result, err := h.bookingService.CheckoutBookingHold(r.Context(), input)
	if err != nil {
		handleBookingError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "checkout successful",
		"booking": result,
	})
}

func handleBookingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrSeatUnavailable):
		response.Error(w, http.StatusConflict, "one or more selected seats are unavailable", nil)
		return
	case errors.Is(err, repository.ErrInvalidSeatSelection):
		response.Error(w, http.StatusBadRequest, "invalid seat selection", nil)
		return
	case errors.Is(err, repository.ErrShowtimeNotFound):
		response.Error(w, http.StatusNotFound, "showtime not found", nil)
		return
	case errors.Is(err, repository.ErrHoldNotFound):
		response.Error(w, http.StatusNotFound, "booking hold not found", nil)
		return
	case errors.Is(err, repository.ErrHoldExpired):
		response.Error(w, http.StatusConflict, "booking hold expired", nil)
		return
	case errors.Is(err, repository.ErrHoldFinalized):
		response.Error(w, http.StatusConflict, "booking hold already finalized", nil)
		return
	case errors.Is(err, payment.ErrPaymentDeclined):
		response.Error(w, http.StatusPaymentRequired, "payment declined", nil)
		return
	case errors.Is(err, payment.ErrGatewayTimeout):
		response.Error(w, http.StatusBadGateway, "payment gateway unavailable, please retry", nil)
		return
	}

	errMessage := err.Error()
	if strings.Contains(errMessage, "is required") || strings.Contains(errMessage, "seatNumbers") || strings.Contains(errMessage, "duplicate seat number") {
		response.Error(w, http.StatusBadRequest, errMessage, nil)
		return
	}

	response.Error(w, http.StatusInternalServerError, "booking operation failed", nil)
}

// GET /api/v1/bookings
func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	bookings, err := h.bookingService.GetUserBookings(r.Context(), identity.UserID)
	if err != nil {
		handleBookingError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"bookings": bookings,
	})
}

// DELETE /api/v1/bookings?bookingId=<bookingId>
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	bookingID := strings.TrimSpace(r.URL.Query().Get("bookingId"))
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "bookingId query parameter is required", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	err := h.bookingService.CancelBooking(r.Context(), bookingID, identity.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			response.Error(w, http.StatusNotFound, "booking not found", nil)
			return
		}
		if errors.Is(err, repository.ErrBookingAlreadyCancelled) {
			response.Error(w, http.StatusConflict, "booking is already cancelled", nil)
			return
		}
		handleBookingError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "booking cancelled successfully",
	})
}

// GET /api/v1/bookings/{bookingId}/ticket
func (h *BookingHandler) DownloadTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	bookingID := strings.TrimSpace(r.PathValue("bookingId"))
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "bookingId path parameter is required", nil)
		return
	}

	identity, ok := middleware.AuthIdentityFromContext(r.Context())
	if !ok || identity.UserID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	pdfBytes, filename, err := h.bookingService.GetTicketPDF(r.Context(), bookingID, identity.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			response.Error(w, http.StatusNotFound, "booking not found", nil)
			return
		}
		if errors.Is(err, repository.ErrBookingNotOwned) {
			response.Error(w, http.StatusForbidden, "you do not have access to this booking", nil)
			return
		}
		errMsg := err.Error()
		if strings.Contains(errMsg, "only available for confirmed bookings") {
			response.Error(w, http.StatusConflict, errMsg, nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to generate ticket", nil)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
