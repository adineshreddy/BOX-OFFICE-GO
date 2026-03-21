package main

import (
	"database/sql"
	"log"
	"net/http"

	"box-office-go/backend/internal/config"
	"box-office-go/backend/internal/database"
	"box-office-go/backend/internal/http/middleware"
	httpRouter "box-office-go/backend/internal/http/router"
	"box-office-go/backend/internal/repository/postgres"
	"box-office-go/backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer func(db *sql.DB) {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("database close error: %v", closeErr)
		}
	}(db)

	userRepository := postgres.NewUserRepository(db)
	movieRepository := postgres.NewMovieRepository(db)

	authService := service.NewAuthService(userRepository)
	movieService := service.NewMovieService(movieRepository)

	api := httpRouter.New(authService, movieService)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORS(api),
	}

	log.Printf("backend server listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server startup failed: %v", err)
	}
}
