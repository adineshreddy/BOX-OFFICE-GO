package config

import "testing"

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "postgres://x")

	cfg := Load()
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("unexpected database url: %s", cfg.DatabaseURL)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://y")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Fatalf("expected custom port 9090, got %s", cfg.Port)
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
