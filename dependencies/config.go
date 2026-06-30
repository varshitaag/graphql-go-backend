package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the application.
// Values are loaded from environment variables with sensible defaults,
// so the app works out of the box with no setup, but can be overridden
// in staging/production via environment variables or a .env file.
type Config struct {
	// Server
	Port string // PORT — default "8080"
	Env  string // APP_ENV — "development" | "production"

	// GraphQL
	PlaygroundEnabled bool   // GQL_PLAYGROUND — enable GraphiQL browser UI
	MaxQueryDepth     int    // GQL_MAX_DEPTH  — prevent deeply nested queries
	MaxQueryComplexity int   // GQL_MAX_COMPLEXITY — prevent expensive queries

	// CORS
	AllowedOrigins string // CORS_ORIGINS — comma-separated list, default "*"
}

// Load reads the environment and returns a populated Config.
func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("APP_ENV", "development"),
		PlaygroundEnabled:  getEnvBool("GQL_PLAYGROUND", true),
		MaxQueryDepth:      getEnvInt("GQL_MAX_DEPTH", 10),
		MaxQueryComplexity: getEnvInt("GQL_MAX_COMPLEXITY", 100),
		AllowedOrigins:     getEnv("CORS_ORIGINS", "*"),
	}
}

// IsDevelopment returns true when running locally.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// ── helpers ──────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}