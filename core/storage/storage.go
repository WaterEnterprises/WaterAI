package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Common constants
const ConversationBaseDir = "sessions"

// FileStore defines the interface for storage backends.
type FileStore interface {
	// Write saves contents to the specified path.
	Write(path string, contents []byte) error
	// Read retrieves contents from the specified path.
	Read(path string) ([]byte, error)
	// List returns a list of files and directories at the given path.
	// Directories are suffixed with a forward slash.
	List(path string) ([]string, error)
	// Delete removes the file or directory (recursively) at the path.
	Delete(path string) error
}

// NewFileStore returns a FileStore instance based on the type.
// storeType: "local" or "memory".
// rootPath: Required for local store, ignored for memory store.
func NewFileStore(storeType string, rootPath string) (FileStore, error) {
	if storeType == "local" {
		if rootPath == "" {
			return nil, errors.New("file store path is required for local file store")
		}
		return NewLocalFileStore(rootPath)
	}
	return NewInMemoryFileStore(nil), nil
}

// GetConversationAgentHistoryFilename returns the formatted path for agent history.
func GetConversationAgentHistoryFilename(sid string) string {
	return filepath.Join(ConversationBaseDir, sid, "agent_state.pkl")
}

// -----------------------------------------------------------------------------
// LocalFileStore Implementation
// -----------------------------------------------------------------------------

type LocalFileStore struct {
	root string
}

// NewLocalFileStore creates a new filesystem-backed store.
func NewLocalFileStore(root string) (*LocalFileStore, error) {
	// Handle ~ expansion
	if strings.HasPrefix(root, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to expand user home dir: %w", err)
		}
		root = filepath.Join(home, root[1:])
	}

	// Ensure root exists
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return &LocalFileStore{root: root}, nil
}

func (s *LocalFileStore) getFullPath(p string) string {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(p)
	if strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "\\") {
		cleanPath = cleanPath[1:]
	}
	return filepath.Join(s.root, cleanPath)
}

func (s *LocalFileStore) Write(path string, contents []byte) error {
	fullPath := s.getFullPath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, contents, 0644)
}

func (s *LocalFileStore) Read(path string) ([]byte, error) {
	fullPath := s.getFullPath(path)
	return os.ReadFile(fullPath)
}

func (s *LocalFileStore) List(path string) ([]string, error) {
	fullPath := s.getFullPath(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		// Construct relative path from the query path
		name := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			name += "/"
		}
		files = append(files, name)
	}
	return files, nil
}

func (s *LocalFileStore) Delete(path string) error {
	fullPath := s.getFullPath(path)
	// Check existence
	if _, err := os.Stat(fullPath); errors.Is(err, fs.ErrNotExist) {
		return nil // Behaves like Python: ignores if not exists
	}
	// RemoveAll handles both files and directories recursively
	return os.RemoveAll(fullPath)
}

// -----------------------------------------------------------------------------
// InMemoryFileStore Implementation
// -----------------------------------------------------------------------------

type InMemoryFileStore struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewInMemoryFileStore creates a new memory-backed store.
func NewInMemoryFileStore(initialFiles map[string][]byte) *InMemoryFileStore {
	files := make(map[string][]byte)
	if initialFiles != nil {
		for k, v := range initialFiles {
			files[k] = v
		}
	}
	return &InMemoryFileStore{files: files}
}

func (s *InMemoryFileStore) Write(path string, contents []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[path] = contents
	return nil
}

func (s *InMemoryFileStore) Read(path string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return val, nil
}

func (s *InMemoryFileStore) List(path string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Normalize path to ensure standard slash handling
	searchPath := path
	if searchPath != "" && !strings.HasSuffix(searchPath, "/") {
		searchPath += "/"
	}

	seenDirs := make(map[string]bool)
	var results []string

	for filePath := range s.files {
		// Check if file starts with the search path
		if !strings.HasPrefix(filePath, searchPath) {
			continue
		}

		// Get the relative suffix
		suffix := strings.TrimPrefix(filePath, searchPath)
		if suffix == "" {
			continue
		}

		parts := strings.Split(suffix, "/")

		if len(parts) == 1 {
			// It's a file in the current directory
			results = append(results, filePath)
		} else {
			// It's a subdirectory
			dirName := filepath.Join(path, parts[0]) + "/"
			if !seenDirs[dirName] {
				results = append(results, dirName)
				seenDirs[dirName] = true
			}
		}
	}

	return results, nil
}

func (s *InMemoryFileStore) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete strict matches and children
	for key := range s.files {
		if key == path || strings.HasPrefix(key, path+"/") {
			delete(s.files, key)
		}
	}
	return nil
}