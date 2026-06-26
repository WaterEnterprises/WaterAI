package contextmanager

import (
	"context"
	"log/slog"
	"testing"
	"strings" // Added for cleaner contains check
)

// MockTokenCounter implements TokenCounter for testing
type MockTokenCounter struct {
	countFunc func(text string) int
}

func (m *MockTokenCounter) CountTokens(text string) int {
	if m.countFunc != nil {
		return m.countFunc(text)
	}
	return len(text) / 3
}

// MockLLMClient implements LLMClient for testing
type MockLLMClient struct {
	generateFunc func(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error)
}

func (m *MockLLMClient) Generate(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, messages, maxTokens, temperature)
	}
	return []ContentBlock{TextResult{Text: "Mock summary"}}, nil
}

func TestTextPromptType(t *testing.T) {
	prompt := TextPrompt{Text: "Hello"}
	if prompt.Type() != "TextPrompt" {
		t.Errorf("Type() = %s; want TextPrompt", prompt.Type())
	}
}

func TestTextResultType(t *testing.T) {
	result := TextResult{Text: "Response"}
	if result.Type() != "TextResult" {
		t.Errorf("Type() = %s; want TextResult", result.Type())
	}
}

func TestToolCallType(t *testing.T) {
	call := ToolCall{ToolInput: map[string]interface{}{"key": "value"}}
	if call.Type() != "ToolCall" {
		t.Errorf("Type() = %s; want ToolCall", call.Type())
	}
}

func TestToolFormattedResultType(t *testing.T) {
	result := ToolFormattedResult{ToolOutput: "output"}
	if result.Type() != "ToolFormattedResult" {
		t.Errorf("Type() = %s; want ToolFormattedResult", result.Type())
	}
}

func TestImageBlockType(t *testing.T) {
	block := ImageBlock{}
	if block.Type() != "ImageBlock" {
		t.Errorf("Type() = %s; want ImageBlock", block.Type())
	}
}

func TestAnthropicThinkingBlockType(t *testing.T) {
	block := AnthropicThinkingBlock{Thinking: "thinking..."}
	if block.Type() != "AnthropicThinkingBlock" {
		t.Errorf("Type() = %s; want AnthropicThinkingBlock", block.Type())
	}
}

func TestAnthropicRedactedThinkingBlockType(t *testing.T) {
	block := AnthropicRedactedThinkingBlock{}
	if block.Type() != "AnthropicRedactedThinkingBlock" {
		t.Errorf("Type() = %s; want AnthropicRedactedThinkingBlock", block.Type())
	}
}

func TestNewManager(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	if manager == nil {
		t.Fatal("New() returned nil")
	}

	// Check default config
	if manager.config.TokenBudget != DefaultTokenBudget {
		t.Errorf("TokenBudget = %d; want %d", manager.config.TokenBudget, DefaultTokenBudget)
	}

	if manager.config.MaxSize != DefaultMaxSize {
		t.Errorf("MaxSize = %d; want %d", manager.config.MaxSize, DefaultMaxSize)
	}
}

func TestNewManagerWithCustomConfig(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		TokenBudget:    10000,
		MaxSize:        50,
		MaxEventLength: 5000,
	}

	manager := New(client, counter, logger, cfg)

	if manager.config.TokenBudget != 10000 {
		t.Errorf("TokenBudget = %d; want 10000", manager.config.TokenBudget)
	}

	if manager.config.MaxSize != 50 {
		t.Errorf("MaxSize = %d; want 50", manager.config.MaxSize)
	}
}

func TestNewManagerWithInvalidMaxSize(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		MaxSize: 0,
	}

	manager := New(client, counter, logger, cfg)

	if manager.config.MaxSize != 1 {
		t.Errorf("MaxSize = %d; want 1 (minimum)", manager.config.MaxSize)
	}
}

