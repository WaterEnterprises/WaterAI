package models

import (
	"testing"
)

func TestSettingsUpdateNilIncoming(t *testing.T) {
	settings := &Settings{}
	settings.Update(nil)
	// Should not panic
}

func TestSettingsUpdateLLMConfigs(t *testing.T) {
	settings := &Settings{
		LLMConfigs: map[string]*LLMConfig{
			"gpt-4": {Model: "gpt-4", APIKey: ""},
		},
	}

	incomig := &Settings{
		LLMConfigs: map[string]*LLMConfig{
			"gpt-4": {Model: "gpt-4", APIKey: "sk-test"},
		},
	}

	settings.Update(incomig)

	if settings.LLMConfigs["gpt-4"].APIKey != "sk-test" {
		t.Errorf("APIKey = %s; want sk-test", settings.LLMConfigs["gpt-4"].APIKey)
	}
}

func TestSettingsUpdateEmptyLLMConfigs(t *testing.T) {
	settings := &Settings{
		LLMConfigs: nil,
	}

	incoming := &Settings{
		LLMConfigs: map[string]*LLMConfig{
			"gpt-4": {Model: "gpt-4", APIKey: "sk-test"},
		},
	}

	settings.Update(incoming)

	if settings.LLMConfigs == nil {
		t.Error("LLMConfigs should not be nil after update")
	}

	if settings.LLMConfigs["gpt-4"].Model != "gpt-4" {
		t.Errorf("Model = %s; want gpt-4", settings.LLMConfigs["gpt-4"].Model)
	}
}

func TestSearchConfigUpdate(t *testing.T) {
	cfg := &SearchConfig{APIKey: ""}
	incoming := &SearchConfig{APIKey: "tavily-123"}

	cfg.Update(incoming)

	if cfg.APIKey != "tavily-123" {
		t.Errorf("APIKey = %s; want tavily-123", cfg.APIKey)
	}
}

func TestMediaConfigUpdate(t *testing.T) {
	cfg := &MediaConfig{OutputDir: ""}
	incoming := &MediaConfig{OutputDir: "/tmp/output"}

	cfg.Update(incoming)

	if cfg.OutputDir != "/tmp/output" {
		t.Errorf("OutputDir = %s; want /tmp/output", cfg.OutputDir)
	}
}

func TestSandboxConfigUpdate(t *testing.T) {
	cfg := &SandboxConfig{}
	// SandboxConfig in settings.go uses 'DockerImage', not 'Mode'
	incoming := &SandboxConfig{DockerImage: "python:3.9"}

	cfg.Update(incoming)

	if cfg.DockerImage != "python:3.9" {
		t.Errorf("DockerImage = %s; want python:3.9", cfg.DockerImage)
	}
}

func TestClientConfigUpdate(t *testing.T) {
	cfg := &ClientConfig{Host: "localhost"}
	incoming := &ClientConfig{Host: "127.0.0.1", Debug: true}

	cfg.Update(incoming)

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %s; want 127.0.0.1", cfg.Host)
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}
}

func TestThirdPartyIntegrationConfigUpdate(t *testing.T) {
	cfg := &ThirdPartyIntegrationConfig{}
	// ThirdPartyIntegrationConfig in settings.go uses 'GithubToken'
	incoming := &ThirdPartyIntegrationConfig{GithubToken: "gh-token"}

	cfg.Update(incoming)

	if cfg.GithubToken != "gh-token" {
		t.Errorf("GithubToken = %s; want gh-token", cfg.GithubToken)
	}
}

func TestSettingsEmpty(t *testing.T) {
	settings := &Settings{}

	if settings.LLMConfigs != nil {
		t.Error("LLMConfigs should be nil")
	}

	if settings.SearchConfig != nil {
		t.Error("SearchConfig should be nil")
	}
}

func TestLLMConfigFields(t *testing.T) {
	cfg := &LLMConfig{
		Model:    "gpt-4",
		APIKey:   "sk-test",
		BaseURL:  "https://api.openai.com",
	}

	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %s; want gpt-4", cfg.Model)
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("APIKey = %s; want sk-test", cfg.APIKey)
	}
}