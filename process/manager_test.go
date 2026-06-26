package process

import (
	"testing"
	"time"
)

func TestManagerConfigDefaults(t *testing.T) {
	cfg := ManagerConfig{
		GatewayPort:  "8080",
		FrontendPort: "8000",
		GatewayPath:  "/usr/local/bin/gateway",
	}

	if cfg.GatewayPort != "8080" {
		t.Errorf("GatewayPort = %s; want 8080", cfg.GatewayPort)
	}

	if cfg.FrontendPort != "8000" {
		t.Errorf("FrontendPort = %s; want 8000", cfg.FrontendPort)
	}
}

func TestManagerConfigWithCustomValues(t *testing.T) {
	cfg := ManagerConfig{
		GatewayPort:         "9090",
		FrontendPort:        "9000",
		GatewayPath:         "/custom/gateway",
		HealthInterval:      10 * time.Second,
		MaxRestartAttempts:  5,
		RestartDelay:        2 * time.Second,
		MaxRestartDelay:     60 * time.Second,
	}

	if cfg.HealthInterval != 10*time.Second {
		t.Errorf("HealthInterval = %v; want 10s", cfg.HealthInterval)
	}

	if cfg.MaxRestartAttempts != 5 {
		t.Errorf("MaxRestartAttempts = %d; want 5", cfg.MaxRestartAttempts)
	}

	if cfg.RestartDelay != 2*time.Second {
		t.Errorf("RestartDelay = %v; want 2s", cfg.RestartDelay)
	}
}

func TestNewManagerSetsDefaults(t *testing.T) {
	cfg := ManagerConfig{
		GatewayPort:  "8080",
		FrontendPort: "8000",
		GatewayPath:  "",
	}

	manager := NewManager(cfg)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	// Check defaults are set
	if manager.config.HealthInterval != 5*time.Second {
		t.Errorf("Default HealthInterval = %v; want 5s", manager.config.HealthInterval)
	}

	if manager.config.MaxRestartAttempts != 10 {
		t.Errorf("Default MaxRestartAttempts = %d; want 10", manager.config.MaxRestartAttempts)
	}

	if manager.config.RestartDelay != 1*time.Second {
		t.Errorf("Default RestartDelay = %v; want 1s", manager.config.RestartDelay)
	}

	if manager.config.MaxRestartDelay != 30*time.Second {
		t.Errorf("Default MaxRestartDelay = %v; want 30s", manager.config.MaxRestartDelay)
	}
}

func TestManagerStruct(t *testing.T) {
	cfg := ManagerConfig{
		GatewayPort:        "8080",
		FrontendPort:       "8000",
		HealthInterval:     5 * time.Second,
		MaxRestartAttempts: 10,
		RestartDelay:       1 * time.Second,
		MaxRestartDelay:    30 * time.Second,
	}

	manager := NewManager(cfg)

	if manager.httpClient == nil {
		t.Error("httpClient should not be nil")
	}

	if manager.stopChan == nil {
		t.Error("stopChan should not be nil")
	}

	if manager.doneChan == nil {
		t.Error("doneChan should not be nil")
	}
}

func TestManagerConfigZeroValues(t *testing.T) {
	cfg := ManagerConfig{}

	// Zero values should be valid
	if cfg.GatewayPort != "" {
		t.Errorf("GatewayPort = %s; want empty", cfg.GatewayPort)
	}

	if cfg.HealthInterval != 0 {
		t.Errorf("HealthInterval = %v; want 0", cfg.HealthInterval)
	}
}
