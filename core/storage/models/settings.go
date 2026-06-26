package models

// ConfigUpdateable is an interface that allows sub-configurations to update themselves.
type ConfigUpdateable[T any] interface {
	Update(newConfig *T)
}

// Settings represents the persisted settings for Water AI sessions.
type Settings struct {
	LLMConfigs                  map[string]*LLMConfig        `json:"llm_configs,omitempty"`
	SearchConfig                *SearchConfig                `json:"search_config,omitempty"`
	MediaConfig                 *MediaConfig                 `json:"media_config,omitempty"`
	AudioConfig                 *AudioConfig                 `json:"audio_config,omitempty"`
	SandboxConfig               *SandboxConfig               `json:"sandbox_config,omitempty"`
	ClientConfig                *ClientConfig                `json:"client_config,omitempty"`
	ThirdPartyIntegrationConfig *ThirdPartyIntegrationConfig `json:"third_party_integration_config,omitempty"`
}

// Update merges the provided settings into the current settings object.
// It specifically handles logic for merging LLM API keys and delegate updates
// to sub-configuration objects.
func (s *Settings) Update(incoming *Settings) {
	if incoming == nil {
		return
	}

	// 1. Merge LLM Configs
	if len(s.LLMConfigs) > 0 && len(incoming.LLMConfigs) > 0 {
		for name, currentCfg := range s.LLMConfigs {
			if incomingCfg, exists := incoming.LLMConfigs[name]; exists {
				// Python Logic: if self.api_key is None and new.api_key exists, copy it.
				if currentCfg.APIKey == "" && incomingCfg.APIKey != "" {
					currentCfg.APIKey = incomingCfg.APIKey
				}
				// Merge other potential fields
				currentCfg.Update(incomingCfg)
			}
		}
	} else if len(s.LLMConfigs) == 0 && len(incoming.LLMConfigs) > 0 {
		// If current is empty, just take the incoming map
		s.LLMConfigs = make(map[string]*LLMConfig)
		for k, v := range incoming.LLMConfigs {
			// Deep copy if necessary, here we just assign pointer
			s.LLMConfigs[k] = v
		}
	}

	// 2. Update all config attributes using helper logic
	// In Go, we explicitly check and update each field instead of using reflection/getattr
	// to maintain type safety and performance.

	if s.SearchConfig == nil {
		s.SearchConfig = incoming.SearchConfig
	} else if incoming.SearchConfig != nil {
		s.SearchConfig.Update(incoming.SearchConfig)
	}

	if s.MediaConfig == nil {
		s.MediaConfig = incoming.MediaConfig
	} else if incoming.MediaConfig != nil {
		s.MediaConfig.Update(incoming.MediaConfig)
	}

	if s.AudioConfig == nil {
		s.AudioConfig = incoming.AudioConfig
	} else if incoming.AudioConfig != nil {
		s.AudioConfig.Update(incoming.AudioConfig)
	}

	if s.SandboxConfig == nil {
		s.SandboxConfig = incoming.SandboxConfig
	} else if incoming.SandboxConfig != nil {
		s.SandboxConfig.Update(incoming.SandboxConfig)
	}

	if s.ClientConfig == nil {
		s.ClientConfig = incoming.ClientConfig
	} else if incoming.ClientConfig != nil {
		s.ClientConfig.Update(incoming.ClientConfig)
	}

	if s.ThirdPartyIntegrationConfig == nil {
		s.ThirdPartyIntegrationConfig = incoming.ThirdPartyIntegrationConfig
	} else if incoming.ThirdPartyIntegrationConfig != nil {
		s.ThirdPartyIntegrationConfig.Update(incoming.ThirdPartyIntegrationConfig)
	}
}

// -----------------------------------------------------------------------------
// Sub-Configuration Structs
// Since the original Python code imported these, we define reasonable defaults
// and the required Update() methods here to keep it in one file.
// -----------------------------------------------------------------------------

type LLMConfig struct {
	Model    string `json:"model,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func (c *LLMConfig) Update(other *LLMConfig) {
	if other.Model != "" {
		c.Model = other.Model
	}
	if other.APIKey != "" {
		c.APIKey = other.APIKey
	}
	if other.BaseURL != "" {
		c.BaseURL = other.BaseURL
	}
	if other.Provider != "" {
		c.Provider = other.Provider
	}
}

type SearchConfig struct {
	Provider string `json:"provider,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	MaxRes   int    `json:"max_results,omitempty"`
}

func (c *SearchConfig) Update(other *SearchConfig) {
	if other.Provider != "" {
		c.Provider = other.Provider
	}
	if other.APIKey != "" {
		c.APIKey = other.APIKey
	}
	if other.MaxRes != 0 {
		c.MaxRes = other.MaxRes
	}
}

type MediaConfig struct {
	OutputDir string `json:"output_dir,omitempty"`
	Format    string `json:"format,omitempty"`
}

func (c *MediaConfig) Update(other *MediaConfig) {
	if other.OutputDir != "" {
		c.OutputDir = other.OutputDir
	}
	if other.Format != "" {
		c.Format = other.Format
	}
}

type AudioConfig struct {
	VoiceID       string  `json:"voice_id,omitempty"`
	Volume        float64 `json:"volume,omitempty"`
	IsVolumeSet   bool    `json:"-"` // Helper to track if volume was explicitly set 0.0
}

func (c *AudioConfig) Update(other *AudioConfig) {
	if other.VoiceID != "" {
		c.VoiceID = other.VoiceID
	}
	// Simple float check (assuming non-zero means update, or use IsVolumeSet flag logic)
	if other.Volume != 0 || other.IsVolumeSet {
		c.Volume = other.Volume
	}
}

type SandboxConfig struct {
	Timeout    int    `json:"timeout,omitempty"`
	DockerImage string `json:"docker_image,omitempty"`
}

func (c *SandboxConfig) Update(other *SandboxConfig) {
	if other.Timeout != 0 {
		c.Timeout = other.Timeout
	}
	if other.DockerImage != "" {
		c.DockerImage = other.DockerImage
	}
}

type ClientConfig struct {
	Host  string `json:"host,omitempty"`
	Debug bool   `json:"debug,omitempty"`
}

func (c *ClientConfig) Update(other *ClientConfig) {
	if other.Host != "" {
		c.Host = other.Host
	}
	c.Debug = other.Debug // Booleans are tricky to merge without pointer or flag, assume overwrite
}

type ThirdPartyIntegrationConfig struct {
	GithubToken  string `json:"github_token,omitempty"`
	SlackToken   string `json:"slack_token,omitempty"`
	JiraAPIToken string `json:"jira_api_token,omitempty"`
}

func (c *ThirdPartyIntegrationConfig) Update(other *ThirdPartyIntegrationConfig) {
	if other.GithubToken != "" {
		c.GithubToken = other.GithubToken
	}
	if other.SlackToken != "" {
		c.SlackToken = other.SlackToken
	}
	if other.JiraAPIToken != "" {
		c.JiraAPIToken = other.JiraAPIToken
	}
}