// Package config reads environment variables (12-factor). godotenv.Load() loads .env for local dev only.
// Call Load() once at startup; fail fast if DATABASE_URL or JWT_SECRET are missing/weak.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from the environment (12-factor style).
// Secrets must never be committed; use .env locally and a secret manager in production.
type Config struct {
	Env            string
	HTTPPort       string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiry      time.Duration
	UploadDir      string
	MaxUploadBytes int64
	RateLimitRPS   float64
	RateBurst      int
	// GinMode is "debug", "release", or "test" (see gin.SetMode).
	GinMode string
}

// Load parses env into Config. Required: DATABASE_URL, JWT_SECRET (≥32 runes). Optional: ports, limits, APP_ENV.
func Load() (*Config, error) {
	_ = godotenv.Load() // .env is optional; production uses real env vars

	jwtExpMin, _ := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "60"))
	maxMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "10"), 10, 64)
	if maxMB <= 0 {
		maxMB = 10
	}
	rps, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "20"), 64)
	if rps <= 0 {
		rps = 20
	}
	burst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "40"))
	if burst <= 0 {
		burst = 40
	}

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	env := getEnv("APP_ENV", "development")
	ginMode := "release"
	if env == "development" {
		ginMode = "debug"
	}

	return &Config{
		Env:            env,
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		DatabaseURL:    dbURL,
		JWTSecret:      secret,
		JWTExpiry:      time.Duration(jwtExpMin) * time.Minute,
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadBytes: maxMB * 1024 * 1024,
		RateLimitRPS:   rps,
		RateBurst:      burst,
		GinMode:        ginMode,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
