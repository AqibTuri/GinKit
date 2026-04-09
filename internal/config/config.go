// Package config reads environment variables (12-factor). godotenv.Load() loads .env for local dev only.
// Call Load() once at startup; fail fast if DATABASE_URL or JWT_SECRET are missing/weak.
package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gin-api/internal/authcookie"
	"github.com/joho/godotenv"
)

// Config holds runtime configuration loaded from the environment (12-factor style).
// Secrets must never be committed; use .env locally and a secret manager in production.
type Config struct {
	Env            string
	HTTPPort       string
	DatabaseURL    string
	JWTSecret         string
	JWTExpiry         time.Duration // access token lifetime
	JWTRefreshExpiry  time.Duration
	CookieAccessName  string
	CookieRefreshName string
	CookiePath        string
	CookieDomain      string
	CookieSecure      bool
	CookieSameSite    string // lax | strict | none
	UploadDir         string
	MaxUploadBytes int64
	RateLimitRPS   float64
	RateBurst      int
	// GinMode is "debug", "release", or "test" (see gin.SetMode).
	GinMode string
}

// Load parses env into Config. Required: DATABASE_URL, JWT_SECRET (≥32 runes). Optional: ports, limits, APP_ENV.
func Load() (*Config, error) {
	_ = godotenv.Load() // .env is optional; production uses real env vars

	jwtExpMin, _ := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "15"))
	refreshDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))
	if refreshDays <= 0 {
		refreshDays = 7
	}
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

	cookieSecure := env != "development"
	switch strings.ToLower(strings.TrimSpace(getEnv("COOKIE_SECURE", ""))) {
	case "true":
		cookieSecure = true
	case "false":
		cookieSecure = false
	}

	return &Config{
		Env:               env,
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       dbURL,
		JWTSecret:         secret,
		JWTExpiry:         time.Duration(jwtExpMin) * time.Minute,
		JWTRefreshExpiry:  time.Duration(refreshDays) * 24 * time.Hour,
		CookieAccessName:  getEnv("AUTH_ACCESS_COOKIE", "access_token"),
		CookieRefreshName: getEnv("AUTH_REFRESH_COOKIE", "refresh_token"),
		CookiePath:        getEnv("AUTH_COOKIE_PATH", "/"),
		CookieDomain:      getEnv("AUTH_COOKIE_DOMAIN", ""),
		CookieSecure:      cookieSecure,
		CookieSameSite:    getEnv("AUTH_COOKIE_SAMESITE", "lax"),
		UploadDir:         getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadBytes:    maxMB * 1024 * 1024,
		RateLimitRPS:      rps,
		RateBurst:         burst,
		GinMode:           ginMode,
	}, nil
}

// AuthCookieSettings builds HTTP cookie options for access/refresh tokens.
func (c *Config) AuthCookieSettings() authcookie.Settings {
	return authcookie.Settings{
		AccessName:  c.CookieAccessName,
		RefreshName: c.CookieRefreshName,
		Path:        c.CookiePath,
		Domain:      c.CookieDomain,
		Secure:      c.CookieSecure,
		SameSite:    c.cookieSameSiteMode(),
	}
}

func (c *Config) cookieSameSiteMode() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(c.CookieSameSite)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
