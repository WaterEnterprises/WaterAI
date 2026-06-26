package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAPITypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		apiType  APIType
		expected string
	}{
		{"OpenAI", APITypeOpenAI, "openai"},
		{"Anthropic", APITypeAnthropic, "anthropic"},
		{"Gemini", APITypeGemini, "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.apiType) != tt.expected {
				t.Errorf("APIType = %s; want %s", tt.apiType, tt.expected)
			}
		})
	}
}

func TestContentBlockText(t *testing.T) {
	block := &ContentBlock{
		Type: ContentTypeText,
		Text: "Hello, world!",
	}

	if block.Type != ContentTypeText {
		t.Errorf("Type = %s; want %s", block.Type, ContentTypeText)
	}

	if block.Text != "Hello, world!" {
		t.Errorf("Text = %s; want Hello, world!", block.Text)
	}
}

func TestContentBlockImage(t *testing.T) {
	source := &ImageSource{
		Type:      "base64",
		MediaType: "image/png",
		Data:      "iVBORw0KGgo=",
	}

	block := &ContentBlock{
		Type:   ContentTypeImage,
		Source: source,
	}

	if block.Type != ContentTypeImage {
		t.Errorf("Type = %s; want %s", block.Type, ContentTypeImage)
	}

	if block.Source == nil || block.Source.Data != "iVBORw0KGgo=" {
		t.Error("Source was not set correctly")
	}
}

func TestContentBlockToolCall(t *testing.T) {
	block := &ContentBlock{
		Type:       ContentTypeToolCall,
		ToolCallID: "call-123",
		ToolName:   "terminal_execute",
		ToolInput: map[string]interface{}{
			"command": "ls -la",
		},
	}

	if block.Type != ContentTypeToolCall {
		t.Errorf("Type = %s; want %s", block.Type, ContentTypeToolCall)
	}

	if block.ToolCallID != "call-123" {
		t.Errorf("ToolCallID = %s; want call-123", block.ToolCallID)
	}

	if block.ToolName != "terminal_execute" {
		t.Errorf("ToolName = %s; want terminal_execute", block.ToolName)
	}
}

func TestContentBlockToolResult(t *testing.T) {
	block := &ContentBlock{
		Type:       ContentTypeToolResult,
		ToolCallID: "call-123",
		ToolName:   "terminal_execute",
		ToolOutput: "total 4\ndrwxr-xr-x  2 user user  4096 Jan  1 00:00 .",
	}

	if block.Type != ContentTypeToolResult {
		t.Errorf("Type = %s; want %s", block.Type, ContentTypeToolResult)
	}

	output, ok := block.ToolOutput.(string)
	if !ok || output != "total 4\ndrwxr-xr-x  2 user user  4096 Jan  1 00:00 ." {
		t.Errorf("ToolOutput = %v; want string output", block.ToolOutput)
	}
}

func TestContentBlockThinking(t *testing.T) {
	block := &ContentBlock{
		Type:      ContentTypeThinking,
		Thinking:  "Let me analyze this problem...",
		Signature: "abc123",
	}

	if block.Type != ContentTypeThinking {
		t.Errorf("Type = %s; want %s", block.Type, ContentTypeThinking)
	}

	if block.Thinking != "Let me analyze this problem..." {
		t.Errorf("Thinking = %s; want analysis text", block.Thinking)
	}
}

func TestGetClientOpenAI(t *testing.T) {
	cfg := LLMConfig{
		APIType: APITypeOpenAI,
		Model:   "gpt-4",
	}

	client, err := GetClient(cfg)
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("GetClient() returned nil")
	}
}

func TestGetClientAnthropic(t *testing.T) {
	cfg := LLMConfig{
		APIType: APITypeAnthropic,
		Model:   "claude-sonnet-4-20250514",
	}

	client, err := GetClient(cfg)
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("GetClient() returned nil")
	}
}

