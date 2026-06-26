package tools

import (
	"context"
	"fmt"
)

// ToolInput represents the generic JSON input from an LLM
type ToolInput map[string]interface{}

// ToolOutput represents the standard response sent back to the LLM
type ToolOutput struct {
	Text      string                 `json:"text"`
	Images    []string               `json:"images,omitempty"` // Base64 encoded images
	Auxiliary map[string]interface{} `json:"auxiliary,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Tool defines the interface that all Water AI tools must implement
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Run(ctx context.Context, input ToolInput) (*ToolOutput, error)
}

// Config holds API keys and settings for tools
type Config struct {
	BrowserHeadless   bool
	WorkspacePath     string
	GeminiAPIKey      string
	SerpAPIKey        string
	TavilyAPIKey      string
	JinaAPIKey        string
	FirecrawlAPIKey   string
	NeonAPIKey        string
}

// Helper to format errors safely
func ErrorOutput(err error) *ToolOutput {
	if err == nil {
		return &ToolOutput{Text: "Error: <nil>", Error: "<nil>"}
	}
	return &ToolOutput{
		Text:  fmt.Sprintf("Error: %v", err),
		Error: err.Error(),
	}
}

// Helper to parse input helper
func GetArg[T any](input ToolInput, key string) (T, error) {
	val, ok := input[key]
	var zero T
	if !ok {
		return zero, fmt.Errorf("missing argument '%s'", key)
	}
	
	// Handle float64 to int conversion typical in JSON decoding
	switch any(zero).(type) {
	case int:
		if f, ok := val.(float64); ok {
			return any(int(f)).(T), nil
		}
	}

	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("argument '%s' has invalid type", key)
	}
	return typed, nil
}