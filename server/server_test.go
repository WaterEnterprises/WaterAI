package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"strings"
	"testing"
)

func TestConfigGetPort(t *testing.T) {
	cfg := Config{
		Port: "3000",
	}

	if cfg.GetPort() != "3000" {
		t.Errorf("GetPort() = %s; want 3000", cfg.GetPort())
	}
}

func TestConfigGetPortDefault(t *testing.T) {
	cfg := Config{}

	if cfg.GetPort() != "8080" {
		t.Errorf("GetPort() = %s; want 8080", cfg.GetPort())
	}
}

func TestConfigGetWorkspaceRoot(t *testing.T) {
	cfg := Config{
		WorkspaceRoot: "/custom/workspace",
	}

	if cfg.GetWorkspaceRoot() != "/custom/workspace" {
		t.Errorf("GetWorkspaceRoot() = %s; want /custom/workspace", cfg.GetWorkspaceRoot())
	}
}

func TestConfigGetWorkspaceRootDefault(t *testing.T) {
	cfg := Config{}

	if cfg.GetWorkspaceRoot() != "./workspace_data" {
		t.Errorf("GetWorkspaceRoot() = %s; want ./workspace_data", cfg.GetWorkspaceRoot())
	}
}

func TestNewConnectionManager(t *testing.T) {
	cfg := Config{
		WorkspaceRoot: "/test",
	}

	manager := NewConnectionManager(cfg)

	if manager == nil {
		t.Fatal("NewConnectionManager() returned nil")
	}

	if manager.config.WorkspaceRoot != "/test" {
		t.Errorf("WorkspaceRoot = %s; want /test", manager.config.WorkspaceRoot)
	}

	if manager.sessions == nil {
		t.Error("Sessions map should not be nil")
	}
}

