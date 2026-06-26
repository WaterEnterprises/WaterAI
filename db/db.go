package db

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global instances to match the Python singleton pattern (Sessions, Events)
var (
	DB       *gorm.DB
	Sessions = &SessionStore{}
	Events   = &EventStore{}
)

// EventType constants (mapped from core/event in the original)
const (
	EventTypeUserMessage = "user_message"
)

// ==========================================
// MODELS
// ==========================================

// Session represents an agent session.
type Session struct {
	ID           string    `gorm:"primaryKey;type:text;length:36"`
	WorkspaceDir string    `gorm:"uniqueIndex;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	DeviceID     *string   `gorm:"index"`
	Name         *string
	SandboxID    *string
	Events       []Event `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

// Event represents a realtime event.
type Event struct {
	ID           string          `gorm:"primaryKey;type:text;length:36"`
	SessionID    string          `gorm:"index;not null;type:text;length:36"`
	Timestamp    time.Time       `gorm:"index;autoCreateTime"`
	EventType    string          `gorm:"not null"`
	EventPayload json.RawMessage `gorm:"type:json;not null"`

	// Associations
	Session Session `gorm:"foreignKey:SessionID"`
}

// BeforeCreate hooks to generate UUIDs automatically
func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return
}

// ==========================================
// INITIALIZATION
// ==========================================

// InitDB initializes the SQLite connection and runs auto-migrations.
// Pass the database path (e.g., "water-ai/water_ai.db").
func InitDB(databaseUrl string) error {
	var err error
	
	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}

	// Connect to SQLite
	// check_same_thread=False is handled automatically by GORM's connection pooling
	DB, err = gorm.Open(sqlite.Open(databaseUrl), config)
	if err != nil {
		return err
	}

	// Run Migrations (equivalent to Alembic upgrade head)
	err = DB.AutoMigrate(&Session{}, &Event{})
	if err != nil {
		log.Printf("Error running migrations: %v", err)
		return err
	}

	return nil
}

// ==========================================
// SESSIONS OPERATIONS
// ==========================================

type SessionStore struct{}

// CreateSession creates a new session.
func (s *SessionStore) CreateSession(
	sessionID uuid.UUID,
	workspacePath string,
	deviceID *string,
	sandboxID *string,
) (uuid.UUID, string, error) {
	
	sess := Session{
		ID:           sessionID.String(),
		WorkspaceDir: workspacePath,
		DeviceID:     deviceID,
		SandboxID:    sandboxID,
	}

	result := DB.Create(&sess)
	if result.Error != nil {
		return uuid.Nil, "", result.Error
	}

	return sessionID, workspacePath, nil
}

// GetSessionByWorkspace gets a session by its workspace directory.
func (s *SessionStore) GetSessionByWorkspace(workspaceDir string) (*Session, error) {
	var sess Session
	result := DB.Where("workspace_dir = ?", workspaceDir).First(&sess)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sess, result.Error
}

// GetSessionByID gets a session by its UUID.
func (s *SessionStore) GetSessionByID(sessionID uuid.UUID) (*Session, error) {
	var sess Session
	result := DB.Where("id = ?", sessionID.String()).First(&sess)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sess, result.Error
}

// GetSessionByDeviceID gets a session by its device ID.
func (s *SessionStore) GetSessionByDeviceID(deviceID string) (*Session, error) {
	var sess Session
	result := DB.Where("device_id = ?", deviceID).First(&sess)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sess, result.Error
}

// UpdateSessionName updates the name of a session.
func (s *SessionStore) UpdateSessionName(sessionID uuid.UUID, name string) error {
	return DB.Model(&Session{}).Where("id = ?", sessionID.String()).Update("name", name).Error
}

// GetSandboxIDBySessionID gets the sandbox_id of a session.
func (s *SessionStore) GetSandboxIDBySessionID(sessionID uuid.UUID) (*string, error) {
	var sess Session
	result := DB.Select("sandbox_id").Where("id = ?", sessionID.String()).First(&sess)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return sess.SandboxID, result.Error
}

// UpdateSessionSandboxID updates the sandbox_id of a session.
func (s *SessionStore) UpdateSessionSandboxID(sessionID uuid.UUID, sandboxID string) error {
	return DB.Model(&Session{}).Where("id = ?", sessionID.String()).Update("sandbox_id", sandboxID).Error
}

// GetSessionsByDeviceID gets all sessions for a specific device ID, sorted by creation time descending.
func (s *SessionStore) GetSessionsByDeviceID(deviceID string) ([]Session, error) {
	var sessions []Session
	err := DB.Where("device_id = ?", deviceID).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

// ==========================================
// EVENTS OPERATIONS
// ==========================================

type EventStore struct{}

// SaveEvent saves an event to the database.
// eventPayload should be a struct or map that can be marshaled to JSON.
func (e *EventStore) SaveEvent(sessionID uuid.UUID, eventType string, eventPayload interface{}) (uuid.UUID, error) {
	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return uuid.Nil, err
	}

	evt := Event{
		SessionID:    sessionID.String(),
		EventType:    eventType,
		EventPayload: payloadBytes,
	}

	if err := DB.Create(&evt).Error; err != nil {
		return uuid.Nil, err
	}

	return uuid.MustParse(evt.ID), nil
}

// GetSessionEvents gets all events for a session.
func (e *EventStore) GetSessionEvents(sessionID uuid.UUID) ([]Event, error) {
	var events []Event
	err := DB.Where("session_id = ?", sessionID.String()).Find(&events).Error
	return events, err
}

// DeleteSessionEvents deletes all events for a session.
func (e *EventStore) DeleteSessionEvents(sessionID uuid.UUID) error {
	return DB.Where("session_id = ?", sessionID.String()).Delete(&Event{}).Error
}

// DeleteEventsFromLastToUserMessage deletes events from the most recent event backwards 
// to the last user message (inclusive).
func (e *EventStore) DeleteEventsFromLastToUserMessage(sessionID uuid.UUID) error {
	var lastUserEvent Event
	
	// Find the last user message event
	err := DB.Where("session_id = ? AND event_type = ?", sessionID.String(), EventTypeUserMessage).
		Order("timestamp DESC").
		First(&lastUserEvent).Error

	if err == nil {
		// Found a user message, delete everything after and including it
		return DB.Where("session_id = ? AND timestamp >= ?", sessionID.String(), lastUserEvent.Timestamp).
			Delete(&Event{}).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// No user message found, delete all events for this session (matching Python logic)
		return e.DeleteSessionEvents(sessionID)
	}

	return err
}

// GetSessionEventsWithDetails gets all events for a session, sorted by timestamp ascending.
// Returns a custom map structure to match the Python API return shape.
func (e *EventStore) GetSessionEventsWithDetails(sessionID string) ([]map[string]interface{}, error) {
	var events []Event
	// Preload Session to get WorkspaceDir
	err := DB.Preload("Session").
		Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, evt := range events {
		// Parse the JSON payload back to object for the return
		var payload interface{}
		_ = json.Unmarshal(evt.EventPayload, &payload)

		data := map[string]interface{}{
			"id":            evt.ID,
			"session_id":    evt.SessionID,
			"timestamp":     evt.Timestamp.Format(time.RFC3339),
			"event_type":    evt.EventType,
			"event_payload": payload,
			"workspace_dir": evt.Session.WorkspaceDir,
		}
		results = append(results, data)
	}

	return results, nil
}