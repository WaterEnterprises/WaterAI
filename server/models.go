package server

import "encoding/json"

// --- WebSocket Messages ---

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type RealtimeEvent struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// Event Types
const (
	EventTypeConnectionEstablished = "connection_established"
	EventTypeError                 = "error"
	EventTypeSystem                = "system"
	EventTypeProcessing            = "processing"
	EventTypeAgentResponse         = "agent_response"
	EventTypeStreamComplete        = "stream_complete"
	EventTypePong                  = "pong"
	EventTypeWorkspaceInfo         = "workspace_info"
	EventTypeAgentInitialized      = "agent_initialized"
)

// --- Request Content Models ---

type InitAgentContent struct {
	ModelName      string                 `json:"model_name"`
	ToolArgs       map[string]interface{} `json:"tool_args"`
	ThinkingTokens int                    `json:"thinking_tokens"`
}

type QueryContent struct {
	Text   string   `json:"text"`
	Resume bool     `json:"resume"`
	Files  []string `json:"files"`
}

type EditQueryContent struct {
	Text   string   `json:"text"`
	Resume bool     `json:"resume"`
	Files  []string `json:"files"`
}

type EnhancePromptContent struct {
	ModelName string   `json:"model_name"`
	Text      string   `json:"text"`
	Files     []string `json:"files"`
}

type ReviewResultContent struct {
	UserInput string `json:"user_input"`
}

// --- REST API Models ---

type FileInfo struct {
	Path    string `json:"path"`
	Content string `json:"content"` // Base64 or text
}

type UploadRequest struct {
	SessionID string   `json:"session_id"`
	File      FileInfo `json:"file"`
}

type SessionInfo struct {
	ID           string `json:"id"`
	WorkspaceDir string `json:"workspace_dir"`
	CreatedAt    string `json:"created_at"`
	DeviceID     string `json:"device_id"`
	Name         string `json:"name"`
}

type SessionResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type EventInfo struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	Timestamp    string                 `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	EventPayload map[string]interface{} `json:"event_payload"`
	WorkspaceDir string                 `json:"workspace_dir"`
}

type EventResponse struct {
	Events []EventInfo `json:"events"`
}

// Settings represents the application configuration
type Settings struct {
	LLMConfigs     map[string]LLMConfig `json:"llm_configs"`
	SearchConfig   *SearchConfig        `json:"search_config,omitempty"`
	// Additional fields omitted for brevity
}

type LLMConfig struct {
	APIKey         *string `json:"api_key,omitempty"`
	Model          string  `json:"model"`
	ThinkingTokens int     `json:"thinking_tokens,omitempty"`
}

type SearchConfig struct {
	APIKey *string `json:"api_key,omitempty"`
}

type GETSettingsModel struct {
	Settings
	LLMAPIKeySet    bool `json:"llm_api_key_set"`
	SearchAPIKeySet bool `json:"search_api_key_set"`
}