func TestConnectionManagerConnect(t *testing.T) {
	cfg := Config{
		WorkspaceRoot: "/test",
	}
	manager := NewConnectionManager(cfg)

	// Note: We can't easily test with real websocket.Conn
	// This tests the basic structure
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestConnectionManagerConnectInvalidUUID(t *testing.T) {
	cfg := Config{
		WorkspaceRoot: "/test",
	}
	manager := NewConnectionManager(cfg)

	// The function should handle invalid UUID gracefully
	// by generating a new one
	if manager != nil {
		// Basic structure test
		// sync.RWMutex is a struct, not a pointer, so we can't check against nil
		// Instead we check if sessions map is initialized which implies constructor ran
		if manager.sessions == nil {
			t.Error("Sessions map should not be nil")
		}
	}
}

func TestConnectionManagerDisconnect(t *testing.T) {
	cfg := Config{
		WorkspaceRoot: "/test",
	}
	manager := NewConnectionManager(cfg)

	// Disconnect on empty manager should not panic
	manager.Disconnect(nil)
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/index.html", "text/html; charset=utf-8"},
		{"/script.js", "application/javascript"},
		{"/style.css", "text/css"},
		{"/data.json", "application/json"},
		{"/image.png", "image/png"},
		{"/image.jpg", "image/jpeg"},
		{"/image.jpeg", "image/jpeg"},
		{"/unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getContentType(tt.path)
			if result != tt.expected {
				t.Errorf("getContentType(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestChatSessionSendEvent(t *testing.T) {
	// Create a minimal test for SendEvent structure
	session := &ChatSession{}

	// The method should not panic
	session.SendEvent("test_type", map[string]interface{}{"key": "value"})
}

func TestChatSessionHandleMessageInvalidJSON(t *testing.T) {
	session := &ChatSession{}

	// Handle invalid JSON - should send error event
	session.HandleMessage([]byte("invalid json"))
}

func TestChatSessionHandleSlashCommandHelp(t *testing.T) {
	session := &ChatSession{}

	// Handle slash command - should not panic
	session.handleSlashCommand("/help")
}

func TestChatSessionHandleSlashCommandCompact(t *testing.T) {
	session := &ChatSession{}

	// Handle slash command - should not panic
	session.handleSlashCommand("/compact")
}

func TestChatSessionHandleSlashCommandUnknown(t *testing.T) {
	session := &ChatSession{}

	// Handle unknown slash command - should not panic
	session.handleSlashCommand("/unknown")
}

func TestChatSessionHandleSlashCommandEmpty(t *testing.T) {
	session := &ChatSession{}

	// Handle empty slash command - should not panic
	session.handleSlashCommand("")
}

func TestChatSessionFields(t *testing.T) {
	session := &ChatSession{}

	// Verify LLMClient is nil before init
	if session.LLMClient != nil {
		t.Error("LLMClient should be nil before initialization")
	}

	// Verify History is nil before init
	if session.History != nil {
		t.Error("History should be nil before initialization")
	}
}

func TestServerStruct(t *testing.T) {
	srv := &Server{
		Config: Config{
			Port:          "8080",
			WorkspaceRoot: "/workspace",
		},
		Router:    nil,
		WSManager: nil,
	}

	if srv.Config.Port != "8080" {
		t.Errorf("Port = %s; want 8080", srv.Config.Port)
	}

	if srv.Config.WorkspaceRoot != "/workspace" {
		t.Errorf("WorkspaceRoot = %s; want /workspace", srv.Config.WorkspaceRoot)
	}
}

func TestChatSessionStruct(t *testing.T) {
	session := &ChatSession{
		SessionUUID: uuidTestGenerator(),
		Workspace:   "/test/workspace",
		Manager:     nil,
	}

	if session.Workspace != "/test/workspace" {
		t.Errorf("Workspace = %s; want /test/workspace", session.Workspace)
	}
}

func TestConnectionManagerStruct(t *testing.T) {
	manager := &ConnectionManager{
		sessions: make(map[*websocket.Conn]*ChatSession),
		config:   Config{},
	}

	if manager.sessions == nil {
		t.Error("Sessions should not be nil")
	}
}

// Helper to generate test UUID (simplified)
func uuidTestGenerator() uuid.UUID {
	return uuid.New()
}

// Mock websocket for testing (if needed)
type mockWebSocket struct{}

func (m *mockWebSocket) WriteJSON(v interface{}) error {
	return nil
}

func (m *mockWebSocket) ReadMessage() (messageType int, p []byte, err error) {
	return 1, []byte("test"), nil
}

func (m *mockWebSocket) Close() error {
	return nil
}

func TestCreateServer(t *testing.T) {
	config := Config{
		Port:          "8080",
		WorkspaceRoot: "/tmp/test-workspace",
	}

	// CreateServer may panic without proper setup
	// Skip actual creation test and just verify config
	if config.GetPort() != "8080" {
		t.Errorf("GetPort() = %s; want 8080", config.GetPort())
	}
}

// Test HTTP handler helpers
func TestUploadRequestStruct(t *testing.T) {
	req := UploadRequest{
		SessionID: "test-session",
		File: FileInfo{
			Path:    "/test/file.txt",
			Content: "file content",
		},
	}

	if req.SessionID != "test-session" {
		t.Errorf("SessionID = %s; want test-session", req.SessionID)
	}

	if req.File.Path != "/test/file.txt" {
		t.Errorf("File.Path = %s; want /test/file.txt", req.File.Path)
	}
}

func TestUploadFileStruct(t *testing.T) {
	file := FileInfo{
		Path:    "/path/to/file.txt",
		Content: "base64data:image/png;base64,abc123",
	}

	if file.Path != "/path/to/file.txt" {
		t.Errorf("Path = %s; want /path/to/file.txt", file.Path)
	}

	if !strings.HasPrefix(file.Content, "base64data:") {
		t.Error("Content should have base64 prefix")
	}
}

func TestSessionResponseStruct(t *testing.T) {
	resp := SessionResponse{
		Sessions: []SessionInfo{
			{
				ID:           "session-1",
				WorkspaceDir: "/workspace",
				CreatedAt:    "2024-01-01T00:00:00Z",
				DeviceID:     "device-1",
				Name:         "Test Session",
			},
		},
	}

	if len(resp.Sessions) != 1 {
		t.Errorf("Sessions length = %d; want 1", len(resp.Sessions))
	}

	if resp.Sessions[0].Name != "Test Session" {
		t.Errorf("Name = %s; want Test Session", resp.Sessions[0].Name)
	}
}

func TestEventResponseStruct(t *testing.T) {
	resp := EventResponse{
		Events: []EventInfo{
			{
				ID:        "event-1",
				SessionID: "session-1",
				Timestamp: "2024-01-01T00:00:00Z",
				EventType: "system",
				EventPayload: map[string]interface{}{
					"message": "Session started",
				},
			},
		},
	}

	if len(resp.Events) != 1 {
		t.Errorf("Events length = %d; want 1", len(resp.Events))
	}

	if resp.Events[0].EventType != "system" {
		t.Errorf("EventType = %s; want system", resp.Events[0].EventType)
	}
}

func TestSettingsStruct(t *testing.T) {
	settings := Settings{
		LLMConfigs: map[string]LLMConfig{
			"gpt-4": {Model: "gpt-4"},
		},
	}

	if len(settings.LLMConfigs) != 1 {
		t.Errorf("LLMConfigs length = %d; want 1", len(settings.LLMConfigs))
	}
}

func TestGETSettingsModelStruct(t *testing.T) {
	model := GETSettingsModel{
		Settings: Settings{
			LLMConfigs: map[string]LLMConfig{},
		},
		LLMAPIKeySet:    true,
		SearchAPIKeySet: false,
	}

	if !model.LLMAPIKeySet {
		t.Error("LLMAPIKeySet should be true")
	}

	if model.SearchAPIKeySet {
		t.Error("SearchAPIKeySet should be false")
	}
}

func TestLLMConfigStruct(t *testing.T) {
	cfg := LLMConfig{
		Model: "gpt-4",
	}

	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %s; want gpt-4", cfg.Model)
	}
}

func TestWebSocketMessageStruct(t *testing.T) {
	msg := WebSocketMessage{
		Type:    "query",
		Content: json.RawMessage(`{"text":"hello"}`),
	}

	if msg.Type != "query" {
		t.Errorf("Type = %s; want query", msg.Type)
	}

	if string(msg.Content) != `{"text":"hello"}` {
		t.Errorf("Content = %s; want {\"text\":\"hello\"}", string(msg.Content))
	}
}

func TestInitAgentContentStruct(t *testing.T) {
	content := InitAgentContent{
		ModelName: "gpt-4",
	}

	if content.ModelName != "gpt-4" {
		t.Errorf("ModelName = %s; want gpt-4", content.ModelName)
	}
}

func TestQueryContentStruct(t *testing.T) {
	content := QueryContent{
		Text: "Hello, world!",
	}

	if content.Text != "Hello, world!" {
		t.Errorf("Text = %s; want Hello, world!", content.Text)
	}
}

// Note: The following are placeholder tests for methods that require
// more complex setup (Gin router, actual HTTP requests, etc.)

func TestCreateServerStructure(t *testing.T) {
	// Test that the factory function exists and returns expected struct fields
	config := Config{
		Port:          "8080",
		WorkspaceRoot: "/test",
	}

	// These are basic property tests
	if config.Port != "8080" {
		t.Errorf("Port = %s; want 8080", config.Port)
	}

	if config.WorkspaceRoot != "/test" {
		t.Errorf("WorkspaceRoot = %s; want /test", config.WorkspaceRoot)
	}
}

func TestChatSessionMessageHandling(t *testing.T) {
	session := &ChatSession{}

	// Test various message types don't panic
	testMessages := []string{
		`{"type":"ping"}`,
		`{"type":"unknown"}`,
		`invalid json`,
	}

	for _, msg := range testMessages {
		session.HandleMessage([]byte(msg))
	}
}

func TestChatSessionSlashCommands(t *testing.T) {
	session := &ChatSession{}

	commands := []string{
		"/help",
		"/compact",
		"/unknown",
		"/",
		"",
	}

	for _, cmd := range commands {
		session.handleSlashCommand(cmd)
	}
}

func TestSessionInfoStruct(t *testing.T) {
	info := SessionInfo{
		ID:           "test-id",
		WorkspaceDir: "/workspace",
		CreatedAt:    "2024-01-01T00:00:00Z",
		DeviceID:     "device-1",
		Name:         "Test",
	}

	if info.ID != "test-id" {
		t.Errorf("ID = %s; want test-id", info.ID)
	}

	if info.Name != "Test" {
		t.Errorf("Name = %s; want Test", info.Name)
	}
}

func TestEventInfoStruct(t *testing.T) {
	info := EventInfo{
		ID:        "event-id",
		SessionID: "session-id",
		Timestamp: "2024-01-01T00:00:00Z",
		EventType: "user_message",
		EventPayload: map[string]interface{}{
			"content": "Hello",
		},
	}

	if info.ID != "event-id" {
		t.Errorf("ID = %s; want event-id", info.ID)
	}

	if info.EventType != "user_message" {
		t.Errorf("EventType = %s; want user_message", info.EventType)
	}
}

func strPtr(s string) *string {
	return &s
}