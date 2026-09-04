package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPPort string
	LogLevel string
}

func Load() Config {
	return Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: normalizeLogLevel(getEnv("LOG_LEVEL", "info")),
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
