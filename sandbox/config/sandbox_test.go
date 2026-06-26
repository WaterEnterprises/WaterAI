package sandbox

import (
	"os"
	"testing"
	"time"
)

func TestSandboxError(t *testing.T) {
	err := SandboxError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Error() = %s; want test error", err.Error())
	}
}

func TestSandboxTimeoutError(t *testing.T) {
	err := SandboxTimeoutError{SandboxError{Message: "timeout"}}
	if err.Error() != "timeout" {
		t.Errorf("Error() = %s; want timeout", err.Error())
	}
}

func TestSandboxResourceError(t *testing.T) {
	err := SandboxResourceError{SandboxError{Message: "resource error"}}
	if err.Error() != "resource error" {
		t.Errorf("Error() = %s; want resource error", err.Error())
	}
}

func TestSandboxUninitializedError(t *testing.T) {
	err := SandboxUninitializedError{SandboxError{Message: "uninitialized"}}
	if err.Error() != "uninitialized" {
		t.Errorf("Error() = %s; want uninitialized", err.Error())
	}
}

func TestNewSettings(t *testing.T) {
	settings := NewSettings()

	if settings == nil {
		t.Fatal("NewSettings() returned nil")
	}

	if settings.WorkDir != "/workspace" {
		t.Errorf("WorkDir = %s; want /workspace", settings.WorkDir)
	}

	if settings.MemoryLimit != "4096mb" {
		t.Errorf("MemoryLimit = %s; want 4096mb", settings.MemoryLimit)
	}

	if settings.CPULimit != 1.0 {
		t.Errorf("CPULimit = %f; want 1.0", settings.CPULimit)
	}

	if settings.Timeout != 600*time.Second {
		t.Errorf("Timeout = %v; want 600s", settings.Timeout)
	}

	if !settings.NetworkEnabled {
		t.Error("NetworkEnabled should be true")
	}
}

func TestNewSettingsWithEnvVars(t *testing.T) {
	// Save original env
	origWorkDir := os.Getenv("SANDBOX_WORK_DIR")
	origMemoryLimit := os.Getenv("SANDBOX_MEMORY_LIMIT")
	defer func() {
		os.Setenv("SANDBOX_WORK_DIR", origWorkDir)
		os.Setenv("SANDBOX_MEMORY_LIMIT", origMemoryLimit)
	}()

	os.Setenv("SANDBOX_WORK_DIR", "/custom/workspace")
	os.Setenv("SANDBOX_MEMORY_LIMIT", "8192mb")

	settings := NewSettings()

	if settings.WorkDir != "/custom/workspace" {
		t.Errorf("WorkDir = %s; want /custom/workspace", settings.WorkDir)
	}

	if settings.MemoryLimit != "8192mb" {
		t.Errorf("MemoryLimit = %s; want 8192mb", settings.MemoryLimit)
	}
}

func TestNewSettingsWithProjectName(t *testing.T) {
	origProject := os.Getenv("COMPOSE_PROJECT_NAME")
	defer os.Setenv("COMPOSE_PROJECT_NAME", origProject)

	os.Setenv("COMPOSE_PROJECT_NAME", "my-project")

	settings := NewSettings()

	if settings.Image != "my-project-sandbox" {
		t.Errorf("Image = %s; want my-project-sandbox", settings.Image)
	}

	if settings.NetworkName != "my-project_water_ai" {
		t.Errorf("NetworkName = %s; want my-project_water_ai", settings.NetworkName)
	}
}

func TestSandboxSettingsFields(t *testing.T) {
	settings := &SandboxSettings{
		Image:          "custom-image",
		SystemShell:    "/bin/zsh",
		WorkDir:        "/app",
		MemoryLimit:    "2048mb",
		CPULimit:       2.0,
		Timeout:        300 * time.Second,
		NetworkEnabled: false,
		NetworkName:    "custom-network",
	}

	if settings.Image != "custom-image" {
		t.Errorf("Image = %s; want custom-image", settings.Image)
	}

	if settings.SystemShell != "/bin/zsh" {
		t.Errorf("SystemShell = %s; want /bin/zsh", settings.SystemShell)
	}

	if settings.WorkDir != "/app" {
		t.Errorf("WorkDir = %s; want /app", settings.WorkDir)
	}

	if settings.MemoryLimit != "2048mb" {
		t.Errorf("MemoryLimit = %s; want 2048mb", settings.MemoryLimit)
	}

	if settings.CPULimit != 2.0 {
		t.Errorf("CPULimit = %f; want 2.0", settings.CPULimit)
	}

	if settings.Timeout != 300*time.Second {
		t.Errorf("Timeout = %v; want 300s", settings.Timeout)
	}

	if settings.NetworkEnabled {
		t.Error("NetworkEnabled should be false")
	}

	if settings.NetworkName != "custom-network" {
		t.Errorf("NetworkName = %s; want custom-network", settings.NetworkName)
	}
}

