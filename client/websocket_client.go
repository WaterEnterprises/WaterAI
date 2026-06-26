package client

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient handles WebSocket communication with the backend
type WebSocketClient struct {
	conn            *websocket.Conn
	url             string
	state           *AppState
	mu              sync.Mutex
	onEvent         func(eventType string, content interface{})
	onStateChange   func()
	onConnected     func()
	onDisconnected  func()
	stopChan        chan struct{}
	reconnect       bool
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(serverURL string, state *AppState) *WebSocketClient {
	return &WebSocketClient{
		url:       serverURL,
		state:     state,
		reconnect: true,
		stopChan:  make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection to the server
func (c *WebSocketClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connectInternal()
}

// connectInternal establishes a WebSocket connection without locking
// This is used internally for reconnection attempts
func (c *WebSocketClient) connectInternal() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	// Add session_uuid query parameter if needed
	q := u.Query()
	q.Set("session_uuid", "")
	u.RawQuery = q.Encode()

	header := http.Header{}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return err
	}

	c.conn = conn
	c.state.IsConnected = true

	// Start the message handler
	go c.handleMessages()

	// Start the ping loop
	go c.pingLoop()

	if c.onConnected != nil {
		c.onConnected()
	}

	if c.onStateChange != nil {
		c.onStateChange()
	}

	return nil
}

// Disconnect closes the WebSocket connection
func (c *WebSocketClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.reconnect = false
	close(c.stopChan)

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.state.IsConnected = false

	if c.onDisconnected != nil {
		c.onDisconnected()
	}

	if c.onStateChange != nil {
		c.onStateChange()
	}
}

// SendMessage sends a message to the server
func (c *WebSocketClient) SendMessage(msgType string, content interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	msg := WebSocketMessage{
		Type:    msgType,
		Content: mustMarshal(content),
	}

	return c.conn.WriteJSON(msg)
}

// handleMessages reads and processes incoming messages
func (c *WebSocketClient) handleMessages() {
	defer func() {
		c.state.IsConnected = false
		if c.onDisconnected != nil {
			c.onDisconnected()
		}
		if c.onStateChange != nil {
			c.onStateChange()
		}

		// Attempt reconnection if enabled
		if c.reconnect {
			go c.reconnectLoop()
		}
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.conn == nil {
				return
			}

			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			c.processMessage(message)
		}
	}
}

// processMessage parses and handles an incoming message
func (c *WebSocketClient) processMessage(data []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	switch msg.Type {
	case EventTypeConnectionEstablished:
		var event ConnectionEstablishedEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			c.state.WorkspacePath = event.WorkspacePath
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeAgentInitialized:
		var event AgentInitializedEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			c.state.IsAgentInitialized = true
			c.state.VSCodeURL = event.VSCodeURL
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeAgentResponse:
		var event AgentResponseEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			// Add or update the assistant message
			newMsg := NewMessage("assistant", event.Text)
			c.state.AddMessage(newMsg)
			if c.onEvent != nil {
				c.onEvent(newMsg.Role, event)
			}
		}

	case EventTypeProcessing:
		var event ProcessingEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			c.state.IsLoading = true
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeStreamComplete:
		c.state.IsLoading = false
		if c.onEvent != nil {
			c.onEvent(msg.Type, nil)
		}

	case EventTypeError:
		var event ErrorEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			log.Printf("Server error: %s", event.Message)
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeSystem:
		var event SystemEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeToolCall:
		var event ToolCallEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	case EventTypeToolResult:
		var event ToolResultEvent
		if err := json.Unmarshal(msg.Content, &event); err == nil {
			if c.onEvent != nil {
				c.onEvent(msg.Type, event)
			}
		}

	default:
		log.Printf("Unknown event type: %s", msg.Type)
	}

	if c.onStateChange != nil {
		c.onStateChange()
	}
}

// pingLoop sends periodic pings to keep the connection alive
func (c *WebSocketClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				c.conn.WriteJSON(WebSocketMessage{Type: "ping"})
			}
			c.mu.Unlock()
		}
	}
}

// reconnectLoop attempts to reconnect to the server
func (c *WebSocketClient) reconnectLoop() {
	// Create a new stopChan for reconnection attempts
	c.mu.Lock()
	c.stopChan = make(chan struct{})
	c.mu.Unlock()

	for i := 0; i < 5 && c.reconnect; i++ {
		time.Sleep(time.Duration(i+1) * time.Second)
		log.Printf("Attempting to reconnect (%d/5)...", i+1)
		
		if err := c.connectInternal(); err == nil {
			log.Println("Reconnected successfully")
			return
		}
	}
	log.Println("Failed to reconnect after 5 attempts")
}

// SetOnEvent sets the event callback
func (c *WebSocketClient) SetOnEvent(callback func(eventType string, content interface{})) {
	c.onEvent = callback
}

// SetOnStateChange sets the state change callback
func (c *WebSocketClient) SetOnStateChange(callback func()) {
	c.onStateChange = callback
}

// SetOnConnected sets the connected callback
func (c *WebSocketClient) SetOnConnected(callback func()) {
	c.onConnected = callback
}

// SetOnDisconnected sets the disconnected callback
func (c *WebSocketClient) SetOnDisconnected(callback func()) {
	c.onDisconnected = callback
}

// InitAgent sends the init_agent message
func (c *WebSocketClient) InitAgent(modelName string, toolArgs map[string]interface{}, thinkingTokens int) error {
	return c.SendMessage("init_agent", InitAgentContent{
		ModelName:      modelName,
		ToolArgs:       toolArgs,
		ThinkingTokens: thinkingTokens,
	})
}

// SendQuery sends a query message
func (c *WebSocketClient) SendQuery(text string, resume bool, files []string) error {
	return c.SendMessage("query", QueryContent{
		Text:   text,
		Resume: resume,
		Files:  files,
	})
}

// EditQuery sends an edit_query message
func (c *WebSocketClient) EditQuery(text string, files []string) error {
	return c.SendMessage("edit_query", EditQueryContent{
		Text:  text,
		Files: files,
	})
}

// CancelQuery sends a cancel message
func (c *WebSocketClient) CancelQuery() error {
	return c.SendMessage("cancel", map[string]interface{}{})
}

// Helper functions

var ErrNotConnected = &ConnectionError{Message: "not connected to server"}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
