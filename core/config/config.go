package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// =============================================================================
// Constants & Enums
// =============================================================================

const (
	MaxOutputTokensPerTurn = 32000
	MaxTurns               = 200
	TokenBudget            = 0 // Assuming 0 as default from original import placeholder
	DefaultModel           = "gpt-4-turbo"
)

type WorkSpaceMode string

const (
	WorkSpaceModeDocker WorkSpaceMode = "docker"
	WorkSpaceModeLocal  WorkSpaceMode = "local"
	WorkSpaceModeE2B    WorkSpaceMode = "e2b"
)

type APIType string

const (
	APITypeOpenAI    APIType = "openai"
	APITypeAnthropic APIType = "anthropic"
	APITypeGemini    APIType = "gemini"
)

// =============================================================================
// Helper Types (Secrets)
// =============================================================================

// SecretString handles sensitive data. It marshals to asterisks by default
// to prevent accidental logging.
type SecretString string

// String returns the redacted value.
func (s SecretString) String() string {
	if s == "" {
		return ""
	}
	return "********"
}

// Reveal returns the actual value.
func (s SecretString) Reveal() string {
	return string(s)
}

// MarshalJSON customizes JSON serialization to hide the secret.
func (s SecretString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// =============================================================================
// Audio Config
// =============================================================================

type AudioConfig struct {
	OpenAIAPIKey    *SecretString `json:"openai_api_key,omitempty"`
	AzureEndpoint   *string       `json:"azure_endpoint,omitempty"`
	AzureAPIVersion *string       `json:"azure_api_version,omitempty"`
}

func (c *AudioConfig) Update(settings AudioConfig) {
	if settings.OpenAIAPIKey != nil && c.OpenAIAPIKey == nil {
		c.OpenAIAPIKey = settings.OpenAIAPIKey
	}
	if settings.AzureEndpoint != nil && c.AzureEndpoint == nil {
		c.AzureEndpoint = settings.AzureEndpoint
	}
	if settings.AzureAPIVersion != nil && c.AzureAPIVersion == nil {
		c.AzureAPIVersion = settings.AzureAPIVersion
	}
}

// =============================================================================
// Database / Third Party Config
// =============================================================================

type ThirdPartyIntegrationConfig struct {
	NeonDBAPIKey *SecretString `json:"neon_db_api_key,omitempty"`
	OpenAIAPIKey *SecretString `json:"openai_api_key,omitempty"`
	VercelAPIKey *SecretString `json:"vercel_api_key,omitempty"`
}

func (c *ThirdPartyIntegrationConfig) Update(settings ThirdPartyIntegrationConfig) {
	if settings.NeonDBAPIKey != nil && c.NeonDBAPIKey == nil {
		c.NeonDBAPIKey = settings.NeonDBAPIKey
	}
	if settings.OpenAIAPIKey != nil && c.OpenAIAPIKey == nil {
		c.OpenAIAPIKey = settings.OpenAIAPIKey
	}
	if settings.VercelAPIKey != nil && c.VercelAPIKey == nil {
		c.VercelAPIKey = settings.VercelAPIKey
	}
}

// =============================================================================
// Client Config
// =============================================================================

type ClientConfig struct {
	ServerURL                      *string `json:"server_url,omitempty"`
	Timeout                        float64 `json:"timeout"`
	IgnoreIndentationForStrReplace bool    `json:"ignore_indentation_for_str_replace"`
	ExpandTabs                     bool    `json:"expand_tabs"`
	DefaultShell                   string  `json:"default_shell"`
	DefaultTimeout                 int     `json:"default_timeout"`
	Cwd                            *string `json:"cwd,omitempty"`
}

func NewClientConfig() ClientConfig {
	return ClientConfig{
		Timeout:        600.0,
		DefaultShell:   "/bin/bash",
		DefaultTimeout: 600,
	}
}

func (c *ClientConfig) Update(settings ClientConfig) {
	// Pass implementation as per original python file
}

// =============================================================================
// Water Agent Config (Formerly IIAgentConfig)
// =============================================================================

type WaterAgentConfig struct {
	FileStore              string        `json:"file_store"`
	FileStorePath          string        `json:"file_store_path"`
	HostWorkspacePath      string        `json:"host_workspace_path"`
	UseContainerWorkspace  WorkSpaceMode `json:"use_container_workspace"`
	MinimizeStdoutLogs     bool          `json:"minimize_stdout_logs"`
	MaxOutputTokensPerTurn int           `json:"max_output_tokens_per_turn"`
	MaxTurns               int           `json:"max_turns"`
	TokenBudget            int           `json:"token_budget"`
	DatabaseURL            *string       `json:"database_url,omitempty"`
}

// NewWaterAgentConfig loads defaults and processes environment variables roughly like Pydantic BaseSettings
func NewWaterAgentConfig() (*WaterAgentConfig, error) {
	// Defaults
	cfg := &WaterAgentConfig{
		FileStore:              getEnv("FILE_STORE", "local"),
		FileStorePath:          getEnv("FILE_STORE_PATH", "~/.water_agent"),
		HostWorkspacePath:      getEnv("HOST_WORKSPACE_PATH", "~/.water_agent/workspace"),
		UseContainerWorkspace:  WorkSpaceMode(getEnv("USE_CONTAINER_WORKSPACE", string(WorkSpaceModeDocker))),
		MinimizeStdoutLogs:     getEnvBool("MINIMIZE_STDOUT_LOGS", false),
		MaxOutputTokensPerTurn: getEnvInt("MAX_OUTPUT_TOKENS_PER_TURN", MaxOutputTokensPerTurn),
		MaxTurns:               getEnvInt("MAX_TURNS", MaxTurns),
		TokenBudget:            getEnvInt("TOKEN_BUDGET", TokenBudget),
	}

	// Expand paths
	var err error
	cfg.FileStorePath, err = expandPath(cfg.FileStorePath)
	if err != nil {
		return nil, err
	}
	
	// Handle Database URL logic
	dbUrl := getEnv("DATABASE_URL", "")
	if dbUrl != "" {
		cfg.DatabaseURL = &dbUrl
	} else {
		// Default SQLite path logic
		generatedURL := "sqlite:///" + filepath.Join(cfg.FileStorePath, "water_agent.db")
		cfg.DatabaseURL = &generatedURL
	}

	return cfg, nil
}

// Computed Properties

func (c *WaterAgentConfig) WorkspaceRoot() string {
	return filepath.Join(c.FileStorePath, "workspace")
}

func (c *WaterAgentConfig) HostWorkspace() (string, error) {
	return expandPath(c.HostWorkspacePath)
}

func (c *WaterAgentConfig) LogsPath() string {
	return filepath.Join(c.FileStorePath, "logs")
}

func (c *WaterAgentConfig) CodeServerPort() int {
	return getEnvInt("CODE_SERVER_PORT", 9000)
}

// =============================================================================
// LLM Config
// =============================================================================

type LLMConfig struct {
	Model           string        `json:"model"`
	APIKey          *SecretString `json:"api_key,omitempty"`
	BaseURL         *string       `json:"base_url,omitempty"`
	MaxRetries      int           `json:"max_retries"`
	MaxMessageChars int           `json:"max_message_chars"`
	Temperature     float64       `json:"temperature"`
	VertexRegion    *string       `json:"vertex_region,omitempty"`
	VertexProjectID *string       `json:"vertex_project_id,omitempty"`
	APIType         APIType       `json:"api_type"`
	ThinkingTokens  int           `json:"thinking_tokens"`
	AzureEndpoint   *string       `json:"azure_endpoint,omitempty"`
	AzureAPIVersion *string       `json:"azure_api_version,omitempty"`
	CoTModel        bool          `json:"cot_model"`
}

func NewLLMConfig() LLMConfig {
	return LLMConfig{
		Model:           DefaultModel,
		MaxRetries:      3,
		MaxMessageChars: 30000,
		Temperature:     0.0,
		APIType:         APITypeAnthropic,
		ThinkingTokens:  0,
		CoTModel:        false,
	}
}

// =============================================================================
// Media Config
// =============================================================================

type MediaConfig struct {
	GCPProjectID          *string       `json:"gcp_project_id,omitempty"`
	GCPLocation           *string       `json:"gcp_location,omitempty"`
	GCSOutputBucket       *string       `json:"gcs_output_bucket,omitempty"`
	GoogleAIStudioAPIKey  *SecretString `json:"google_ai_studio_api_key,omitempty"`
}

func (c *MediaConfig) Update(settings MediaConfig) {
	if settings.GCPProjectID != nil && c.GCPProjectID == nil {
		c.GCPProjectID = settings.GCPProjectID
	}
	if settings.GCPLocation != nil && c.GCPLocation == nil {
		c.GCPLocation = settings.GCPLocation
	}
	if settings.GCSOutputBucket != nil && c.GCSOutputBucket == nil {
		c.GCSOutputBucket = settings.GCSOutputBucket
	}
	if settings.GoogleAIStudioAPIKey != nil && c.GoogleAIStudioAPIKey == nil {
		c.GoogleAIStudioAPIKey = settings.GoogleAIStudioAPIKey
	}
}

// =============================================================================
// Sandbox Config
// =============================================================================

type SandboxConfig struct {
	Mode          WorkSpaceMode `json:"mode"`
	TemplateID    *string       `json:"template_id,omitempty"`
	SandboxAPIKey *SecretString `json:"sandbox_api_key,omitempty"`
	ServicePort   int           `json:"service_port"`
}

func NewSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Mode:        WorkSpaceModeDocker,
		ServicePort: 17300,
	}
}

