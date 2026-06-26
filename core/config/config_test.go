package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSecretStringString(t *testing.T) {
	tests := []struct {
		name     string
		input    SecretString
		expected string
	}{
		{"empty", "", ""},
		{"normal", "secret_value", "********"},
		{"api_key", "sk-123456", "********"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %s; want %s", result, tt.expected)
			}
		})
	}
}

func TestSecretStringReveal(t *testing.T) {
	tests := []struct {
		name     string
		input    SecretString
		expected string
	}{
		{"empty", "", ""},
		{"normal", "secret_value", "secret_value"},
		{"api_key", "sk-123456", "sk-123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Reveal()
			if result != tt.expected {
				t.Errorf("Reveal() = %s; want %s", result, tt.expected)
			}
		})
	}
}

func TestSecretStringMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    SecretString
		expected string
	}{
		{"empty", "", `""`},
		{"normal", "secret", `"********"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("Marshal() = %s; want %s", string(data), tt.expected)
			}
		})
	}
}


func TestGetEnv(t *testing.T) {
	// Save original env
	origVal := os.Getenv("WATER_TEST_KEY")
	defer os.Setenv("WATER_TEST_KEY", origVal)

	// Test when not set
	os.Unsetenv("WATER_TEST_KEY")
	result := getEnv("WATER_TEST_KEY", "default_value")
	if result != "default_value" {
		t.Errorf("getEnv() = %s; want default_value", result)
	}

	// Test when set
	os.Setenv("WATER_TEST_KEY", "actual_value")
	result = getEnv("WATER_TEST_KEY", "default_value")
	if result != "actual_value" {
		t.Errorf("getEnv() = %s; want actual_value", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	// Save original env
	origVal := os.Getenv("WATER_TEST_INT")
	defer os.Setenv("WATER_TEST_INT", origVal)

	tests := []struct {
		name     string
		envVal   string
		fallback int
		expected int
		wantErr  bool
	}{
		{"not_set", "", 10, 10, false},
		{"valid", "42", 10, 42, false},
		{"invalid", "abc", 10, 10, false},
		{"negative", "-5", 10, -5, false},
		{"zero", "0", 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("WATER_TEST_INT")
			} else {
				os.Setenv("WATER_TEST_INT", tt.envVal)
			}
			result := getEnvInt("WATER_TEST_INT", tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnvInt() = %d; want %d", result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	// Save original env
	origVal := os.Getenv("WATER_TEST_BOOL")
	defer os.Setenv("WATER_TEST_BOOL", origVal)

	tests := []struct {
		name     string
		envVal   string
		fallback bool
		expected bool
	}{
		{"not_set", "", false, false},
		{"not_set_true", "", true, true},
		{"true_lower", "true", false, true},
		{"True_upper", "True", false, true},
		{"TRUE_mixed", "TRUE", false, true},
		{"false_lower", "false", true, false},
		{"False_upper", "False", true, false},
		{"1", "1", false, true},
		{"0", "0", true, false},
		{"invalid", "invalid", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("WATER_TEST_BOOL")
			} else {
				os.Setenv("WATER_TEST_BOOL", tt.envVal)
			}
			result := getEnvBool("WATER_TEST_BOOL", tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnvBool() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	val := "test_value"
	ptr := StringPtr(val)

	if ptr == nil {
		t.Fatal("StringPtr returned nil")
	}

	if *ptr != val {
		t.Errorf("*StringPtr() = %s; want %s", *ptr, val)
	}
}

func TestSecretPtr(t *testing.T) {
	val := "secret_value"
	ptr := SecretPtr(val)

	if ptr == nil {
		t.Fatal("SecretPtr returned nil")
	}

	if *ptr != SecretString(val) {
		t.Errorf("*SecretPtr() = %s; want %s", *ptr, val)
	}
}

func TestNewWaterAgentConfig(t *testing.T) {
	// Save original env
	origEnv := map[string]string{
		"FILE_STORE":              os.Getenv("FILE_STORE"),
		"FILE_STORE_PATH":        os.Getenv("FILE_STORE_PATH"),
		"HOST_WORKSPACE_PATH":    os.Getenv("HOST_WORKSPACE_PATH"),
		"USE_CONTAINER_WORKSPACE": os.Getenv("USE_CONTAINER_WORKSPACE"),
		"MINIMIZE_STDOUT_LOGS":   os.Getenv("MINIMIZE_STDOUT_LOGS"),
		"DATABASE_URL":           os.Getenv("DATABASE_URL"),
	}
	defer func() {
		for k, v := range origEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Clear all test env vars
	os.Unsetenv("FILE_STORE")
	os.Unsetenv("FILE_STORE_PATH")
	os.Unsetenv("HOST_WORKSPACE_PATH")
	os.Unsetenv("USE_CONTAINER_WORKSPACE")
	os.Unsetenv("MINIMIZE_STDOUT_LOGS")
	os.Unsetenv("DATABASE_URL")

	cfg, err := NewWaterAgentConfig()
	if err != nil {
		t.Fatalf("NewWaterAgentConfig() error = %v", err)
	}

	if cfg.FileStore != "local" {
		t.Errorf("FileStore = %s; want local", cfg.FileStore)
	}

	if cfg.MaxOutputTokensPerTurn != MaxOutputTokensPerTurn {
		t.Errorf("MaxOutputTokensPerTurn = %d; want %d", cfg.MaxOutputTokensPerTurn, MaxOutputTokensPerTurn)
	}

	if cfg.MaxTurns != MaxTurns {
		t.Errorf("MaxTurns = %d; want %d", cfg.MaxTurns, MaxTurns)
	}
}

func TestWaterAgentConfigWorkspaceRoot(t *testing.T) {
	cfg := &WaterAgentConfig{
		FileStorePath: "/tmp/test",
	}

	expected := "/tmp/test/workspace"
	if got := cfg.WorkspaceRoot(); got != expected {
		t.Errorf("WorkspaceRoot() = %s; want %s", got, expected)
	}
}

func TestWaterAgentConfigLogsPath(t *testing.T) {
	cfg := &WaterAgentConfig{
		FileStorePath: "/tmp/test",
	}

	expected := "/tmp/test/logs"
	if got := cfg.LogsPath(); got != expected {
		t.Errorf("LogsPath() = %s; want %s", got, expected)
	}
}

func TestWaterAgentConfigCodeServerPort(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		expected int
	}{
		{"default", "", 9000},
		{"custom", "9001", 9001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origVal := os.Getenv("CODE_SERVER_PORT")
			defer func() {
				if origVal == "" {
					os.Unsetenv("CODE_SERVER_PORT")
				} else {
					os.Setenv("CODE_SERVER_PORT", origVal)
				}
			}()

			if tt.envVal == "" {
				os.Unsetenv("CODE_SERVER_PORT")
			} else {
				os.Setenv("CODE_SERVER_PORT", tt.envVal)
			}

			cfg := &WaterAgentConfig{}
			if got := cfg.CodeServerPort(); got != tt.expected {
				t.Errorf("CodeServerPort() = %d; want %d", got, tt.expected)
			}
		})
	}
}

func TestNewLLMConfig(t *testing.T) {
	cfg := NewLLMConfig()

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %s; want %s", cfg.Model, DefaultModel)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d; want 3", cfg.MaxRetries)
	}

	if cfg.MaxMessageChars != 30000 {
		t.Errorf("MaxMessageChars = %d; want 30000", cfg.MaxMessageChars)
	}

	if cfg.Temperature != 0.0 {
		t.Errorf("Temperature = %f; want 0.0", cfg.Temperature)
	}

	if cfg.APIType != APITypeAnthropic {
		t.Errorf("APIType = %s; want %s", cfg.APIType, APITypeAnthropic)
	}

	if cfg.ThinkingTokens != 0 {
		t.Errorf("ThinkingTokens = %d; want 0", cfg.ThinkingTokens)
	}

	if cfg.CoTModel != false {
		t.Error("CoTModel should be false")
	}
}

func TestNewSandboxConfig(t *testing.T) {
	cfg := NewSandboxConfig()

	if cfg.Mode != WorkSpaceModeDocker {
		t.Errorf("Mode = %s; want %s", cfg.Mode, WorkSpaceModeDocker)
	}

	if cfg.ServicePort != 17300 {
		t.Errorf("ServicePort = %d; want 17300", cfg.ServicePort)
	}
}

func TestClientConfigNew(t *testing.T) {
	cfg := NewClientConfig()

	if cfg.Timeout != 600.0 {
		t.Errorf("Timeout = %f; want 600.0", cfg.Timeout)
	}

	if cfg.DefaultShell != "/bin/bash" {
		t.Errorf("DefaultShell = %s; want /bin/bash", cfg.DefaultShell)
	}

	if cfg.DefaultTimeout != 600 {
		t.Errorf("DefaultTimeout = %d; want 600", cfg.DefaultTimeout)
	}
}

func TestAudioConfigUpdate(t *testing.T) {
	cfg := &AudioConfig{}
	apiKey := SecretPtr("sk-test")
	azureEndpoint := StringPtr("https://test.openai.azure.com/")
	azureVersion := StringPtr("2023-05-15")

	cfg.Update(AudioConfig{
		OpenAIAPIKey:    apiKey,
		AzureEndpoint:   azureEndpoint,
		AzureAPIVersion: azureVersion,
	})

	if cfg.OpenAIAPIKey == nil || *cfg.OpenAIAPIKey != "sk-test" {
		t.Error("OpenAIAPIKey was not updated")
	}

	if cfg.AzureEndpoint == nil || *cfg.AzureEndpoint != "https://test.openai.azure.com/" {
		t.Error("AzureEndpoint was not updated")
	}

	if cfg.AzureAPIVersion == nil || *cfg.AzureAPIVersion != "2023-05-15" {
		t.Error("AzureAPIVersion was not updated")
	}
}

func TestSearchConfigUpdate(t *testing.T) {
	cfg := &SearchConfig{}

	firecrawl := SecretPtr("fc-test")
	serpapi := SecretPtr("serp-test")
	tavily := SecretPtr("tv-test")
	jina := SecretPtr("jn-test")

	cfg.Update(SearchConfig{
		FirecrawlAPIKey: firecrawl,
		SerpapiAPIKey:   serpapi,
		TavilyAPIKey:    tavily,
		JinaAPIKey:      jina,
	})

	if cfg.FirecrawlAPIKey == nil || *cfg.FirecrawlAPIKey != "fc-test" {
		t.Error("FirecrawlAPIKey was not updated")
	}

	if cfg.SerpapiAPIKey == nil || *cfg.SerpapiAPIKey != "serp-test" {
		t.Error("SerpapiAPIKey was not updated")
	}

	if cfg.TavilyAPIKey == nil || *cfg.TavilyAPIKey != "tv-test" {
		t.Error("TavilyAPIKey was not updated")
	}

	if cfg.JinaAPIKey == nil || *cfg.JinaAPIKey != "jn-test" {
		t.Error("JinaAPIKey was not updated")
	}
}

func TestSandboxConfigUpdate(t *testing.T) {
	cfg := &SandboxConfig{}

	template := StringPtr("test-template")
	apiKey := SecretPtr("sandbox-key")

	cfg.Update(SandboxConfig{
		Mode:        WorkSpaceModeLocal,
		TemplateID:  template,
		SandboxAPIKey: apiKey,
		ServicePort: 8080,
	})

	if cfg.Mode != WorkSpaceModeLocal {
		t.Errorf("Mode = %s; want %s", cfg.Mode, WorkSpaceModeLocal)
	}

	if cfg.TemplateID == nil || *cfg.TemplateID != "test-template" {
		t.Error("TemplateID was not updated")
	}

	if cfg.SandboxAPIKey == nil || *cfg.SandboxAPIKey != "sandbox-key" {
		t.Error("SandboxAPIKey was not updated")
	}

	if cfg.ServicePort != 8080 {
		t.Errorf("ServicePort = %d; want 8080", cfg.ServicePort)
	}
}

func TestMediaConfigUpdate(t *testing.T) {
	cfg := &MediaConfig{}

	projectID := StringPtr("test-project")
	location := StringPtr("us-central1")
	bucket := StringPtr("test-bucket")
	aiStudioKey := SecretPtr("ai-studio-key")

	cfg.Update(MediaConfig{
		GCPProjectID:         projectID,
		GCPLocation:          location,
		GCSOutputBucket:      bucket,
		GoogleAIStudioAPIKey: aiStudioKey,
	})

	if cfg.GCPProjectID == nil || *cfg.GCPProjectID != "test-project" {
		t.Error("GCPProjectID was not updated")
	}

	if cfg.GCPLocation == nil || *cfg.GCPLocation != "us-central1" {
		t.Error("GCPLocation was not updated")
	}

	if cfg.GCSOutputBucket == nil || *cfg.GCSOutputBucket != "test-bucket" {
		t.Error("GCSOutputBucket was not updated")
	}

	if cfg.GoogleAIStudioAPIKey == nil || *cfg.GoogleAIStudioAPIKey != "ai-studio-key" {
		t.Error("GoogleAIStudioAPIKey was not updated")
	}
}
