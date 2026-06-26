package chat

import (
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ChatView represents the chat interface
type ChatView struct {
	widget.BaseWidget

	state    *client.AppState
	wsClient *client.WebSocketClient

	// UI Components
	messageList   *MessageList
	inputArea     *InputArea
	scroll        *container.Scroll
	loadingLabel  *widget.Label
	loadingBox    *fyne.Container
}

// NewChatView creates a new chat view
func NewChatView(state *client.AppState, wsClient *client.WebSocketClient) *ChatView {
	cv := &ChatView{
		state:    state,
		wsClient: wsClient,
	}

	cv.ExtendBaseWidget(cv)
	cv.createUI()

	return cv
}

// createUI creates the chat UI components
func (cv *ChatView) createUI() {
	// Create message list
	cv.messageList = NewMessageList(cv.state)

	// Create scroll container for messages
	cv.scroll = container.NewScroll(cv.messageList)
	cv.scroll.SetMinSize(fyne.NewSize(400, 500))

	// Create loading indicator
	cv.loadingLabel = widget.NewLabel("Thinking...")
	cv.loadingLabel.Importance = widget.LowImportance
	cv.loadingBox = container.NewHBox(
		widget.NewActivity(),
		cv.loadingLabel,
	)
	cv.loadingBox.Hide()

	// Create input area
	cv.inputArea = NewInputArea(cv.state, cv.wsClient)
	cv.inputArea.OnSubmit = cv.handleSubmit
}

// handleSubmit handles message submission
func (cv *ChatView) handleSubmit(text string) {
	if text == "" {
		return
	}

	// Add user message to state
	msg := client.NewMessage("user", text)
	cv.state.AddMessage(msg)

	// Clear input
	cv.inputArea.SetText("")

	// Get attached files
	files := cv.inputArea.GetAttachedFiles()
	cv.inputArea.ClearAttachedFiles()

	// Show loading indicator
	cv.state.IsLoading = true
	cv.loadingBox.Show()

	// Initialize agent if not already done
	if !cv.state.IsAgentInitialized {
		cv.wsClient.InitAgent(cv.state.SelectedModel, map[string]interface{}{}, 0)
	}

	// Send query with files
	cv.wsClient.SendQuery(text, len(cv.state.Messages) > 1, files)

	// Refresh UI
	cv.Refresh()

	// Scroll to bottom
	cv.scrollToBottom()
}

// SetLoadingText sets the loading indicator text
func (cv *ChatView) SetLoadingText(text string) {
	cv.loadingLabel.SetText(text)
}

// ShowLoading shows the loading indicator
func (cv *ChatView) ShowLoading() {
	cv.loadingBox.Show()
}

// HideLoading hides the loading indicator
func (cv *ChatView) HideLoading() {
	cv.loadingBox.Hide()
}

// scrollToBottom scrolls the message list to the bottom
func (cv *ChatView) scrollToBottom() {
	cv.scroll.ScrollToBottom()
}

// Refresh refreshes the chat view
func (cv *ChatView) Refresh() {
	cv.messageList.Refresh()
	cv.inputArea.Refresh()

	// Update loading state
	if cv.state.IsLoading {
		cv.loadingBox.Show()
	} else {
		cv.loadingBox.Hide()
	}

	cv.BaseWidget.Refresh()

	// Scroll to bottom when new messages arrive
	cv.scrollToBottom()
}

// CreateRenderer creates the widget renderer
func (cv *ChatView) CreateRenderer() fyne.WidgetRenderer {
	// Create the layout with loading indicator
	content := container.NewBorder(
		nil,                        // top
		container.NewVBox(          // bottom
			cv.loadingBox,
			cv.inputArea,
		),
		nil,                        // left
		nil,                        // right
		cv.scroll,                  // center
	)

	return widget.NewSimpleRenderer(content)
}
