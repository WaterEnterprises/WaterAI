package browser

import (
	"testing"
)

func TestTabInfo(t *testing.T) {
	tab := TabInfo{
		PageID: 1,
		URL:    "https://example.com",
		Title:  "Example",
	}

	if tab.PageID != 1 {
		t.Errorf("PageID = %d; want 1", tab.PageID)
	}

	if tab.URL != "https://example.com" {
		t.Errorf("URL = %s; want https://example.com", tab.URL)
	}

	if tab.Title != "Example" {
		t.Errorf("Title = %s; want Example", tab.Title)
	}
}

func TestCoordinates(t *testing.T) {
	coords := Coordinates{
		X:      100.5,
		Y:      200.5,
		Width:  50.0,
		Height: 30.0,
	}

	if coords.X != 100.5 {
		t.Errorf("X = %f; want 100.5", coords.X)
	}

	if coords.Y != 200.5 {
		t.Errorf("Y = %f; want 200.5", coords.Y)
	}

	if coords.Width != 50.0 {
		t.Errorf("Width = %f; want 50.0", coords.Width)
	}

	if coords.Height != 30.0 {
		t.Errorf("Height = %f; want 30.0", coords.Height)
	}
}

func TestRect(t *testing.T) {
	rect := Rect{
		Left:   0.0,
		Top:    0.0,
		Right:  100.0,
		Bottom: 50.0,
		Width:  100.0,
		Height: 50.0,
	}

	if rect.Left != 0.0 {
		t.Errorf("Left = %f; want 0.0", rect.Left)
	}

	if rect.Width != 100.0 {
		t.Errorf("Width = %f; want 100.0", rect.Width)
	}
}

func TestInteractiveElement(t *testing.T) {
	element := InteractiveElement{
		Index:   1,
		TagName: "button",
		Text:    "Click me",
		Attributes: map[string]string{
			"id":    "submit-btn",
			"class": "btn-primary",
		},
		Viewport: Coordinates{X: 10, Y: 20},
		Page:     Coordinates{X: 10, Y: 20},
		Center:   Coordinates{X: 35, Y: 30},
		Weight:   1.5,
		Rect:     Rect{Left: 10, Top: 20, Right: 60, Bottom: 40},
		ZIndex:   10,
	}

	if element.Index != 1 {
		t.Errorf("Index = %d; want 1", element.Index)
	}

	if element.TagName != "button" {
		t.Errorf("TagName = %s; want button", element.TagName)
	}

	if element.Attributes["id"] != "submit-btn" {
		t.Errorf("Attributes[id] = %s; want submit-btn", element.Attributes["id"])
	}
}

func TestViewport(t *testing.T) {
	viewport := Viewport{
		Width:            1920,
		Height:           1080,
		ScrollX:          0,
		ScrollY:          100,
		DevicePixelRatio: 2.0,
	}

	if viewport.Width != 1920 {
		t.Errorf("Width = %d; want 1920", viewport.Width)
	}

	if viewport.Height != 1080 {
		t.Errorf("Height = %d; want 1080", viewport.Height)
	}

	if viewport.DevicePixelRatio != 2.0 {
		t.Errorf("DevicePixelRatio = %f; want 2.0", viewport.DevicePixelRatio)
	}
}

func TestInteractiveElementsData(t *testing.T) {
	data := InteractiveElementsData{
		Viewport: Viewport{Width: 1024, Height: 768},
		Elements: []InteractiveElement{
			{Index: 0, TagName: "a", Text: "Link"},
			{Index: 1, TagName: "button", Text: "Button"},
		},
	}

	if data.Viewport.Width != 1024 {
		t.Errorf("Viewport.Width = %d; want 1024", data.Viewport.Width)
	}

	if len(data.Elements) != 2 {
		t.Errorf("Elements length = %d; want 2", len(data.Elements))
	}
}

