package agents

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

type ReviewerAgent struct {
	BaseAgent
	SystemPrompt    string
	Client          LLMClient
	Tools           []LLMTool // In Python this is wrapped in AgentToolManager
	MessageQueue    chan RealtimeEvent
	Logger          *log.Logger
	ContextManager  ContextManager
	MaxOutputTokens int
	MaxTurns        int
	Websocket       WebSocket
	History         MessageHistory
	
	interrupted      bool
	cachedToolParams []ToolParam
}

func NewReviewerAgent(
	systemPrompt string,
	client LLMClient,
	tools []LLMTool,
	messageQueue chan RealtimeEvent,
	logger *log.Logger,
	contextManager ContextManager,
	history MessageHistory,
	maxOutputTokens int,
	maxTurns int,
	websocket WebSocket,
) *ReviewerAgent {
	return &ReviewerAgent{
		BaseAgent: BaseAgent{
			Name: "reviewer_agent",
			Description: `A comprehensive reviewer agent that evaluates and reviews the results/websites/slides created by general agent, 
then provides detailed feedback and improvement suggestions with special focus on functionality testing.

This agent conducts thorough reviews with emphasis on:
- Testing ALL interactive elements (buttons, forms, navigation, etc.)
- Verifying website functionality and user experience
- Providing detailed, natural language feedback without format restrictions
- Identifying specific issues and areas for improvement
`,
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task":          map[string]string{"type": "string", "description": "The task that the general agent is trying to solve"},
					"workspace_dir": map[string]string{"type": "string", "description": "The workspace directory of the general agent execution to review"},
				},
				"required": []string{"task", "workspace_dir"},
			},
		},
		SystemPrompt:    systemPrompt,
		Client:          client,
		Tools:           tools,
		MessageQueue:    messageQueue,
		Logger:          logger,
		ContextManager:  contextManager,
		History:         history,
		MaxOutputTokens: maxOutputTokens,
		MaxTurns:        maxTurns,
		Websocket:       websocket,
	}
}

// StartMessageProcessing - Python implementation is empty pass
func (r *ReviewerAgent) StartMessageProcessing() {
	// No-op to match Python
}

func (r *ReviewerAgent) validateToolParameters() ([]ToolParam, error) {
	if r.cachedToolParams != nil {
		return r.cachedToolParams, nil
	}

	var params []ToolParam
	names := make([]string, 0)
	
	for _, tool := range r.Tools {
		p := tool.GetToolParam()
		params = append(params, p)
		names = append(names, p.Name)
	}

	sort.Strings(names)
	for i := 0; i < len(names)-1; i++ {
		if names[i] == names[i+1] {
			return nil, fmt.Errorf("tool %s is duplicated", names[i])
		}
	}

	r.cachedToolParams = params
	return params, nil
}

func (r *ReviewerAgent) generateLLMResponse(ctx context.Context, messages []Message, tools []ToolParam) ([]interface{}, error) {
	start := time.Now()
	
	// Centralized LLM response generation with timing metrics
	response, err := r.Client.Generate(ctx, messages, r.MaxOutputTokens, tools, r.SystemPrompt)
	
	elapsed := time.Since(start)
	r.Logger.Printf("LLM generation took %.2fs", elapsed.Seconds())
	return response, err
}

