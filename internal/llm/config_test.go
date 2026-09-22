package llm

import (
	"testing"
)

func TestLoadConfigRequiresAPIKey(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() should fail without an API key")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv(EnvAPIKey, "or-test-key")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.APIKey != "or-test-key" {
		t.Errorf("APIKey = %q", cfg.APIKey)
	}
	if cfg.ModelDM != DefaultModelDM {
		t.Errorf("ModelDM = %q, want %q", cfg.ModelDM, DefaultModelDM)
	}
	if cfg.ModelFast != DefaultModelFast {
		t.Errorf("ModelFast = %q, want %q", cfg.ModelFast, DefaultModelFast)
	}
	if cfg.ModelCampaign != DefaultModelCampaign {
		t.Errorf("ModelCampaign = %q, want %q", cfg.ModelCampaign, DefaultModelCampaign)
	}
	if cfg.ModelAdvisor != DefaultModelAdvisor {
		t.Errorf("ModelAdvisor = %q, want %q", cfg.ModelAdvisor, DefaultModelAdvisor)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv(EnvAPIKey, "or-test-key")
	t.Setenv(EnvModelDM, "opus")
	t.Setenv(EnvModelFast, "anthropic/claude-haiku-4.5")
	t.Setenv(EnvModelAdvisor, "opus")
	t.Setenv(EnvHTTPReferer, "https://example.org")
	t.Setenv(EnvAppName, "SkillsWeaver")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.ModelDM != ModelOpus5 {
		t.Errorf("ModelDM = %q, want alias resolved to %q", cfg.ModelDM, ModelOpus5)
	}
	if cfg.ModelFast != ModelHaiku45 {
		t.Errorf("ModelFast = %q", cfg.ModelFast)
	}
	if cfg.ModelAdvisor != ModelOpus5 {
		t.Errorf("ModelAdvisor = %q", cfg.ModelAdvisor)
	}
	if cfg.HTTPReferer != "https://example.org" {
		t.Errorf("HTTPReferer = %q", cfg.HTTPReferer)
	}
	if cfg.AppName != "SkillsWeaver" {
		t.Errorf("AppName = %q", cfg.AppName)
	}
}

func TestNewClientRequiresKey(t *testing.T) {
	if _, err := NewOpenRouterClient(Config{}); err == nil {
		t.Fatal("NewOpenRouterClient without API key should fail")
	}
}
