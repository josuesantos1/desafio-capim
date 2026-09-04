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
