package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Settings holds configuration for all tools (API keys, paths, etc.)
type Settings struct {
	WorkspaceRoot    string
	OpenAIKey        string
	AzureEndpoint    string
	AzureAPIVersion  string
	GoogleAPIKey     string
	GCPProjectID     string
	GCPLocation      string
	GCSOutputBucket  string
	SearchAPIKey     string // e.g., Serper or Bing
}


// ToolResult represents the standardized output returned to the LLM
type ToolResult struct {
	Output        string                 `json:"output"`
	ResultMessage string                 `json:"result_message"`
	Success       bool                   `json:"success"`
	AuxiliaryData map[string]interface{} `json:"auxiliary_data,omitempty"`
}

// SystemTool is the interface all Water AI tools must implement
type SystemTool interface {
	Name() string
	Description() string
	// Schema returns the JSON schema for the tool's input
	Schema() map[string]interface{}
	// Run executes the tool logic
	Run(ctx context.Context, input ToolInput) (ToolResult, error)
}

// Manager handles tool registration and execution
type Manager struct {
	tools    map[string]SystemTool
	Settings Settings
}

func NewManager(settings Settings) *Manager {
	return &Manager{
		tools:    make(map[string]SystemTool),
		Settings: settings,
	}
}

func (m *Manager) Register(tools ...SystemTool) {
	for _, t := range tools {
		m.tools[t.Name()] = t
	}
}

func (m *Manager) GetTool(name string) (SystemTool, bool) {
	t, ok := m.tools[name]
	return t, ok
}

func (m *Manager) GetAllTools() []SystemTool {
	var list []SystemTool
	for _, t := range m.tools {
		list = append(list, t)
	}
	return list
}

func (m *Manager) ExecuteTool(ctx context.Context, name string, rawInput string) (ToolResult, error) {
	tool, exists := m.tools[name]
	if !exists {
		return ToolResult{Success: false}, fmt.Errorf("tool %s not found", name)
	}

	var input ToolInput
	if err := json.Unmarshal([]byte(rawInput), &input); err != nil {
		return ToolResult{Success: false, Output: "Invalid JSON input"}, err
	}

	log.Printf("Running tool: %s", name)
	result, err := tool.Run(ctx, input)
	if err != nil {
		// Return the error as a result output so the LLM sees it
		return ToolResult{
			Output:        fmt.Sprintf("Error executing tool: %v", err),
			ResultMessage: "Tool execution failed",
			Success:       false,
			AuxiliaryData: map[string]interface{}{"error": err.Error()},
		}, nil
	}
	return result, nil
}