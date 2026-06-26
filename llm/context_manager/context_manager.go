package contextmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// Defaults based on the original code
const (
	DefaultTokenBudget     = 20000 // Placeholder for TOKEN_BUDGET
	DefaultSummaryMaxToken = 4000
	DefaultMaxSize         = 100
	DefaultMaxEventLength  = 10000
	KeepFirst              = 1
	ImageTokenCost         = 1000
)

const summaryPromptTemplate = `
Your task is to create a detailed summary of the conversation so far, paying close attention to the user's explicit requests and your previous actions.
This summary should be thorough in capturing technical details, code patterns, and architectural decisions that would be essential for continuing development work without losing context.

[... Full prompt redacted for brevity, insert original prompt text here ...]

Now summarize the events using the rules above.`

// ============================================================================
// Types & Interfaces
// ============================================================================

// ContentBlock is a marker interface for all message block types.
// In a real integration, these might be imported from water/llm/base.
type ContentBlock interface {
	Type() string
}

// Concrete block types
type TextPrompt struct {
	Text string
}
func (t TextPrompt) Type() string { return "TextPrompt" }

type TextResult struct {
	Text string
}
func (t TextResult) Type() string { return "TextResult" }

type ToolCall struct {
	ToolInput interface{}
}
func (t ToolCall) Type() string { return "ToolCall" }

type ToolFormattedResult struct {
	ToolOutput string
}
func (t ToolFormattedResult) Type() string { return "ToolFormattedResult" }

type ImageBlock struct {
	// Image data omitted for brevity
}
func (t ImageBlock) Type() string { return "ImageBlock" }

type AnthropicThinkingBlock struct {
	Thinking string
}
func (t AnthropicThinkingBlock) Type() string { return "AnthropicThinkingBlock" }

type AnthropicRedactedThinkingBlock struct{}
func (t AnthropicRedactedThinkingBlock) Type() string { return "AnthropicRedactedThinkingBlock" }

// TokenCounter abstracts the token counting logic.
type TokenCounter interface {
	CountTokens(text string) int
}

// LLMClient abstracts the generation capability.
type LLMClient interface {
	Generate(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error)
}

// Config holds configuration for the manager.
type Config struct {
	TokenBudget    int
	MaxSize        int
	MaxEventLength int
}

// ============================================================================
// Context Manager Implementation
// ============================================================================

type Manager struct {
	client       LLMClient
	tokenCounter TokenCounter
	logger       *slog.Logger
	config       Config
}

// New creates a new ContextManager.
func New(client LLMClient, counter TokenCounter, logger *slog.Logger, cfg *Config) *Manager {
	if cfg == nil {
		cfg = &Config{
			TokenBudget:    DefaultTokenBudget,
			MaxSize:        DefaultMaxSize,
			MaxEventLength: DefaultMaxEventLength,
		}
	}
	if cfg.MaxSize < 1 {
		cfg.MaxSize = 1
	}

	return &Manager{
		client:       client,
		tokenCounter: counter,
		logger:       logger,
		config:       *cfg,
	}
}

// CountTokens counts tokens in the conversation history.
// It ignores thinking blocks unless they are in the very last turn.
func (m *Manager) CountTokens(messageLists [][]ContentBlock) int {
	totalTokens := 0
	numTurns := len(messageLists)

	for i, messageList := range messageLists {
		isLastTurn := i == numTurns-1
		for _, msg := range messageList {
			switch v := msg.(type) {
			case TextPrompt:
				totalTokens += m.tokenCounter.CountTokens(v.Text)
			case TextResult:
				totalTokens += m.tokenCounter.CountTokens(v.Text)
			case ToolFormattedResult:
				totalTokens += m.tokenCounter.CountTokens(v.ToolOutput)
			case ToolCall:
				// Basic counting of input JSON
				bytes, err := json.Marshal(v.ToolInput)
				if err != nil {
					m.logger.Warn("Could not serialize tool input for token counting", "error", err)
					totalTokens += 100 // Arbitrary penalty
				} else {
					totalTokens += m.tokenCounter.CountTokens(string(bytes))
				}
			case ImageBlock:
				totalTokens += ImageTokenCost
			case AnthropicRedactedThinkingBlock:
				// Always 0
			case AnthropicThinkingBlock:
				if isLastTurn {
					totalTokens += m.tokenCounter.CountTokens(v.Thinking)
				}
			default:
				m.logger.Warn("Unhandled message type for token counting", "type", fmt.Sprintf("%T", msg))
			}
		}
	}
	return totalTokens
}

