package ui

import (
	"water-ai/client"
	"water-ai/resources"
	"water-ai/ui/chat"
	"water-ai/ui/panels"
	"water-ai/ui/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	serverURL = "ws://localhost:7777/ws"
)

// MainWindow represents the main application window
type MainWindow struct {
	app         fyne.App
	window      fyne.Window
	state       *client.AppState
	wsClient    *client.WebSocketClient

	// UI Components
	chatView       *chat.ChatView
	browserPanel   *panels.BrowserPanel
	codePanel      *panels.CodePanel
	terminalPanel  *panels.TerminalPanel
	settingsDialog *settings.SettingsDialog

	// Tabs
	panelTabs *container.AppTabs

	// Status
	connectionStatus *widget.Label
	connectionIcon   *widget.Icon
	workspaceLabel   *widget.Label
}

// NewMainWindow creates a new main window
func NewMainWindow(app fyne.App) *MainWindow {
	mw := &MainWindow{
		app:   app,
		state: client.NewAppState(),
	}

	// Initialize WebSocket client
	mw.wsClient = client.NewWebSocketClient(serverURL, mw.state)
	mw.wsClient.SetOnStateChange(mw.onStateChange)
	mw.wsClient.SetOnEvent(mw.onEvent)
	mw.wsClient.SetOnConnected(mw.onConnected)
	mw.wsClient.SetOnDisconnected(mw.onDisconnected)

	// Create the window
	mw.window = app.NewWindow("Water AI")

	// Set window size
	mw.window.Resize(fyne.NewSize(1200, 800))

	// Set window icon
	mw.window.SetIcon(resources.GetLogoOnly())

	// Create UI components
	mw.createUI()

	// Set up window close handler
	mw.window.SetCloseIntercept(mw.onClose)

	// Set up keyboard shortcuts
	mw.setupKeyboardShortcuts()

	return mw
}

// createUI creates all UI components
func (mw *MainWindow) createUI() {
	// Create chat view
	mw.chatView = chat.NewChatView(mw.state, mw.wsClient)

	// Create panel views
	mw.browserPanel = panels.NewBrowserPanel(mw.state)
	mw.codePanel = panels.NewCodePanel(mw.state)
	mw.terminalPanel = panels.NewTerminalPanel(mw.state)

	// Create settings dialog
	mw.settingsDialog = settings.NewSettingsDialog(mw.window, mw.state, mw.wsClient)

	// Create tabbed panel
	mw.panelTabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Browser", theme.ComputerIcon(), mw.browserPanel),
		container.NewTabItemWithIcon("Code", theme.FileTextIcon(), mw.codePanel),
		container.NewTabItemWithIcon("Terminal", theme.DocumentIcon(), mw.terminalPanel),
	)

	// Create header
	header := mw.createHeader()

	// Create status bar
	statusBar := mw.createStatusBar()

	// Create main content with split layout
	// Left side: Chat view (40%)
	// Right side: Tabbed panels (60%)
	content := container.NewHSplit(
		mw.chatView,
		mw.panelTabs,
	)
	content.SetOffset(0.4)

	// Main layout
	mainLayout := container.NewBorder(
		header,      // top
		statusBar,   // bottom
		nil,         // left
		nil,         // right
		content,     // center
	)

	mw.window.SetContent(mainLayout)
}

// createHeader creates the header bar
func (mw *MainWindow) createHeader() fyne.CanvasObject {
	// Create logo image
	logoImg := canvas.NewImageFromResource(resources.GetLogoOnly())
	logoImg.SetMinSize(fyne.NewSize(32, 32))
	logoImg.FillMode = canvas.ImageFillContain

	title := widget.NewLabelWithStyle("Water AI", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// New chat button
	newChatBtn := widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), mw.onNewChat)

	// Settings button
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		mw.settingsDialog.Show()
	})

	return container.NewBorder(
		nil, nil,
		container.NewHBox(
			logoImg,
			title,
		),
		container.NewHBox(
			newChatBtn,
			settingsBtn,
		),
	)
}

// createStatusBar creates the status bar
func (mw *MainWindow) createStatusBar() fyne.CanvasObject {
	// Connection status
	mw.connectionIcon = widget.NewIcon(theme.CancelIcon())
	mw.connectionStatus = widget.NewLabel("Disconnected")
	mw.connectionStatus.Importance = widget.LowImportance

	// Workspace path
	workspaceLabel := widget.NewLabel("")
	workspaceLabel.Importance = widget.LowImportance

	// Store reference for updates
	mw.workspaceLabel = workspaceLabel

	return container.NewBorder(
		nil, nil,
		container.NewHBox(
			mw.connectionIcon,
			mw.connectionStatus,
			widget.NewSeparator(),
			workspaceLabel,
		),
		nil,
	)
}

// setupKeyboardShortcuts sets up keyboard shortcuts
func (mw *MainWindow) setupKeyboardShortcuts() {
	// Ctrl+N: New chat
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}, func(_ fyne.Shortcut) {
		mw.onNewChat()
	})

	// Ctrl+,: Settings
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierControl}, func(_ fyne.Shortcut) {
		mw.settingsDialog.Show()
	})

	// Ctrl+Q: Quit
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}, func(_ fyne.Shortcut) {
		mw.onClose()
	})

	// F5: Refresh connection
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF5}, func(_ fyne.Shortcut) {
		mw.reconnect()
	})
}

