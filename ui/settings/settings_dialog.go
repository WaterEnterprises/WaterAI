package settings

import (
	"water-ai/client"
	"water-ai/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SettingsDialog represents the settings dialog
type SettingsDialog struct {
	parent   fyne.Window
	state    *client.AppState
	wsClient *client.WebSocketClient

	// UI Components
	dialog      dialog.Dialog
	modelEntry  *widget.Select
	apiKeyEntry *widget.Entry
}

// NewSettingsDialog creates a new settings dialog
func NewSettingsDialog(parent fyne.Window, state *client.AppState, wsClient *client.WebSocketClient) *SettingsDialog {
	sd := &SettingsDialog{
		parent:   parent,
		state:    state,
		wsClient: wsClient,
	}
	sd.createUI()
	return sd
}

// createUI creates the settings dialog UI
func (sd *SettingsDialog) createUI() {
	// Create logo
	logoImg := canvas.NewImageFromResource(resources.GetLogoOnly())
	logoImg.SetMinSize(fyne.NewSize(64, 64))
	logoImg.FillMode = canvas.ImageFillContain

	// Title
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Model selection
	sd.modelEntry = widget.NewSelect([]string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-3-5-sonnet",
		"gemini-pro",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
	}, func(selected string) {
		sd.state.SelectedModel = selected
	})
	sd.modelEntry.SetSelected(sd.state.SelectedModel)

	modelFormItem := widget.NewFormItem("Model", sd.modelEntry)

	// API Key entry
	sd.apiKeyEntry = widget.NewPasswordEntry()
	sd.apiKeyEntry.SetPlaceHolder("Enter your API key...")

	apiKeyFormItem := widget.NewFormItem("API Key", sd.apiKeyEntry)

	// Connection status
	connectionStatus := widget.NewLabel("Disconnected")
	if sd.state.IsConnected {
		connectionStatus.SetText("Connected")
	}

	connectionFormItem := widget.NewFormItem("Status", connectionStatus)

	// Workspace path
	workspacePath := widget.NewLabel(sd.state.WorkspacePath)
	if sd.state.WorkspacePath == "" {
		workspacePath.SetText("Not set")
	}

	workspaceFormItem := widget.NewFormItem("Workspace", workspacePath)

	// Form
	form := widget.NewForm(
		modelFormItem,
		apiKeyFormItem,
		connectionFormItem,
		workspaceFormItem,
	)

	// VS Code button
	vscodeBtn := widget.NewButtonWithIcon("Open VS Code", theme.ComputerIcon(), func() {
		// TODO: Open VS Code
		if sd.state.VSCodeURL != "" {
			// Open VS Code URL in browser
		}
	})

	// Save button
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		sd.saveSettings()
	})

	// Cancel button
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		sd.dialog.Hide()
	})

	// Button row
	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	// Main content
	content := container.NewVBox(
		container.NewCenter(logoImg),
		container.NewCenter(title),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		vscodeBtn,
		buttonRow,
	)

	// Create custom dialog
	sd.dialog = dialog.NewCustomWithoutButtons(
		"Settings",
		container.NewVScroll(content),
		sd.parent,
	)
}

// saveSettings saves the settings
func (sd *SettingsDialog) saveSettings() {
	// TODO: Implement settings persistence
	// For now, just update the state
	sd.state.SelectedModel = sd.modelEntry.Selected

	sd.dialog.Hide()
}

// Show shows the settings dialog
func (sd *SettingsDialog) Show() {
	// Update connection status before showing
	if sd.state.IsConnected {
		// Connection status label is set during creation
	}
	sd.dialog.Show()
	sd.dialog.Resize(fyne.NewSize(400, 500))
}

// Hide hides the settings dialog
func (sd *SettingsDialog) Hide() {
	sd.dialog.Hide()
}