// Run implements the LLMTool interface (run_impl in Python)
func (r *ReviewerAgent) Run(ctx context.Context, toolInput map[string]interface{}, history MessageHistory) (ToolImplOutput, error) {
	task, _ := toolInput["task"].(string)
	workspaceDir, _ := toolInput["workspace_dir"].(string)
	result, _ := toolInput["result"].(string)
	
	userInputDelimiter := "--------------------------------------------- REVIEWER INPUT ---------------------------------------------"
	r.Logger.Printf("\n%s\nReviewing agent logs and output...\n", userInputDelimiter)

	reviewInstruction := fmt.Sprintf(`You are a reviewer agent tasked with evaluating the work done by an general agent. 
You have access to all the same tools that the general agent has.

Here is the task that the general agent is trying to solve:
%s

Here is the result of the general agent's execution:
%s

Here is the workspace directory of the general agent's execution:
%s

Now your turn to review the general agent's work.
`, task, result, workspaceDir)

	r.History.AddUserPrompt(reviewInstruction, nil)
	r.interrupted = false

	remainingTurns := r.MaxTurns

	for remainingTurns > 0 {
		select {
		case <-ctx.Done():
			r.interrupted = true
		default:
		}

		remainingTurns--
		delimiter := "--------------------------------------------- REVIEWER TURN ---------------------------------------------"
		r.Logger.Printf("\n%s\n", delimiter)

		toolParams, err := r.validateToolParameters()
		if err != nil {
			return ToolImplOutput{}, err
		}

		if r.interrupted {
			return ToolImplOutput{ToolOutput: "Reviewer interrupted", ToolResultMessage: "Reviewer interrupted by user"}, nil
		}

		currentMessages := r.History.GetMessagesForLLM()
		currentTokCount := r.ContextManager.CountTokens(currentMessages)
		r.Logger.Printf("(Current token count: %d)\n", currentTokCount)
		
		maxContext := r.ContextManager.GetMaxContextLength()
		if maxContext > 0 && float64(currentTokCount) > float64(maxContext)*0.9 {
			r.Logger.Printf("WARNING: Approaching token limit: %d/%d", currentTokCount, maxContext)
		}

		truncatedMessages := r.ContextManager.ApplyTruncationIfNeeded(currentMessages)
		
		// Note: Python sets history message list here, but in Go interfaces usually handle state internally.
		// We proceed with truncatedMessages for generation.

		modelResponse, err := r.generateLLMResponse(ctx, truncatedMessages, toolParams)
		if err != nil {
			return ToolImplOutput{ToolOutput: "Error calling LLM"}, err
		}

		if len(modelResponse) == 0 {
			modelResponse = []interface{}{TextResult{Text: "No response from model"}}
		}

		r.History.AddAssistantTurn(modelResponse)

		pendingTools := r.History.GetPendingToolCalls()
		if len(pendingTools) > 1 {
			return ToolImplOutput{}, errors.New("only one tool call per turn is supported")
		}

		if len(pendingTools) == 1 {
			toolCall := pendingTools[0]

			for _, item := range modelResponse {
				if tr, ok := item.(TextResult); ok {
					r.Logger.Printf("Reviewer planning next step: %s\n", tr.Text)
					break
				}
			}

			if r.interrupted {
				r.History.AddToolCallResult(toolCall, "Tool execution interrupted")
				return ToolImplOutput{ToolOutput: "Reviewer interrupted", ToolResultMessage: "Reviewer interrupted during tool execution"}, nil
			}

			// Run Tool
			var toolOutputStr string
			var foundTool bool
			for _, t := range r.Tools {
				if t.GetToolParam().Name == toolCall.Name {
					res, err := t.Run(ctx, toolCall.Arguments, r.History)
					if err != nil {
						toolOutputStr = fmt.Sprintf("Error: %v", err)
					} else {
						toolOutputStr = res.ToolOutput
					}
					foundTool = true
					break
				}
			}
			if !foundTool {
				toolOutputStr = "Tool not found"
			}

			r.History.AddToolCallResult(toolCall, toolOutputStr)

			if toolCall.Name == "return_control_to_general_agent" {
				summarizeReview := "Now based on your review, please rewrite detailed feedback to the general agent."
				r.History.AddUserPrompt(summarizeReview, nil)
				
				currentMessages = r.History.GetMessagesForLLM()
				truncatedMessages = r.ContextManager.ApplyTruncationIfNeeded(currentMessages)
				
				summaryResponse, err := r.generateLLMResponse(ctx, truncatedMessages, toolParams)
				if err != nil {
					return ToolImplOutput{}, err
				}

				var finalText string
				for _, msg := range summaryResponse {
					if tr, ok := msg.(TextResult); ok {
						finalText = tr.Text
						break
					}
				}

				if finalText != "" {
					return ToolImplOutput{
						ToolOutput: finalText,
						ToolResultMessage: "Reviewer completed comprehensive review",
					}, nil
				} else {
					r.Logger.Println("Error: No text output in model response for review summary")
					return ToolImplOutput{
						ToolOutput: "ERROR: Reviewer did not provide text feedback",
						ToolResultMessage: "Review incomplete - no text response",
					}, nil
				}
			}
		}
	}

	return ToolImplOutput{
		ToolOutput: "ERROR: Reviewer did not complete review within maximum turns. The review process was interrupted or took too long to complete.",
		ToolResultMessage: "Review incomplete - maximum turns reached",
	}, nil
}

// RunAgent is the synchronous convenience wrapper (mimics run_agent)
func (r *ReviewerAgent) RunAgent(task, result, workspaceDir string, resume bool) (string, error) {
	// In Go, usually run synchronously, or use StartMessageProcessing for background
	
	// Reset tool logic if implemented in a manager
	// r.ToolManager.Reset() 
	
	if resume {
		// assert r.History.IsNextTurnUser()
	} else {
		r.History.Clear()
		r.interrupted = false
	}

	toolInput := map[string]interface{}{
		"task":          task,
		"workspace_dir": workspaceDir,
		"result":        result,
	}

	// Create a background context or use TODO
	output, err := r.Run(context.Background(), toolInput, r.History)
	return output.ToolOutput, err
}

func (r *ReviewerAgent) Cancel() {
	r.interrupted = true
	r.Logger.Println("Reviewer cancellation requested")
}

func (r *ReviewerAgent) Clear() {
	r.History.Clear()
	r.interrupted = false
	r.cachedToolParams = nil
}