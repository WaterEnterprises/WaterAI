# Unit Test Plan for Water AI

## Overview
This document outlines the comprehensive unit test strategy for the Water AI codebase. The goal is to achieve complete test coverage across all Go backend packages without modifying existing source code.

## Project Structure Analysis

### Go Backend Packages
| Package | Path | Priority | Complexity |
|---------|------|----------|------------|
| core | `core/` | High | Low |
| config | `core/config/` | High | Medium |
| storage | `core/storage/` | High | Medium |
| db | `db/` | High | Medium |
| llm | `llm/` | High | Medium |
| context_manager | `llm/context_manager/` | High | Medium |
| agents | `agents/` | Medium | Medium |
| sandbox | `sandbox/` | Medium | Low |
| server | `server/` | High | Medium |
| tools | `tools/` | High | Medium |
| utils | `utils/` | Medium | Low |

## Test Structure

### Testing Framework
- **Go Testing**: `testing` package (built-in)
- **Mocking**: `go.uber.org/mock` (available in go.mod)
- **Assertions**: Standard Go testing patterns with `require`/`assert` style helpers

### Test File Naming Convention
- `*_test.go` for each source file
- Test files placed alongside source files
- `testdata/` subdirectories for fixtures when needed

## Detailed Test Plans by Package

### 1. Core Package (`core/`)

#### `core/even.go` Tests
**File**: `core/even_test.go`

```go
// Tests for EventType constants
func TestEventTypeConstants(t *testing.T)
func TestRealtimeEventCreation(t *testing.T)
func TestRealtimeEventContent(t *testing.T)
```

#### `core/logger.go` Tests
**File**: `core/logger_test.go`

```go
// Logger initialization tests
func TestLoggerInitDefault(t *testing.T)
func TestLoggerInitDebug(t *testing.T)
func TestLoggerInitError(t *testing.T)
func TestLoggerLevelFromEnv(t *testing.T)
```

---

### 2. Config Package (`core/config/`)

**File**: `core/config/config_test.go`

```go
// SecretString tests
func TestSecretStringString(t *testing.T)
func TestSecretStringReveal(t *testing.T)
func TestSecretStringMarshalJSON(t *testing.T)

// Config tests
func TestNewWaterAgentConfig(t *testing.T)
func TestWaterAgentConfigWorkspaceRoot(t *testing.T)
func TestWaterAgentConfigHostWorkspace(t *testing.T)
func TestWaterAgentConfigLogsPath(t *testing.T)
func TestWaterAgentConfigCodeServerPort(t *testing.T)
func TestNewLLMConfig(t *testing.T)
func TestNewSandboxConfig(t *testing.T)
func TestAudioConfigUpdate(t *testing.T)
func TestMediaConfigUpdate(t *testing.T)
func TestSearchConfigUpdate(t *testing.T)
func TestThirdPartyIntegrationConfigUpdate(t *testing.T)

// Utility tests
func TestExpandPath(t *testing.T)
func TestGetEnv(t *testing.T)
func TestGetEnvInt(t *testing.T)
func TestGetEnvBool(t *testing.T)
func TestStringPtr(t *testing.T)
func TestSecretPtr(t *testing.T)
```

---

### 3. Storage Package (`core/storage/`)

**File**: `core/storage/storage_test.go`

```go
// FileStore interface tests
func TestNewFileStoreLocal(t *testing.T)
func TestNewFileStoreMemory(t *testing.T)
func TestNewFileStoreLocalEmptyPath(t *testing.T)

// LocalFileStore tests
func TestLocalFileStoreWrite(t *testing.T)
func TestLocalFileStoreRead(t *testing.T)
func TestLocalFileStoreList(t *testing.T)
func TestLocalFileStoreDelete(t *testing.T)
func TestLocalFileStoreDirectoryTraversal(t *testing.T)
func TestLocalFileStoreDeleteNonExistent(t *testing.T)

// InMemoryFileStore tests
func TestInMemoryFileStoreWrite(t *testing.T)
func TestInMemoryFileStoreRead(t *testing.T)
func TestTestInMemoryFileStoreList(t *testing.T)
func TestInMemoryFileStoreDelete(t *testing.T)
func TestInMemoryFileStoreReadNonExistent(t *testing.T)
func TestInMemoryFileStoreConcurrency(t *testing.T)

// Helper tests
func TestGetConversationAgentHistoryFilename(t *testing.T)
```

