package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"water-ai/llm"
	"water-ai/prompts"
)

// --- Configuration & Global State ---

type Config struct {
	WorkspaceRoot string
	Port          string
}

// GetPort returns the configured port or default
func (c Config) GetPort() string {
	if c.Port == "" {
		return "8080"
	}
	return c.Port
}

// GetWorkspaceRoot returns the configured workspace or default
func (c Config) GetWorkspaceRoot() string {
	if c.WorkspaceRoot == "" {
		return "./workspace_data"
	}
	return c.WorkspaceRoot
}

// Server holds the dependencies for the application
type Server struct {
	Config     Config
	Router     *gin.Engine
	WSManager  *ConnectionManager
	// Stub for DB/FileStore interfaces
	FileStore  interface{} 
}

// --- WebSocket Manager ---

type ConnectionManager struct {
	sessions map[*websocket.Conn]*ChatSession
	mu       sync.RWMutex
	config   Config
}

func NewConnectionManager(cfg Config) *ConnectionManager {
	return &ConnectionManager{
		sessions: make(map[*websocket.Conn]*ChatSession),
		config:   cfg,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// --- Chat Session Logic ---

type ChatSession struct {
	Conn        *websocket.Conn
	SessionUUID uuid.UUID
	Workspace   string
	Manager     *ConnectionManager
	LLMClient    llm.Client
	History      *llm.MessageHistory
	SystemPrompt string
	mu           sync.Mutex
}

func (s *ChatSession) SendEvent(eventType string, content interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.Conn == nil {
		return
	}

	msg := RealtimeEvent{
		Type:    eventType,
		Content: content,
	}
	if err := s.Conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending event: %v", err)
	}
}

func (s *ChatSession) StartLoop() {
	defer func() {
		s.Manager.Disconnect(s.Conn)
		s.Conn.Close()
	}()

	// Handshake
	s.SendEvent(EventTypeConnectionEstablished, gin.H{
		"message":        "Connected to Water AI Server",
		"workspace_path": s.Workspace,
	})

	for {
		_, messageData, err := s.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS Error: %v", err)
			}
			break
		}
		go s.HandleMessage(messageData)
	}
}

func (s *ChatSession) HandleMessage(data []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		s.SendEvent(EventTypeError, gin.H{"message": "Invalid JSON"})
		return
	}

	switch msg.Type {
	case "init_agent":
		var content InitAgentContent
		_ = json.Unmarshal(msg.Content, &content)
		s.handleInitAgent(content)
	case "query":
		var content QueryContent
		_ = json.Unmarshal(msg.Content, &content)
		s.handleQuery(content)
	case "ping":
		s.SendEvent(EventTypePong, gin.H{})
	case "workspace_info":
		s.SendEvent(EventTypeWorkspaceInfo, gin.H{"path": s.Workspace})
	case "cancel":
		s.SendEvent(EventTypeSystem, gin.H{"message": "Query cancelled"})
	// Add other handlers (edit_query, etc.) as needed
	default:
		s.SendEvent(EventTypeError, gin.H{"message": "Unknown message type"})
	}
}

func (s *ChatSession) handleInitAgent(content InitAgentContent) {
	// Create workspace if needed
	os.MkdirAll(s.Workspace, 0755)

	// Determine API type from model name
	modelName := content.ModelName
	if modelName == "" {
		modelName = "gpt-4-turbo"
	}

	apiType := llm.APITypeOpenAI
	if strings.Contains(modelName, "claude") || strings.Contains(modelName, "anthropic") {
		apiType = llm.APITypeAnthropic
	} else if strings.Contains(modelName, "gemini") {
		apiType = llm.APITypeGemini
	}

	// Read API key from environment
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		switch apiType {
		case llm.APITypeOpenAI:
			apiKey = os.Getenv("OPENAI_API_KEY")
		case llm.APITypeAnthropic:
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case llm.APITypeGemini:
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
	}

	cfg := llm.LLMConfig{
		APIType:        apiType,
		Model:          modelName,
		APIKey:         apiKey,
		MaxRetries:     3,
		ThinkingTokens: content.ThinkingTokens,
	}

	client, err := llm.GetClient(cfg)
	if err != nil {
		s.SendEvent(EventTypeError, gin.H{"message": fmt.Sprintf("Failed to initialize LLM client: %v", err)})
		return
	}

	s.LLMClient = client
	s.History = llm.NewMessageHistory()
	s.SystemPrompt = prompts.GetSystemPrompt(prompts.WorkspaceModeLocal, false)

	s.SendEvent(EventTypeAgentInitialized, gin.H{
		"message": "Agent initialized",
	})
}

