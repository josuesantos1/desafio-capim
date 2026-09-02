package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg := Load()

	assertEqual(t, "HTTPPort", cfg.HTTPPort, "8080")
	assertEqual(t, "LogLevel", cfg.LogLevel, "info")
	assertEqual(t, "DB.Host", cfg.DB.Host, "localhost")
	assertEqual(t, "DB.Port", cfg.DB.Port, "5432")
	assertEqual(t, "DB.User", cfg.DB.User, "postgres")
	assertEqual(t, "DB.Password", cfg.DB.Password, "")
	assertEqual(t, "DB.Name", cfg.DB.Name, "clinic")
	assertEqual(t, "DB.SSLMode", cfg.DB.SSLMode, "disable")
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "clinic_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "clinic_prod")
	t.Setenv("DB_SSLMODE", "require")

	cfg := Load()

	assertEqual(t, "HTTPPort", cfg.HTTPPort, "9090")
	assertEqual(t, "LogLevel", cfg.LogLevel, "debug")
	assertEqual(t, "DB.Host", cfg.DB.Host, "db.internal")
	assertEqual(t, "DB.Port", cfg.DB.Port, "6543")
	assertEqual(t, "DB.User", cfg.DB.User, "clinic_user")
	assertEqual(t, "DB.Password", cfg.DB.Password, "secret")
	assertEqual(t, "DB.Name", cfg.DB.Name, "clinic_prod")
	assertEqual(t, "DB.SSLMode", cfg.DB.SSLMode, "require")
}

func TestLoad_InvalidLogLevelFallsBackToInfo(t *testing.T) {
	clearEnv(t)
	t.Setenv("LOG_LEVEL", "verbose")

	cfg := Load()

	assertEqual(t, "LogLevel", cfg.LogLevel, "info")
}

func TestIsValidLogLevel(t *testing.T) {
	valid := []string{"debug", "info", "warn", "error", "DEBUG", "Info", "WARN", "Error"}
	for _, v := range valid {
		if !IsValidLogLevel(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"verbose", "trace", "", "informational"}
	for _, v := range invalid {
		if IsValidLogLevel(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"HTTP_PORT", "LOG_LEVEL",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
	}
	for _, k := range keys {
		original, hadOriginal := os.LookupEnv(k)
		os.Unsetenv(k)
		t.Cleanup(func() {
			if hadOriginal {
				os.Setenv(k, original)
			}
		})
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}
