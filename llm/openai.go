package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type OpenAIClient struct {
	config LLMConfig
	client *http.Client
}

func NewOpenAIClient(cfg LLMConfig) *OpenAIClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	return &OpenAIClient{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Internal structures for OpenAI API Payload
type oaMessage struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content,omitempty"` // string or []oaContentPart
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type oaContentPart struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *oaImageURL       `json:"image_url,omitempty"`
}

type oaImageURL struct {
	URL string `json:"url"`
}

type oaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function oaFunction `json:"function"`
}

type oaFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type oaRequest struct {
	Model       string      `json:"model"`
	Messages    []oaMessage `json:"messages"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Temperature float64     `json:"temperature"`
	Tools       []oaToolDef `json:"tools,omitempty"`
	ToolChoice  interface{} `json:"tool_choice,omitempty"`
}

type oaToolDef struct {
	Type     string `json:"type"`
	Function ToolParam `json:"function"`
}

func (c *OpenAIClient) Generate(
	messages []*Message,
	maxTokens int,
	systemPrompt string,
	temperature float64,
	tools []*ToolParam,
	toolChoice *ToolChoice,
	thinkingTokens *int,
) (*GenerateResponse, error) {

	// 1. Prepare Messages
	var oaMsgs []oaMessage

	if systemPrompt != "" && !c.config.CotModel {
		oaMsgs = append(oaMsgs, oaMessage{Role: "system", Content: systemPrompt})
	}

	for _, msg := range messages {
		// OpenAI User/Assistant Logic
		if msg.Role == "user" {
			// Check for text only or multimodal
			hasImage := false
			for _, b := range msg.Content {
				if b.Type == ContentTypeImage {
					hasImage = true
					break
				}
			}

			if !hasImage {
				// Combine text parts
				var textContent string
				for _, b := range msg.Content {
					if b.Type == ContentTypeText {
						textContent += b.Text
					}
				}
				// Handle COT special case (System Prompt inside User Msg)
				if c.config.CotModel && systemPrompt != "" && len(oaMsgs) == 0 {
					textContent = systemPrompt + "\n\n" + textContent
				}
				oaMsgs = append(oaMsgs, oaMessage{Role: "user", Content: textContent})
			} else {
				// Multimodal
				parts := []oaContentPart{}
				for _, b := range msg.Content {
					if b.Type == ContentTypeText {
						parts = append(parts, oaContentPart{Type: "text", Text: b.Text})
					} else if b.Type == ContentTypeImage {
						// Assuming base64 source
						url := fmt.Sprintf("data:%s;base64,%s", b.Source.MediaType, b.Source.Data)
						parts = append(parts, oaContentPart{Type: "image_url", ImageURL: &oaImageURL{URL: url}})
					}
				}
				oaMsgs = append(oaMsgs, oaMessage{Role: "user", Content: parts})
			}
		} else if msg.Role == "assistant" {
			var textContent string
			var toolCalls []oaToolCall
			
			for _, b := range msg.Content {
				if b.Type == ContentTypeText {
					textContent += b.Text
				} else if b.Type == ContentTypeToolCall {
					argsJSON, _ := json.Marshal(b.ToolInput)
					toolCalls = append(toolCalls, oaToolCall{
						ID:   b.ToolCallID,
						Type: "function",
						Function: oaFunction{
							Name:      b.ToolName,
							Arguments: string(argsJSON),
						},
					})
				}
			}
			
			m := oaMessage{Role: "assistant"}
			if textContent != "" {
				m.Content = textContent
			}
			if len(toolCalls) > 0 {
				m.ToolCalls = toolCalls
			}
			oaMsgs = append(oaMsgs, m)

		} else {
			// Tool Results (Role = "tool" in OpenAI, but usually mapped from "user" with ToolResult blocks in our struct)
			// Need to iterate content blocks to see if they are results
			for _, b := range msg.Content {
				if b.Type == ContentTypeToolResult {
					outputStr := ""
					if s, ok := b.ToolOutput.(string); ok {
						outputStr = s
					} else {
						jsonBytes, _ := json.Marshal(b.ToolOutput)
						outputStr = string(jsonBytes)
					}
					
					oaMsgs = append(oaMsgs, oaMessage{
						Role:       "tool",
						ToolCallID: b.ToolCallID,
						Content:    outputStr,
					})
				}
			}
		}
	}

	// 2. Prepare Tools
	var oaTools []oaToolDef
	for _, t := range tools {
		oaTools = append(oaTools, oaToolDef{Type: "function", Function: *t})
	}

	// 3. Prepare Request
	reqBody := oaRequest{
		Model:       c.config.Model,
		Messages:    oaMsgs,
		Temperature: temperature,
	}
	
	if c.config.CotModel {
		// O1 models don't support temperature/max_tokens in the standard way mostly
		// reqBody.MaxCompletionTokens = maxTokens // struct field needs adding if strict
	} else {
		reqBody.MaxTokens = maxTokens
	}

	if len(oaTools) > 0 {
		reqBody.Tools = oaTools
		if toolChoice != nil {
			if toolChoice.Type == "any" {
				reqBody.ToolChoice = "required"
			} else if toolChoice.Type == "auto" {
				reqBody.ToolChoice = "auto"
			} else if toolChoice.Type == "tool" {
				reqBody.ToolChoice = map[string]interface{}{
					"type": "function",
					"function": map[string]string{"name": toolChoice.Name},
				}
			}
		}
	}

	// 4. Execute
	jsonBody, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	var resp *http.Response
	var err error

	for i := 0; i < c.config.MaxRetries; i++ {
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			break
		}
		if i < c.config.MaxRetries-1 {
			time.Sleep(time.Duration(10 * (i + 1)) * time.Second) // Simple backoff
		}
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	// 5. Parse Response
	var result struct {
		Choices []struct {
			Message oaMessage `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	oaRespMsg := result.Choices[0].Message
	
	// Convert back to ContentBlocks
	var blocks []*ContentBlock

	// Content
	if oaRespMsg.Content != nil {
		if text, ok := oaRespMsg.Content.(string); ok && text != "" {
			blocks = append(blocks, &ContentBlock{Type: ContentTypeText, Text: text})
		}
	}

	// Tool Calls
	for _, tc := range oaRespMsg.ToolCalls {
		var args map[string]interface{}
		// OpenAI returns stringified JSON for arguments
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			log.Printf("Error unmarshaling tool args: %v", err)
			continue
		}
		blocks = append(blocks, &ContentBlock{
			Type:       ContentTypeToolCall,
			ToolCallID: tc.ID,
			ToolName:   tc.Function.Name,
			ToolInput:  args,
		})
	}

	return &GenerateResponse{
		Content: blocks,
		Usage: UsageMetadata{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
			RawResponse:  result,
		},
	}, nil
}