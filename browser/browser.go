package browser

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/playwright-community/playwright-go"
)

// Browser responsible for interacting with the browser via Playwright.
type Browser struct {
	Config            BrowserConfig
	CloseContext      bool
	playwright        *playwright.Playwright
	playwrightBrowser playwright.Browser
	context           playwright.BrowserContext
	currentPage       playwright.Page
	state             *BrowserState
	cdpSession        playwright.CDPSession
	detector          Detector
	
	ScreenshotScaleFactor float64
}

// NewBrowser initializes the browser structure.
func NewBrowser(config BrowserConfig, closeContext bool) *Browser {
	return &Browser{
		Config:       config,
		CloseContext: closeContext,
		detector:     config.Detector,
		state:        initState(""),
	}
}

func initState(url string) *BrowserState {
	return &BrowserState{
		URL:                 url,
		Tabs:                []TabInfo{},
		InteractiveElements: make(map[int]InteractiveElement),
		Viewport:            DefaultViewport(),
	}
}

// Init initializes the Playwright instance, browser, and context.
func (b *Browser) Init() error {
	log.Println("Initializing browser")
	var err error

	if b.playwright == nil {
		b.playwright, err = playwright.Run()
		if err != nil {
			return fmt.Errorf("could not start playwright: %w", err)
		}
	}

	if b.playwrightBrowser == nil {
		if b.Config.CDPURL != "" {
			log.Printf("Connecting to remote browser via CDP %s", b.Config.CDPURL)
			err = retry.Do(
				func() error {
					b.playwrightBrowser, err = b.playwright.Chromium.ConnectOverCDP(b.Config.CDPURL, playwright.BrowserTypeConnectOverCDPOptions{
						Timeout: playwright.Float(2500),
					})
					return err
				},
				retry.Attempts(3),
				retry.Delay(1*time.Second),
			)
			if err != nil {
				return fmt.Errorf("failed to connect over CDP: %w", err)
			}
		} else {
			log.Println("Launching new browser instance")
			b.playwrightBrowser, err = b.playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
				Headless: playwright.Bool(false),
				Args: []string{
					"--no-sandbox",
					"--disable-blink-features=AutomationControlled",
					"--disable-web-security",
					"--disable-site-isolation-trials",
					"--disable-features=IsolateOrigins,site-per-process",
					fmt.Sprintf("--window-size=%d,%d", b.Config.ViewportSize.Width, b.Config.ViewportSize.Height),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to launch browser: %w", err)
			}
		}
	}

	if b.context == nil {
		if len(b.playwrightBrowser.Contexts()) > 0 {
			b.context = b.playwrightBrowser.Contexts()[0]
		} else {
			b.context, err = b.playwrightBrowser.NewContext(playwright.BrowserNewContextOptions{
				Viewport: &playwright.Size{
					Width:  b.Config.ViewportSize.Width,
					Height: b.Config.ViewportSize.Height,
				},
				UserAgent:         playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36"),
				JavaScriptEnabled: playwright.Bool(true),
				BypassCSP:         playwright.Bool(true),
				IgnoreHttpsErrors: playwright.Bool(true),
			})
			if err != nil {
				return fmt.Errorf("failed to create context: %w", err)
			}
		}
		if err := b.applyAntiDetectionScripts(); err != nil {
			return err
		}
	}

	b.context.On("page", func(page playwright.Page) {
		b.onPageChange(page)
	})

	// Add Cookies from storage state if present
	if cookies, ok := b.Config.StorageState["cookies"]; ok {
		// Needs robust type assertion in real world, assuming generic map slice here
		if cookieList, ok := cookies.([]interface{}); ok {
			var pCookies []playwright.OptionalCookie
			// conversion logic omitted for brevity, essentially marshalling to OptionalCookie struct
			// In Go, usually easiest to marshal to JSON then back to proper struct
			bytes, _ := json.Marshal(cookieList)
			json.Unmarshal(bytes, &pCookies)
			b.context.AddCookies(pCookies)
		}
	}

	if b.currentPage == nil {
		if len(b.context.Pages()) > 0 {
			b.currentPage = b.context.Pages()[len(b.context.Pages())-1]
		} else {
			b.currentPage, err = b.context.NewPage()
			if err != nil {
				return fmt.Errorf("failed to create page: %w", err)
			}
		}
	}

	return nil
}

