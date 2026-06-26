package core

import (
	"os"
	"testing"
)

func TestLoggerInitDefault(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	// Clear env to test default
	os.Unsetenv("LOG_LEVEL")

	// Re-initialize logger using helper
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init")
	}
}

func TestLoggerInitDebug(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	os.Setenv("LOG_LEVEL", "debug")

	// Re-initialize logger
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init with DEBUG level")
	}
}

func TestLoggerInitInfo(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	os.Setenv("LOG_LEVEL", "info")

	// Re-initialize logger
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init with INFO level")
	}
}

func TestLoggerInitWarning(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	// Test WARN variant
	os.Setenv("LOG_LEVEL", "WARN")
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init with WARN level")
	}
}

func TestLoggerInitError(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	os.Setenv("LOG_LEVEL", "error")

	// Re-initialize logger
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init with ERROR level")
	}
}

func TestLoggerInitCritical(t *testing.T) {
	// Save original env
	origEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", origEnv)

	os.Setenv("LOG_LEVEL", "critical")

	// Re-initialize logger
	Initialize()

	if Logger == nil {
		t.Fatal("Logger should not be nil after init with CRITICAL level")
	}
}

func TestLoggerLevelFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVal    string
		wantPanic bool
	}{
		{"empty", "", false},
		{"debug", "DEBUG", false},
		{"info", "INFO", false},
		{"warning", "WARNING", false},
		{"warn", "WARN", false},
		{"error", "ERROR", false},
		{"critical", "CRITICAL", false},
		{"invalid", "NOT_A_LEVEL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.envVal)
			Initialize()
			if Logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}