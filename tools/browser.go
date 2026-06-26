package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

type BrowserManager struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
}

func NewBrowserManager(headless bool) (*BrowserManager, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		return nil, err
	}
	ctx, err := browser.NewContext()
	if err != nil {
		return nil, err
	}
	page, err := ctx.NewPage()
	if err != nil {
		return nil, err
	}
	return &BrowserManager{pw: pw, browser: browser, context: ctx, page: page}, nil
}

func (b *BrowserManager) captureState() (string, string, error) {
	// Returns screenshot (base64) and simple text message
	screenshot, err := b.page.Screenshot(playwright.PageScreenshotOptions{
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(screenshot), b.page.URL(), nil
}

// --- Tool Implementations ---

type BrowserNavigateTool struct{ Manager *BrowserManager }

func (t *BrowserNavigateTool) Name() string        { return "browser_navigate" }
func (t *BrowserNavigateTool) Description() string { return "Navigate to a URL" }
func (t *BrowserNavigateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]string{"type": "string"},
		},
		"required": []string{"url"},
	}
}
func (t *BrowserNavigateTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	url, err := GetArg[string](input, "url")
	if err != nil {
		return ErrorOutput(err), nil
	}
	if _, err := t.Manager.page.Goto(url); err != nil {
		return ErrorOutput(err), nil
	}
	img, _, _ := t.Manager.captureState()
	return &ToolOutput{Text: "Navigated to " + url, Images: []string{img}}, nil
}

type BrowserClickTool struct{ Manager *BrowserManager }

func (t *BrowserClickTool) Name() string        { return "browser_click" }
func (t *BrowserClickTool) Description() string { return "Click at coordinates (x, y)" }
func (t *BrowserClickTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"x": map[string]string{"type": "number"},
			"y": map[string]string{"type": "number"},
		},
		"required": []string{"x", "y"},
	}
}
func (t *BrowserClickTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	x, err := GetArg[float64](input, "x")
	if err != nil { return ErrorOutput(err), nil }
	y, err := GetArg[float64](input, "y")
	if err != nil { return ErrorOutput(err), nil }

	if err := t.Manager.page.Mouse().Click(x, y); err != nil {
		return ErrorOutput(err), nil
	}
	time.Sleep(1 * time.Second) // Wait for reaction
	img, _, _ := t.Manager.captureState()
	return &ToolOutput{Text: fmt.Sprintf("Clicked at %f, %f", x, y), Images: []string{img}}, nil
}

type BrowserScrollTool struct {
	Manager *BrowserManager
	Direction string // "up" or "down"
}

func (t *BrowserScrollTool) Name() string { return "browser_scroll_" + t.Direction }
func (t *BrowserScrollTool) Description() string { return "Scroll the page " + t.Direction }
func (t *BrowserScrollTool) InputSchema() map[string]interface{} { return map[string]interface{}{} }

func (t *BrowserScrollTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
	deltaY := 500.0
	if t.Direction == "up" {
		deltaY = -500.0
	}
	if err := t.Manager.page.Mouse().Wheel(0, deltaY); err != nil {
		return ErrorOutput(err), nil
	}
	time.Sleep(500 * time.Millisecond)
	img, _, _ := t.Manager.captureState()
	return &ToolOutput{Text: "Scrolled " + t.Direction, Images: []string{img}}, nil
}

type BrowserTypeTool struct{ Manager *BrowserManager }
func (t *BrowserTypeTool) Name() string { return "browser_enter_text" }
func (t *BrowserTypeTool) Description() string { return "Type text into the focused element" }
func (t *BrowserTypeTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "text": map[string]string{"type": "string"},
            "press_enter": map[string]string{"type": "boolean"},
        },
        "required": []string{"text"},
    }
}
func (t *BrowserTypeTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
    text, _ := GetArg[string](input, "text")
    pressEnter, _ := GetArg[bool](input, "press_enter")
    
    t.Manager.page.Keyboard().Type(text)
    if pressEnter {
        t.Manager.page.Keyboard().Press("Enter")
    }
    time.Sleep(1 * time.Second)
    img, _, _ := t.Manager.captureState()
    return &ToolOutput{Text: fmt.Sprintf("Typed '%s'", text), Images: []string{img}}, nil
}

type BrowserViewTool struct{ Manager *BrowserManager }
func (t *BrowserViewTool) Name() string { return "browser_view_interactive_elements" }
func (t *BrowserViewTool) Description() string { return "Get text structure of interactive elements" }
func (t *BrowserViewTool) InputSchema() map[string]interface{} { return map[string]interface{}{} }
func (t *BrowserViewTool) Run(ctx context.Context, input ToolInput) (*ToolOutput, error) {
    // Simplified: in production, you'd run a JS evaluation script to tag elements
    img, url, _ := t.Manager.captureState()
    return &ToolOutput{
        Text: fmt.Sprintf("Current URL: %s. Viewport captured.", url),
        Images: []string{img},
    }, nil
}