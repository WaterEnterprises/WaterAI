package core

// EventType represents the category of the system event.
type EventType string

const (
	EventTypeConnectionEstablished      EventType = "connection_established"
	EventTypeAgentInitialized           EventType = "agent_initialized"
	EventTypeWorkspaceInfo              EventType = "workspace_info"
	EventTypeProcessing                 EventType = "processing"
	EventTypeAgentThinking              EventType = "agent_thinking"
	EventTypeToolCall                   EventType = "tool_call"
	EventTypeToolResult                 EventType = "tool_result"
	EventTypeAgentResponse              EventType = "agent_response"
	EventTypeAgentResponseInterrupted   EventType = "agent_response_interrupted"
	EventTypeStreamComplete             EventType = "stream_complete"
	EventTypeError                      EventType = "error"
	EventTypeSystem                     EventType = "system"
	EventTypePong                       EventType = "pong"
	EventTypeUploadSuccess              EventType = "upload_success"
	EventTypeBrowserUse                 EventType = "browser_use"
	EventTypeFileEdit                   EventType = "file_edit"
	EventTypeUserMessage                EventType = "user_message"
	EventTypePromptGenerated            EventType = "prompt_generated"
)

// RealtimeEvent represents a unified event structure exchanging data.
// Content uses map[string]any to replicate Python's dict[str, Any].
type RealtimeEvent struct {
	Type    EventType      `json:"type"`
	Content map[string]any `json:"content"`
}