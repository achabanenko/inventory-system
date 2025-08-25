package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiry       time.Duration
	RefreshExpiry   time.Duration
	CORSOrigins     []string
	Environment     string
	LogLevel        string
	MaxPageSize     int
	DefaultPageSize int
	// Google OAuth Configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		MaxPageSize:     getEnvAsInt("MAX_PAGE_SIZE", 100),
		DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 20),
		// Google OAuth Configuration
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:5173/auth/google/callback"),
	}

	jwtExpiry := getEnvAsInt("JWT_EXPIRY_MINUTES", 15)
	cfg.JWTExpiry = time.Duration(jwtExpiry) * time.Minute

	refreshExpiry := getEnvAsInt("REFRESH_EXPIRY_DAYS", 7)
	cfg.RefreshExpiry = time.Duration(refreshExpiry) * 24 * time.Hour

	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000,http://localhost:3001")
	if corsOrigins != "" {
		// Split comma-separated origins
		origins := make([]string, 0)
		for _, origin := range strings.Split(corsOrigins, ",") {
			origins = append(origins, strings.TrimSpace(origin))
		}
		cfg.CORSOrigins = origins
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
