package agents

import (
	"context"
	"errors"
)

// BaseAgent provides common fields for all agents.
type BaseAgent struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// GetToolParam returns the definition required for the LLM to use this agent as a tool.
func (b *BaseAgent) GetToolParam() ToolParam {
	return ToolParam{
		Name:        b.Name,
		Description: b.Description,
		Schema:      b.InputSchema,
	}
}

// Run is the interface method. Concrete agents (Reviewer, FunctionCall) must override this.
func (b *BaseAgent) Run(ctx context.Context, input map[string]interface{}, history MessageHistory) (ToolImplOutput, error) {
	return ToolImplOutput{}, errors.New("Run method not implemented in base agent")
}