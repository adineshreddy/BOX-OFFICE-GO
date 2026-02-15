package router

import (
	"net/http"

	"box-office-go/backend/internal/handler"
	"box-office-go/backend/internal/http/response"
	"box-office-go/backend/internal/service"
)

func New(authService *service.AuthService) http.Handler {
	mux := http.NewServeMux()
	authHandler := handler.NewAuthHandler(authService)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/auth/signup", authHandler.Signup)

	return mux
}