func (b *Browser) onPageChange(page playwright.Page) {
	log.Printf("Current page changed to %s", page.URL())
	var err error
	b.cdpSession, err = b.context.NewCDPSession(page)
	if err != nil {
		log.Printf("Failed to create CDP session on page change: %v", err)
		return
	}

	// Set metrics
	params := map[string]interface{}{
		"width":             b.Config.ViewportSize.Width,
		"height":            b.Config.ViewportSize.Height,
		"deviceScaleFactor": 1,
		"mobile":            false,
	}
	b.cdpSession.Send("Emulation.setDeviceMetricsOverride", params)
	
	b.cdpSession.Send("Emulation.setVisibleSize", map[string]interface{}{
		"width":  b.Config.ViewportSize.Width,
		"height": b.Config.ViewportSize.Height,
	})

	b.currentPage = page
}

func (b *Browser) applyAntiDetectionScripts() error {
	script := `
		// Webdriver property
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		// Languages
		Object.defineProperty(navigator, 'languages', { get: () => ['en-US'] });
		// Plugins
		Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
		// Chrome runtime
		window.chrome = { runtime: {} };
		// Permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
		(function () {
			const originalAttachShadow = Element.prototype.attachShadow;
			Element.prototype.attachShadow = function attachShadow(options) {
				return originalAttachShadow.call(this, { ...options, mode: "open" });
			};
		})();
	`
	err := b.context.AddInitScript(playwright.Script{Content: playwright.String(script)})
	return err
}

func (b *Browser) Close() error {
	log.Println("Closing browser")
	if b.cdpSession != nil {
		b.cdpSession.Detach()
		b.cdpSession = nil
	}
	if b.context != nil {
		b.context.Close()
		b.context = nil
	}
	if b.playwrightBrowser != nil {
		b.playwrightBrowser.Close()
		b.playwrightBrowser = nil
	}
	if b.playwright != nil {
		b.playwright.Stop()
		b.playwright = nil
	}
	b.currentPage = nil
	b.state = nil
	return nil
}

func (b *Browser) Restart() error {
	b.Close()
	return b.Init()
}