---

### 4. Database Package (`db/`)

**File**: `db/db_test.go`

```go
// Mock setup for database tests
func setupTestDB(t *testing.T) *gorm.DB
func teardownTestDB(db *gorm.DB)

// Session tests
func TestSessionBeforeCreate(t *testing.T)
func TestCreateSession(t *testing.T)
func TestGetSessionByWorkspace(t *testing.T)
func TestGetSessionByID(t *testing.T)
func TestGetSessionByDeviceID(t *testing.T)
func TestUpdateSessionName(t *testing.T)
func TestGetSandboxIDBySessionID(t *testing.T)
func TestUpdateSessionSandboxID(t *testing.T)
func TestGetSessionsByDeviceID(t *testing.T)

// Event tests
func TestEventBeforeCreate(t *testing.T)
func TestSaveEvent(t *testing.T)
func TestGetSessionEvents(t *testing.T)
func TestDeleteSessionEvents(t *testing.T)
func TestDeleteEventsFromLastToUserMessage(t *testing.T)
func TestDeleteEventsFromLastToUserMessageNoUserMessage(t *testing.T)
func TestGetSessionEventsWithDetails(t *testing.T)
```

---

### 5. LLM Package (`llm/`)

**File**: `llm/commom_test.go`

```go
// APIType tests
func TestAPITypeConstants(t *testing.T)

// ContentBlock tests
func TestContentBlockText(t *testing.T)
func TestContentBlockImage(t *testing.T)
func TestContentBlockToolCall(t *testing.T)
func TestContentBlockToolResult(t *testing.T)
func TestContentBlockThinking(t *testing.T)

// Client factory tests
func TestGetClientOpenAI(t *testing.T)
func TestGetClientAnthropic(t *testing.T)
func TestGetClientGemini(t *testing.T)
func TestGetClientUnknown(t *testing.T)

// MessageHistory tests
func TestNewMessageHistory(t *testing.T)
func TestMessageHistoryAddUserPrompt(t *testing.T)
func TestMessageHistoryAddUserPromptWithImages(t *testing.T)
func TestMessageHistoryAddAssistantTurn(t *testing.T)
func TestMessageHistoryAddToolResult(t *testing.T)
func TestMessageHistoryGetMessages(t *testing.T)
func TestMessageHistoryClear(t *testing.T)
func TestMessageHistoryEnsureToolCallIntegrity(t *testing.T)
func TestMessageHistorySaveToFile(t *testing.T)
func TestMessageHistoryLoadFromFile(t *testing.T)

// Utility tests
func TestGenerateID(t *testing.T)
func TestCountTokens(t *testing.T)
```

---

### 6. Context Manager Package (`llm/context_manager/`)

**File**: `llm/context_manager/context_manager_test.go`

```go
// Content block type tests
func TestTextPromptType(t *testing.T)
func TestTextResultType(t *testing.T)
func TestToolCallType(t *testing.T)
func TestToolFormattedResultType(t *testing.T)
func TestImageBlockType(t *testing.T)
func TestAnthropicThinkingBlockType(t *testing.T)
func TestAnthropicRedactedThinkingBlockType(t *testing.T)

// Manager tests
func TestNewManager(t *testing.T)
func TestManagerCountTokens(t *testing.T)
func TestManagerCountTokensWithImages(t *testing.T)
func TestManagerCountTokensWithThinking(t *testing.T)
func TestApplyTruncationIfNeededNoTruncation(t *testing.T)
func TestApplyTruncationIfNeededWithTruncation(t *testing.T)
func TestApplyTruncationIfNeededExceedsMaxSize(t *testing.T)

// Truncation strategy tests
func TestTruncateWithThinkingBlocks(t *testing.T)
func TestTruncateStandard(t *testing.T)
func TestTruncateStandardWithExistingSummary(t *testing.T)
func TestGenerateSummary(t *testing.T)
func TestGenerateCompleteConversationSummary(t *testing.T)
func TestGenerateCompleteConversationSummaryEmpty(t *testing.T)

// Helper tests
func TestTruncateContent(t *testing.T)
func TestMessageListToString(t *testing.T)
func TestHasThinkingBlocks(t *testing.T)
func TestFindLastTextPromptIndex(t *testing.T)
```

