package panels

import (
	"encoding/base64"
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BrowserPanel displays browser screenshots and web content
type BrowserPanel struct {
	widget.BaseWidget

	state *client.AppState

	// UI Components
	urlEntry    *widget.Entry
	image       *canvas.Image
	statusLabel *widget.Label
	scroll      *container.Scroll
	emptyLabel  *widget.Label
}

// NewBrowserPanel creates a new browser panel
func NewBrowserPanel(state *client.AppState) *BrowserPanel {
	bp := &BrowserPanel{
		state: state,
	}
	bp.ExtendBaseWidget(bp)
	bp.createUI()
	return bp
}

// createUI creates the browser panel UI components
func (bp *BrowserPanel) createUI() {
	// URL entry
	bp.urlEntry = widget.NewEntry()
	bp.urlEntry.SetPlaceHolder("URL will appear here...")
	bp.urlEntry.Disable()

	// Empty state label
	bp.emptyLabel = widget.NewLabel("No browser activity yet.\n\nWhen the AI uses a browser, screenshots will appear here.")
	bp.emptyLabel.Alignment = fyne.TextAlignCenter
	bp.emptyLabel.Importance = widget.LowImportance

	// Image display
	bp.image = canvas.NewImageFromResource(theme.ComputerIcon())
	bp.image.FillMode = canvas.ImageFillContain
	bp.image.SetMinSize(fyne.NewSize(600, 400))

	// Status label
	bp.statusLabel = widget.NewLabel("Ready")
	bp.statusLabel.Alignment = fyne.TextAlignCenter

	// Scroll container for image
	bp.scroll = container.NewScroll(bp.emptyLabel)
	bp.scroll.SetMinSize(fyne.NewSize(600, 400))
}

// SetScreenshot sets the browser screenshot from base64 data
func (bp *BrowserPanel) SetScreenshot(base64Data string) {
	// Decode base64 data
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		bp.statusLabel.SetText("Error decoding screenshot")
		return
	}

	// Create image from data
	staticResource := fyne.NewStaticResource("screenshot.png", imageData)
	bp.image.Resource = staticResource
	bp.image.Refresh()
	bp.scroll.Content = bp.image
	bp.statusLabel.SetText("Screenshot updated")
}

// SetURL sets the current browser URL
func (bp *BrowserPanel) SetURL(url string) {
	bp.urlEntry.SetText(url)
}

// Refresh updates the browser panel
func (bp *BrowserPanel) Refresh() {
	if bp.state.BrowserURL != "" {
		bp.urlEntry.SetText(bp.state.BrowserURL)
	}

	if len(bp.state.BrowserScreenshot) > 0 {
		staticResource := fyne.NewStaticResource("screenshot.png", bp.state.BrowserScreenshot)
		bp.image.Resource = staticResource
		bp.scroll.Content = bp.image
	}

	bp.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (bp *BrowserPanel) CreateRenderer() fyne.WidgetRenderer {
	// URL bar
	urlBar := container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.FileApplicationIcon()),
		nil,
		bp.urlEntry,
	)

	// Toolbar
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		// TODO: Implement refresh
	})
	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		// TODO: Implement back navigation
	})
	forwardBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		// TODO: Implement forward navigation
	})
	openBrowserBtn := widget.NewButtonWithIcon("Open in Browser", theme.MailForwardIcon(), func() {
		// TODO: Open URL in system browser
	})

	toolbar := container.NewHBox(
		backBtn,
		forwardBtn,
		refreshBtn,
		widget.NewSeparator(),
		openBrowserBtn,
		layout.NewSpacer(),
	)

	// Main content
	content := container.NewBorder(
		container.NewVBox(urlBar, toolbar),  // top
		bp.statusLabel,                       // bottom
		nil,                                  // left
		nil,                                  // right
		bp.scroll,                            // center
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size
func (bp *BrowserPanel) MinSize() fyne.Size {
	return fyne.NewSize(600, 500)
}