func (s *ChatSession) handleQuery(content QueryContent) {
	if strings.HasPrefix(content.Text, "/") {
		s.handleSlashCommand(content.Text)
		return
	}

	if s.LLMClient == nil {
		s.SendEvent(EventTypeError, gin.H{"message": "Agent not initialized. Send init_agent first."})
		return
	}

	s.SendEvent(EventTypeProcessing, gin.H{"message": "Processing request..."})

	// Add user message to history
	s.History.AddUserPrompt(content.Text, nil)

	// Call the real LLM client
	resp, err := s.LLMClient.Generate(
		s.History.GetMessages(),
		4096,
		s.SystemPrompt,
		0.0,
		nil,  // tools
		nil,  // toolChoice
		nil,  // thinkingTokens
	)
	if err != nil {
		log.Printf("LLM Generate error: %v", err)
		s.SendEvent(EventTypeError, gin.H{"message": fmt.Sprintf("LLM error: %v", err)})
		s.SendEvent(EventTypeStreamComplete, gin.H{})
		return
	}

	// Add assistant response to history
	s.History.AddAssistantTurn(resp.Content)

	// Extract text from response blocks and send to client
	var responseText string
	for _, block := range resp.Content {
		if block.Type == llm.ContentTypeText && block.Text != "" {
			responseText += block.Text
		}
	}

	if responseText != "" {
		s.SendEvent(EventTypeAgentResponse, gin.H{"text": responseText})
	}
	s.SendEvent(EventTypeStreamComplete, gin.H{})
}

func (s *ChatSession) handleSlashCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/help":
		s.SendEvent(EventTypeSystem, gin.H{"message": "Available commands: /help, /compact"})
		s.SendEvent(EventTypeStreamComplete, gin.H{})
	case "/compact":
		s.SendEvent(EventTypeProcessing, gin.H{"message": "Compacting memory..."})
		time.Sleep(500 * time.Millisecond)
		s.SendEvent(EventTypeSystem, gin.H{"message": "Memory compacted."})
		s.SendEvent(EventTypeStreamComplete, gin.H{})
	default:
		s.SendEvent(EventTypeError, gin.H{"message": "Unknown command"})
	}
}

func (m *ConnectionManager) Connect(conn *websocket.Conn, sessionUUIDStr string) *ChatSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid, err := uuid.Parse(sessionUUIDStr)
	if err != nil {
		uid = uuid.New()
	}

	// Resolve workspace path
	workspacePath := filepath.Join(m.config.WorkspaceRoot, uid.String())

	session := &ChatSession{
		Conn:        conn,
		SessionUUID: uid,
		Workspace:   workspacePath,
		Manager:     m,
	}

	m.sessions[conn] = session
	log.Printf("New Session: %s", uid.String())
	return session
}

func (m *ConnectionManager) Disconnect(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[conn]; ok {
		delete(m.sessions, conn)
	}
}

// --- HTTP Handlers ---

// UploadHandler handles file uploads (base64 or text)
func (s *Server) UploadHandler(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	// Path logic
	workspace := filepath.Join(s.Config.WorkspaceRoot, req.SessionID)
	uploadDir := filepath.Join(workspace, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Handle path normalization
	baseName := filepath.Base(req.File.Path)
	fullPath := filepath.Join(uploadDir, baseName)

	// Collision handling
	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)
	counter := 1
	for {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			break
		}
		fullPath = filepath.Join(uploadDir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		counter++
	}

	// Write content
	var contentBytes []byte
	var err error

	if strings.HasPrefix(req.File.Content, "data:") {
		// Handle Base64
		parts := strings.SplitN(req.File.Content, ",", 2)
		if len(parts) == 2 {
			contentBytes, err = base64.StdEncoding.DecodeString(parts[1])
		} else {
			err = fmt.Errorf("invalid data URI")
		}
	} else {
		// Handle Text
		contentBytes = []byte(req.File.Content)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode content"})
		return
	}

	if err := os.WriteFile(fullPath, contentBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	relPath, _ := filepath.Rel(workspace, fullPath)
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file": gin.H{
			"path":       "/" + relPath,
			"saved_path": fullPath,
		},
	})
}

