package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// --- Bash Tool ---

type BashTool struct {
	WorkspaceRoot string
}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) Description() string { return "Execute a bash command in the workspace." }
func (t *BashTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]string{"type": "string", "description": "The command to run"},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	cmdStr, ok := input["command"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("command is required")
	}

	// Security: In a real app, strict filtering/sandboxing is required here.
	if strings.Contains(cmdStr, "rm -rf /") {
		return ToolResult{Output: "Command blocked for safety", Success: false}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)
	if t.WorkspaceRoot != "" {
		cmd.Dir = t.WorkspaceRoot
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return ToolResult{
			Output:        fmt.Sprintf("Error: %v\nOutput: %s", err, outputStr),
			ResultMessage: "Command failed",
			Success:       false,
		}, nil
	}

	return ToolResult{
		Output:        outputStr,
		ResultMessage: "Command executed successfully",
		Success:       true,
	}, nil
}

// --- String Replace / File Editor Tool ---

type SystemFileEditorTool struct {
	WorkspaceRoot string
}

func (t *SystemFileEditorTool) Name() string        { return "str_replace_editor" }
func (t *SystemFileEditorTool) Description() string { return "View, create, or replace text in files." }
func (t *SystemFileEditorTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command":   map[string]string{"type": "string", "enum": "view, create, str_replace"},
			"path":      map[string]string{"type": "string"},
			"file_text": map[string]string{"type": "string", "description": "Content for create"},
			"old_str":   map[string]string{"type": "string", "description": "Text to replace"},
			"new_str":   map[string]string{"type": "string", "description": "Replacement text"},
		},
		"required": []string{"command", "path"},
	}
}

func (t *SystemFileEditorTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	cmd, _ := input["command"].(string)
	path, _ := input["path"].(string)
	
	fullPath := filepath.Join(t.WorkspaceRoot, path)

	switch cmd {
	case "view":
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return ToolResult{Output: err.Error(), Success: false}, nil
		}
		return ToolResult{Output: string(content), Success: true}, nil

	case "create":
		content, _ := input["file_text"].(string)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return ToolResult{Output: err.Error(), Success: false}, nil
		}
		return ToolResult{Output: "File created", Success: true}, nil

	case "str_replace":
		oldStr, _ := input["old_str"].(string)
		newStr, _ := input["new_str"].(string)
		
		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return ToolResult{Output: err.Error(), Success: false}, nil
		}
		content := string(contentBytes)

		if strings.Count(content, oldStr) != 1 {
			return ToolResult{Output: "old_str must occur exactly once in the file", Success: false}, nil
		}

		newContent := strings.Replace(content, oldStr, newStr, 1)
		if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
			return ToolResult{Output: err.Error(), Success: false}, nil
		}
		return ToolResult{Output: "File updated", Success: true}, nil
	}

	return ToolResult{Output: "Unknown command", Success: false}, nil
}