---

### 7. Agents Package (`agents/`)

**File**: `agents/base_test.go`
**File**: `agents/types_test.go`

```go
// BaseAgent tests
func TestBaseAgentGetToolParam(t *testing.T)
func TestBaseAgentRunNotImplemented(t *testing.T)

// EventType tests
func TestEventTypeConstants(t *testing.T)

// ToolImplOutput tests
func TestToolImplOutput(t *testing.T)
func TestToolImplOutputIsFinal(t *testing.T)

// ToolCallParameters tests
func TestToolCallParameters(t *testing.T)

// ToolParam tests
func TestToolParam(t *testing.T)

// Message tests
func TestMessage(t *testing.T)
func TestMessageWithImages(t *testing.T)

// RealtimeEvent tests
func TestRealtimeEvent(t *testing.T)
```

---

### 8. Sandbox Package (`sandbox/`)

**File**: `sandbox/sandbox_test.go`

```go
// WorkSpaceMode tests
func TestWorkSpaceModeConstants(t *testing.T)

// Base tests
func TestBaseGetHostURL(t *testing.T)
func TestBaseGetHostURLEmpty(t *testing.T)
func TestBaseGetSandboxID(t *testing.T)
func TestBaseGetSandboxIDEmpty(t *testing.T)

// Registry tests
func TestRegister(t *testing.T)
func TestCreate(t *testing.T)
func TestCreateUnknown(t *testing.T)
func TestRegisterConcurrent(t *testing.T)
```

---

### 9. Server Package (`server/`)

**File**: `server/server_test.go`

```go
// Config tests
func TestConfigGetPort(t *testing.T)
func TestConfigGetPortDefault(t *testing.T)
func TestConfigGetWorkspaceRoot(t *testing.T)
func TestConfigGetWorkspaceRootDefault(t *testing.T)

// ConnectionManager tests
func TestNewConnectionManager(t *testing.T)
func TestConnectionManagerConnect(t *testing.T)
func TestConnectionManagerConnectInvalidUUID(t *testing.T)
func TestConnectionManagerDisconnect(t *testing.T)

// ChatSession tests
func TestChatSessionSendEvent(t *testing.T)
func TestChatSessionHandleMessage(t *testing.T)
func TestChatSessionHandleInitAgent(t *testing.T)
func TestChatSessionHandleQuery(t *testing.T)
func TestChatSessionHandleSlashCommandHelp(t *testing.T)
func TestChatSessionHandleSlashCommandCompact(t *testing.T)
func TestChatSessionHandleSlashCommandUnknown(t *testing.T)
func TestChatSessionStartLoop(t *testing.T)

// HTTP handlers tests
func TestUploadHandler(tfunc TestUploadHandler *testing.T)
MissingSessionID(t *testing.T)
func TestUploadHandlerBase64(t *testing.T)
func TestUploadHandlerFilenameCollision(t *testing.T)
func TestGetSessionsHandler(t *testing.T)
func TestGetEventsHandler(t *testing.T)
func TestSessionsHandler(t *testing.T)
func TestGetSettingsHandler(t *testing.T)
func TestPostSettingsHandler(t *testing.T)

// Factory tests
func TestCreateServer(t *testing.T)
func TestGetContentType(t *testing.T)
```

---

### 10. Tools Package (`tools/`)

**File**: `tools/base_test.go`
**File**: `tools/terminal_test.go`

