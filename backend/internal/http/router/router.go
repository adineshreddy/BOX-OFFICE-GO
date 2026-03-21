package router

import (
	"net/http"

	"box-office-go/backend/internal/handler"
	"box-office-go/backend/internal/http/response"
	"box-office-go/backend/internal/service"
)

func New(authService *service.AuthService, movieService *service.MovieService) http.Handler {
	mux := http.NewServeMux()
	authHandler := handler.NewAuthHandler(authService)
	movieHandler := handler.NewMovieHandler(movieService)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/v1/movies", movieHandler.ListMovies)
	mux.HandleFunc("GET /api/v1/movies/{movieId}", movieHandler.GetMovie)
	mux.HandleFunc("GET /api/v1/shows/details", movieHandler.GetShowDetailsBySelection)
	mux.HandleFunc("GET /api/v1/shows/seat-map", movieHandler.GetSeatMapBySelection)
	mux.HandleFunc("GET /api/v1/shows/seat-map/availability", movieHandler.RefreshSeatAvailability)
	mux.HandleFunc("GET /api/v1/movies/{movieId}/theaters", movieHandler.ListTheatersByMovie)

	return mux
}
