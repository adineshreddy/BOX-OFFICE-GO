package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	AuthTokenTTLMinutes int
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	authTokenTTLMinutes := 1440
	if raw := strings.TrimSpace(os.Getenv("AUTH_TOKEN_TTL_MINUTES")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			authTokenTTLMinutes = parsed
		}
	}

	return Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		AuthTokenTTLMinutes: authTokenTTLMinutes,
	}
}
