package db

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&Session{}, &Event{})
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set global DB for tests
	DB = db

	return db
}

func teardownTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func TestSessionBeforeCreate(t *testing.T) {
	sess := &Session{
		WorkspaceDir: "/test/workspace",
	}

	err := sess.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate() error = %v", err)
	}

	if sess.ID == "" {
		t.Error("BeforeCreate() should set ID")
	}

	// Verify it's a valid UUID
	_, err = uuid.Parse(sess.ID)
	if err != nil {
		t.Errorf("BeforeCreate() should set valid UUID; got = %s, error = %v", sess.ID, err)
	}
}

func TestSessionBeforeCreateWithExistingID(t *testing.T) {
	existingID := uuid.New().String()
	sess := &Session{
		ID:           existingID,
		WorkspaceDir: "/test/workspace",
	}

	err := sess.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate() error = %v", err)
	}

	if sess.ID != existingID {
		t.Errorf("BeforeCreate() should not change existing ID; got = %s, want = %s", sess.ID, existingID)
	}
}

func TestEventBeforeCreate(t *testing.T) {
	evt := &Event{
		SessionID: uuid.New().String(),
		EventType: "test_event",
	}

	err := evt.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate() error = %v", err)
	}

	if evt.ID == "" {
		t.Error("BeforeCreate() should set ID")
	}

	// Verify it's a valid UUID
	_, err = uuid.Parse(evt.ID)
	if err != nil {
		t.Errorf("BeforeCreate() should set valid UUID; got = %s, error = %v", evt.ID, err)
	}
}

func TestCreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"
	sandboxID := "sandbox-456"

	resultID, resultPath, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, &sandboxID)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if resultID != sessionID {
		t.Errorf("CreateSession() ID = %s; want %s", resultID, sessionID)
	}

	if resultPath != workspacePath {
		t.Errorf("CreateSession() path = %s; want %s", resultPath, workspacePath)
	}
}

func TestGetSessionByWorkspace(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	workspacePath := "/test/workspace/" + uuid.New().String()
	sessionID := uuid.New()
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	sess, err := Sessions.GetSessionByWorkspace(workspacePath)
	if err != nil {
		t.Fatalf("GetSessionByWorkspace() error = %v", err)
	}

	if sess == nil {
		t.Fatal("GetSessionByWorkspace() returned nil")
	}

	if sess.WorkspaceDir != workspacePath {
		t.Errorf("GetSessionByWorkspace() WorkspaceDir = %s; want %s", sess.WorkspaceDir, workspacePath)
	}
}

func TestGetSessionByWorkspaceNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sess, err := Sessions.GetSessionByWorkspace("/non/existent/path")
	if err != nil {
		t.Fatalf("GetSessionByWorkspace() error = %v", err)
	}

	if sess != nil {
		t.Error("GetSessionByWorkspace() should return nil for non-existent workspace")
	}
}

func TestGetSessionByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace/" + uuid.New().String()
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	sess, err := Sessions.GetSessionByID(sessionID)
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}

	if sess == nil {
		t.Fatal("GetSessionByID() returned nil")
	}

	if sess.ID != sessionID.String() {
		t.Errorf("GetSessionByID() ID = %s; want %s", sess.ID, sessionID.String())
	}
}

func TestGetSessionByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sess, err := Sessions.GetSessionByID(uuid.New())
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}

	if sess != nil {
		t.Error("GetSessionByID() should return nil for non-existent ID")
	}
}

func TestGetSessionByDeviceID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	deviceID := "device-" + uuid.New().String()
	sessionID := uuid.New()

	_, _, err := Sessions.CreateSession(sessionID, "/test/workspace", &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	sess, err := Sessions.GetSessionByDeviceID(deviceID)
	if err != nil {
		t.Fatalf("GetSessionByDeviceID() error = %v", err)
	}

	if sess == nil {
		t.Fatal("GetSessionByDeviceID() returned nil")
	}

	if sess.DeviceID == nil || *sess.DeviceID != deviceID {
		t.Errorf("GetSessionByDeviceID() DeviceID = %v; want %s", sess.DeviceID, deviceID)
	}
}

func TestUpdateSessionName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	newName := "Updated Session Name"
	err = Sessions.UpdateSessionName(sessionID, newName)
	if err != nil {
		t.Fatalf("UpdateSessionName() error = %v", err)
	}

	sess, err := Sessions.GetSessionByID(sessionID)
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}

	if sess.Name == nil || *sess.Name != newName {
		t.Errorf("GetSessionByID() Name = %v; want %s", sess.Name, newName)
	}
}

func TestGetSandboxIDBySessionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"
	sandboxID := "sandbox-789"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, &sandboxID)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	result, err := Sessions.GetSandboxIDBySessionID(sessionID)
	if err != nil {
		t.Fatalf("GetSandboxIDBySessionID() error = %v", err)
	}

	if result == nil || *result != sandboxID {
		t.Errorf("GetSandboxIDBySessionID() = %v; want %s", result, sandboxID)
	}
}

func TestUpdateSessionSandboxID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	newSandboxID := "new-sandbox-id"
	err = Sessions.UpdateSessionSandboxID(sessionID, newSandboxID)
	if err != nil {
		t.Fatalf("UpdateSessionSandboxID() error = %v", err)
	}

	result, err := Sessions.GetSandboxIDBySessionID(sessionID)
	if err != nil {
		t.Fatalf("GetSandboxIDBySessionID() error = %v", err)
	}

	if result == nil || *result != newSandboxID {
		t.Errorf("GetSandboxIDBySessionID() = %v; want %s", result, newSandboxID)
	}
}