func TestGetClientGemini(t *testing.T) {
	cfg := LLMConfig{
		APIType: APITypeGemini,
		Model:   "gemini-pro",
	}

	client, err := GetClient(cfg)
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("GetClient() returned nil")
	}
}

func TestGetClientUnknown(t *testing.T) {
	cfg := LLMConfig{
		APIType: APIType("unknown"),
		Model:   "some-model",
	}

	_, err := GetClient(cfg)
	if err == nil {
		t.Error("GetClient() should return error for unknown API type")
	}
}

func TestNewMessageHistory(t *testing.T) {
	history := NewMessageHistory()

	if history == nil {
		t.Fatal("NewMessageHistory() returned nil")
	}

	if history.Messages == nil {
		t.Error("Messages should not be nil")
	}

	if len(history.Messages) != 0 {
		t.Errorf("Messages length = %d; want 0", len(history.Messages))
	}
}

func TestMessageHistoryAddUserPrompt(t *testing.T) {
	history := NewMessageHistory()

	history.AddUserPrompt("Hello, world!", nil)

	if len(history.Messages) != 1 {
		t.Errorf("Messages length = %d; want 1", len(history.Messages))
	}

	msg := history.Messages[0]
	if msg.Role != "user" {
		t.Errorf("Role = %s; want user", msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Errorf("Content length = %d; want 1", len(msg.Content))
	}

	if msg.Content[0].Type != ContentTypeText {
		t.Errorf("Content[0].Type = %s; want %s", msg.Content[0].Type, ContentTypeText)
	}

	if msg.Content[0].Text != "Hello, world!" {
		t.Errorf("Content[0].Text = %s; want Hello, world!", msg.Content[0].Text)
	}
}

func TestMessageHistoryAddUserPromptWithImages(t *testing.T) {
	history := NewMessageHistory()

	images := []*ImageSource{
		{Type: "base64", MediaType: "image/png", Data: "image-data-1"},
	}

	history.AddUserPrompt("Look at this image!", images)

	if len(history.Messages) != 1 {
		t.Errorf("Messages length = %d; want 1", len(history.Messages))
	}

	msg := history.Messages[0]
	if len(msg.Content) != 2 {
		t.Errorf("Content length = %d; want 2", len(msg.Content))
	}

	if msg.Content[0].Type != ContentTypeImage {
		t.Errorf("Content[0].Type = %s; want %s", msg.Content[0].Type, ContentTypeImage)
	}

	if msg.Content[1].Type != ContentTypeText {
		t.Errorf("Content[1].Type = %s; want %s", msg.Content[1].Type, ContentTypeText)
	}

	if msg.Content[1].Text != "Look at this image!" {
		t.Errorf("Content[1].Text = %s; want Look at this image!", msg.Content[1].Text)
	}
}

func TestMessageHistoryAddAssistantTurn(t *testing.T) {
	history := NewMessageHistory()

	blocks := []*ContentBlock{
		{Type: ContentTypeText, Text: "I can help you with that."},
	}

	history.AddAssistantTurn(blocks)

	if len(history.Messages) != 1 {
		t.Errorf("Messages length = %d; want 1", len(history.Messages))
	}

	msg := history.Messages[0]
	if msg.Role != "assistant" {
		t.Errorf("Role = %s; want assistant", msg.Role)
	}
}

func TestMessageHistoryAddToolResult(t *testing.T) {
	history := NewMessageHistory()

	history.AddToolResult("call-123", "terminal_execute", "command output")

	if len(history.Messages) != 1 {
		t.Errorf("Messages length = %d; want 1", len(history.Messages))
	}

	msg := history.Messages[0]
	if msg.Role != "user" {
		t.Errorf("Role = %s; want user", msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Errorf("Content length = %d; want 1", len(msg.Content))
	}

	if msg.Content[0].Type != ContentTypeToolResult {
		t.Errorf("Content[0].Type = %s; want %s", msg.Content[0].Type, ContentTypeToolResult)
	}
}

func TestMessageHistoryGetMessages(t *testing.T) {
	history := NewMessageHistory()

	history.AddUserPrompt("Hello!", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "Hi there!"}})

	msgs := history.GetMessages()

	if len(msgs) != 2 {
		t.Errorf("GetMessages() returned %d messages; want 2", len(msgs))
	}
}

