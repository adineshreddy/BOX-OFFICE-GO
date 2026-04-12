package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

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
	bookingRepository := postgres.NewBookingRepository(db)

	authService := service.NewAuthServiceWithConfig(
		userRepository,
		cfg.JWTSecret,
		time.Duration(cfg.AuthTokenTTLMinutes)*time.Minute,
	)
	movieService := service.NewMovieService(movieRepository)
	bookingService := service.NewBookingService(bookingRepository)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			if err := bookingService.ReleaseExpiredHolds(context.Background()); err != nil {
				log.Printf("expired hold cleanup failed: %v", err)
			}

			<-ticker.C
		}
	}()

	api := httpRouter.New(authService, movieService, bookingService)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: middleware.CORS(api),
	}

	log.Printf("backend server listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server startup failed: %v", err)
	}
}
