package tools

import (
	"context"
	"encoding/json"
)

// --- Sequential Thinking Tool ---
type SequentialThinkingTool struct {
	History []map[string]interface{}
}

func (t *SequentialThinkingTool) Name() string        { return "sequential_thinking" }
func (t *SequentialThinkingTool) Description() string { return "Break down complex problems step-by-step." }
func (t *SequentialThinkingTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"thought":           map[string]string{"type": "string"},
			"thoughtNumber":     map[string]string{"type": "integer"},
			"totalThoughts":     map[string]string{"type": "integer"},
			"nextThoughtNeeded": map[string]string{"type": "boolean"},
		},
		"required": []string{"thought", "thoughtNumber", "totalThoughts"},
	}
}

func (t *SequentialThinkingTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	t.History = append(t.History, input)
	
	historyJSON, _ := json.MarshalIndent(t.History, "", "  ")
	
	return ToolResult{
		Output:        string(historyJSON),
		ResultMessage: "Thought recorded",
		Success:       true,
	}, nil
}

// --- Complete Tool ---
type CompleteTool struct{}

func (t *CompleteTool) Name() string        { return "complete" }
func (t *CompleteTool) Description() string { return "Signal task completion." }
func (t *CompleteTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"answer": map[string]string{"type": "string", "description": "Final answer or summary"},
		},
		"required": []string{"answer"},
	}
}

func (t *CompleteTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	answer, _ := input["answer"].(string)
	return ToolResult{
		Output:        answer,
		ResultMessage: "Task Completed",
		Success:       true,
	}, nil
}

// --- Message Tool ---
type MessageTool struct{}

func (t *MessageTool) Name() string        { return "message_user" }
func (t *MessageTool) Description() string { return "Send a message to the user." }
func (t *MessageTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]string{"type": "string"},
		},
		"required": []string{"text"},
	}
}

func (t *MessageTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	text, _ := input["text"].(string)
	return ToolResult{
		Output:        text,
		ResultMessage: "Message sent to user",
		Success:       true,
	}, nil
}