func TestMessageHistoryClear(t *testing.T) {
	history := NewMessageHistory()

	history.AddUserPrompt("Hello!", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "Hi there!"}})

	history.Clear()

	if len(history.Messages) != 0 {
		t.Errorf("Messages length after Clear() = %d; want 0", len(history.Messages))
	}
}

func TestMessageHistoryEnsureToolCallIntegrity(t *testing.T) {
	history := NewMessageHistory()

	// Add user prompt
	history.AddUserPrompt("Run a command", nil)

	// Add assistant turn with tool call
	toolCallBlock := &ContentBlock{
		Type:       ContentTypeToolCall,
		ToolCallID: "call-1",
		ToolName:   "terminal_execute",
		ToolInput:  map[string]interface{}{"command": "ls"},
	}
	history.AddAssistantTurn([]*ContentBlock{toolCallBlock})

	// Add tool result
	history.AddToolResult("call-1", "terminal_execute", "output")

	// Call integrity check
	history.EnsureToolCallIntegrity()

	// Verify the history still contains the tool call and result
	if len(history.Messages) != 3 {
		t.Errorf("Messages length = %d; want 3", len(history.Messages))
	}
}

func TestMessageHistorySaveToFile(t *testing.T) {
	history := NewMessageHistory()

	history.AddUserPrompt("Hello!", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "Hi there!"}})

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "history.json")

	err := history.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("SaveToFile() should create the file")
	}
}

func TestMessageHistoryLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "history.json")

	// Create a history file
	history := NewMessageHistory()
	history.AddUserPrompt("Hello!", nil)
	err := history.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Load into new history
	newHistory := NewMessageHistory()
	err = newHistory.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if len(newHistory.Messages) != 1 {
		t.Errorf("Loaded history has %d messages; want 1", len(newHistory.Messages))
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID("test")
	id2 := generateID("test")

	if id1 == id2 {
		t.Error("generateID() should generate unique IDs")
	}

	// Check format: prefix_timestamp_random
	if len(id1) < 10 {
		t.Errorf("ID too short: %s", id1)
	}
}

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"short", "hello", 1},
		{"medium", "hello world", 2},
		{"long", "this is a longer text for testing token counting", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countTokens(tt.text)
			// Token counting is approximate: len/3
			if result != tt.expected {
				t.Errorf("countTokens(%s) = %d; want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestMessageHistoryWithMultipleTurns(t *testing.T) {
	history := NewMessageHistory()

	// Simulate a conversation
	history.AddUserPrompt("What is 2+2?", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "2+2 equals 4."}})

	history.AddUserPrompt("What about 3+3?", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "3+3 equals 6."}})

	if len(history.Messages) != 4 {
		t.Errorf("Messages length = %d; want 4", len(history.Messages))
	}

	if history.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %s; want user", history.Messages[0].Role)
	}

	if history.Messages[1].Role != "assistant" {
		t.Errorf("Messages[1].Role = %s; want assistant", history.Messages[1].Role)
	}
}

func TestContentBlockJSONMarshaling(t *testing.T) {
	block := &ContentBlock{
		Type:       ContentTypeToolCall,
		ToolCallID: "call-123",
		ToolName:   "test_tool",
		ToolInput:  map[string]interface{}{"param": "value"},
	}

	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("Failed to marshal ContentBlock: %v", err)
	}

	var decoded ContentBlock
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ContentBlock: %v", err)
	}

	if decoded.Type != block.Type {
		t.Errorf("Decoded Type = %s; want %s", decoded.Type, block.Type)
	}

	if decoded.ToolCallID != block.ToolCallID {
		t.Errorf("Decoded ToolCallID = %s; want %s", decoded.ToolCallID, block.ToolCallID)
	}
}