func TestManagerCountTokens(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{countFunc: func(text string) int { return len(text) }}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "Hello world"}},
		{TextResult{Text: "Response"}},
	}

	tokens := manager.CountTokens(messageLists)

	if tokens < 19 { // Minimal check based on mock logic
		t.Errorf("CountTokens() = %d; want >= 19", tokens)
	}
}

func TestManagerCountTokensWithImages(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	messageLists := [][]ContentBlock{
		{ImageBlock{}},
	}

	tokens := manager.CountTokens(messageLists)

	if tokens != ImageTokenCost {
		t.Errorf("CountTokens() with image = %d; want %d", tokens, ImageTokenCost)
	}
}

func TestManagerCountTokensWithThinking(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	messageLists := [][]ContentBlock{
		{AnthropicThinkingBlock{Thinking: "thinking content"}},
	}

	tokens := manager.CountTokens(messageLists)

	if tokens == 0 {
		t.Error("CountTokens() should count thinking content in last turn")
	}
}

func TestApplyTruncationIfNeededNoTruncation(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		TokenBudget: 10000,
		MaxSize:     100,
	}

	manager := New(client, counter, logger, cfg)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "Hello"}},
		{TextResult{Text: "Hi"}},
	}

	result, err := manager.ApplyTruncationIfNeeded(context.Background(), messageLists)
	if err != nil {
		t.Fatalf("ApplyTruncationIfNeeded() error = %v", err)
	}

	if len(result) != len(messageLists) {
		t.Errorf("Length changed = %d; want %d", len(result), len(messageLists))
	}
}

func TestApplyTruncationIfNeededWithTruncation(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		TokenBudget: 10,
		MaxSize:     2,
	}

	manager := New(client, counter, logger, cfg)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "Hello world this is a long message"}},
		{TextResult{Text: "Response one"}},
		{TextPrompt{Text: "Second prompt"}},
		{TextResult{Text: "Response two"}},
	}

	result, err := manager.ApplyTruncationIfNeeded(context.Background(), messageLists)
	if err != nil {
		t.Fatalf("ApplyTruncationIfNeeded() error = %v", err)
	}

	if len(result) >= len(messageLists) {
		t.Error("Truncation should reduce message count")
	}
}

func TestApplyTruncationIfNeededExceedsMaxSize(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		TokenBudget: 100000,
		MaxSize:     2,
	}

	manager := New(client, counter, logger, cfg)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "First"}},
		{TextResult{Text: "Second"}},
		{TextPrompt{Text: "Third"}},
		{TextResult{Text: "Fourth"}},
	}

	result, err := manager.ApplyTruncationIfNeeded(context.Background(), messageLists)
	if err != nil {
		t.Fatalf("ApplyTruncationIfNeeded() error = %v", err)
	}

	if len(result) > cfg.MaxSize {
		t.Errorf("Result length = %d; want <= %d", len(result), cfg.MaxSize)
	}
}

func TestHasThinkingBlocks(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	tests := []struct {
		name        string
		messages    [][]ContentBlock
		hasThinking bool
	}{
		{"no_thinking", [][]ContentBlock{{TextPrompt{Text: "Hello"}}}, false},
		{"has_thinking", [][]ContentBlock{{AnthropicThinkingBlock{Thinking: "thinking"}}}, true},
		{"has_redacted", [][]ContentBlock{{AnthropicRedactedThinkingBlock{}}}, true},
		{"mixed", [][]ContentBlock{{TextPrompt{Text: "Hello"}, AnthropicThinkingBlock{Thinking: "thinking"}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.hasThinkingBlocks(tt.messages)
			if result != tt.hasThinking {
				t.Errorf("hasThinkingBlocks() = %v; want %v", result, tt.hasThinking)
			}
		})
	}
}

func TestFindLastTextPromptIndex(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	tests := []struct {
		name     string
		messages [][]ContentBlock
		expected int
	}{
		{"single", [][]ContentBlock{{TextPrompt{Text: "Hello"}}}, 0},
		{"result_first", [][]ContentBlock{{TextResult{Text: "Response"}}, {TextPrompt{Text: "Hello"}}}, 1},
		{"no_prompt", [][]ContentBlock{{TextResult{Text: "Response"}}}, 0},
		{"last_prompt", [][]ContentBlock{{TextResult{Text: "R1"}}, {TextResult{Text: "R2"}}, {TextPrompt{Text: "Last"}}}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.findLastTextPromptIndex(tt.messages)
			if result != tt.expected {
				t.Errorf("findLastTextPromptIndex() = %d; want %d", result, tt.expected)
			}
		})
	}
}

