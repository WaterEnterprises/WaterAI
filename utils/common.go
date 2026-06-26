// --- START OF FILE utils/common.go ---
package utils

import (
	"path/filepath"
	"strings"
)

// --- Constants ---

const (
	UploadFolderName           = "uploaded_files"
	CompleteMessage            = "Completed the task."
	DefaultModel               = "claude-sonnet-4@20250514"
	TokenBudget                = 120_000
	SummaryMaxTokens           = 32_000
	VisitWebPageMaxOutputLength = 40_000
	SnippetLines               = 4
	MaxResponseLen             = 200_000
	TruncatedMessage           = "<response clipped><NOTE>To save on context only part of this file has been shown...</NOTE>"
)

// --- Enums ---

type WorkspaceMode string

const (
	ModeDocker WorkspaceMode = "docker"
	ModeE2B    WorkspaceMode = "e2b"
	ModeLocal  WorkspaceMode = "local"
)

// --- Models (DTOs) ---

type SessionResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type StrReplaceResponse struct {
	Success     bool   `json:"success"`
	FileContent string `json:"file_content"`
}

// --- Configuration ---

type SandboxSettings struct {
	WorkDir     string
	Mode        WorkspaceMode
	SystemShell string
}

func NewSandboxSettings() *SandboxSettings {
	return &SandboxSettings{
		WorkDir:     "/workspace",
		Mode:        ModeLocal,
		SystemShell: "/bin/bash",
	}
}

// --- Workspace Manager ---

type WorkspaceManager struct {
	Root          string
	SessionID     string
	Mode          WorkspaceMode
	ContainerWork string
}

func NewWorkspaceManager(parentDir, sessionID string, settings *SandboxSettings) *WorkspaceManager {
	root := filepath.Join(parentDir, sessionID)
	// Note: Directory creation should happen in the caller or init
	
	wm := &WorkspaceManager{
		Root:      root,
		SessionID: sessionID,
		Mode:      settings.Mode,
	}

	if settings.Mode != ModeLocal {
		wm.ContainerWork = settings.WorkDir
	}
	return wm
}

func (w *WorkspaceManager) IsLocal() bool {
	return w.Mode == ModeLocal
}

// WorkspacePath returns the absolute local path
func (w *WorkspaceManager) WorkspacePath(pathStr string) string {
	if filepath.IsAbs(pathStr) && !strings.HasPrefix(pathStr, w.Root) {
		// If it's absolute but not in our root, treat it as relative to container workspace logic
		// simpler approach for Go: join with root
		rel := filepath.Base(pathStr) // Fallback safety
		return filepath.Join(w.Root, rel)
	}
	// Simple join for relative paths
	return filepath.Join(w.Root, pathStr)
}

func (w *WorkspaceManager) RootPath() string {
	if !w.IsLocal() {
		return w.ContainerWork
	}
	return w.Root
}