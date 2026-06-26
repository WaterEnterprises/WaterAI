package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

// --- Web Search Tool ---
type WebWebSearchTool struct {
	APIKey string
}

func (t *WebWebSearchTool) Name() string        { return "web_search" }
func (t *WebWebSearchTool) Description() string { return "Search the web for information." }
func (t *WebWebSearchTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]string{"type": "string"},
		},
		"required": []string{"query"},
	}
}

func (t *WebWebSearchTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	query, _ := input["query"].(string)
	// Mock implementation. In production, use http.Get to Serper/Google API using t.APIKey
	return ToolResult{
		Output:        fmt.Sprintf("Mock search results for: %s", query),
		ResultMessage: "Search completed",
		Success:       true,
	}, nil
}

// --- Visit Webpage Tool ---
type VisitWebpageTool struct{}

func (t *VisitWebpageTool) Name() string        { return "visit_webpage" }
func (t *VisitWebpageTool) Description() string { return "Visit a URL and extract text." }
func (t *VisitWebpageTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]string{"type": "string"},
		},
		"required": []string{"url"},
	}
}

func (t *VisitWebpageTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	url, _ := input["url"].(string)
	resp, err := http.Get(url)
	if err != nil {
		return ToolResult{Output: err.Error(), Success: false}, nil
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	// Simplified: In production, strip HTML tags properly
	return ToolResult{
		Output:        string(body)[:min(len(body), 5000)], // Truncate
		ResultMessage: "Page visited",
		Success:       true,
	}, nil
}

// --- YouTube Transcript Tool ---
type YouTubeTranscriptTool struct{}

func (t *YouTubeTranscriptTool) Name() string        { return "youtube_transcript" }
func (t *YouTubeTranscriptTool) Description() string { return "Get transcript from YouTube video." }
func (t *YouTubeTranscriptTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]string{"type": "string"},
		},
		"required": []string{"url"},
	}
}

func (t *YouTubeTranscriptTool) Run(ctx context.Context, input ToolInput) (ToolResult, error) {
	url, _ := input["url"].(string)
	
	// Uses yt-dlp CLI which must be installed on the system
	cmd := exec.CommandContext(ctx, "yt-dlp", "--write-auto-sub", "--skip-download", "--sub-lang", "en", "--output", "-", url)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return ToolResult{Output: string(output), Success: false, ResultMessage: "Failed to download subtitles"}, nil
	}

	return ToolResult{
		Output:        "Transcript extraction simulated (yt-dlp integration required)",
		ResultMessage: "Transcript extracted",
		Success:       true,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}