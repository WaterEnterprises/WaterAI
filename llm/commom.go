package llm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// ==========================================
// CONFIGURATION
// ==========================================

type APIType string

const (
	APITypeOpenAI    APIType = "openai"
	APITypeAnthropic APIType = "anthropic"
	APITypeGemini    APIType = "gemini"
)

type LLMConfig struct {
	APIType          APIType
	Model            string
	APIKey           string
	BaseURL          string // Optional
	MaxRetries       int
	AzureEndpoint    string // Optional
	AzureAPIVersion  string // Optional
	VertexProjectID  string // Optional
	VertexRegion     string // Optional
	ThinkingTokens   int    // Optional (Anthropic)
	CotModel         bool   // Optional (OpenAI o1/o3)
}

// ==========================================
// TYPES & DATA STRUCTURES
// ==========================================

type ContentType string

const (
	ContentTypeText             ContentType = "text"
	ContentTypeImage            ContentType = "image"
	ContentTypeToolCall         ContentType = "tool_call"
	ContentTypeToolResult       ContentType = "tool_result"
	ContentTypeThinking         ContentType = "thinking"
	ContentTypeRedactedThinking ContentType = "redacted_thinking"
)

// ContentBlock is a unified structure handling all types of content
type ContentBlock struct {
	Type ContentType `json:"type"`

	// Text Content
	Text string `json:"text,omitempty"`

	// Image Content
	Source *ImageSource `json:"source,omitempty"`

	// Tool Call
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolInput  map[string]interface{} `json:"tool_input,omitempty"`

	// Tool Result
	ToolOutput interface{} `json:"tool_output,omitempty"` // string or []ContentBlock

	// Thinking (Anthropic)
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
	Data      string `json:"data,omitempty"` // For redacted thinking
}

type ImageSource struct {
	Type      string `json:"type"` // e.g. "base64"
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ToolParam struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type Message struct {
	Role    string          `json:"role"` // "user", "assistant", "model", "tool"
	Content []*ContentBlock `json:"content"`
}

type GenerateResponse struct {
	Content []*ContentBlock
	Usage   UsageMetadata
}

type UsageMetadata struct {
	InputTokens  int
	OutputTokens int
	RawResponse  interface{}
}

type ToolChoice struct {
	Type string // "auto", "any", "tool"
	Name string // Used if Type is "tool"
}

// ==========================================
// INTERFACE & FACTORY
// ==========================================

type Client interface {
	Generate(
		messages []*Message,
		maxTokens int,
		systemPrompt string,
		temperature float64,
		tools []*ToolParam,
		toolChoice *ToolChoice,
		thinkingTokens *int,
	) (*GenerateResponse, error)
}

func GetClient(cfg LLMConfig) (Client, error) {
	switch cfg.APIType {
	case APITypeOpenAI:
		return NewOpenAIClient(cfg), nil
	case APITypeAnthropic:
		return NewAnthropicClient(cfg), nil
	case APITypeGemini:
		return NewGeminiClient(cfg), nil
	default:
		return nil, fmt.Errorf("unknown api type: %s", cfg.APIType)
	}
}

// ==========================================
// MESSAGE HISTORY
// ==========================================

type MessageHistory struct {
	Messages []*Message
}

func NewMessageHistory() *MessageHistory {
	return &MessageHistory{
		Messages: make([]*Message, 0),
	}
}

func (h *MessageHistory) AddUserPrompt(prompt string, images []*ImageSource) {
	blocks := []*ContentBlock{}
	
	if images != nil {
		for _, img := range images {
			blocks = append(blocks, &ContentBlock{
				Type:   ContentTypeImage,
				Source: img,
			})
		}
	}
	
	blocks = append(blocks, &ContentBlock{
		Type: ContentTypeText,
		Text: prompt,
	})

	h.Messages = append(h.Messages, &Message{
		Role:    "user",
		Content: blocks,
	})
}

func (h *MessageHistory) AddAssistantTurn(blocks []*ContentBlock) {
	// Simple validation to ensure only one tool call per turn if strictness is required,
	// though Go slices handle multiple fine.
	h.Messages = append(h.Messages, &Message{
		Role:    "assistant",
		Content: blocks,
	})
}

func (h *MessageHistory) AddToolResult(toolCallID, toolName string, output interface{}) {
	block := &ContentBlock{
		Type:       ContentTypeToolResult,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		ToolOutput: output,
	}
	// Tool results are typically user-side messages in many APIs (or "tool" role)
	// We append a new message turn for results
	h.Messages = append(h.Messages, &Message{
		Role:    "user", // Or "tool", logic handled in specific clients
		Content: []*ContentBlock{block},
	})
}

func (h *MessageHistory) GetMessages() []*Message {
	// Deep copy could be implemented here if needed
	return h.Messages
}

func (h *MessageHistory) Clear() {
	h.Messages = make([]*Message, 0)
}

// EnsureToolCallIntegrity removes tool calls that don't have matching results and vice versa.
// Simplified version of the Python logic.
func (h *MessageHistory) EnsureToolCallIntegrity() {
	callIDs := make(map[string]bool)
	resultIDs := make(map[string]bool)

	// Pass 1: Collect IDs
	for _, msg := range h.Messages {
		for _, block := range msg.Content {
			if block.Type == ContentTypeToolCall {
				callIDs[block.ToolCallID] = true
			}
			if block.Type == ContentTypeToolResult {
				resultIDs[block.ToolCallID] = true
			}
		}
	}

	validIDs := make(map[string]bool)
	for id := range callIDs {
		if resultIDs[id] {
			validIDs[id] = true
		}
	}

	// Pass 2: Filter
	var cleanMessages []*Message
	for _, msg := range h.Messages {
		var cleanBlocks []*ContentBlock
		// keepMsg := true
		
		for _, block := range msg.Content {
			if block.Type == ContentTypeToolCall {
				if validIDs[block.ToolCallID] {
					cleanBlocks = append(cleanBlocks, block)
				}
			} else if block.Type == ContentTypeToolResult {
				if validIDs[block.ToolCallID] {
					cleanBlocks = append(cleanBlocks, block)
				}
			} else {
				cleanBlocks = append(cleanBlocks, block)
			}
		}

		if len(cleanBlocks) > 0 {
			msg.Content = cleanBlocks
			cleanMessages = append(cleanMessages, msg)
		}
	}
	h.Messages = cleanMessages
}

// Save/Load using JSON instead of Pickle
func (h *MessageHistory) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(h.Messages, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (h *MessageHistory) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.Messages)
}

// ==========================================
// UTILS
// ==========================================

func generateID(prefix string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	random := rand.Intn(9000) + 1000
	return fmt.Sprintf("%s_%d_%d", prefix, timestamp, random)
}

func countTokens(text string) int {
	// Rough approximation: 1 token ~= 4 chars
	return len(text) / 4
}