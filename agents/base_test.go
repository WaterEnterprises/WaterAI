package agents

import (
	"context"
	"testing"
)

func TestBaseAgentGetToolParam(t *testing.T) {
	agent := &BaseAgent{
		Name:        "test_agent",
		Description: "A test agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	toolParam := agent.GetToolParam()

	if toolParam.Name != "test_agent" {
		t.Errorf("Name = %s; want test_agent", toolParam.Name)
	}

	if toolParam.Description != "A test agent" {
		t.Errorf("Description = %s; want A test agent", toolParam.Description)
	}

	if toolParam.Schema["type"] != "object" {
		t.Errorf("Schema type = %v; want object", toolParam.Schema["type"])
	}
}

func TestBaseAgentRunNotImplemented(t *testing.T) {
	agent := &BaseAgent{
		Name:        "base_agent",
		Description: "Base agent",
	}

	ctx := context.Background()
	input := map[string]interface{}{"key": "value"}

	_, err := agent.Run(ctx, input, nil)

	if err == nil {
		t.Error("Run() should return error for base agent")
	}

	expected := "Run method not implemented in base agent"
	if err.Error() != expected {
		t.Errorf("Error message = %s; want %s", err.Error(), expected)
	}
}

func TestToolImplOutput(t *testing.T) {
	output := ToolImplOutput{
		ToolOutput:        "Command output",
		ToolResultMessage: "Result message",
		IsFinal:           true,
	}

	if output.ToolOutput != "Command output" {
		t.Errorf("ToolOutput = %s; want Command output", output.ToolOutput)
	}

	if output.ToolResultMessage != "Result message" {
		t.Errorf("ToolResultMessage = %s; want Result message", output.ToolResultMessage)
	}

	if !output.IsFinal {
		t.Error("IsFinal should be true")
	}
}

func TestToolCallParameters(t *testing.T) {
	params := ToolCallParameters{
		ID:        "call-123",
		Name:      "terminal_execute",
		Arguments: map[string]interface{}{"command": "ls"},
	}

	if params.ID != "call-123" {
		t.Errorf("ID = %s; want call-123", params.ID)
	}

	if params.Name != "terminal_execute" {
		t.Errorf("Name = %s; want terminal_execute", params.Name)
	}

	if params.Arguments["command"] != "ls" {
		t.Errorf("Arguments[command] = %v; want ls", params.Arguments["command"])
	}
}

func TestToolParam(t *testing.T) {
	param := ToolParam{
		Name:        "test_tool",
		Description: "Test tool description",
		Schema: map[string]interface{}{
			"type": "object",
		},
	}

	if param.Name != "test_tool" {
		t.Errorf("Name = %s; want test_tool", param.Name)
	}

	if param.Description != "Test tool description" {
		t.Errorf("Description = %s; want Test tool description", param.Description)
	}
}

func TestRealtimeEvent(t *testing.T) {
	event := RealtimeEvent{
		Type:    "agent_response",
		Content: map[string]interface{}{"text": "Hello!"},
	}

	if event.Type != "agent_response" {
		t.Errorf("Type = %s; want agent_response", event.Type)
	}

	if event.Content["text"] != "Hello!" {
		t.Errorf("Content[text] = %v; want Hello!", event.Content["text"])
	}
}

func TestTextResult(t *testing.T) {
	result := TextResult{
		Text: "Response text",
	}

	if result.Text != "Response text" {
		t.Errorf("Text = %s; want Response text", result.Text)
	}
}

func TestThinkingBlock(t *testing.T) {
	block := ThinkingBlock{
		Thinking: "Let me analyze this...",
	}

	if block.Thinking != "Let me analyze this..." {
		t.Errorf("Thinking = %s; want Let me analyze this...", block.Thinking)
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	if msg.Role != "user" {
		t.Errorf("Role = %s; want user", msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("Content = %s; want Hello, world!", msg.Content)
	}
}

func TestMessageWithComplexContent(t *testing.T) {
	msg := Message{
		Role: "assistant",
		Content: []map[string]interface{}{
			{"type": "text", "text": "Response"},
			{"type": "tool_call", "tool_name": "test", "tool_input": map[string]interface{}{}},
		},
	}

	if msg.Role != "assistant" {
		t.Errorf("Role = %s; want assistant", msg.Role)
	}

	contentSlice, ok := msg.Content.([]map[string]interface{})
	if !ok {
		t.Fatal("Content should be []map[string]interface{}")
	}

	if len(contentSlice) != 2 {
		t.Errorf("Content length = %d; want 2", len(contentSlice))
	}
}

func TestLLMClientInterface(t *testing.T) {
	// Verify LLMClient is an interface
	var _ LLMClient = (*mockLLMClient)(nil)
}

type mockLLMClient struct{}

func (m *mockLLMClient) Generate(ctx context.Context, messages []Message, maxTokens int, tools []ToolParam, systemPrompt string) ([]interface{}, error) {
	return nil, nil
}

func TestLLMToolInterface(t *testing.T) {
	// Verify LLMTool is an interface
	var _ LLMTool = (*mockLLMTool)(nil)
}

type mockLLMTool struct{}

func (m *mockLLMTool) GetToolParam() ToolParam {
	return ToolParam{Name: "mock"}
}

func (m *mockLLMTool) Run(ctx context.Context, input map[string]interface{}, history MessageHistory) (ToolImplOutput, error) {
	return ToolImplOutput{}, nil
}

func TestToolManagerInterface(t *testing.T) {
	// Verify ToolManager is an interface
	var _ ToolManager = (*mockToolManager)(nil)
}

type mockToolManager struct{}

func (m *mockToolManager) GetTools() []LLMTool {
	return nil
}

func (m *mockToolManager) RunTool(ctx context.Context, call ToolCallParameters, history MessageHistory) (ToolImplOutput, error) {
	return ToolImplOutput{}, nil
}

func (m *mockToolManager) Reset() {}

func TestMessageHistoryInterface(t *testing.T) {
	// Verify MessageHistory is an interface
	var _ MessageHistory = (*mockMessageHistory)(nil)
}

type mockMessageHistory struct{}

func (m *mockMessageHistory) AddUserPrompt(prompt string, images []interface{}) {}
func (m *mockMessageHistory) AddAssistantTurn(responses []interface{})          {}
func (m *mockMessageHistory) AddToolCallResult(toolCall ToolCallParameters, result string) {}
func (m *mockMessageHistory) GetMessagesForLLM() []Message                        { return nil }
func (m *mockMessageHistory) GetPendingToolCalls() []ToolCallParameters            { return nil }
func (m *mockMessageHistory) GetLastAssistantTextResponse() string               { return "" }
func (m *mockMessageHistory) Clear()                                              {}
func (m *mockMessageHistory) Truncate()                                            {}
func (m *mockMessageHistory) CountTokens() int                                     { return 0 }
func (m *mockMessageHistory) IsNextTurnUser() bool                                 { return true }

func TestContextManagerInterface(t *testing.T) {
	// Verify ContextManager is an interface
	var _ ContextManager = (*mockContextManager)(nil)
}

type mockContextManager struct{}

func (m *mockContextManager) CountTokens(messages []Message) int {
	return 0
}

func (m *mockContextManager) ApplyTruncationIfNeeded(messages []Message) []Message {
	return messages
}

func (m *mockContextManager) GetMaxContextLength() int {
	return 100000
}

func TestWorkspaceManagerInterface(t *testing.T) {
	// Verify WorkspaceManager is an interface
	var _ WorkspaceManager = (*mockWorkspaceManager)(nil)
}

type mockWorkspaceManager struct{}

func (m *mockWorkspaceManager) RelativePath(path string) string {
	return path
}

func (m *mockWorkspaceManager) WorkspacePath(path string) string {
	return path
}

func (m *mockWorkspaceManager) SessionID() string {
	return "test-session"
}

func TestWebSocketInterface(t *testing.T) {
	// Verify WebSocket is an interface
	var _ WebSocket = (*mockWebSocket)(nil)
}

type mockWebSocket struct{}

func (m *mockWebSocket) SendJSON(v interface{}) error {
	return nil
}

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		eventType string
		expected string
	}{
		{"UserMessage", EventTypeUserMessage, "user_message"},
		{"AgentResponse", EventTypeAgentResponse, "agent_response"},
		{"AgentThinking", EventTypeAgentThinking, "agent_thinking"},
		{"ToolCall", EventTypeToolCall, "tool_call"},
		{"ToolResult", EventTypeToolResult, "tool_result"},
		{"ResponseInterrupt", EventTypeResponseInterrupt, "agent_response_interrupted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("EventType = %s; want %s", tt.eventType, tt.expected)
			}
		})
	}
}