// ApplyTruncationIfNeeded checks if truncation is required and applies it.
func (m *Manager) ApplyTruncationIfNeeded(ctx context.Context, messageLists [][]ContentBlock) ([][]ContentBlock, error) {
	currentCount := m.CountTokens(messageLists)
	
	// Check if we exceed budget OR max number of turns
	if currentCount <= m.config.TokenBudget && len(messageLists) <= m.config.MaxSize {
		return messageLists, nil
	}

	m.logger.Warn("Token limit or max size exceeded, applying truncation", 
		"current_tokens", currentCount, 
		"turns", len(messageLists), 
		"budget", m.config.TokenBudget)

	truncatedLists, err := m.applyTruncation(ctx, messageLists)
	if err != nil {
		return messageLists, err
	}

	newCount := m.CountTokens(truncatedLists)
	m.logger.Info("Truncation completed", "saved_tokens", currentCount-newCount, "new_count", newCount)

	return truncatedLists, nil
}

// applyTruncation routes to the specific truncation strategy.
func (m *Manager) applyTruncation(ctx context.Context, messageLists [][]ContentBlock) ([][]ContentBlock, error) {
	if m.hasThinkingBlocks(messageLists) {
		return m.truncateWithThinkingBlocks(ctx, messageLists)
	}
	return m.truncateStandard(ctx, messageLists)
}

// truncateWithThinkingBlocks applies logic preserving context around the last user prompt.
func (m *Manager) truncateWithThinkingBlocks(ctx context.Context, messageLists [][]ContentBlock) ([][]ContentBlock, error) {
	lastPromptIdx := m.findLastTextPromptIndex(messageLists)

	if lastPromptIdx <= 0 {
		return messageLists, nil
	}

	targetSize := min(m.config.MaxSize, len(messageLists)) / 2
	
	// Ensure we don't cut past the last prompt
	lastSummaryIdx := min(lastPromptIdx, KeepFirst+targetSize)

	eventsToSummarize := messageLists[KeepFirst:lastSummaryIdx]
	eventsToKeep := messageLists[lastSummaryIdx:]

	if len(eventsToSummarize) <= 1 {
		m.logger.Info("Not enough events to summarize")
		return messageLists, nil
	}

	summary, err := m.generateSummary(ctx, eventsToSummarize, "No events summarized")
	if err != nil {
		return nil, err
	}

	// Rebuild conversation: Head + Summary + Tail (from last prompt onwards)
	result := make([][]ContentBlock, 0)
	result = append(result, messageLists[:KeepFirst]...)
	result = append(result, []ContentBlock{TextResult{Text: "Conversation Summary: " + summary}})
	result = append(result, eventsToKeep...)

	m.logger.Info("Truncated with thinking blocks", 
		"original_len", len(messageLists), 
		"new_len", len(result))

	return result, nil
}

// truncateStandard applies standard sliding window summarization.
func (m *Manager) truncateStandard(ctx context.Context, messageLists [][]ContentBlock) ([][]ContentBlock, error) {
	head := messageLists[:KeepFirst]
	targetSize := min(m.config.MaxSize, len(messageLists)) / 2
	
	// Calculate how many items to keep from the end
	eventsFromTail := targetSize - len(head) - 1
	if eventsFromTail < 0 {
		eventsFromTail = 0
	}

	// Determine where to start summarizing. 
	// If a summary already exists at Head+1, we might merge into it.
	summaryStartIdx := KeepFirst
	prevSummaryContent := "No events summarized"

	// Check for existing summary (Simple heuristic: Second message is a TextResult starting with "Conversation Summary")
	if len(messageLists) > KeepFirst && len(messageLists[KeepFirst]) > 0 {
		if tr, ok := messageLists[KeepFirst][0].(TextResult); ok {
			if strings.HasPrefix(tr.Text, "Conversation Summary:") {
				prevSummaryContent = tr.Text
				summaryStartIdx = KeepFirst + 1
			} else if tp, ok := messageLists[KeepFirst][0].(TextPrompt); ok {
				// The python code checks TextPrompt for summary, though usually summary is Assistant (TextResult).
				// We support the Python logic here.
				if strings.HasPrefix(tp.Text, "Conversation Summary:") {
					prevSummaryContent = tp.Text
					summaryStartIdx = KeepFirst + 1
				}
			}
		}
	}

	endIdx := len(messageLists)
	if eventsFromTail > 0 {
		endIdx = len(messageLists) - eventsFromTail
	}

	forgottenEvents := messageLists[summaryStartIdx:endIdx]

	if len(forgottenEvents) == 0 {
		return messageLists, nil
	}

	summary, err := m.generateSummary(ctx, forgottenEvents, prevSummaryContent)
	if err != nil {
		return nil, err
	}

	// Rebuild conversation
	result := make([][]ContentBlock, 0)
	result = append(result, head...)
	result = append(result, []ContentBlock{TextResult{Text: "Conversation Summary: " + summary}})
	
	if eventsFromTail > 0 {
		result = append(result, messageLists[len(messageLists)-eventsFromTail:]...)
	}

	m.logger.Info("Standard truncation applied", 
		"original_len", len(messageLists), 
		"new_len", len(result))

	return result, nil
}

