package agents

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	ToolResultInterruptMsg       = "Tool execution interrupted by user."
	AgentInterruptMsg            = "Agent interrupted by user."
	ToolCallInterruptFakeRsp     = "Tool execution interrupted by user. You can resume by providing a new instruction."
	AgentInterruptFakeRsp        = "Agent interrupted by user. You can resume by providing a new instruction."
	CompleteMessage              = "Task Completed"
)

// SystemPromptBuilder interface
type SystemPromptBuilder interface {
	GetSystemPrompt() string
}

type FunctionCallAgent struct {
	BaseAgent
	SystemPromptBuilder SystemPromptBuilder
	Client              LLMClient
	Tools               []LLMTool
	History             MessageHistory
	WorkspaceManager    WorkspaceManager
	MessageQueue        chan RealtimeEvent
	Logger              *log.Logger
	MaxOutputTokens     int
	MaxTurns            int
	Websocket           WebSocket
	
	interrupted         bool
	sessionID           string
}

func NewFunctionCallAgent(
	systemPromptBuilder SystemPromptBuilder,
	client LLMClient,
	tools []LLMTool,
	initHistory MessageHistory,
	workspaceManager WorkspaceManager,
	messageQueue chan RealtimeEvent,
	logger *log.Logger,
	maxOutputTokens int,
	maxTurns int,
	websocket WebSocket,
) *FunctionCallAgent {
	return &FunctionCallAgent{
		BaseAgent: BaseAgent{
			Name: "general_agent",
			Description: `A general agent that can accomplish tasks and answer questions.

If you are faced with a task that involves more than a few steps, or if the task is complex, or if the instructions are very long,
try breaking down the task into smaller steps. After call this tool to update or create a plan, use write_file or str_replace_tool to update the plan to todo.md`,
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"instruction": map[string]string{
						"type": "string", 
						"description": "The instruction to the agent.",
					},
				},
				"required": []string{"instruction"},
			},
		},
		SystemPromptBuilder: systemPromptBuilder,
		Client:              client,
		Tools:               tools,
		History:             initHistory,
		WorkspaceManager:    workspaceManager,
		MessageQueue:        messageQueue,
		Logger:              logger,
		MaxOutputTokens:     maxOutputTokens,
		MaxTurns:            maxTurns,
		Websocket:           websocket,
		sessionID:           workspaceManager.SessionID(),
	}
}

func (a *FunctionCallAgent) StartMessageProcessing(ctx context.Context) {
	go func() {
		defer a.Logger.Println("Message processor stopped")
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-a.MessageQueue:
				// Note: Database saving would happen here using Events.SaveEvent
				if a.sessionID != "" {
					// Events.SaveEvent(a.sessionID, msg)
				} else {
					a.Logger.Printf("No session ID, skipping event save: %v", msg)
				}

				if msg.Type != EventTypeUserMessage && a.Websocket != nil {
					if err := a.Websocket.SendJSON(msg); err != nil {
						a.Logger.Printf("Failed to send message to websocket: %v", err)
						a.Websocket = nil
					}
				}
			}
		}
	}()
}

