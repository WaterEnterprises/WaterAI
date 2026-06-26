package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type SimpleMemoryTool struct {
	mu      sync.Mutex
	content string
}

func NewMemoryTool() *SimpleMemoryTool {
	return &SimpleMemoryTool{}
}

func (t *SimpleMemoryTool) Name() string { return "simple_memory" }
func (t *SimpleMemoryTool) Description() string { return "Read/Write/Edit persistent string memory" }
func (t *SimpleMemoryTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action":     map[string]interface{}{"type": "string", "enum": []string{"read", "write", "edit"}},
			"content":    map[string]interface{}{"type": "string"},
			"old_string": map[string]interface{}{"type": "string"},
			"new_string": map[string]interface{}{"type": "string"},
		},
		"required": []string{"action"},
	}
}

func (t *SimpleMemoryTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	action, _ := GetArg[string](input, "action")

	switch action {
	case "read":
		return &ToolOutput{Text: t.content}, nil
	case "write":
		newContent, _ := GetArg[string](input, "content")
		prev := t.content
		t.content = newContent
		return &ToolOutput{Text: fmt.Sprintf("Memory overwritten. Previous size: %d", len(prev))}, nil
	case "edit":
		oldStr, _ := GetArg[string](input, "old_string")
		newStr, _ := GetArg[string](input, "new_string")
		if strings.Count(t.content, oldStr) != 1 {
			return ErrorOutput(fmt.Errorf("old_string must occur exactly once, found %d", strings.Count(t.content, oldStr))), nil
		}
		t.content = strings.Replace(t.content, oldStr, newStr, 1)
		return &ToolOutput{Text: "Memory edited successfully"}, nil
	}
	return ErrorOutput(fmt.Errorf("invalid action")), nil
}

// CompactifyMemoryTool in Go is usually logic inside the Agent loop rather than a tool, 
// but if exposed as a tool:
type CompactifyMemoryTool struct {}

func (t *CompactifyMemoryTool) Name() string { return "compactify_memory" }
func (t *CompactifyMemoryTool) Description() string { return "Trigger context truncation" }
func (t *CompactifyMemoryTool) InputSchema() map[string]interface{} { return map[string]interface{}{} }
func (t *CompactifyMemoryTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	return &ToolOutput{
		Text: "Memory compaction signal sent.", 
		Auxiliary: map[string]interface{}{"success": true, "signal": "truncate_history"},
	}, nil
}