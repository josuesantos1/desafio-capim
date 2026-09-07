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
	if cfg.SeedData {
		t.Errorf("SeedData = true, want false")
	}
}

func TestLoad_SeedDataEnabled(t *testing.T) {
	clearEnv(t)
	t.Setenv("SEED_DATA", "true")

	cfg := Load()

	if !cfg.SeedData {
		t.Errorf("SeedData = false, want true")
	}
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "DEBUG")

	cfg := Load()

	assertEqual(t, "HTTPPort", cfg.HTTPPort, "9090")
	assertEqual(t, "LogLevel", cfg.LogLevel, "debug")
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
	keys := []string{"HTTP_PORT", "LOG_LEVEL", "SEED_DATA"}
	for _, k := range keys {
		original, hadOriginal := os.LookupEnv(k)
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("failed to unset %s: %v", k, err)
		}
		t.Cleanup(func() {
			if hadOriginal {
				if err := os.Setenv(k, original); err != nil {
					t.Fatalf("failed to restore %s: %v", k, err)
				}
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