func TestBrowserState(t *testing.T) {
	state := BrowserState{
		URL: "https://example.com/page",
		Tabs: []TabInfo{
			{PageID: 0, URL: "https://example.com/page", Title: "Page"},
		},
		Viewport: Viewport{Width: 1280, Height: 720},
		Screenshot: "base64imagedata",
		ScreenshotWithHighlights: "base64highlighted",
		InteractiveElements: map[int]InteractiveElement{
			1: {Index: 1, TagName: "button"},
		},
	}

	if state.URL != "https://example.com/page" {
		t.Errorf("URL = %s; want https://example.com/page", state.URL)
	}

	if len(state.Tabs) != 1 {
		t.Errorf("Tabs length = %d; want 1", len(state.Tabs))
	}

	if len(state.InteractiveElements) != 1 {
		t.Errorf("InteractiveElements length = %d; want 1", len(state.InteractiveElements))
	}
}

func TestViewportSize(t *testing.T) {
	size := ViewportSize{
		Width:  1268,
		Height: 951,
	}

	if size.Width != 1268 {
		t.Errorf("Width = %d; want 1268", size.Width)
	}

	if size.Height != 951 {
		t.Errorf("Height = %d; want 951", size.Height)
	}
}

func TestBrowserConfig(t *testing.T) {
	config := BrowserConfig{
		CDPURL:       "http://localhost:9222",
		ViewportSize: ViewportSize{Width: 1920, Height: 1080},
		StorageState: map[string]interface{}{"cookies": []interface{}{}},
	}

	if config.CDPURL != "http://localhost:9222" {
		t.Errorf("CDPURL = %s; want http://localhost:9222", config.CDPURL)
	}

	if config.ViewportSize.Width != 1920 {
		t.Errorf("ViewportSize.Width = %d; want 1920", config.ViewportSize.Width)
	}
}

func TestDefaultBrowserConfig(t *testing.T) {
	config := DefaultBrowserConfig()

	if config.ViewportSize.Width != 1268 {
		t.Errorf("Default ViewportSize.Width = %d; want 1268", config.ViewportSize.Width)
	}

	if config.ViewportSize.Height != 951 {
		t.Errorf("Default ViewportSize.Height = %d; want 951", config.ViewportSize.Height)
	}
}

func TestInteractiveElementWithInputType(t *testing.T) {
	element := InteractiveElement{
		Index:     1,
		TagName:   "input",
		InputType: "text",
	}

	if element.InputType != "text" {
		t.Errorf("InputType = %s; want text", element.InputType)
	}
}

func TestBrowserStateEmpty(t *testing.T) {
	state := BrowserState{}

	if state.URL != "" {
		t.Errorf("URL = %s; want empty", state.URL)
	}

	if state.Tabs != nil {
		t.Error("Tabs should be nil")
	}

	if state.InteractiveElements != nil {
		t.Error("InteractiveElements should be nil")
	}
}

func TestCoordinatesOptionalFields(t *testing.T) {
	coords := Coordinates{
		X: 100.0,
		Y: 200.0,
		// Width and Height are optional
	}

	if coords.Width != 0 {
		t.Errorf("Width = %f; want 0 (optional)", coords.Width)
	}

	if coords.Height != 0 {
		t.Errorf("Height = %f; want 0 (optional)", coords.Height)
	}
}

func TestBrowserConfigWithDetector(t *testing.T) {
	config := BrowserConfig{
		ViewportSize: ViewportSize{Width: 1024, Height: 768},
		Detector:     nil, // Detector interface can be nil
	}

	if config.Detector != nil {
		t.Error("Detector should be nil")
	}
}

func TestDefaultViewport(t *testing.T) {
	vp := DefaultViewport()

	if vp.Width != 1268 {
		t.Errorf("Default Viewport.Width = %d; want 1268", vp.Width)
	}

	if vp.Height != 951 {
		t.Errorf("Default Viewport.Height = %d; want 951", vp.Height)
	}

	if vp.DevicePixelRatio != 1.0 {
		t.Errorf("Default Viewport.DevicePixelRatio = %f; want 1.0", vp.DevicePixelRatio)
	}
}