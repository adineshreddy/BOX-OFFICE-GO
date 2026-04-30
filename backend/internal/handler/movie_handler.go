package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/http/response"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type MovieHandler struct {
	movieService *service.MovieService
}

func NewMovieHandler(movieService *service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

func (h *MovieHandler) ListMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	q := r.URL.Query()

	// ── page ──────────────────────────────────────────────────────────
	page := 1
	if raw := strings.TrimSpace(q.Get("page")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			response.Error(w, http.StatusBadRequest, "page must be a positive integer", nil)
			return
		}
		page = parsed
	}

	// ── limit ─────────────────────────────────────────────────────────
	limit := 20
	if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			response.Error(w, http.StatusBadRequest, "limit must be a positive integer", nil)
			return
		}
		if parsed > 100 {
			response.Error(w, http.StatusBadRequest, "limit must not exceed 100", nil)
			return
		}
		limit = parsed
	}

	// ── sort ──────────────────────────────────────────────────────────
	sortBy := domain.MovieSortByReleaseDate
	if raw := strings.TrimSpace(q.Get("sort")); raw != "" {
		if !domain.ValidMovieSortFields[raw] {
			response.Error(w, http.StatusBadRequest, "sort must be one of: title, release_date, rating", nil)
			return
		}
		sortBy = raw
	}

	movieQuery := domain.MovieListQuery{
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Title:  strings.TrimSpace(q.Get("title")),
		Genre:  strings.TrimSpace(q.Get("genre")),
	}

	result, err := h.movieService.ListMovies(r.Context(), movieQuery)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch movies", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"movies":  result.Movies,
		"page":    result.Page,
		"limit":   result.Limit,
		"total":   result.Total,
		"hasNext": result.HasNext,
	})
}

func (h *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	movieID := strings.TrimSpace(r.PathValue("movieId"))
	if movieID == "" {
		response.Error(w, http.StatusBadRequest, "movieId path parameter is required", nil)
		return
	}

	movie, err := h.movieService.GetMovieByID(r.Context(), movieID)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			response.Error(w, http.StatusNotFound, "movie not found", nil)
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to fetch movie", nil)
		return
	}

	response.JSON(w, http.StatusOK, movie)
}

func (h *MovieHandler) GetShowDetailsBySelection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	query := r.URL.Query()
	movieID := strings.TrimSpace(query.Get("movieId"))
	theaterID := strings.TrimSpace(query.Get("theaterId"))
	showTime := strings.TrimSpace(query.Get("showTime"))

	details, err := h.movieService.GetShowDetailsBySelection(r.Context(), movieID, theaterID, showTime)
	if err != nil {
		handleShowSelectionError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, details)
}

func (h *MovieHandler) GetSeatMapBySelection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	query := r.URL.Query()
	movieID := strings.TrimSpace(query.Get("movieId"))
	theaterID := strings.TrimSpace(query.Get("theaterId"))
	showTime := strings.TrimSpace(query.Get("showTime"))

	seatMap, err := h.movieService.GetSeatMapBySelection(r.Context(), movieID, theaterID, showTime)
	if err != nil {
		handleShowSelectionError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, seatMap)
}

func (h *MovieHandler) RefreshSeatAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	query := r.URL.Query()
	movieID := strings.TrimSpace(query.Get("movieId"))
	theaterID := strings.TrimSpace(query.Get("theaterId"))
	showTime := strings.TrimSpace(query.Get("showTime"))

	availability, err := h.movieService.GetSeatAvailabilityBySelection(r.Context(), movieID, theaterID, showTime)
	if err != nil {
		handleShowSelectionError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, availability)
}

func (h *MovieHandler) ListTheatersByMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	movieID := strings.TrimSpace(r.PathValue("movieId"))
	if movieID == "" {
		response.Error(w, http.StatusBadRequest, "movieId path parameter is required", nil)
		return
	}

	dateQuery := strings.TrimSpace(r.URL.Query().Get("date"))
	theaterList, err := h.movieService.ListTheatersByMovie(r.Context(), movieID, dateQuery)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			response.Error(w, http.StatusNotFound, "movie not found", nil)
			return
		}

		errMessage := err.Error()
		if strings.Contains(errMessage, "date must be in YYYY-MM-DD format") ||
			strings.Contains(errMessage, "movie id is required") ||
			strings.Contains(errMessage, "date cannot be in the past") ||
			strings.Contains(errMessage, "bookings are only available within the next 14 days") {
			response.Error(w, http.StatusBadRequest, errMessage, nil)
			return
		}

		if strings.Contains(errMessage, "invalid showtime configuration") {
			response.Error(w, http.StatusConflict, errMessage, nil)
			return
		}

		response.Error(w, http.StatusInternalServerError, "failed to fetch theaters for movie", nil)
		return
	}

	response.JSON(w, http.StatusOK, theaterList)
}

func handleShowSelectionError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrShowtimeNotFound) {
		response.Error(w, http.StatusNotFound, "showtime not found for provided movieId, theaterId, and showTime", nil)
		return
	}

	errMessage := err.Error()
	if strings.Contains(errMessage, "query parameter is required") || strings.Contains(errMessage, "showTime must be in RFC3339 format") {
		response.Error(w, http.StatusBadRequest, errMessage, nil)
		return
	}

	response.Error(w, http.StatusInternalServerError, "failed to fetch show information", nil)
}
