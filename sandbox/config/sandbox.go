package sandbox

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// --- Errors (from sandbox/model/exception.py) ---

// SandboxError is the base error for all sandbox-related issues.
type SandboxError struct {
	Message string
}

func (e SandboxError) Error() string { return e.Message }

// Custom error types for specific failure scenarios.
type SandboxTimeoutError struct{ SandboxError }
type SandboxResourceError struct{ SandboxError }
type SandboxUninitializedError struct{ SandboxError }

// --- Configuration (from sandbox/config/config.py) ---

// SandboxSettings defines the configuration for the execution sandbox.
type SandboxSettings struct {
	Image          string        // Base image name
	SystemShell    string        // System shell to use
	WorkDir        string        // Container working directory
	MemoryLimit    string        // Memory limit (e.g., "4096mb")
	CPULimit       float64       // CPU limit (units of CPU)
	Timeout        time.Duration // Default command timeout
	NetworkEnabled bool          // Whether network access is allowed
	NetworkName    string        // Name of the Docker network
}

// NewSettings initializes SandboxSettings with default values and environment overrides.
// It uses "water-ai" as the default project name if COMPOSE_PROJECT_NAME is not set.
func NewSettings() *SandboxSettings {
	projectName := getEnv("COMPOSE_PROJECT_NAME", "water-ai")

	return &SandboxSettings{
		Image:          fmt.Sprintf("%s-sandbox", projectName),
		SystemShell:    getEnv("SANDBOX_SHELL", "system_shell"),
		WorkDir:        getEnv("SANDBOX_WORK_DIR", "/workspace"),
		MemoryLimit:    getEnv("SANDBOX_MEMORY_LIMIT", "4096mb"),
		CPULimit:       getEnvFloat("SANDBOX_CPU_LIMIT", 1.0),
		Timeout:        time.Duration(getEnvInt("SANDBOX_TIMEOUT", 600)) * time.Second,
		NetworkEnabled: getEnvBool("SANDBOX_NETWORK_ENABLED", true),
		NetworkName:    fmt.Sprintf("%s_water_ai", projectName), // Replaced _ii with _water_ai
	}
}

// --- Internal Helpers for Environment Parsing ---

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}