func (b *Browser) Goto(url string) error {
	page, err := b.GetCurrentPage()
	if err != nil {
		return err
	}
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

func (b *Browser) GetTabsInfo() ([]TabInfo, error) {
	var tabs []TabInfo
	for i, page := range b.context.Pages() {
		title, _ := page.Title()
		tabs = append(tabs, TabInfo{
			PageID: i,
			URL:    page.URL(),
			Title:  title,
		})
	}
	return tabs, nil
}

func (b *Browser) SwitchToTab(pageID int) error {
	if b.context == nil {
		if err := b.Init(); err != nil {
			return err
		}
	}
	pages := b.context.Pages()
	if pageID >= len(pages) || pageID < 0 {
		return fmt.Errorf("no tab found with page_id: %d", pageID)
	}
	
	page := pages[pageID]
	b.currentPage = page
	page.BringToFront()
	page.WaitForLoadState()
	return nil
}

func (b *Browser) CreateNewTab(url string) error {
	if b.context == nil {
		if err := b.Init(); err != nil {
			return err
		}
	}
	newPage, err := b.context.NewPage()
	if err != nil {
		return err
	}
	b.currentPage = newPage
	newPage.WaitForLoadState()
	if url != "" {
		_, err := newPage.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		return err
	}
	return nil
}

func (b *Browser) CloseCurrentTab() error {
	if b.currentPage == nil {
		return nil
	}
	b.currentPage.Close()
	
	if b.context != nil && len(b.context.Pages()) > 0 {
		return b.SwitchToTab(0)
	}
	return nil
}

func (b *Browser) GetCurrentPage() (playwright.Page, error) {
	if b.currentPage == nil {
		if err := b.Init(); err != nil {
			return nil, err
		}
	}
	return b.currentPage, nil
}

func (b *Browser) GetState() *BrowserState {
	return b.state
}

func (b *Browser) UpdateState() (*BrowserState, error) {
	var err error
	b.state, err = b.updateStateInternal()
	return b.state, err
}

func (b *Browser) updateStateInternal() (*BrowserState, error) {
	var state *BrowserState

	err := retry.Do(
		func() error {
			if b.currentPage == nil {
				if err := b.Init(); err != nil {
					return err
				}
			}
			url := b.currentPage.URL()
			detectSheets := strings.Contains(url, "docs.google.com/spreadsheets/d")

			screenshotB64, err := b.FastScreenshot()
			if err != nil {
				return err
			}

			data, err := b.GetInteractiveElements(screenshotB64, detectSheets)
			if err != nil {
				return err
			}

			interactiveElements := make(map[int]InteractiveElement)
			for _, el := range data.Elements {
				interactiveElements[el.Index] = el
			}

			highlightScreenshot := PutHighlightElementsOnScreenshot(interactiveElements, screenshotB64)
			tabs, _ := b.GetTabsInfo()

			state = &BrowserState{
				URL:                      url,
				Tabs:                     tabs,
				ScreenshotWithHighlights: highlightScreenshot,
				Screenshot:               screenshotB64,
				Viewport:                 data.Viewport,
				InteractiveElements:      interactiveElements,
			}
			return nil
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
	)

	if err != nil {
		log.Printf("Failed to update state after multiple attempts: %v", err)
		if b.state != nil {
			return b.state, nil
		}
		return nil, err
	}
	return state, nil
}

func (b *Browser) DetectBrowserElements() (InteractiveElementsData, error) {
	page, err := b.GetCurrentPage()
	if err != nil {
		return InteractiveElementsData{}, err
	}

	result, err := page.Evaluate(InteractiveElementsJSCode)
	if err != nil {
		return InteractiveElementsData{}, err
	}

	// Unmarshal result map to struct
	bytes, _ := json.Marshal(result)
	var data InteractiveElementsData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return InteractiveElementsData{}, err
	}
	return data, nil
}

func (b *Browser) GetInteractiveElements(screenshotB64 string, detectSheets bool) (InteractiveElementsData, error) {
	var elements []InteractiveElement
	browserData, err := b.DetectBrowserElements()
	if err != nil {
		return InteractiveElementsData{}, err
	}

	if b.detector != nil {
		scaleFactor := float64(browserData.Viewport.Width) / 1024.0
		cvElements, err := b.detector.DetectFromImage(screenshotB64, scaleFactor, detectSheets)
		if err == nil {
			elements = append(browserData.Elements, cvElements...)
			elements = FilterElements(elements, 0.7)
		} else {
			elements = browserData.Elements
		}
	} else {
		elements = browserData.Elements
	}

	return InteractiveElementsData{
		Viewport: browserData.Viewport,
		Elements: elements,
	}, nil
}

func (b *Browser) GetCDPSession() (playwright.CDPSession, error) {
	// Simplified check: Playwright Go doesn't expose _page easily, 
	// relying on onPageChange management
	if b.cdpSession == nil {
		var err error
		b.cdpSession, err = b.context.NewCDPSession(b.currentPage)
		if err != nil {
			return nil, err
		}
	}
	return b.cdpSession, nil
}

func (b *Browser) FastScreenshot() (string, error) {
	session, err := b.GetCDPSession()
	if err != nil {
		return "", err
	}

	params := map[string]interface{}{
		"format":                "png",
		"fromSurface":           false,
		"captureBeyondViewport": false,
	}

	result, err := session.Send("Page.captureScreenshot", params)
	if err != nil {
		return "", err
	}

	// Result data is base64 string in "data" key
	var resultData struct {
		Data string `json:"data"`
	}
	
	// playwright-go Send returns interface{}, we need to handle it.
	// Actually, Send returns (interface{}, error). The underlying implementation unmarshals JSON.
	// We need to marshal and unmarshal if it comes back as map[string]interface{}
	jsonBytes, _ := json.Marshal(result)
	json.Unmarshal(jsonBytes, &resultData)

	return ScaleB64Image(resultData.Data, b.ScreenshotScaleFactor), nil
}

func (b *Browser) HandlePDFURLNavigation() (*BrowserState, error) {
	page, err := b.GetCurrentPage()
	if err != nil {
		return nil, err
	}
	
	if IsPDFURL(page.URL()) {
		time.Sleep(5 * time.Second)
		page.Keyboard().Press("Escape")
		time.Sleep(100 * time.Millisecond)
		page.Keyboard().Press("Control+\\")
		time.Sleep(100 * time.Millisecond)
		
		// Click specific coordinates
		w := float64(b.Config.ViewportSize.Width)
		h := float64(b.Config.ViewportSize.Height)
		page.Mouse().Click(w*0.75, h*0.25)
	}

	return b.UpdateState()
}

// GetCookies returns cookies
func (b *Browser) GetCookies() ([]playwright.Cookie, error) {
	if b.context != nil {
		return b.context.Cookies()
	}
	return nil, nil
}

// GetStorageState returns cookies wrapper
func (b *Browser) GetStorageState() (map[string]interface{}, error) {
	if b.context != nil {
		cookies, err := b.context.Cookies()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"cookies": cookies}, nil
	}
	return map[string]interface{}{}, nil
}