func (a *FunctionCallAgent) validateToolParameters() ([]ToolParam, error) {
	var params []ToolParam
	names := make([]string, 0)

	for _, tool := range a.Tools {
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
	return params, nil
}

// encodeImage Helper (simulates ii_agent.tools.utils.encode_image)
func encodeImage(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *FunctionCallAgent) Run(ctx context.Context, toolInput map[string]interface{}, history MessageHistory) (ToolImplOutput, error) {
	instruction, _ := toolInput["instruction"].(string)
	
	// Handle Files Input
	var files []string
	if fList, ok := toolInput["files"].([]interface{}); ok {
		for _, f := range fList {
			if fStr, ok := f.(string); ok {
				files = append(files, fStr)
			}
		}
	} else if fListStr, ok := toolInput["files"].([]string); ok {
		files = fListStr
	}

	delimiter := "--------------------------------------------- USER INPUT ---------------------------------------------\n" + instruction
	a.Logger.Printf("\n%s\n", delimiter)

	var imageBlocks []interface{}

	if len(files) > 0 {
		instruction += "\n\nAttached files:\n"
		for _, file := range files {
			relPath := a.WorkspaceManager.RelativePath(file)
			instruction += fmt.Sprintf(" - %s\n", relPath)
			a.Logger.Printf("Attached file: %s", relPath)

			// Process images
			ext := ""
			if parts := strings.Split(file, "."); len(parts) > 1 {
				ext = parts[len(parts)-1]
			}
			if ext == "jpg" { ext = "jpeg" }
			
			if ext == "png" || ext == "jpeg" || ext == "gif" || ext == "webp" {
				fullPath := a.WorkspaceManager.WorkspacePath(file)
				b64Data, err := encodeImage(fullPath)
				if err == nil {
					imageBlocks = append(imageBlocks, map[string]interface{}{
						"source": map[string]interface{}{
							"type": "base64",
							"media_type": fmt.Sprintf("image/%s", ext),
							"data": b64Data,
						},
					})
				} else {
					a.Logger.Printf("Failed to encode image %s: %v", fullPath, err)
				}
			}
		}
	}

	a.History.AddUserPrompt(instruction, imageBlocks)
	a.interrupted = false

	remainingTurns := a.MaxTurns
	for remainingTurns > 0 {
		a.History.Truncate()
		remainingTurns--

		a.Logger.Println("\n--------------------------------------------- NEW TURN ---------------------------------------------")

		toolParams, err := a.validateToolParameters()
		if err != nil {
			return ToolImplOutput{}, err
		}

		if a.interrupted {
			a.addFakeAssistantTurn(AgentInterruptFakeRsp)
			return ToolImplOutput{ToolOutput: AgentInterruptMsg, ToolResultMessage: AgentInterruptMsg}, nil
		}

		a.Logger.Printf("(Current token count: %d)\n", a.History.CountTokens())

		// Generate
		modelResponse, err := a.Client.Generate(
			ctx,
			a.History.GetMessagesForLLM(),
			a.MaxOutputTokens,
			toolParams,
			a.SystemPromptBuilder.GetSystemPrompt(),
		)

		if err != nil {
			return ToolImplOutput{ToolOutput: "Error calling LLM"}, err
		}

		if len(modelResponse) == 0 {
			modelResponse = []interface{}{TextResult{Text: CompleteMessage}}
		}

		a.History.AddAssistantTurn(modelResponse)

		// Check if we are done (no tools called)
		pendingTools := a.History.GetPendingToolCalls()
		if len(pendingTools) == 0 {
			a.Logger.Println("[no tools were called]")
			a.emitEvent(EventTypeAgentResponse, map[string]interface{}{"text": "Task completed"})
			return ToolImplOutput{
				ToolOutput: a.History.GetLastAssistantTextResponse(),
				ToolResultMessage: "Task completed",
			}, nil
		}

		// Process Thinking and Text
		for _, item := range modelResponse {
			if tb, ok := item.(ThinkingBlock); ok {
				// Format thinking block logic from Python
				wrappedThinking := ""
				words := strings.Fields(tb.Thinking)
				for i := 0; i < len(words); i += 8 {
					end := i + 8
					if end > len(words) {
						end = len(words)
					}
					wrappedThinking += strings.Join(words[i:end], " ") + "\n"
				}
				formatted := fmt.Sprintf("```Thinking:\n%s\n```", strings.TrimSpace(wrappedThinking))
				
				a.Logger.Printf("Top-level agent planning next step: %s\n", formatted)
				a.emitEvent(EventTypeAgentThinking, map[string]interface{}{"text": formatted})
			} else if tr, ok := item.(TextResult); ok {
				a.Logger.Printf("Top-level agent planning next step: %s\n", tr.Text)
				a.emitEvent(EventTypeAgentThinking, map[string]interface{}{"text": tr.Text})
			}
		}

		if len(pendingTools) > 1 {
			return ToolImplOutput{}, errors.New("only one tool call per turn is supported")
		}

		toolCall := pendingTools[0]
		a.emitEvent(EventTypeToolCall, map[string]interface{}{
			"tool_call_id": toolCall.ID,
			"tool_name":    toolCall.Name,
			"tool_input":   toolCall.Arguments,
		})

		// Handle interruption before tool run
		if a.interrupted {
			a.addToolCallResult(toolCall, ToolResultInterruptMsg)
			a.addFakeAssistantTurn(ToolCallInterruptFakeRsp)
			return ToolImplOutput{ToolOutput: ToolResultInterruptMsg, ToolResultMessage: ToolResultInterruptMsg}, nil
		}

		// Execute Tool
		var selectedTool LLMTool
		for _, t := range a.Tools {
			if t.GetToolParam().Name == toolCall.Name {
				selectedTool = t
				break
			}
		}

		var toolOutput ToolImplOutput
		if selectedTool != nil {
			toolOutput, err = selectedTool.Run(ctx, toolCall.Arguments, a.History)
			if err != nil {
				// Log error, but return generic failure string to history
				a.Logger.Printf("Tool execution error: %v", err)
				toolOutput = ToolImplOutput{
					ToolOutput: fmt.Sprintf("Error executing tool: %v", err),
					IsFinal: false,
				}
			}
		} else {
			toolOutput = ToolImplOutput{ToolOutput: "Tool not found", IsFinal: false}
		}

		a.addToolCallResult(toolCall, toolOutput.ToolOutput)
		
		// Check for Final Answer (should_stop logic)
		if toolOutput.IsFinal {
			finalAnswer := toolOutput.ToolOutput 
			// In Python: self.tool_manager.get_final_answer()
			a.addFakeAssistantTurn(finalAnswer)
			return ToolImplOutput{
				ToolOutput: finalAnswer,
				ToolResultMessage: "Task completed",
			}, nil
		}
	}

	agentAnswer := "Agent did not complete after max turns"
	a.emitEvent(EventTypeAgentResponse, map[string]interface{}{"text": agentAnswer})
	return ToolImplOutput{ToolOutput: agentAnswer, ToolResultMessage: agentAnswer}, nil
}

// RunAgent is the convenience wrapper (mimics run_agent logic)
func (a *FunctionCallAgent) RunAgent(instruction string, files []string, resume bool, orientationInstruction string) (string, error) {
	// Reset tool logic if implemented via manager
	// a.ToolManager.Reset()

	if !resume {
		a.History.Clear()
		a.interrupted = false
	}

	toolInput := map[string]interface{}{
		"instruction": instruction,
		"files":       files,
	}
	if orientationInstruction != "" {
		toolInput["orientation_instruction"] = orientationInstruction
	}

	output, err := a.Run(context.Background(), toolInput, a.History)
	return output.ToolOutput, err
}

func (a *FunctionCallAgent) emitEvent(eventType string, content map[string]interface{}) {
	a.MessageQueue <- RealtimeEvent{
		Type:    eventType,
		Content: content,
	}
}

func (a *FunctionCallAgent) addToolCallResult(toolCall ToolCallParameters, result string) {
	a.History.AddToolCallResult(toolCall, result)
	a.emitEvent(EventTypeToolResult, map[string]interface{}{
		"tool_call_id": toolCall.ID,
		"tool_name":    toolCall.Name,
		"result":       result,
	})
}

func (a *FunctionCallAgent) addFakeAssistantTurn(text string) {
	a.History.AddAssistantTurn([]interface{}{TextResult{Text: text}})
	evtType := EventTypeAgentResponse
	if a.interrupted {
		evtType = EventTypeResponseInterrupt
	}
	a.emitEvent(evtType, map[string]interface{}{"text": text})
}

func (a *FunctionCallAgent) Cancel() {
	a.interrupted = true
	a.Logger.Println("Agent cancellation requested")
}

func (a *FunctionCallAgent) Clear() {
	a.History.Clear()
	a.interrupted = false
}