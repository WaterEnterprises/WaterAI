package tools

import (
	"context"
	"testing"
)

func TestToolInput(t *testing.T) {
	input := ToolInput{
		"command": "ls -la",
		"timeout": 30,
	}

	if input["command"] != "ls -la" {
		t.Errorf("command = %v; want ls -la", input["command"])
	}

	if input["timeout"] != 30 {
		t.Errorf("timeout = %v; want 30", input["timeout"])
	}
}

func TestToolOutput(t *testing.T) {
	output := ToolOutput{
		Text:      "Command output",
		Images:    []string{"img1", "img2"},
		Auxiliary: map[string]interface{}{"key": "value"},
		Error:     "",
	}

	if output.Text != "Command output" {
		t.Errorf("Text = %s; want Command output", output.Text)
	}

	if len(output.Images) != 2 {
		t.Errorf("Images length = %d; want 2", len(output.Images))
	}

	if output.Auxiliary["key"] != "value" {
		t.Errorf("Auxiliary[key] = %v; want value", output.Auxiliary["key"])
	}
}

func TestToolOutputWithError(t *testing.T) {
	output := ToolOutput{
		Text:  "Error occurred",
		Error: "connection refused",
	}

	if output.Error != "connection refused" {
		t.Errorf("Error = %s; want connection refused", output.Error)
	}
}

func TestToolOutputWithEmptyImages(t *testing.T) {
	output := ToolOutput{
		Text:   "No images",
		Images: nil,
	}

	if output.Images != nil {
		t.Errorf("Images = %v; want nil", output.Images)
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		BrowserHeadless:  true,
		WorkspacePath:    "/workspace",
		GeminiAPIKey:     "gemini-key",
		SerpAPIKey:       "serp-key",
		TavilyAPIKey:     "tavily-key",
		JinaAPIKey:       "jina-key",
		FirecrawlAPIKey:  "firecrawl-key",
		NeonAPIKey:       "neon-key",
	}

	if !cfg.BrowserHeadless {
		t.Error("BrowserHeadless should be true")
	}

	if cfg.WorkspacePath != "/workspace" {
		t.Errorf("WorkspacePath = %s; want /workspace", cfg.WorkspacePath)
	}

	if cfg.GeminiAPIKey != "gemini-key" {
		t.Errorf("GeminiAPIKey = %s; want gemini-key", cfg.GeminiAPIKey)
	}
}

func TestErrorOutput(t *testing.T) {
	err := &testError{"test error message"}
	output := ErrorOutput(err)

	if output.Text != "Error: test error message" {
		t.Errorf("Text = %s; want Error: test error message", output.Text)
	}

	if output.Error != "test error message" {
		t.Errorf("Error = %s; want test error message", output.Error)
	}
}

