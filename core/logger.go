package core

import (
	"log/slog"
	"os"
	"strings"
)

// Logger is the global logger instance for water-ai.
var Logger *slog.Logger

func init() {
	Initialize()
}

// Initialize sets up the logger based on environment variables.
// This is exported or kept package-private to allow manual re-init in tests.
func Initialize() {
	// Determine log level from environment variable
	logLevelStr := os.Getenv("LOG_LEVEL")
	var level slog.Level

	switch strings.ToUpper(logLevelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARNING", "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	case "CRITICAL":
		// slog doesn't have a specific critical level, mapping to Error+4
		level = slog.LevelError + 4
	default:
		// Default to INFO if not set or unrecognized
		level = slog.LevelInfo
	}

	// Configure the handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create a TextHandler
	handler := slog.NewTextHandler(os.Stderr, opts)

	// Initialize the logger with the service name "water_ai"
	Logger = slog.New(handler).With("service", "water_ai")

	// Set as the default global logger for the application
	slog.SetDefault(Logger)
}