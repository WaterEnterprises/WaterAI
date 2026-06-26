package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Search Tool ---
type WebSearchTool struct {
	Config Config
}

func (t *WebSearchTool) Name() string { return "web_search" }
func (t *WebSearchTool) Description() string { return "Search the web (Tavily/Jina/DuckDuckGo)" }
func (t *WebSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{"query": map[string]string{"type": "string"}},
		"required": []string{"query"},
	}
}

func (t *WebSearchTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	query, _ := GetArg[string](input, "query")

	// Priority: Tavily -> Jina -> DuckDuckGo (Placeholder logic)
	if t.Config.TavilyAPIKey != "" {
		// Mock Tavily Call
		return t.tavilySearch(query)
	}
	
	// Fallback to DuckDuckGo (simplified simulation)
	return &ToolOutput{Text: fmt.Sprintf("Searching DuckDuckGo for: %s (Implementation pending real API call)", query)}, nil
}

func (t *WebSearchTool) tavilySearch(query string) (*ToolOutput, error) {
	// Implement HTTP Post to Tavily API
	return &ToolOutput{Text: "Tavily results for: " + query}, nil
}

// --- Webpage Visit Tool ---
type WebVisitTool struct {
	Config Config
}

func (t *WebVisitTool) Name() string { return "visit_webpage" }
func (t *WebVisitTool) Description() string { return "Visit a webpage and extract markdown" }
func (t *WebVisitTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{"url": map[string]string{"type": "string"}},
		"required": []string{"url"},
	}
}

func (t *WebVisitTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	targetURL, _ := GetArg[string](input, "url")
	
	// Use Jina Reader if Key exists (cleanest markdown)
	if t.Config.JinaAPIKey != "" {
		jinaURL := "https://r.jina.ai/" + targetURL
		req, _ := http.NewRequest("GET", jinaURL, nil)
		req.Header.Set("Authorization", "Bearer "+t.Config.JinaAPIKey)
		
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil { return ErrorOutput(err), nil }
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		return &ToolOutput{Text: string(body)}, nil
	}
	
	return &ToolOutput{Text: "No Jina API key provided, cannot convert to Markdown reliably."}, nil
}