```go
// ToolInput tests
func TestToolInput(t *testing.T)

// ToolOutput tests
func TestToolOutput(t *testing.T)
func TestToolOutputWithImages(t *testing.T)
func TestToolOutputWithAuxiliary(t *testing.T)
func TestToolOutputWithError(t *testing.T)

// Config tests
func TestConfig(t *testing.T)

// ErrorOutput tests
func TestErrorOutput(t *testing.T)
func TestErrorOutputNil(t *testing.T)

// GetArg tests
func TestGetArgString(t *testing.T)
func TestGetArgInt(t *testing.T)
func TestGetArgMissing(t *testing.T)
func TestGetArgInvalidType(t *testing.T)
func TestGetArgFloat64ToInt(t *testing.T)

// FileEditorTool tests
func TestFileEditorToolName(t *testing.T)
func TestFileEditorToolDescription(t *testing.T)
func TestFileEditorToolInputSchema(t *testing.T)
func TestFileEditorToolRunRead(t *testing.T)
func TestFileEditorToolRunWrite(t *testing.T)
func TestFileEditorToolRunStrReplace(t *testing.T)
func TestFileEditorToolRunStrReplaceMultiple(t *testing.T)
func TestFileEditorToolRunStrReplaceNotFound(t *testing.T)
func TestFileEditorToolRunUnknownAction(t *testing.T)
func TestFileEditorToolRunDirectoryTraversal(t *testing.T)

// TerminalTool tests
func TestTerminalToolName(t *testing.T)
func TestTerminalToolDescription(t *testing.T)
func TestTerminalToolInputSchema(t *testing.T)
func TestTerminalToolRun(t *testing.T)
func TestTerminalToolRunWithTimeout(t *testing.T)
func TestTerminalToolRunTimeout(t *testing.T)
func TestTerminalToolRunCommandError(t *testing.T)
```

---

### 11. Utils Package (`utils/`)

**File**: `utils/common_test.go`

```go
// Constants tests
func TestConstants(t *testing.T)

// WorkspaceMode tests
func TestWorkspaceMode(t *testing.T)

// SessionResult tests
func TestSessionResult(t *testing.T)
func TestSessionResultSuccess(t *testing.T)
func TestSessionResultFailure(t *testing.T)

// StrReplaceResponse tests
func TestStrReplaceResponse(t *testing.T)
func TestStrReplaceResponseSuccess(t *testing.T)

// SandboxSettings tests
func TestNewSandboxSettings(t *testing.T)
func TestSandboxSettingsDefaults(t *testing.T)

// WorkspaceManager tests
func TestNewWorkspaceManager(t *testing.T)
func TestWorkspaceManagerIsLocal(t *testing.T)
func TestWorkspaceManagerIsLocalDocker(t *testing.T)
func TestWorkspaceManagerWorkspacePath(t *testing.T)
func TestWorkspaceManagerRootPath(t *testing.T)
func TestWorkspaceManagerRootPathNonLocal(t *testing.T)
```

---

## Test Utilities and Helpers

### Mock Implementations
For packages requiring external dependencies (database, LLM clients), create mock implementations:

```go
// Mock LLM Client
type MockLLMClient struct {
    GenerateFunc func(ctx context.Context, messages [][]ContentBlock, maxTokens int, temperature float64) ([]ContentBlock, error)
}

// Mock Token Counter
type MockTokenCounter struct {
    CountTokensFunc func(text string) int
}

// Mock WebSocket
type MockWebSocket struct {
    SendJSONFunc func(v interface{}) error
}
```

### Test Fixtures
- Temporary directories for file operations
- In-memory SQLite databases for DB tests
- Mock configurations for config tests

## Test Execution

### Makefile Integration
Add to `Makefile`:
```makefile
.PHONY: test test-coverage

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
```

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test ./core/... -v

# Run tests with verbose output
go test ./... -v -count=1
```

## Coverage Goals

| Package | Target Coverage |
|---------|----------------|
| core | 100% |
| config | 100% |
| storage | 100% |
| db | 90%+ |
| llm | 90%+ |
| context_manager | 90%+ |
| agents | 85%+ |
| sandbox | 100% |
| server | 85%+ |
| tools | 90%+ |
| utils | 100% |

## Implementation Order

1. **Phase 1: Core Infrastructure**
   - core/even.go
   - core/logger.go
   - core/config/
   - core/storage/

2. **Phase 2: Data Layer**
   - db/
   - utils/

3. **Phase 3: Business Logic**
   - llm/
   - llm/context_manager/
   - agents/

4. **Phase 4: Services**
   - sandbox/
   - tools/
   - server/

## Notes

- All tests should be independent and not rely on external services
- Use table-driven tests for comprehensive coverage
- Include edge cases in test scenarios
- Ensure tests are fast (< 1s per package where possible)
- Use subtests for related test cases
- Mock external dependencies (browser, Docker, LLM APIs)