// onNewChat handles new chat action
func (mw *MainWindow) onNewChat() {
	// Clear messages
	mw.state.ClearMessages()
	mw.state.IsAgentInitialized = false

	// Refresh UI
	mw.chatView.Refresh()
}

// reconnect attempts to reconnect to the server
func (mw *MainWindow) reconnect() {
	if mw.state.IsConnected {
		mw.wsClient.Disconnect()
	}

	mw.updateConnectionStatus(false, "Reconnecting...")

	go func() {
		if err := mw.wsClient.Connect(); err != nil {
			mw.app.SendNotification(&fyne.Notification{
				Title:   "Connection Error",
				Content: "Failed to connect to server: " + err.Error(),
			})
		}
	}()
}

// updateConnectionStatus updates the connection status display
func (mw *MainWindow) updateConnectionStatus(connected bool, status string) {
	mw.connectionStatus.SetText(status)
	if connected {
		mw.connectionIcon.SetResource(theme.ConfirmIcon())
	} else {
		mw.connectionIcon.SetResource(theme.CancelIcon())
	}
}

// onStateChange handles state changes
func (mw *MainWindow) onStateChange() {
	// Update UI components on the main thread
	fyne.Do(func() {
		mw.chatView.Refresh()
		mw.browserPanel.Refresh()
		mw.codePanel.Refresh()
		mw.terminalPanel.Refresh()

		// Update workspace label
		if mw.workspaceLabel != nil && mw.state.WorkspacePath != "" {
			mw.workspaceLabel.SetText("📁 " + mw.state.WorkspacePath)
		}
	})
}

// onEvent handles WebSocket events
func (mw *MainWindow) onEvent(eventType string, content interface{}) {
	// Handle specific events on the main thread
	fyne.Do(func() {
		switch eventType {
		case client.EventTypeToolCall:
			if tc, ok := content.(client.ToolCallEvent); ok {
				mw.handleToolCall(tc)
			}
		case client.EventTypeToolResult:
			if tr, ok := content.(client.ToolResultEvent); ok {
				mw.handleToolResult(tr)
			}
		case client.EventTypeProcessing:
			mw.chatView.SetLoadingText("Processing...")
			mw.chatView.ShowLoading()
		case client.EventTypeStreamComplete:
			mw.chatView.HideLoading()
			mw.state.IsLoading = false
		}
	})
}

// handleToolCall handles tool call events
func (mw *MainWindow) handleToolCall(tc client.ToolCallEvent) {
	// Switch to appropriate tab based on tool
	switch tc.ToolName {
	case "browser_view", "browser_click", "browser_enter_text", "browser_navigate", "browser_screenshot":
		mw.panelTabs.SelectIndex(0) // Browser tab
		mw.chatView.SetLoadingText("Browsing...")
	case "write_file", "read_file", "edit_file":
		mw.panelTabs.SelectIndex(1) // Code tab
		mw.chatView.SetLoadingText("Working on code...")
	case "execute_command":
		mw.panelTabs.SelectIndex(2) // Terminal tab
		mw.chatView.SetLoadingText("Running command...")
	default:
		mw.chatView.SetLoadingText("Working...")
	}
}

// handleToolResult handles tool result events
func (mw *MainWindow) handleToolResult(tr client.ToolResultEvent) {
	// Update panels based on tool result
	switch tr.ToolName {
	case "browser_view", "browser_screenshot":
		if screenshot, ok := tr.Result.(string); ok {
			mw.browserPanel.SetScreenshot(screenshot)
		}
	case "write_file", "read_file":
		if content, ok := tr.Result.(string); ok {
			mw.codePanel.SetContent(content)
		}
	case "execute_command":
		if output, ok := tr.Result.(string); ok {
			mw.terminalPanel.AppendOutput(output)
		}
	}
}

// onConnected handles connection established
func (mw *MainWindow) onConnected() {
	fyne.Do(func() {
		mw.updateConnectionStatus(true, "Connected")
	})
}

// onDisconnected handles disconnection
func (mw *MainWindow) onDisconnected() {
	fyne.Do(func() {
		mw.updateConnectionStatus(false, "Disconnected")
	})
}

// onClose handles window close
func (mw *MainWindow) onClose() {
	// Show confirmation dialog
	dialog.ShowConfirm(
		"Quit Water AI?",
		"Are you sure you want to quit?",
		func(confirmed bool) {
			if confirmed {
				mw.wsClient.Disconnect()
				mw.window.Close()
			}
		},
		mw.window,
	)
}

// ShowAndRun shows the window and runs the application
func (mw *MainWindow) ShowAndRun() {
	// Show the window
	mw.window.Show()

	// Set initial connection status
	mw.updateConnectionStatus(false, "Connecting...")

	// Attempt to connect to the server
	go func() {
		if err := mw.wsClient.Connect(); err != nil {
			// Show error dialog on main thread
			mw.app.SendNotification(&fyne.Notification{
				Title:   "Connection Error",
				Content: "Failed to connect to server: " + err.Error(),
			})
		}
	}()

	// Run the application
	mw.app.Run()
}

// Refresh refreshes the UI
func (mw *MainWindow) Refresh() {
	mw.window.Content().Refresh()
}
