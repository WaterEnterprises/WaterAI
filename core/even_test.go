package core

import (
	"encoding/json"
	"testing"
)

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		eventType EventType
		expected string
	}{
		{"ConnectionEstablished", EventTypeConnectionEstablished, "connection_established"},
		{"AgentInitialized", EventTypeAgentInitialized, "agent_initialized"},
		{"WorkspaceInfo", EventTypeWorkspaceInfo, "workspace_info"},
		{"Processing", EventTypeProcessing, "processing"},
		{"AgentThinking", EventTypeAgentThinking, "agent_thinking"},
		{"ToolCall", EventTypeToolCall, "tool_call"},
		{"ToolResult", EventTypeToolResult, "tool_result"},
		{"AgentResponse", EventTypeAgentResponse, "agent_response"},
		{"AgentResponseInterrupted", EventTypeAgentResponseInterrupted, "agent_response_interrupted"},
		{"StreamComplete", EventTypeStreamComplete, "stream_complete"},
		{"Error", EventTypeError, "error"},
		{"System", EventTypeSystem, "system"},
		{"Pong", EventTypePong, "pong"},
		{"UploadSuccess", EventTypeUploadSuccess, "upload_success"},
		{"BrowserUse", EventTypeBrowserUse, "browser_use"},
		{"FileEdit", EventTypeFileEdit, "file_edit"},
		{"UserMessage", EventTypeUserMessage, "user_message"},
		{"PromptGenerated", EventTypePromptGenerated, "prompt_generated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("EventType %s = %s; want %s", tt.name, tt.eventType, tt.expected)
			}
		})
	}
}

func TestRealtimeEventCreation(t *testing.T) {
	event := RealtimeEvent{
		Type: EventTypeAgentResponse,
		Content: map[string]any{
			"text": "Hello, world!",
		},
	}

	if event.Type != EventTypeAgentResponse {
		t.Errorf("Expected Type = %s; got %s", EventTypeAgentResponse, event.Type)
	}

	if event.Content["text"] != "Hello, world!" {
		t.Errorf("Expected Content[text] = 'Hello, world!'; got %v", event.Content["text"])
	}
}

func TestRealtimeEventJSONMarshaling(t *testing.T) {
	event := RealtimeEvent{
		Type: EventTypeToolCall,
		Content: map[string]any{
			"tool_name": "terminal_execute",
			"arguments": map[string]any{
				"command": "ls -la",
			},
		},
	}

	// Test that the event can be marshaled to JSON
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal RealtimeEvent: %v", err)
	}

	// Test that the JSON can be unmarshaled back
	var decoded RealtimeEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RealtimeEvent: %v", err)
	}

	if decoded.Type != event.Type {
		t.Errorf("Decoded Type = %s; want %s", decoded.Type, event.Type)
	}

	if decoded.Content["tool_name"] != event.Content["tool_name"] {
		t.Errorf("Decoded Content[tool_name] = %v; want %v", decoded.Content["tool_name"], event.Content["tool_name"])
	}
}

func TestRealtimeEventEmptyContent(t *testing.T) {
	event := RealtimeEvent{
		Type:    EventTypeSystem,
		Content: nil,
	}

	if event.Type != EventTypeSystem {
		t.Errorf("Expected Type = %s; got %s", EventTypeSystem, event.Type)
	}

	if event.Content != nil {
		t.Errorf("Expected Content = nil; got %v", event.Content)
	}
}

func TestRealtimeEventWithMultipleContent(t *testing.T) {
	event := RealtimeEvent{
		Type: EventTypeAgentResponse,
		Content: map[string]any{
			"text":         "Response text",
			"is_complete": true,
			"turn":         5,
		},
	}

	// Verify all content fields
	text, ok := event.Content["text"].(string)
	if !ok || text != "Response text" {
		t.Errorf("Expected text = 'Response text'; got %v", event.Content["text"])
	}

	isComplete, ok := event.Content["is_complete"].(bool)
	if !ok || !isComplete {
		t.Errorf("Expected is_complete = true; got %v", event.Content["is_complete"])
	}

	turn, ok := event.Content["turn"].(int)
	if !ok || turn != 5 {
		t.Errorf("Expected turn = 5; got %v", event.Content["turn"])
	}
}
