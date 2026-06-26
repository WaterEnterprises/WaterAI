package sandbox

import (
	"context"
	"testing"
)

func TestWorkSpaceModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     WorkSpaceMode
		expected string
	}{
		{"Docker", ModeDocker, "docker"},
		{"E2B", ModeE2B, "e2b"},
		{"Local", ModeLocal, "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("Mode = %s; want %s", tt.mode, tt.expected)
			}
		})
	}
}

func TestBaseGetHostURL(t *testing.T) {
	base := &Base{
		HostURL: "http://localhost:8080",
	}

	url, err := base.GetHostURL()
	if err != nil {
		t.Fatalf("GetHostURL() error = %v", err)
	}

	if url != "http://localhost:8080" {
		t.Errorf("GetHostURL() = %s; want http://localhost:8080", url)
	}
}

func TestBaseGetHostURLEmpty(t *testing.T) {
	base := &Base{}

	_, err := base.GetHostURL()
	if err == nil {
		t.Error("GetHostURL() should return error when HostURL is empty")
	}

	expected := "host URL is not set"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestBaseGetSandboxID(t *testing.T) {
	base := &Base{
		SandboxID: "sandbox-123",
	}

	id, err := base.GetSandboxID()
	if err != nil {
		t.Fatalf("GetSandboxID() error = %v", err)
	}

	if id != "sandbox-123" {
		t.Errorf("GetSandboxID() = %s; want sandbox-123", id)
	}
}

func TestBaseGetSandboxIDEmpty(t *testing.T) {
	base := &Base{}

	_, err := base.GetSandboxID()
	if err == nil {
		t.Error("GetSandboxID() should return error when SandboxID is empty")
	}

	expected := "sandbox ID is not set"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestRegister(t *testing.T) {
	factory := func(sessionID string, settings *Settings) Sandbox {
		return &mockSandbox{}
	}

	// Clear global registry first
	globalRegistry.factories = make(map[WorkSpaceMode]func(string, *Settings) Sandbox)

	Register(ModeDocker, factory)

	// Verify registration
	if _, ok := globalRegistry.factories[ModeDocker]; !ok {
		t.Error("Register() should add factory to registry")
	}
}

func TestCreate(t *testing.T) {
	// Clear global registry
	globalRegistry.factories = make(map[WorkSpaceMode]func(string, *Settings) Sandbox)

	// Register a mock factory
	factory := func(sessionID string, settings *Settings) Sandbox {
		return &mockSandbox{sandboxID: "test-sandbox"}
	}
	Register(ModeLocal, factory)

	settings := &Settings{}
	sandbox, err := Create(ModeLocal, "test-session", settings)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sandbox == nil {
		t.Fatal("Create() returned nil")
	}

	id, err := sandbox.GetSandboxID()
	if err != nil {
		t.Fatalf("GetSandboxID() error = %v", err)
	}

	if id != "test-sandbox" {
		t.Errorf("SandboxID = %s; want test-sandbox", id)
	}
}

func TestCreateUnknown(t *testing.T) {
	// Clear global registry
	globalRegistry.factories = make(map[WorkSpaceMode]func(string, *Settings) Sandbox)

	_, err := Create(ModeDocker, "test-session", &Settings{})
	if err == nil {
		t.Error("Create() should return error for unknown sandbox type")
	}

	expected := "unknown sandbox type: docker"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestRegisterConcurrent(t *testing.T) {
	// Clear global registry
	globalRegistry.factories = make(map[WorkSpaceMode]func(string, *Settings) Sandbox)

	done := make(chan bool)

	// Register concurrently
	for i := 0; i < 10; i++ {
		mode := WorkSpaceMode(string(rune('a' + i)))
		go func(m WorkSpaceMode) {
			factory := func(sessionID string, settings *Settings) Sandbox {
				return nil
			}
			Register(m, factory)
			done <- true
		}(mode)
	}

	// Wait for all registrations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all registrations succeeded
	if len(globalRegistry.factories) != 10 {
		t.Errorf("Registry has %d factories; want 10", len(globalRegistry.factories))
	}
}

func TestSandboxInterface(t *testing.T) {
	// Verify Sandbox is an interface
	var _ Sandbox = (*mockSandbox)(nil)
}

type mockSandbox struct {
	sandboxID string
}

func (m *mockSandbox) Connect(ctx context.Context) error             { return nil }
func (m *mockSandbox) Create(ctx context.Context) error              { return nil }
func (m *mockSandbox) Start(ctx context.Context) error               { return nil }
func (m *mockSandbox) Stop(ctx context.Context) error                { return nil }
func (m *mockSandbox) Cleanup(ctx context.Context) error             { return nil }
func (m *mockSandbox) ExposePort(port int) string                    { return "" }
func (m *mockSandbox) GetHostURL() (string, error)                   { return "", nil }
func (m *mockSandbox) GetSandboxID() (string, error)                 { return m.sandboxID, nil }

func TestSettings(t *testing.T) {
	settings := &Settings{}

	// Verify nested struct exists
	if settings.SandboxConfig.ServicePort != 0 {
		t.Errorf("ServicePort = %d; want 0", settings.SandboxConfig.ServicePort)
	}

	// Set values
	settings.SandboxConfig.ServicePort = 17300
	settings.SandboxConfig.TemplateID = "test-template"
	settings.SandboxConfig.MemoryLimit = 1024
	settings.SandboxConfig.CPULimit = 2.0

	if settings.SandboxConfig.ServicePort != 17300 {
		t.Errorf("ServicePort = %d; want 17300", settings.SandboxConfig.ServicePort)
	}

	if settings.SandboxConfig.TemplateID != "test-template" {
		t.Errorf("TemplateID = %s; want test-template", settings.SandboxConfig.TemplateID)
	}

	if settings.SandboxConfig.MemoryLimit != 1024 {
		t.Errorf("MemoryLimit = %d; want 1024", settings.SandboxConfig.MemoryLimit)
	}

	if settings.SandboxConfig.CPULimit != 2.0 {
		t.Errorf("CPULimit = %f; want 2.0", settings.SandboxConfig.CPULimit)
	}
}

func TestRegistryStruct(t *testing.T) {
	registry := &Registry{
		factories: make(map[WorkSpaceMode]func(string, *Settings) Sandbox),
	}

	if registry.factories == nil {
		t.Error("Factories map should not be nil")
	}

	// Add factory
	factory := func(sessionID string, settings *Settings) Sandbox {
		return &mockSandbox{}
	}
	registry.factories[ModeDocker] = factory

	if _, ok := registry.factories[ModeDocker]; !ok {
		t.Error("Factory should be in registry")
	}
}
