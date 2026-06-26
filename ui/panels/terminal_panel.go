package panels

import (
	"strings"
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TerminalPanel displays terminal output
type TerminalPanel struct {
	widget.BaseWidget

	state *client.AppState

	// UI Components
	output      *widget.Label
	scroll      *container.Scroll
	outputText  strings.Builder
	emptyLabel  *widget.Label
}

// NewTerminalPanel creates a new terminal panel
func NewTerminalPanel(state *client.AppState) *TerminalPanel {
	tp := &TerminalPanel{
		state: state,
	}
	tp.ExtendBaseWidget(tp)
	tp.createUI()
	return tp
}

// createUI creates the terminal panel UI components
func (tp *TerminalPanel) createUI() {
	// Empty state label
	tp.emptyLabel = widget.NewLabel("No terminal output yet.\n\nWhen the AI runs commands, output will appear here.")
	tp.emptyLabel.Alignment = fyne.TextAlignCenter
	tp.emptyLabel.Importance = widget.LowImportance

	// Output label (monospace)
	tp.output = widget.NewLabel("")
	tp.output.TextStyle = fyne.TextStyle{Monospace: true}
	tp.output.Wrapping = fyne.TextWrapWord
	tp.output.Alignment = fyne.TextAlignLeading

	// Scroll container
	tp.scroll = container.NewVScroll(tp.emptyLabel)
	tp.scroll.SetMinSize(fyne.NewSize(600, 400))
}

// AppendOutput appends text to the terminal output
func (tp *TerminalPanel) AppendOutput(text string) {
	tp.outputText.WriteString(text)
	tp.outputText.WriteString("\n")
	tp.output.SetText(tp.outputText.String())
	tp.scroll.Content = tp.output
	tp.scroll.ScrollToBottom()
}

// ClearOutput clears the terminal output
func (tp *TerminalPanel) ClearOutput() {
	tp.outputText.Reset()
	tp.output.SetText("")
	tp.scroll.Content = tp.emptyLabel
}

// Refresh updates the terminal panel
func (tp *TerminalPanel) Refresh() {
	if tp.state.TerminalOutput != "" {
		tp.output.SetText(tp.state.TerminalOutput)
		tp.scroll.Content = tp.output
	}
	tp.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (tp *TerminalPanel) CreateRenderer() fyne.WidgetRenderer {
	// Clear button
	clearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		tp.ClearOutput()
	})

	// Copy button
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if tp.output.Text != "" {
			fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(tp.output.Text)
		}
	})

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewIcon(theme.DocumentIcon()),
		widget.NewLabel("Terminal Output"),
		layout.NewSpacer(),
		copyBtn,
		clearBtn,
	)

	// Main content with dark background
	content := container.NewBorder(
		toolbar,   // top
		nil,       // bottom
		nil,       // left
		nil,       // right
		tp.scroll, // center
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size
func (tp *TerminalPanel) MinSize() fyne.Size {
	return fyne.NewSize(600, 500)
}
