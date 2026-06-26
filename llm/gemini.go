package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GeminiClient struct {
	config LLMConfig
	client *http.Client
}

func NewGeminiClient(cfg LLMConfig) *GeminiClient {
	return &GeminiClient{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Internal structures for Gemini JSON
type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string              `json:"text,omitempty"`
	InlineData       *geminiBlob         `json:"inlineData,omitempty"` // For images
	FunctionCall     *geminiFuncCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFuncResponse `json:"functionResponse,omitempty"`
}

type geminiBlob struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiFuncCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFuncResponse struct {
	Name     string `json:"name"`
	Response struct {
		Result interface{} `json:"result"`
	} `json:"response"`
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	Tools            []interface{}   `json:"tools,omitempty"`
	SystemInstr      *geminiContent  `json:"systemInstruction,omitempty"`
	GenerationConfig struct {
		Temperature     float64 `json:"temperature"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

func (c *GeminiClient) Generate(
	messages []*Message,
	maxTokens int,
	systemPrompt string,
	temperature float64,
	tools []*ToolParam,
	toolChoice *ToolChoice,
	thinkingTokens *int,
) (*GenerateResponse, error) {

	// 1. Convert Messages
	var gemContents []geminiContent

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model" // Gemini uses 'model'
		}
		
		var parts []geminiPart
		for _, b := range msg.Content {
			switch b.Type {
			case ContentTypeText:
				parts = append(parts, geminiPart{Text: b.Text})
			case ContentTypeImage:
				parts = append(parts, geminiPart{InlineData: &geminiBlob{
					MimeType: b.Source.MediaType,
					Data:     b.Source.Data,
				}})
			case ContentTypeToolCall:
				parts = append(parts, geminiPart{FunctionCall: &geminiFuncCall{
					Name: b.ToolName,
					Args: b.ToolInput,
				}})
			case ContentTypeToolResult:
				parts = append(parts, geminiPart{FunctionResponse: &geminiFuncResponse{
					Name: b.ToolName,
					Response: struct {
						Result interface{} `json:"result"`
					}{Result: b.ToolOutput},
				}})
			}
		}
		gemContents = append(gemContents, geminiContent{Role: role, Parts: parts})
	}

	// 2. Prepare Tools
	var gemTools []interface{}
	if len(tools) > 0 {
		funcs := []map[string]interface{}{}
		for _, t := range tools {
			funcs = append(funcs, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.InputSchema,
			})
		}
		gemTools = append(gemTools, map[string]interface{}{"function_declarations": funcs})
	}

	// 3. Prepare Request
	reqBody := geminiRequest{
		Contents: gemContents,
		Tools:    gemTools,
	}
	reqBody.GenerationConfig.Temperature = temperature
	reqBody.GenerationConfig.MaxOutputTokens = maxTokens

	if systemPrompt != "" {
		reqBody.SystemInstr = &geminiContent{
			Role: "user", 
			Parts: []geminiPart{{Text: systemPrompt}},
		}
	}

	// 4. Execute
	jsonBody, _ := json.Marshal(reqBody)
	
	// Assuming API Key auth. Vertex Logic skipped for brevity as per rewrite constraints.
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.config.Model, c.config.APIKey)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini Error %d: %s", resp.StatusCode, string(b))
	}

	// 5. Parse Response
	var result struct {
		Candidates []struct {
			Content geminiContent `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var blocks []*ContentBlock
	if len(result.Candidates) > 0 {
		cand := result.Candidates[0]
		for _, p := range cand.Content.Parts {
			if p.Text != "" {
				blocks = append(blocks, &ContentBlock{Type: ContentTypeText, Text: p.Text})
			}
			if p.FunctionCall != nil {
				// Gemini doesn't always provide IDs, generate one
				id := generateID("call")
				blocks = append(blocks, &ContentBlock{
					Type:       ContentTypeToolCall,
					ToolCallID: id,
					ToolName:   p.FunctionCall.Name,
					ToolInput:  p.FunctionCall.Args,
				})
			}
		}
	}

	return &GenerateResponse{
		Content: blocks,
		Usage: UsageMetadata{
			InputTokens:  result.UsageMetadata.PromptTokenCount,
			OutputTokens: result.UsageMetadata.CandidatesTokenCount,
			RawResponse:  result,
		},
	}, nil
}