func (c *SandboxConfig) Update(settings SandboxConfig) {
	if settings.SandboxAPIKey != nil && c.SandboxAPIKey == nil {
		c.SandboxAPIKey = settings.SandboxAPIKey
	}
	if settings.ServicePort != 0 { 
		// Go int zero value is 0, assuming 0 means not set here, 
		// though logic differs slightly from None
		c.ServicePort = settings.ServicePort
	}
	if settings.Mode != "" {
		c.Mode = settings.Mode
	}
	if settings.TemplateID != nil && c.TemplateID == nil {
		c.TemplateID = settings.TemplateID
	}
}

// =============================================================================
// Search Config
// =============================================================================

type SearchConfig struct {
	FirecrawlAPIKey *SecretString `json:"firecrawl_api_key,omitempty"`
	SerpapiAPIKey   *SecretString `json:"serpapi_api_key,omitempty"`
	TavilyAPIKey    *SecretString `json:"tavily_api_key,omitempty"`
	JinaAPIKey      *SecretString `json:"jina_api_key,omitempty"`
}

func (c *SearchConfig) Update(settings SearchConfig) {
	if settings.FirecrawlAPIKey != nil && c.FirecrawlAPIKey == nil {
		c.FirecrawlAPIKey = settings.FirecrawlAPIKey
	}
	if settings.SerpapiAPIKey != nil && c.SerpapiAPIKey == nil {
		c.SerpapiAPIKey = settings.SerpapiAPIKey
	}
	if settings.TavilyAPIKey != nil && c.TavilyAPIKey == nil {
		c.TavilyAPIKey = settings.TavilyAPIKey
	}
	if settings.JinaAPIKey != nil && c.JinaAPIKey == nil {
		c.JinaAPIKey = settings.JinaAPIKey
	}
}

// =============================================================================
// Utils
// =============================================================================

// expandPath expands the tilde (~) in paths to the user's home directory.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt retrieves an environment variable as an int or returns a default.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

// getEnvBool retrieves an environment variable as a bool or returns a default.
func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}

// Helper for creating string pointers
func StringPtr(s string) *string {
	return &s
}

// Helper for creating SecretString pointers
func SecretPtr(s string) *SecretString {
	ss := SecretString(s)
	return &ss
}