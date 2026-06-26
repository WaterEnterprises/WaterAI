package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config represents the configuration required to initialize the store.
// In the original code, this was IIAgentConfig.
type Config struct {
	FileStorePath string
}

// Settings represents the user settings data model.
// Adjust the struct fields below to match the JSON structure you expect.
type Settings struct {
	// Example fields based on typical AI agent needs
	UserID    string                 `json:"user_id,omitempty"`
	Theme     string                 `json:"theme,omitempty"`
	APIKeys   map[string]string      `json:"api_keys,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Store is the interface for storing and retrieving user settings.
// This replaces the abstract base class SettingsStore.
type Store interface {
	Load(ctx context.Context) (*Settings, error)
	Save(ctx context.Context, settings *Settings) error
}

// FileStore implements Store using the local file system.
// This replaces FileSettingsStore.
type FileStore struct {
	path string
	mu   sync.RWMutex // mutex ensures thread-safe access to the file
}

// NewFileStore creates a new instance of a file-based settings store.
// This replaces the get_instance class method.
func NewFileStore(cfg Config, userID string) (*FileStore, error) {
	if cfg.FileStorePath == "" {
		return nil, errors.New("file store path is required")
	}

	// Ensure the directory exists
	if err := os.MkdirAll(cfg.FileStorePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Construct the full path (e.g., path/to/store/settings.json)
	// You might want to make the filename unique per user if userID is provided
	filename := "settings.json"
	if userID != "" {
		filename = fmt.Sprintf("settings_%s.json", userID)
	}

	return &FileStore{
		path: filepath.Join(cfg.FileStorePath, filename),
	}, nil
}

// Load reads the settings from the file.
// Returns nil if the file does not exist (similar to the Python FileNotFoundError handling).
func (fs *FileStore) Load(ctx context.Context) (*Settings, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// check context for cancellation before doing I/O
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fs.path)
	if os.IsNotExist(err) {
		// Return nil, nil if file doesn't exist, as per original logic
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings json: %w", err)
	}

	return &settings, nil
}

// Save writes the settings to the file.
func (fs *FileStore) Save(ctx context.Context, settings *Settings) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	// Indent for readability, similar to how dump_json usually works
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// 0600 permissions: read/write only by owner (secure for secrets)
	if err := os.WriteFile(fs.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}