func TestMessageJSONMarshaling(t *testing.T) {
	msg := &Message{
		Role: "user",
		Content: []*ContentBlock{
			{Type: ContentTypeText, Text: "Hello!"},
			{Type: ContentTypeImage, Source: &ImageSource{Type: "base64", MediaType: "image/png", Data: "data"}},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal Message: %v", err)
	}

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message: %v", err)
	}

	if decoded.Role != msg.Role {
		t.Errorf("Decoded Role = %s; want %s", decoded.Role, msg.Role)
	}

	if len(decoded.Content) != 2 {
		t.Errorf("Decoded Content length = %d; want 2", len(decoded.Content))
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	history := NewMessageHistory()

	history.AddUserPrompt("Hello!", nil)
	history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "Hi there!"}})

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_history.json")

	// Save
	err := history.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Load into new instance
	loadedHistory := NewMessageHistory()
	err = loadedHistory.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Verify
	if len(loadedHistory.Messages) != 2 {
		t.Errorf("Loaded history has %d messages; want 2", len(loadedHistory.Messages))
	}

	// Verify content
	if loadedHistory.Messages[0].Content[0].Text != "Hello!" {
		t.Errorf("Loaded message[0].Content[0].Text = %s; want Hello!", loadedHistory.Messages[0].Content[0].Text)
	}
}

func TestMessageHistoryPreservesOrder(t *testing.T) {
	history := NewMessageHistory()

	expectedOrder := []string{"Hello", "How are you?", "I'm doing well, thanks!"}
	for _, msg := range expectedOrder {
		history.AddUserPrompt(msg, nil)
		history.AddAssistantTurn([]*ContentBlock{{Type: ContentTypeText, Text: "Response to: " + msg}})
	}

	msgs := history.GetMessages()

	if len(msgs) != 6 {
		t.Fatalf("Expected 6 messages, got %d", len(msgs))
	}

	// Verify alternating user/assistant pattern
	for i, msg := range msgs {
		expectedRole := "user"
		if i%2 == 1 {
			expectedRole = "assistant"
		}
		if msg.Role != expectedRole {
			t.Errorf("Message %d role = %s; want %s", i, msg.Role, expectedRole)
		}
	}
}

func TestMessageHistoryToolResultIntegration(t *testing.T) {
	history := NewMessageHistory()

	// Add user prompt
	history.AddUserPrompt("List files", nil)

	// Add assistant with tool call
	toolCall := &ContentBlock{
		Type:       ContentTypeToolCall,
		ToolCallID: "call-1",
		ToolName:   "terminal_execute",
		ToolInput:  map[string]interface{}{"command": "ls"},
	}
	history.AddAssistantTurn([]*ContentBlock{toolCall})

	// Add tool result
	history.AddToolResult("call-1", "terminal_execute", "file1.txt\nfile2.txt")

	// Verify structure
	if len(history.Messages) != 3 {
		t.Errorf("Messages length = %d; want 3", len(history.Messages))
	}

	// Verify tool result
	resultMsg := history.Messages[2]
	if resultMsg.Role != "user" {
		t.Errorf("Tool result message role = %s; want user", resultMsg.Role)
	}

	if resultMsg.Content[0].Type != ContentTypeToolResult {
		t.Errorf("Tool result content type = %s; want %s", resultMsg.Content[0].Type, ContentTypeToolResult)
	}
}

func TestLoadFromFileNonExistent(t *testing.T) {
	history := NewMessageHistory()

	err := history.LoadFromFile("/non/existent/path.json")
	if err == nil {
		t.Error("LoadFromFile() should return error for non-existent file")
	}
}

func TestSaveToFileCreatesDirectories(t *testing.T) {
	history := NewMessageHistory()
	history.AddUserPrompt("Test", nil)

	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "dirs")
	filePath := filepath.Join(nestedDir, "history.json")

	err := history.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("SaveToFile() should create nested directories")
	}
}
