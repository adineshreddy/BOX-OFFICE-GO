package config

import "testing"

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("AUTH_TOKEN_TTL_MINUTES", "")

	cfg := Load()
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("unexpected database url: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret == "" {
		t.Fatal("expected default JWT secret")
	}
	if cfg.AuthTokenTTLMinutes != 1440 {
		t.Fatalf("expected default auth token ttl 1440, got %d", cfg.AuthTokenTTLMinutes)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://y")
	t.Setenv("JWT_SECRET", "custom-secret")
	t.Setenv("AUTH_TOKEN_TTL_MINUTES", "45")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Fatalf("expected custom port 9090, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "custom-secret" {
		t.Fatalf("expected custom jwt secret, got %s", cfg.JWTSecret)
	}
	if cfg.AuthTokenTTLMinutes != 45 {
		t.Fatalf("expected ttl 45, got %d", cfg.AuthTokenTTLMinutes)
	}
}

func TestLoad_EmptyDatabaseURLAllowed(t *testing.T) {
	t.Setenv("PORT", "8081")
	t.Setenv("DATABASE_URL", "")

	cfg := Load()
	if cfg.DatabaseURL != "" {
		t.Fatalf("expected empty database url, got %s", cfg.DatabaseURL)
	}
}
