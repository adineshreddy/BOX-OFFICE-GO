package handler

import (
	"net/http"

	"box-office-go/backend/internal/http/response"
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

	query := r.URL.Query()
	titleQuery := query.Get("title")
	genreQuery := query.Get("genre")

	movies, err := h.movieService.ListMovies(r.Context(), titleQuery, genreQuery)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch movies", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"movies": movies,
	})
}
