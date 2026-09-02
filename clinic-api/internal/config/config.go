package config

import (
	"os"
	"strings"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Config struct {
	HTTPPort string
	LogLevel string
	DB       DBConfig
}

func Load() Config {
	return Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: normalizeLogLevel(getEnv("LOG_LEVEL", "info")),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "clinic"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// IsValidLogLevel reports whether level is a recognized log level
// (case-insensitive). Exposed so callers (e.g. main) can decide whether
// to warn about an invalid LOG_LEVEL without duplicating the valid set.
func IsValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func normalizeLogLevel(level string) string {
	lower := strings.ToLower(level)
	if IsValidLogLevel(lower) {
		return lower
	}
	return "info"
}
