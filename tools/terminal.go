package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// --- File System Tools ---

type FileEditorTool struct {
	BaseDir string
}

func (t *FileEditorTool) Name() string { return "file_editor" }
func (t *FileEditorTool) Description() string { return "Read, Write, or Edit files (replace string)" }
func (t *FileEditorTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action":   map[string]interface{}{"type": "string", "enum": []string{"read", "write", "str_replace"}},
			"path":     map[string]interface{}{"type": "string"},
			"content":  map[string]interface{}{"type": "string"},
			"old_str":  map[string]interface{}{"type": "string"},
			"new_str":  map[string]interface{}{"type": "string"},
		},
		"required": []string{"action", "path"},
	}
}

func (t *FileEditorTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	action, _ := GetArg[string](input, "action")
	relPath, _ := GetArg[string](input, "path")
	
	// Security: Prevent directory traversal
	fullPath := filepath.Join(t.BaseDir, relPath)
	if !strings.HasPrefix(fullPath, t.BaseDir) {
		return ErrorOutput(fmt.Errorf("access denied to path outside workspace")), nil
	}

	switch action {
	case "read":
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return ErrorOutput(err), nil
		}
		return &ToolOutput{Text: string(data)}, nil

	case "write":
		content, _ := GetArg[string](input, "content")
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return ErrorOutput(err), nil
		}
		return &ToolOutput{Text: "File written successfully."}, nil

	case "str_replace":
		oldStr, _ := GetArg[string](input, "old_str")
		newStr, _ := GetArg[string](input, "new_str")
		
		data, err := os.ReadFile(fullPath)
		if err != nil { return ErrorOutput(err), nil }
		
		content := string(data)
		if strings.Count(content, oldStr) > 1 {
			return ErrorOutput(fmt.Errorf("multiple occurrences of old_str found, please be more specific")), nil
		}
		if !strings.Contains(content, oldStr) {
			return ErrorOutput(fmt.Errorf("old_str not found in file")), nil
		}
		
		newContent := strings.Replace(content, oldStr, newStr, 1)
		if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
			return ErrorOutput(err), nil
		}
		return &ToolOutput{Text: "File patched successfully."}, nil
	}

	return ErrorOutput(fmt.Errorf("unknown action")), nil
}

// --- Terminal Tools ---

type TerminalTool struct {
	WorkDir string
}

func (t *TerminalTool) Name() string { return "terminal_execute" }
func (t *TerminalTool) Description() string { return "Execute a shell command" }
func (t *TerminalTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]string{"type": "string"},
			"timeout": map[string]string{"type": "integer"},
		},
		"required": []string{"command"},
	}
}

func (t *TerminalTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	cmdStr, _ := GetArg[string](input, "command")
	timeoutSec, _ := GetArg[int](input, "timeout")
	if timeoutSec == 0 { timeoutSec = 30 }

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)
	cmd.Dir = t.WorkDir
	
	output, err := cmd.CombinedOutput()
	resultText := string(output)
	
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			resultText += "\n[Error: Command timed out]"
		} else {
			resultText += fmt.Sprintf("\n[Error: %v]", err)
		}
	}
	
	return &ToolOutput{Text: resultText}, nil
}

// --- Search Tool ---
type TerminalWebSearchTool struct {
	Config Config
}

func (t *TerminalWebSearchTool) Name() string { return "web_search" }
func (t *TerminalWebSearchTool) Description() string { return "Search the web (Tavily/Jina/DuckDuckGo)" }
func (t *TerminalWebSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}},
		"required": []string{"query"},
	}
}

func (t *TerminalWebSearchTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	query, _ := GetArg[string](input, "query")

	// Priority: Tavily -> Jina -> DuckDuckGo (Placeholder logic)
	if t.Config.TavilyAPIKey != "" {
		// Mock Tavily Call
		return t.tavilySearch(query)
	}
	
	// Fallback to DuckDuckGo (simplified simulation)
	return &ToolOutput{Text: fmt.Sprintf("Searching DuckDuckGo for: %s (Implementation pending real API call)", query)}, nil
}

func (t *TerminalWebSearchTool) tavilySearch(query string) (*ToolOutput, error) {
	// Implement HTTP Post to Tavily API
	return &ToolOutput{Text: "Tavily results for: " + query}, nil
}

// --- Webpage Visit Tool ---
type TerminalWebVisitTool struct {
	Config Config
}

func (t *TerminalWebVisitTool) Name() string { return "visit_webpage" }
func (t *TerminalWebVisitTool) Description() string { return "Visit a webpage and extract markdown" }
func (t *TerminalWebVisitTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{"url": map[string]interface{}{"type": "string"}},
		"required": []string{"url"},
	}
}

func (t *TerminalWebVisitTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	targetURL, _ := GetArg[string](input, "url")
	
	// Use Jina Reader if Key exists (cleanest markdown)
	if t.Config.JinaAPIKey != "" {
		jinaURL := "https://r.jina.ai/" + targetURL
		req, _ := http.NewRequest("GET", jinaURL, nil)
		req.Header.Set("Authorization", "Bearer "+t.Config.JinaAPIKey)
		
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil { return ErrorOutput(err), nil }
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		return &ToolOutput{Text: string(body)}, nil
	}
	
	return &ToolOutput{Text: "No Jina API key provided, cannot convert to Markdown reliably."}, nil
}