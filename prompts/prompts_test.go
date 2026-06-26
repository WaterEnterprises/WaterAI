package prompts

import (
	"testing"
)

func TestWorkspaceModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     WorkspaceMode
		expected string
	}{
		{"Local", WorkspaceModeLocal, "local"},
		{"Sandbox", WorkspaceModeSandbox, "sandbox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("Mode = %s; want %s", tt.mode, tt.expected)
			}
		})
	}
}

func TestWaterAIConstants(t *testing.T) {
	if AgentName != "Water AI" {
		t.Errorf("AgentName = %s; want Water AI", AgentName)
	}

	if TeamName != "Water AI Team" {
		t.Errorf("TeamName = %s; want Water AI Team", TeamName)
	}
}

func TestNewSystemPromptBuilder(t *testing.T) {
	builder := NewSystemPromptBuilder(WorkspaceModeLocal, false)

	if builder == nil {
		t.Fatal("NewSystemPromptBuilder returned nil")
	}

	if builder.WorkspaceMode != WorkspaceModeLocal {
		t.Errorf("WorkspaceMode = %s; want %s", builder.WorkspaceMode, WorkspaceModeLocal)
	}

	if builder.SequentialThinking != false {
		t.Error("SequentialThinking should be false")
	}

	if builder.DefaultPrompt == "" {
		t.Error("DefaultPrompt should not be empty")
	}
}

func TestNewSystemPromptBuilderWithSeqThinking(t *testing.T) {
	builder := NewSystemPromptBuilder(WorkspaceModeSandbox, true)

	if builder.SequentialThinking != true {
		t.Error("SequentialThinking should be true")
	}
}

func TestSystemPromptBuilderReset(t *testing.T) {
	builder := NewSystemPromptBuilder(WorkspaceModeLocal, false)
	originalPrompt := builder.CurrentPrompt

	// Modify the prompt
	builder.CurrentPrompt = "modified prompt"

	// Reset
	builder.Reset()

	if builder.CurrentPrompt != originalPrompt {
		t.Error("Reset() should restore CurrentPrompt to DefaultPrompt")
	}
}

func TestSystemPromptBuilderGetPrompt(t *testing.T) {
	builder := NewSystemPromptBuilder(WorkspaceModeLocal, false)

	prompt := builder.GetPrompt()

	if prompt == "" {
		t.Error("GetPrompt() should not return empty string")
	}

	if prompt != builder.CurrentPrompt {
		t.Error("GetPrompt() should return CurrentPrompt")
	}
}

func TestSystemPromptBuilderStruct(t *testing.T) {
	builder := &SystemPromptBuilder{
		WorkspaceMode:      WorkspaceModeLocal,
		SequentialThinking: true,
		DefaultPrompt:      "default",
		CurrentPrompt:      "current",
	}

	if builder.WorkspaceMode != WorkspaceModeLocal {
		t.Errorf("WorkspaceMode = %s; want %s", builder.WorkspaceMode, WorkspaceModeLocal)
	}

	if !builder.SequentialThinking {
		t.Error("SequentialThinking should be true")
	}
}

func TestGetSystemPromptLocal(t *testing.T) {
	prompt := GetSystemPrompt(WorkspaceModeLocal, false)

	if prompt == "" {
		t.Error("GetSystemPrompt should not return empty string")
	}
}

func TestGetSystemPromptSandbox(t *testing.T) {
	prompt := GetSystemPrompt(WorkspaceModeSandbox, false)

	if prompt == "" {
		t.Error("GetSystemPrompt should not return empty string")
	}
}

func TestGetSystemPromptWithSeqThinking(t *testing.T) {
	prompt := GetSystemPrompt(WorkspaceModeLocal, true)

	if prompt == "" {
		t.Error("GetSystemPrompt should not return empty string")
	}
}
