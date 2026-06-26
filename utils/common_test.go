package utils

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if UploadFolderName != "uploaded_files" {
		t.Errorf("UploadFolderName = %s; want uploaded_files", UploadFolderName)
	}

	if CompleteMessage != "Completed the task." {
		t.Errorf("CompleteMessage = %s; want Completed the task.", CompleteMessage)
	}

	if DefaultModel != "claude-sonnet-4@20250514" {
		t.Errorf("DefaultModel = %s; want claude-sonnet-4@20250514", DefaultModel)
	}

	if TokenBudget != 120_000 {
		t.Errorf("TokenBudget = %d; want 120000", TokenBudget)
	}

	if SummaryMaxTokens != 32_000 {
		t.Errorf("SummaryMaxTokens = %d; want 32000", SummaryMaxTokens)
	}

	if VisitWebPageMaxOutputLength != 40_000 {
		t.Errorf("VisitWebPageMaxOutputLength = %d; want 40000", VisitWebPageMaxOutputLength)
	}

	if SnippetLines != 4 {
		t.Errorf("SnippetLines = %d; want 4", SnippetLines)
	}

	if MaxResponseLen != 200_000 {
		t.Errorf("MaxResponseLen = %d; want 200000", MaxResponseLen)
	}
}

func TestWorkspaceMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     WorkspaceMode
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

func TestSessionResult(t *testing.T) {
	result := SessionResult{
		Success: true,
		Output:  "Task completed",
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.Output != "Task completed" {
		t.Errorf("Output = %s; want Task completed", result.Output)
	}
}

func TestSessionResultSuccess(t *testing.T) {
	result := SessionResult{
		Success: true,
		Output:  "Success output",
	}

	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestSessionResultFailure(t *testing.T) {
	result := SessionResult{
		Success: false,
		Output:  "Error occurred",
	}

	if result.Success {
		t.Error("Success should be false")
	}
}

func TestStrReplaceResponse(t *testing.T) {
	response := StrReplaceResponse{
		Success:     true,
		FileContent: "new file content",
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	if response.FileContent != "new file content" {
		t.Errorf("FileContent = %s; want new file content", response.FileContent)
	}
}

func TestStrReplaceResponseSuccess(t *testing.T) {
	response := StrReplaceResponse{
		Success:     true,
		FileContent: "content",
	}

	if !response.Success {
		t.Error("Success should be true")
	}
}

func TestNewSandboxSettings(t *testing.T) {
	settings := NewSandboxSettings()

	if settings.WorkDir != "/workspace" {
		t.Errorf("WorkDir = %s; want /workspace", settings.WorkDir)
	}

	if settings.Mode != ModeLocal {
		t.Errorf("Mode = %s; want %s", settings.Mode, ModeLocal)
	}

	if settings.SystemShell != "/bin/bash" {
		t.Errorf("SystemShell = %s; want /bin/bash", settings.SystemShell)
	}
}

func TestSandboxSettingsDefaults(t *testing.T) {
	settings := &SandboxSettings{}

	if settings.WorkDir != "" {
		t.Errorf("WorkDir = %s; want empty", settings.WorkDir)
	}

	if settings.Mode != "" {
		t.Errorf("Mode = %s; want empty", settings.Mode)
	}

	if settings.SystemShell != "" {
		t.Errorf("SystemShell = %s; want empty", settings.SystemShell)
	}
}

func TestNewWorkspaceManager(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	if manager.Root != "/parent/session-123" {
		t.Errorf("Root = %s; want /parent/session-123", manager.Root)
	}

	if manager.SessionID != "session-123" {
		t.Errorf("SessionID = %s; want session-123", manager.SessionID)
	}

	if manager.Mode != ModeLocal {
		t.Errorf("Mode = %s; want %s", manager.Mode, ModeLocal)
	}
}

func TestWorkspaceManagerIsLocal(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	if !manager.IsLocal() {
		t.Error("IsLocal() should return true for ModeLocal")
	}
}

func TestWorkspaceManagerIsLocalDocker(t *testing.T) {
	settings := &SandboxSettings{
		WorkDir:     "/workspace",
		Mode:        ModeDocker,
		SystemShell: "/bin/bash",
	}
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	if manager.IsLocal() {
		t.Error("IsLocal() should return false for ModeDocker")
	}
}

func TestWorkspaceManagerWorkspacePath(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"relative", "file.txt", "/parent/session-123/file.txt"},
		{"nested", "dir/file.txt", "/parent/session-123/dir/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.WorkspacePath(tt.path)
			if result != tt.expected {
				t.Errorf("WorkspacePath(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestWorkspaceManagerRootPath(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	rootPath := manager.RootPath()

	if rootPath != manager.Root {
		t.Errorf("RootPath() = %s; want %s", rootPath, manager.Root)
	}
}

func TestWorkspaceManagerRootPathNonLocal(t *testing.T) {
	settings := &SandboxSettings{
		WorkDir:     "/container/workspace",
		Mode:        ModeDocker,
		SystemShell: "/bin/bash",
	}
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	rootPath := manager.RootPath()

	if rootPath != "/container/workspace" {
		t.Errorf("RootPath() = %s; want /container/workspace", rootPath)
	}
}

func TestTruncatedMessage(t *testing.T) {
	expected := "<response clipped><NOTE>To save on context only part of this file has been shown...</NOTE>"

	if TruncatedMessage != expected {
		t.Errorf("TruncatedMessage = %s; want %s", TruncatedMessage, expected)
	}
}

func TestSandboxSettingsWithCustomValues(t *testing.T) {
	settings := &SandboxSettings{
		WorkDir:     "/custom/workspace",
		Mode:        ModeE2B,
		SystemShell: "/bin/zsh",
	}

	if settings.WorkDir != "/custom/workspace" {
		t.Errorf("WorkDir = %s; want /custom/workspace", settings.WorkDir)
	}

	if settings.Mode != ModeE2B {
		t.Errorf("Mode = %s; want %s", settings.Mode, ModeE2B)
	}

	if settings.SystemShell != "/bin/zsh" {
		t.Errorf("SystemShell = %s; want /bin/zsh", settings.SystemShell)
	}
}

func TestWorkspaceManagerSessionID(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "test-session-id", settings)

	if manager.SessionID != "test-session-id" {
		t.Errorf("SessionID = %s; want test-session-id", manager.SessionID)
	}
}

func TestWorkspaceManagerContainerWork(t *testing.T) {
	settings := &SandboxSettings{
		WorkDir:     "/container/work",
		Mode:        ModeDocker,
		SystemShell: "/bin/bash",
	}
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	if manager.ContainerWork != "/container/work" {
		t.Errorf("ContainerWork = %s; want /container/work", manager.ContainerWork)
	}
}

func TestWorkspaceManagerContainerWorkEmptyForLocal(t *testing.T) {
	settings := NewSandboxSettings()
	manager := NewWorkspaceManager("/parent", "session-123", settings)

	if manager.ContainerWork != "" {
		t.Errorf("ContainerWork = %s; want empty for local mode", manager.ContainerWork)
	}
}
