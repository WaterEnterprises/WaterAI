package settings

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStore(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "test-user")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	if store == nil {
		t.Fatal("NewFileStore() returned nil")
	}
}

func TestNewFileStoreEmptyPath(t *testing.T) {
	cfg := Config{FileStorePath: ""}

	_, err := NewFileStore(cfg, "test-user")
	if err == nil {
		t.Error("NewFileStore() should return error for empty path")
	}
}

func TestNewFileStoreCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "nested", "dir")
	cfg := Config{FileStorePath: storePath}

	_, err := NewFileStore(cfg, "")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Error("NewFileStore() should create directory")
	}
}

func TestFileStoreLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "nonexistent")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	settings, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if settings != nil {
		t.Error("Load() should return nil for non-existent file")
	}
}

func TestFileStoreSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "test-user")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	// Save settings
	settings := &Settings{
		UserID:  "test-user",
		Theme:   "dark",
		APIKeys: map[string]string{"openai": "sk-test"},
	}

	err = store.Save(context.Background(), settings)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load settings
	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	if loaded.UserID != "test-user" {
		t.Errorf("UserID = %s; want test-user", loaded.UserID)
	}

	if loaded.Theme != "dark" {
		t.Errorf("Theme = %s; want dark", loaded.Theme)
	}

	if loaded.APIKeys["openai"] != "sk-test" {
		t.Errorf("APIKeys[openai] = %s; want sk-test", loaded.APIKeys["openai"])
	}
}

func TestFileStoreLoadInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "test-user")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	// Write invalid JSON
	invalidJSON := []byte("{invalid json}")
	err = os.WriteFile(store.path, invalidJSON, 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = store.Load(context.Background())
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestFileStoreSaveNilSettings(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "test-user")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	err = store.Save(context.Background(), nil)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

func TestFileStoreLoadCancelledContext(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "test-user")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = store.Load(ctx)
	if err == nil {
		t.Error("Load() should return error for cancelled context")
	}
}

func TestSettingsStruct(t *testing.T) {
	settings := Settings{
		UserID:  "user123",
		Theme:   "light",
		APIKeys: map[string]string{"key1": "value1"},
		Variables: map[string]interface{}{
			"var1": "val1",
		},
	}

	if settings.UserID != "user123" {
		t.Errorf("UserID = %s; want user123", settings.UserID)
	}

	if settings.Theme != "light" {
		t.Errorf("Theme = %s; want light", settings.Theme)
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		FileStorePath: "/path/to/store",
	}

	if cfg.FileStorePath != "/path/to/store" {
		t.Errorf("FileStorePath = %s; want /path/to/store", cfg.FileStorePath)
	}
}

func TestStoreInterface(t *testing.T) {
	// Verify FileStore implements Store interface
	var _ Store = (*FileStore)(nil)
}

func TestNewFileStoreWithoutUserID(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	// Should use default filename
	expectedPath := filepath.Join(tempDir, "settings.json")
	if store.path != expectedPath {
		t.Errorf("path = %s; want %s", store.path, expectedPath)
	}
}

func TestNewFileStoreWithUserID(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{FileStorePath: tempDir}

	store, err := NewFileStore(cfg, "user123")
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	// Should use user-specific filename
	expectedPath := filepath.Join(tempDir, "settings_user123.json")
	if store.path != expectedPath {
		t.Errorf("path = %s; want %s", store.path, expectedPath)
	}
}