func TestErrorOutputNil(t *testing.T) {
	output := ErrorOutput(nil)

	if output.Text != "Error: <nil>" {
		t.Errorf("Text = %s; want Error: <nil>", output.Text)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestGetArgString(t *testing.T) {
	input := ToolInput{
		"name": "John Doe",
	}

	result, err := GetArg[string](input, "name")
	if err != nil {
		t.Fatalf("GetArg() error = %v", err)
	}

	if result != "John Doe" {
		t.Errorf("result = %s; want John Doe", result)
	}
}

func TestGetArgInt(t *testing.T) {
	input := ToolInput{
		"count": 42,
	}

	result, err := GetArg[int](input, "count")
	if err != nil {
		t.Fatalf("GetArg() error = %v", err)
	}

	if result != 42 {
		t.Errorf("result = %d; want 42", result)
	}
}

func TestGetArgMissing(t *testing.T) {
	input := ToolInput{}

	_, err := GetArg[string](input, "missing")
	if err == nil {
		t.Error("GetArg() should return error for missing key")
	}

	expected := "missing argument 'missing'"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestGetArgInvalidType(t *testing.T) {
	input := ToolInput{
		"value": "not an int",
	}

	_, err := GetArg[int](input, "value")
	if err == nil {
		t.Error("GetArg() should return error for invalid type")
	}

	expected := "argument 'value' has invalid type"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestGetArgFloat64ToInt(t *testing.T) {
	// JSON numbers are decoded as float64 by default
	input := ToolInput{
		"count": float64(42),
	}

	result, err := GetArg[int](input, "count")
	if err != nil {
		t.Fatalf("GetArg() error = %v", err)
	}

	if result != 42 {
		t.Errorf("result = %d; want 42", result)
	}
}

func TestToolInterface(t *testing.T) {
	// Verify Tool is an interface
	var _ Tool = (*mockTool)(nil)
}

type mockTool struct{}

func (m *mockTool) Name() string                                      { return "mock" }
func (m *mockTool) Description() string                               { return "Mock tool" }
func (m *mockTool) InputSchema() map[string]interface{}               { return nil }
func (m *mockTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	return &ToolOutput{}, nil
}

func TestFileEditorToolName(t *testing.T) {
	tool := &FileEditorTool{}
	if tool.Name() != "file_editor" {
		t.Errorf("Name = %s; want file_editor", tool.Name())
	}
}

func TestFileEditorToolDescription(t *testing.T) {
	tool := &FileEditorTool{}
	if tool.Description() != "Read, Write, or Edit files (replace string)" {
		t.Errorf("Description = %s", tool.Description())
	}
}

func TestFileEditorToolInputSchema(t *testing.T) {
	tool := &FileEditorTool{}
	schema := tool.InputSchema()

	if schema["type"] != "object" {
		t.Errorf("Schema type = %v; want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be map[string]interface{}")
	}

	if _, ok := props["action"]; !ok {
		t.Error("action property should exist")
	}

	if _, ok := props["path"]; !ok {
		t.Error("path property should exist")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required should be []string")
	}

	if len(required) != 2 {
		t.Errorf("required length = %d; want 2", len(required))
	}
}

func TestTerminalToolName(t *testing.T) {
	tool := &TerminalTool{}
	if tool.Name() != "terminal_execute" {
		t.Errorf("Name = %s; want terminal_execute", tool.Name())
	}
}

func TestTerminalToolDescription(t *testing.T) {
	tool := &TerminalTool{}
	if tool.Description() != "Execute a shell command" {
		t.Errorf("Description = %s", tool.Description())
	}
}

func TestTerminalToolInputSchema(t *testing.T) {
	tool := &TerminalTool{}
	schema := tool.InputSchema()

	if schema["type"] != "object" {
		t.Errorf("Schema type = %v; want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be map[string]interface{}")
	}

	if _, ok := props["command"]; !ok {
		t.Error("command property should exist")
	}
}

func TestToolOutputWithOnlyText(t *testing.T) {
	output := ToolOutput{
		Text: "Simple output",
	}

	if output.Text != "Simple output" {
		t.Errorf("Text = %s; want Simple output", output.Text)
	}

	if output.Images != nil {
		t.Error("Images should be nil")
	}

	if output.Auxiliary != nil {
		t.Error("Auxiliary should be nil")
	}
}

func TestToolInputWithVariousTypes(t *testing.T) {
	input := ToolInput{
		"string_val": "text",
		"int_val":    100,
		"bool_val":   true,
		"float_val":  3.14,
		"array_val":  []string{"a", "b", "c"},
	}

	if input["string_val"] != "text" {
		t.Errorf("string_val = %v; want text", input["string_val"])
	}

	if input["int_val"] != 100 {
		t.Errorf("int_val = %v; want 100", input["int_val"])
	}

	if input["bool_val"] != true {
		t.Errorf("bool_val = %v; want true", input["bool_val"])
	}
}

func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{}

	if cfg.BrowserHeadless {
		t.Error("BrowserHeadless should be false by default")
	}

	if cfg.WorkspacePath != "" {
		t.Errorf("WorkspacePath = %s; want empty", cfg.WorkspacePath)
	}
}
