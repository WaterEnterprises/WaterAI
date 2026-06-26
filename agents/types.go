package agents

import (
	"context"
)

// EventType constants matching ii_agent
const (
	EventTypeUserMessage       = "user_message"
	EventTypeAgentResponse     = "agent_response"
	EventTypeAgentThinking     = "agent_thinking"
	EventTypeToolCall          = "tool_call"
	EventTypeToolResult        = "tool_result"
	EventTypeResponseInterrupt = "agent_response_interrupted"
)

// --- Tooling & LLM Interfaces ---

// ToolImplOutput represents the result of a tool execution.
// Added IsFinal to mimic ToolManager.should_stop() logic.
type ToolImplOutput struct {
	ToolOutput        string
	ToolResultMessage string
	IsFinal           bool   // If true, the agent loop terminates (e.g., finish_task)
}

// ToolCallParameters represents a request from the LLM to call a tool
type ToolCallParameters struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

// ToolParam describes the tool definition sent to the LLM
type ToolParam struct {
	Name        string
	Description string
	Schema      map[string]interface{}
}

// LLMTool represents the base interface for any tool (including Agents)
type LLMTool interface {
	GetToolParam() ToolParam
	// Run executes the tool. Logic typically found in run_impl.
	Run(ctx context.Context, input map[string]interface{}, history MessageHistory) (ToolImplOutput, error)
}

// ToolManager encapsulates tool execution and state (reset, should_stop)
type ToolManager interface {
	GetTools() []LLMTool
	RunTool(ctx context.Context, call ToolCallParameters, history MessageHistory) (ToolImplOutput, error)
	Reset()
}

// LLMClient interface for generating responses
type LLMClient interface {
	Generate(ctx context.Context, messages []Message, maxTokens int, tools []ToolParam, systemPrompt string) ([]interface{}, error)
}

// --- Message History & Context ---

type Message struct {
	Role    string
	Content interface{} // string or []map[string]interface{} (for images)
}

type MessageHistory interface {
	AddUserPrompt(prompt string, images []interface{})
	AddAssistantTurn(responses []interface{})
	AddToolCallResult(toolCall ToolCallParameters, result string)
	GetMessagesForLLM() []Message
	GetPendingToolCalls() []ToolCallParameters
	GetLastAssistantTextResponse() string
	Clear()
	Truncate()
	CountTokens() int
	IsNextTurnUser() bool // for Resume logic
}

type ContextManager interface {
	CountTokens(messages []Message) int
	ApplyTruncationIfNeeded(messages []Message) []Message
	GetMaxContextLength() int
}

// --- Workspace & Environment ---

type WorkspaceManager interface {
	RelativePath(path string) string
	WorkspacePath(path string) string
	SessionID() string
}

// --- Events & Communication ---

type RealtimeEvent struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

type WebSocket interface {
	SendJSON(v interface{}) error
}

// --- LLM Result Types ---

type TextResult struct {
	Text string
}

type ThinkingBlock struct {
	Thinking string
}