func TestGetEnv(t *testing.T) {
	origVal := os.Getenv("TEST_GET_ENV")
	defer os.Setenv("TEST_GET_ENV", origVal)

	// Test with env set
	os.Setenv("TEST_GET_ENV", "value")
	result := getEnv("TEST_GET_ENV", "default")
	if result != "value" {
		t.Errorf("getEnv() = %s; want value", result)
	}

	// Test with env not set
	os.Unsetenv("TEST_GET_ENV")
	result = getEnv("TEST_GET_ENV", "default")
	if result != "default" {
		t.Errorf("getEnv() = %s; want default", result)
	}
}

func TestGetEnvFloat(t *testing.T) {
	origVal := os.Getenv("TEST_GET_ENV_FLOAT")
	defer os.Setenv("TEST_GET_ENV_FLOAT", origVal)

	// Test with valid float
	os.Setenv("TEST_GET_ENV_FLOAT", "3.14")
	result := getEnvFloat("TEST_GET_ENV_FLOAT", 1.0)
	if result != 3.14 {
		t.Errorf("getEnvFloat() = %f; want 3.14", result)
	}

	// Test with invalid float
	os.Setenv("TEST_GET_ENV_FLOAT", "invalid")
	result = getEnvFloat("TEST_GET_ENV_FLOAT", 1.0)
	if result != 1.0 {
		t.Errorf("getEnvFloat() = %f; want 1.0 (fallback)", result)
	}

	// Test with env not set
	os.Unsetenv("TEST_GET_ENV_FLOAT")
	result = getEnvFloat("TEST_GET_ENV_FLOAT", 2.5)
	if result != 2.5 {
		t.Errorf("getEnvFloat() = %f; want 2.5", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	origVal := os.Getenv("TEST_GET_ENV_INT")
	defer os.Setenv("TEST_GET_ENV_INT", origVal)

	// Test with valid int
	os.Setenv("TEST_GET_ENV_INT", "42")
	result := getEnvInt("TEST_GET_ENV_INT", 10)
	if result != 42 {
		t.Errorf("getEnvInt() = %d; want 42", result)
	}

	// Test with invalid int
	os.Setenv("TEST_GET_ENV_INT", "invalid")
	result = getEnvInt("TEST_GET_ENV_INT", 10)
	if result != 10 {
		t.Errorf("getEnvInt() = %d; want 10 (fallback)", result)
	}

	// Test with env not set
	os.Unsetenv("TEST_GET_ENV_INT")
	result = getEnvInt("TEST_GET_ENV_INT", 20)
	if result != 20 {
		t.Errorf("getEnvInt() = %d; want 20", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	origVal := os.Getenv("TEST_GET_ENV_BOOL")
	defer os.Setenv("TEST_GET_ENV_BOOL", origVal)

	// Test with true
	os.Setenv("TEST_GET_ENV_BOOL", "true")
	result := getEnvBool("TEST_GET_ENV_BOOL", false)
	if !result {
		t.Error("getEnvBool() = false; want true")
	}

	// Test with false
	os.Setenv("TEST_GET_ENV_BOOL", "false")
	result = getEnvBool("TEST_GET_ENV_BOOL", true)
	if result {
		t.Error("getEnvBool() = true; want false")
	}

	// Test with invalid bool
	os.Setenv("TEST_GET_ENV_BOOL", "invalid")
	result = getEnvBool("TEST_GET_ENV_BOOL", true)
	if !result {
		t.Error("getEnvBool() should return fallback for invalid value")
	}

	// Test with env not set
	os.Unsetenv("TEST_GET_ENV_BOOL")
	result = getEnvBool("TEST_GET_ENV_BOOL", false)
	if result {
		t.Error("getEnvBool() = true; want false (fallback)")
	}
}
