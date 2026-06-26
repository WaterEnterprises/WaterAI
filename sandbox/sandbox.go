package sandbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// WorkSpaceMode defines the type of sandbox environment
type WorkSpaceMode string

const (
	ModeDocker WorkSpaceMode = "docker"
	ModeE2B    WorkSpaceMode = "e2b"
	ModeLocal  WorkSpaceMode = "local"
)

// Settings represents the configuration required for sandboxes
type Settings struct {
	SandboxConfig struct {
		ServicePort   int
		TemplateID    string
		SandboxAPIKey string
		Image         string
		MemoryLimit   int64
		CPULimit      float64
		NetworkName   string
	}
}

// Sandbox defines the behavior every sandbox implementation must provide
type Sandbox interface {
	Connect(ctx context.Context) error
	Create(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
	ExposePort(port int) string
	GetHostURL() (string, error)
	GetSandboxID() (string, error)
}

// Base holds common fields for all sandbox types
type Base struct {
	SessionID  string
	Settings   *Settings
	SandboxID  string
	HostURL    string
}

func (b *Base) GetHostURL() (string, error) {
	if b.HostURL == "" {
		return "", errors.New("host URL is not set")
	}
	return b.HostURL, nil
}

func (b *Base) GetSandboxID() (string, error) {
	if b.SandboxID == "" {
		return "", errors.New("sandbox ID is not set")
	}
	return b.SandboxID, nil
}

// Registry manages the available sandbox implementations
type Registry struct {
	mu        sync.RWMutex
	factories map[WorkSpaceMode]func(string, *Settings) Sandbox
}

var globalRegistry = &Registry{
	factories: make(map[WorkSpaceMode]func(string, *Settings) Sandbox),
}

func Register(mode WorkSpaceMode, factory func(string, *Settings) Sandbox) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[mode] = factory
}

func Create(mode WorkSpaceMode, sessionID string, settings *Settings) (Sandbox, error) {
	globalRegistry.mu.RLock()
	factory, ok := globalRegistry.factories[mode]
	globalRegistry.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown sandbox type: %s", mode)
	}
	return factory(sessionID, settings), nil
}