// GetSessionsHandler (Mock Implementation)
func (s *Server) GetSessionsHandler(c *gin.Context) {
	deviceID := c.Param("device_id")
	// Note: In a real implementation, you would query SQLite/Postgres here.
	// Returning a mock response for demonstration.
	c.JSON(http.StatusOK, SessionResponse{
		Sessions: []SessionInfo{
			{
				ID:           uuid.New().String(),
				WorkspaceDir: s.Config.WorkspaceRoot,
				CreatedAt:    time.Now().Format(time.RFC3339),
				DeviceID:     deviceID,
				Name:         "Demo Session",
			},
		},
	})
}

// GetEventsHandler (Mock Implementation)
func (s *Server) GetEventsHandler(c *gin.Context) {
	sessionID := c.Param("session_id")
	c.JSON(http.StatusOK, EventResponse{
		Events: []EventInfo{
			{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				Timestamp: time.Now().Format(time.RFC3339),
				EventType: "system",
				EventPayload: map[string]interface{}{
					"message": "Session started",
				},
			},
		},
	})
}

// SessionsHandler handles both /sessions/:device_id and /sessions/:session_id/events
func (s *Server) SessionsHandler(c *gin.Context) {
	path := c.Param("path")
	// Remove leading slash if present
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if strings.HasPrefix(path, "events") {
		// Handle /sessions/:session_id/events
		sessionID := strings.TrimPrefix(path, "events")
		sessionID = strings.TrimPrefix(sessionID, "/")
		c.JSON(http.StatusOK, EventResponse{
			Events: []EventInfo{
				{
					ID:        uuid.New().String(),
					SessionID: sessionID,
					Timestamp: time.Now().Format(time.RFC3339),
					EventType: "system",
					EventPayload: map[string]interface{}{
						"message": "Session started",
					},
				},
			},
		})
	} else {
		// Handle /sessions/:device_id
		deviceID := path
		c.JSON(http.StatusOK, SessionResponse{
			Sessions: []SessionInfo{
				{
					ID:           uuid.New().String(),
					WorkspaceDir: s.Config.WorkspaceRoot,
					CreatedAt:    time.Now().Format(time.RFC3339),
					DeviceID:     deviceID,
					Name:         "Demo Session",
				},
			},
		})
	}
}

// GetSettingsHandler
func (s *Server) GetSettingsHandler(c *gin.Context) {
	// Mock loading settings
	settings := GETSettingsModel{
		Settings: Settings{
			LLMConfigs: map[string]LLMConfig{
				"gpt-4": {Model: "gpt-4"},
			},
		},
		LLMAPIKeySet:    true,
		SearchAPIKeySet: false,
	}
	c.JSON(http.StatusOK, settings)
}

// PostSettingsHandler
func (s *Server) PostSettingsHandler(c *gin.Context) {
	var settings Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Logic to save settings to file/DB goes here
	c.JSON(http.StatusOK, gin.H{"message": "Settings stored"})
}

// --- Factory ---

func CreateServer(config Config) *Server {
	router := gin.Default()
	
	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	manager := NewConnectionManager(config)

	srv := &Server{
		Config:    config,
		Router:    router,
		WSManager: manager,
	}

	// API Routes
	api := router.Group("/api")
	{
		api.POST("/upload", srv.UploadHandler)
		api.GET("/sessions/*path", srv.SessionsHandler)
		api.GET("/settings", srv.GetSettingsHandler)
		api.POST("/settings", srv.PostSettingsHandler)
	}

	// Workspace Static Files
	// Create root if it doesn't exist
	os.MkdirAll(config.WorkspaceRoot, 0755)
	router.StaticFS("/workspace", gin.Dir(config.WorkspaceRoot, true))

	// WebSocket Endpoint
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to upgrade WS:", err)
			return
		}
		
		sessionID := c.Query("session_uuid")
		session := manager.Connect(conn, sessionID)
		go session.StartLoop()
	})

	// Frontend Static Files (SPA fallback for client-side routing)
	// Note: The Fyne GUI connects directly via WebSocket, no web frontend needed
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Water AI Backend API - Use the Fyne GUI client to connect")
	})

	return srv
}

// getContentType returns the appropriate Content-Type for static files
func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".ico":
		return "image/x-icon"
	default:
		return ""
	}
}