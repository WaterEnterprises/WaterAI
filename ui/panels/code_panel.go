package panels

import (
	"fmt"
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CodePanel displays code files with syntax highlighting
type CodePanel struct {
	widget.BaseWidget

	state *client.AppState

	// UI Components
	fileLabel   *widget.Label
	codeEntry   *widget.Entry
	scroll      *container.Scroll
	emptyLabel  *widget.Label
	lineNumbers *widget.Label
}

// NewCodePanel creates a new code panel
func NewCodePanel(state *client.AppState) *CodePanel {
	cp := &CodePanel{
		state: state,
	}
	cp.ExtendBaseWidget(cp)
	cp.createUI()
	return cp
}

// createUI creates the code panel UI components
func (cp *CodePanel) createUI() {
	// File label
	cp.fileLabel = widget.NewLabel("No file selected")
	cp.fileLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Empty state label
	cp.emptyLabel = widget.NewLabel("No code to display yet.\n\nWhen the AI reads or writes code, it will appear here.")
	cp.emptyLabel.Alignment = fyne.TextAlignCenter
	cp.emptyLabel.Importance = widget.LowImportance

	// Line numbers
	cp.lineNumbers = widget.NewLabel("")
	cp.lineNumbers.TextStyle = fyne.TextStyle{Monospace: true}
	cp.lineNumbers.Importance = widget.LowImportance

	// Code entry (read-only)
	cp.codeEntry = widget.NewMultiLineEntry()
	cp.codeEntry.SetPlaceHolder("Code will appear here...")
	cp.codeEntry.Wrapping = fyne.TextWrapWord
	cp.codeEntry.TextStyle = fyne.TextStyle{Monospace: true}
	cp.codeEntry.Disable() // Read-only

	// Scroll container
	cp.scroll = container.NewScroll(cp.emptyLabel)
	cp.scroll.SetMinSize(fyne.NewSize(600, 400))
}

// SetContent sets the code content
func (cp *CodePanel) SetContent(content string) {
	cp.codeEntry.SetText(content)
	cp.scroll.Content = container.NewHSplit(
		cp.lineNumbers,
		cp.codeEntry,
	)
	cp.scroll.Content.(*container.Split).SetOffset(0.05)
	cp.updateLineNumbers(content)
}

// SetFile sets the current file name
func (cp *CodePanel) SetFile(filename string) {
	cp.fileLabel.SetText(filename)
}

// updateLineNumbers updates the line numbers display
func (cp *CodePanel) updateLineNumbers(content string) {
	if content == "" {
		cp.lineNumbers.SetText("")
		return
	}

	// Count lines
	lines := 1
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}

	// Generate line numbers
	numStr := ""
	for i := 1; i <= lines; i++ {
		numStr += fmt.Sprintf("%d\n", i)
	}
	cp.lineNumbers.SetText(numStr)
}

// Refresh updates the code panel
func (cp *CodePanel) Refresh() {
	if cp.state.CodeContent != "" {
		cp.codeEntry.SetText(cp.state.CodeContent)
		cp.updateLineNumbers(cp.state.CodeContent)
		cp.scroll.Content = container.NewHSplit(
			cp.lineNumbers,
			cp.codeEntry,
		)
	}

	if cp.state.CodeFile != "" {
		cp.fileLabel.SetText(cp.state.CodeFile)
	}

	cp.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (cp *CodePanel) CreateRenderer() fyne.WidgetRenderer {
	// Copy button
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		// Copy to clipboard
		if cp.codeEntry.Text != "" {
			fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(cp.codeEntry.Text)
		}
	})

	// Language label
	langLabel := widget.NewLabel("")
	langLabel.Importance = widget.LowImportance

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewIcon(theme.FileTextIcon()),
		cp.fileLabel,
		layout.NewSpacer(),
		langLabel,
		copyBtn,
	)

	// Main content
	content := container.NewBorder(
		toolbar,   // top
		nil,       // bottom
		nil,       // left
		nil,       // right
		cp.scroll, // center
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size
func (cp *CodePanel) MinSize() fyne.Size {
	return fyne.NewSize(600, 500)
}
