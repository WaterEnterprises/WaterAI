package chat

import (
	"fmt"
	"strings"
	"water-ai/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MessageList displays a list of chat messages
type MessageList struct {
	widget.BaseWidget

	state *client.AppState
	box   *fyne.Container
}

// NewMessageList creates a new message list
func NewMessageList(state *client.AppState) *MessageList {
	ml := &MessageList{
		state: state,
		box:   container.NewVBox(),
	}
	ml.ExtendBaseWidget(ml)
	return ml
}

// Refresh updates the message list
func (ml *MessageList) Refresh() {
	// Clear existing items
	ml.box.Objects = nil

	// Add all visible messages
	for _, msg := range ml.state.Messages {
		if msg.IsHidden {
			continue
		}
		ml.box.Add(NewMessageItem(msg))
	}

	ml.BaseWidget.Refresh()
}

// CreateRenderer creates the widget renderer
func (ml *MessageList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ml.box)
}

// MinSize returns the minimum size
func (ml *MessageList) MinSize() fyne.Size {
	return fyne.NewSize(400, 500)
}

// MessageItem represents a single message in the chat
type MessageItem struct {
	widget.BaseWidget

	message client.Message
}

// NewMessageItem creates a new message item
func NewMessageItem(msg client.Message) *MessageItem {
	mi := &MessageItem{
		message: msg,
	}
	mi.ExtendBaseWidget(mi)
	return mi
}

// CreateRenderer creates the widget renderer
func (mi *MessageItem) CreateRenderer() fyne.WidgetRenderer {
	// Determine style based on role
	var icon fyne.Resource
	var roleLabel string

	switch mi.message.Role {
	case "user":
		icon = theme.AccountIcon()
		roleLabel = "You"
	case "assistant":
		icon = theme.ComputerIcon()
		roleLabel = "Water AI"
	default:
		icon = theme.InfoIcon()
		roleLabel = "System"
	}

	// Create header with icon and role
	header := container.NewHBox(
		widget.NewIcon(icon),
		widget.NewLabelWithStyle(roleLabel, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create content with basic markdown support
	content := NewMarkdownLabel(mi.message.Content)

	// Create message container
	messageContainer := container.NewVBox(
		header,
		widget.NewSeparator(),
		content,
	)

	// Create card-like appearance
	card := widget.NewCard("", "", messageContainer)

	return widget.NewSimpleRenderer(card)
}

// MinSize returns the minimum size for the message item
func (mi *MessageItem) MinSize() fyne.Size {
	return fyne.NewSize(350, 80)
}

// String returns a string representation
func (mi *MessageItem) String() string {
	content := mi.message.Content
	if len(content) > 50 {
		content = content[:50]
	}
	return fmt.Sprintf("Message[%s]: %s", mi.message.Role, content)
}

// MarkdownLabel is a label that supports basic markdown formatting
type MarkdownLabel struct {
	widget.BaseWidget

	text    string
	content *fyne.Container
}

// NewMarkdownLabel creates a new markdown label
func NewMarkdownLabel(text string) *MarkdownLabel {
	ml := &MarkdownLabel{
		text:    text,
		content: container.NewVBox(),
	}
	ml.ExtendBaseWidget(ml)
	ml.parseMarkdown()
	return ml
}

// parseMarkdown parses basic markdown and creates widgets
func (ml *MarkdownLabel) parseMarkdown() {
	ml.content.Objects = nil

	// Split into lines
	lines := strings.Split(ml.text, "\n")

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			// Skip code block markers for now
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "### ") {
			text := strings.TrimPrefix(line, "### ")
			ml.content.Add(widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			text := strings.TrimPrefix(line, "## ")
			ml.content.Add(widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			continue
		}
		if strings.HasPrefix(line, "# ") {
			text := strings.TrimPrefix(line, "# ")
			ml.content.Add(widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			continue
		}

		// Handle bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			text := "  • " + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			ml.content.Add(widget.NewLabel(text))
			continue
		}

		// Handle numbered lists
		if len(line) > 2 && line[1] == '.' && line[0] >= '0' && line[0] <= '9' {
			ml.content.Add(widget.NewLabel("  " + line))
			continue
		}

		// Handle inline code
		if strings.Contains(line, "`") {
			line = ml.processInlineCode(line)
		}

		// Handle bold text
		if strings.Contains(line, "**") {
			line = ml.processBold(line)
		}

		// Regular text
		if line == "" {
			ml.content.Add(widget.NewLabel(" "))
		} else {
			label := widget.NewLabel(line)
			label.Wrapping = fyne.TextWrapWord
			ml.content.Add(label)
		}
	}
}

// processInlineCode processes inline code markers
func (ml *MarkdownLabel) processInlineCode(text string) string {
	// Simple replacement - just remove backticks for now
	return strings.ReplaceAll(text, "`", "'")
}

// processBold processes bold text markers
func (ml *MarkdownLabel) processBold(text string) string {
	// Simple replacement - just remove ** for now
	return strings.ReplaceAll(text, "**", "")
}

// SetText sets the label text
func (ml *MarkdownLabel) SetText(text string) {
	ml.text = text
	ml.parseMarkdown()
	ml.Refresh()
}

// CreateRenderer creates the widget renderer
func (ml *MarkdownLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ml.content)
}