func TestGetSessionsByDeviceID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	deviceID := "device-" + uuid.New().String()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		sessionID := uuid.New()
		_, _, err := Sessions.CreateSession(sessionID, "/test/workspace/"+uuid.New().String(), &deviceID, nil)
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}
	}

	sessions, err := Sessions.GetSessionsByDeviceID(deviceID)
	if err != nil {
		t.Fatalf("GetSessionsByDeviceID() error = %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("GetSessionsByDeviceID() returned %d sessions; want 3", len(sessions))
	}
}

func TestSaveEvent(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	eventType := "test_event"
	payload := map[string]interface{}{
		"message": "test message",
		"count":   42,
	}

	eventID, err := Events.SaveEvent(sessionID, eventType, payload)
	if err != nil {
		t.Fatalf("SaveEvent() error = %v", err)
	}

	if eventID == uuid.Nil {
		t.Error("SaveEvent() should return valid UUID")
	}
}

func TestGetSessionEvents(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Save multiple events
	for i := 0; i < 3; i++ {
		_, err := Events.SaveEvent(sessionID, "event_type_"+string(rune('A'+i)), map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("SaveEvent() error = %v", err)
		}
	}

	events, err := Events.GetSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionEvents() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("GetSessionEvents() returned %d events; want 3", len(events))
	}
}

func TestDeleteSessionEvents(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Save events
	for i := 0; i < 3; i++ {
		_, err := Events.SaveEvent(sessionID, "event_type_"+string(rune('A'+i)), map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("SaveEvent() error = %v", err)
		}
	}

	err = Events.DeleteSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("DeleteSessionEvents() error = %v", err)
	}

	events, err := Events.GetSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionEvents() error = %v", err)
	}

	if len(events) != 0 {
		t.Errorf("GetSessionEvents() returned %d events after delete; want 0", len(events))
	}
}

func TestDeleteEventsFromLastToUserMessage(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Save events: some before user message, then user message, then after
	eventTypes := []string{"event_1", "event_2", EventTypeUserMessage, "event_4", "event_5"}
	for i, eventType := range eventTypes {
		_, err := Events.SaveEvent(sessionID, eventType, map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("SaveEvent() error = %v", err)
		}
	}

	err = Events.DeleteEventsFromLastToUserMessage(sessionID)
	if err != nil {
		t.Fatalf("DeleteEventsFromLastToUserMessage() error = %v", err)
	}

	events, err := Events.GetSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionEvents() error = %v", err)
	}

	// Should only have events before user message
	if len(events) != 2 {
		t.Errorf("GetSessionEvents() returned %d events; want 2", len(events))
	}
}

func TestDeleteEventsFromLastToUserMessageNoUserMessage(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Save events without user message
	for i := 0; i < 3; i++ {
		_, err := Events.SaveEvent(sessionID, "event_type_"+string(rune('A'+i)), map[string]interface{}{"index": i})
		if err != nil {
			t.Fatalf("SaveEvent() error = %v", err)
		}
	}

	err = Events.DeleteEventsFromLastToUserMessage(sessionID)
	if err != nil {
		t.Fatalf("DeleteEventsFromLastToUserMessage() error = %v", err)
	}

	events, err := Events.GetSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionEvents() error = %v", err)
	}

	// Should delete all events when no user message exists
	if len(events) != 0 {
		t.Errorf("GetSessionEvents() returned %d events; want 0", len(events))
	}
}

func TestGetSessionEventsWithDetails(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace/" + uuid.New().String()
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	payload := map[string]interface{}{
		"message": "test message",
		"count":   42,
	}
	_, err = Events.SaveEvent(sessionID, "test_event", payload)
	if err != nil {
		t.Fatalf("SaveEvent() error = %v", err)
	}

	results, err := Events.GetSessionEventsWithDetails(sessionID.String())
	if err != nil {
		t.Fatalf("GetSessionEventsWithDetails() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("GetSessionEventsWithDetails() returned %d results; want 1", len(results))
	}

	if results[0]["event_type"] != "test_event" {
		t.Errorf("GetSessionEventsWithDetails() event_type = %v; want test_event", results[0]["event_type"])
	}

	if results[0]["workspace_dir"] != workspacePath {
		t.Errorf("GetSessionEventsWithDetails() workspace_dir = %v; want %s", results[0]["workspace_dir"], workspacePath)
	}
}

func TestEventJSONPayload(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	sessionID := uuid.New()
	workspacePath := "/test/workspace"
	deviceID := "device-123"

	_, _, err := Sessions.CreateSession(sessionID, workspacePath, &deviceID, nil)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	payload := map[string]interface{}{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
	}

	_, err = Events.SaveEvent(sessionID, "json_test", payload)
	if err != nil {
		t.Fatalf("SaveEvent() error = %v", err)
	}

	events, err := Events.GetSessionEvents(sessionID)
	if err != nil {
		t.Fatalf("GetSessionEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("GetSessionEvents() returned %d events; want 1", len(events))
	}

	// Verify JSON payload can be unmarshaled
	var decodedPayload map[string]interface{}
	err = json.Unmarshal(events[0].EventPayload, &decodedPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal event payload: %v", err)
	}

	if decodedPayload["string_val"] != "hello" {
		t.Errorf("Payload string_val = %v; want hello", decodedPayload["string_val"])
	}
}
