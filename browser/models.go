package browser

// models.go

type TabInfo struct {
	PageID int    `json:"pageId"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

type Coordinates struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

type Rect struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type InteractiveElement struct {
	Index          int               `json:"index"`
	TagName        string            `json:"tagName"`
	Text           string            `json:"text"`
	Attributes     map[string]string `json:"attributes"`
	Viewport       Coordinates       `json:"viewport"`
	Page           Coordinates       `json:"page"`
	Center         Coordinates       `json:"center"`
	Weight         float64           `json:"weight"`
	BrowserAgentID string            `json:"browserAgentId"`
	InputType      string            `json:"inputType,omitempty"`
	Rect           Rect              `json:"rect"`
	ZIndex         int               `json:"zIndex"`
}

type Viewport struct {
	Width                       int     `json:"width"`
	Height                      int     `json:"height"`
	ScrollX                     int     `json:"scrollX"`
	ScrollY                     int     `json:"scrollY"`
	DevicePixelRatio            float64 `json:"devicePixelRatio"`
	ScrollDistanceAboveViewport int     `json:"scrollDistanceAboveViewport"`
	ScrollDistanceBelowViewport int     `json:"scrollDistanceBelowViewport"`
}

// InteractiveElementsData represents the JS evaluation result
type InteractiveElementsData struct {
	Viewport Viewport             `json:"viewport"`
	Elements []InteractiveElement `json:"elements"`
}

type BrowserState struct {
	URL                      string                     `json:"url"`
	Tabs                     []TabInfo                  `json:"tabs"`
	Viewport                 Viewport                   `json:"viewport"`
	ScreenshotWithHighlights string                     `json:"screenshotWithHighlights,omitempty"`
	Screenshot               string                     `json:"screenshot,omitempty"`
	InteractiveElements      map[int]InteractiveElement `json:"interactiveElements"`
}

type ViewportSize struct {
	Width  int
	Height int
}

type BrowserConfig struct {
	CDPURL       string
	ViewportSize ViewportSize
	StorageState map[string]interface{}
	Detector     Detector
}

func DefaultBrowserConfig() BrowserConfig {
	return BrowserConfig{
		ViewportSize: ViewportSize{Width: 1268, Height: 951},
	}
}

func DefaultViewport() Viewport {
	return Viewport{
		Width:            1268,
		Height:           951,
		DevicePixelRatio: 1,
	}
}