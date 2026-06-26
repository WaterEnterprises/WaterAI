package chat

import (
	"fmt"
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// InputArea represents the chat input area
type InputArea struct {
	widget.BaseWidget

	state    *client.AppState
	wsClient *client.WebSocketClient

	// UI Components
	entry      *widget.Entry
	sendBtn    *widget.Button
	cancelBtn  *widget.Button
	attachBtn  *widget.Button
	fileLabel  *widget.Label

	// State
	attachedFiles []string

	// Callbacks
	OnSubmit func(text string)
}

// NewInputArea creates a new input area
func NewInputArea(state *client.AppState, wsClient *client.WebSocketClient) *InputArea {
	ia := &InputArea{
		state:         state,
		wsClient:      wsClient,
		attachedFiles: []string{},
	}
	ia.ExtendBaseWidget(ia)
	ia.createUI()
	return ia
}

// createUI creates the input area UI components
func (ia *InputArea) createUI() {
	// Create multi-line entry
	ia.entry = widget.NewMultiLineEntry()
	ia.entry.SetPlaceHolder("Give Water AI a task to work on...")
	ia.entry.Wrapping = fyne.TextWrapWord
	ia.entry.SetMinRowsVisible(3)

	// Handle Enter key (submit on Enter, newline on Shift+Enter)
	ia.entry.OnSubmitted = func(text string) {
		if ia.OnSubmit != nil {
			ia.OnSubmit(text)
		}
	}

	// Create file label
	ia.fileLabel = widget.NewLabel("")
	ia.fileLabel.Importance = widget.LowImportance

	// Create send button
	ia.sendBtn = widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
		if ia.OnSubmit != nil {
			ia.OnSubmit(ia.entry.Text)
		}
	})

	// Create cancel button
	ia.cancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		ia.wsClient.CancelQuery()
	})

	// Create attach button
	ia.attachBtn = widget.NewButtonWithIcon("", theme.DocumentIcon(), ia.showFilePicker)
}

// showFilePicker shows a file picker dialog
func (ia *InputArea) showFilePicker() {
	// Get the window from the current focus
	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if uc == nil {
			return
		}
		defer uc.Close()

		// Add file to attached files
		ia.attachedFiles = append(ia.attachedFiles, uc.URI().Path())
		ia.updateFileLabel()
	}, win)
}

// updateFileLabel updates the file label text
func (ia *InputArea) updateFileLabel() {
	if len(ia.attachedFiles) == 0 {
		ia.fileLabel.SetText("")
	} else if len(ia.attachedFiles) == 1 {
		ia.fileLabel.SetText("📎 1 file attached")
	} else {
		ia.fileLabel.SetText(fmt.Sprintf("📎 %d files attached", len(ia.attachedFiles)))
	}
}

// SetText sets the entry text
func (ia *InputArea) SetText(text string) {
	ia.entry.SetText(text)
}

// GetText returns the entry text
func (ia *InputArea) GetText() string {
	return ia.entry.Text
}

// GetAttachedFiles returns the list of attached files
func (ia *InputArea) GetAttachedFiles() []string {
	return ia.attachedFiles
}

// ClearAttachedFiles clears the attached files
func (ia *InputArea) ClearAttachedFiles() {
	ia.attachedFiles = []string{}
	ia.updateFileLabel()
}

// Refresh updates the input area state
func (ia *InputArea) Refresh() {
	// Update button states based on loading state
	if ia.state.IsLoading {
		ia.sendBtn.Disable()
		ia.cancelBtn.Enable()
	} else {
		ia.sendBtn.Enable()
		ia.cancelBtn.Disable()
	}

	// Update based on connection state
	if !ia.state.IsConnected {
		ia.sendBtn.Disable()
		ia.entry.SetPlaceHolder("Connecting to server...")
	} else {
		ia.entry.SetPlaceHolder("Give Water AI a task to work on...")
	}

	ia.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (ia *InputArea) CreateRenderer() fyne.WidgetRenderer {
	// Create button row
	buttonRow := container.NewHBox(
		ia.attachBtn,
		ia.fileLabel,
		container.NewCenter(widget.NewLabel("Shift+Enter for newline")),
		layout.NewSpacer(),
		ia.cancelBtn,
		ia.sendBtn,
	)

	// Create main layout
	content := container.NewBorder(
		nil,        // top
		buttonRow,  // bottom
		nil,        // left
		nil,        // right
		ia.entry,   // center
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size
func (ia *InputArea) MinSize() fyne.Size {
	return fyne.NewSize(400, 100)
}

// FileDropHandler handles file drops
func (ia *InputArea) FileDropHandler() func([]fyne.URI) {
	return func(uris []fyne.URI) {
		for _, uri := range uris {
			ia.attachedFiles = append(ia.attachedFiles, uri.Path())
		}
		ia.updateFileLabel()
		ia.Refresh()
	}
}

// OnKeyDown handles keyboard shortcuts
func (ia *InputArea) OnKeyDown(key fyne.KeyName) {
	switch key {
	case fyne.KeyReturn:
		// Submit on Enter (without Shift)
		if ia.OnSubmit != nil && ia.entry.Text != "" {
			ia.OnSubmit(ia.entry.Text)
		}
	case fyne.KeyEscape:
		// Clear input on Escape
		ia.entry.SetText("")
		ia.ClearAttachedFiles()
	}
}