// generateSummary calls the LLM to summarize specific events.
func (m *Manager) generateSummary(ctx context.Context, events [][]ContentBlock, prevSummary string) (string, error) {
	var sb strings.Builder
	sb.WriteString(summaryPromptTemplate)

	// Clean previous summary tag
	cleanPrev := strings.Replace(prevSummary, "Conversation Summary: ", "", 1)
	if cleanPrev == "No events summarized" {
		cleanPrev = ""
	}
	
	fmt.Fprintf(&sb, "<PREVIOUS SUMMARY>\n%s\n</PREVIOUS SUMMARY>\n\n", m.truncateContent(cleanPrev))

	for i, event := range events {
		eventContent := m.messageListToString(event)
		fmt.Fprintf(&sb, "<EVENT id=%d>\n%s\n</EVENT>\n", i, m.truncateContent(eventContent))
	}

	sb.WriteString("\nNow summarize the events using the rules above.")

	prompt := []ContentBlock{TextPrompt{Text: sb.String()}}
	
	// Call LLM
	response, err := m.client.Generate(ctx, [][]ContentBlock{prompt}, DefaultSummaryMaxToken, 0.0)
	if err != nil {
		m.logger.Error("Failed to generate summary", "error", err)
		return fmt.Sprintf("Failed to summarize %d events due to error: %v", len(events), err), nil
	}

	summary := ""
	for _, block := range response {
		if txt, ok := block.(TextResult); ok {
			summary += txt.Text
		}
	}

	return summary, nil
}

// GenerateCompleteConversationSummary creates a summary of the entire history (for /compact commands).
func (m *Manager) GenerateCompleteConversationSummary(ctx context.Context, messageLists [][]ContentBlock) (string, error) {
	if len(messageLists) == 0 {
		return "No conversation history to summarize.", nil
	}

	var sb strings.Builder
	sb.WriteString(summaryPromptTemplate)
	sb.WriteString("<CONVERSATION>\n")
	
	for i, list := range messageLists {
		content := m.messageListToString(list)
		fmt.Fprintf(&sb, "<TURN id=%d>\n%s\n</TURN>\n\n", i, content)
	}
	sb.WriteString("</CONVERSATION>\n\n")
	sb.WriteString("Now summarize the conversation using the rules above.")

	prompt := []ContentBlock{TextPrompt{Text: sb.String()}}

	response, err := m.client.Generate(ctx, [][]ContentBlock{prompt}, DefaultSummaryMaxToken, 0.0)
	if err != nil {
		return "", err
	}

	summary := ""
	for _, block := range response {
		if txt, ok := block.(TextResult); ok {
			summary += txt.Text
		}
	}
	return summary, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (m *Manager) truncateContent(content string) string {
	if len(content) <= m.config.MaxEventLength {
		return content
	}
	return content[:m.config.MaxEventLength] + "... [truncated]"
}

func (m *Manager) messageListToString(list []ContentBlock) string {
	var parts []string
	for _, msg := range list {
		switch v := msg.(type) {
		case TextPrompt:
			parts = append(parts, "USER: "+v.Text)
		case TextResult:
			parts = append(parts, "ASSISTANT: "+v.Text)
		case AnthropicThinkingBlock:
			parts = append(parts, "ASSISTANT (Thinking): "+v.Thinking)
		case AnthropicRedactedThinkingBlock:
			continue
		case ToolCall:
			js, _ := json.Marshal(v.ToolInput)
			parts = append(parts, "ToolCall: "+string(js))
		case ToolFormattedResult:
			parts = append(parts, "ToolResult: "+v.ToolOutput)
		default:
			parts = append(parts, fmt.Sprintf("%T: %v", v, v))
		}
	}
	return strings.Join(parts, "\n")
}

func (m *Manager) hasThinkingBlocks(lists [][]ContentBlock) bool {
	for _, list := range lists {
		for _, msg := range list {
			switch msg.(type) {
			case AnthropicThinkingBlock, AnthropicRedactedThinkingBlock:
				return true
			}
		}
	}
	return false
}

func (m *Manager) findLastTextPromptIndex(lists [][]ContentBlock) int {
	for i := len(lists) - 1; i >= 0; i-- {
		for _, msg := range lists[i] {
			if _, ok := msg.(TextPrompt); ok {
				return i
			}
		}
	}
	return len(lists) - 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}