func TestTruncateContent(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	cfg := &Config{
		MaxEventLength: 10,
	}

	manager := New(client, counter, logger, cfg)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"short", "Hello", "Hello"},
		{"exact_length", "1234567890", "1234567890"},
		{"long", "12345678901", "1234567890... [truncated]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.truncateContent(tt.content)
			if result != tt.expected {
				t.Errorf("truncateContent(%s) = %s; want %s", tt.content, result, tt.expected)
			}
		})
	}
}

func TestMessageListToString(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	tests := []struct {
		name     string
		list     []ContentBlock
		contains []string
	}{
		{"text_prompt", []ContentBlock{TextPrompt{Text: "Hello"}}, []string{"USER: Hello"}},
		{"text_result", []ContentBlock{TextResult{Text: "Response"}}, []string{"ASSISTANT: Response"}},
		{"thinking", []ContentBlock{AnthropicThinkingBlock{Thinking: "thinking..."}}, []string{"ASSISTANT (Thinking): thinking..."}},
		{"redacted", []ContentBlock{AnthropicRedactedThinkingBlock{}}, []string{}}, // Fixed brace syntax
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.messageListToString(tt.list)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("messageListToString() should contain %q; got %s", expected, result)
				}
			}
		})
	}
}

func TestGenerateCompleteConversationSummaryEmpty(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{}

	manager := New(client, counter, logger, nil)

	result, err := manager.GenerateCompleteConversationSummary(context.Background(), [][]ContentBlock{})
	if err != nil {
		t.Fatalf("GenerateCompleteConversationSummary() error = %v", err)
	}

	expected := "No conversation history to summarize."
	if result != expected {
		t.Errorf("Summary = %s; want %s", result, expected)
	}
}

func TestGenerateCompleteConversationSummary(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{
		generateFunc: func(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error) {
			return []ContentBlock{TextResult{Text: "Summary of conversation"}}, nil
		},
	}

	manager := New(client, counter, logger, nil)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "Hello"}},
		{TextResult{Text: "Hi there!"}},
	}

	result, err := manager.GenerateCompleteConversationSummary(context.Background(), messageLists)
	if err != nil {
		t.Fatalf("GenerateCompleteConversationSummary() error = %v", err)
	}

	if result == "" {
		t.Error("Summary should not be empty")
	}
}

func TestTruncateStandard(t *testing.T) {
	logger := slog.Default()
	counter := &MockTokenCounter{}
	client := &MockLLMClient{
		generateFunc: func(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error) {
			return []ContentBlock{TextResult{Text: "Generated summary"}}, nil
		},
	}

	cfg := &Config{
		TokenBudget:    100000,
		MaxSize:        10,
		MaxEventLength: 1000,
	}

	manager := New(client, counter, logger, cfg)

	messageLists := [][]ContentBlock{
		{TextPrompt{Text: "First"}},
		{TextResult{Text: "Response 1"}},
		{TextPrompt{Text: "Second"}},
		{TextResult{Text: "Response 2"}},
		{TextPrompt{Text: "Third"}},
		{TextResult{Text: "Response 3"}},
	}

	result, err := manager.truncateStandard(context.Background(), messageLists)
	if err != nil {
		t.Fatalf("truncateStandard() error = %v", err)
	}

	if len(result) >= len(messageLists) {
		t.Error("Standard truncation should reduce message count")
	}
}