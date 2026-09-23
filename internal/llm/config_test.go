package llm

import (
	"testing"
)

func clearModelEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{EnvModelDM, EnvModelNested, EnvRulesKeeper, EnvCharacterCreator, EnvWorldKeeper, EnvModelFast, EnvModelCampaign, EnvModelAdvisor} {
		t.Setenv(name, "")
	}
}

func TestLoadConfigRequiresAPIKey(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() should fail without an API key")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearModelEnv(t)
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
	if cfg.ModelNested != DefaultModelNested {
		t.Errorf("ModelNested = %q, want %q", cfg.ModelNested, DefaultModelNested)
	}
	for _, name := range nestedAgentNames {
		model, err := cfg.ModelForAgent(name)
		if err != nil || model != ModelSonnet5 {
			t.Errorf("ModelForAgent(%q) = %q, %v", name, model, err)
		}
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
	clearModelEnv(t)
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

func TestLoadConfigNestedAgentEnvironmentOverrides(t *testing.T) {
	clearModelEnv(t)
	t.Setenv(EnvAPIKey, "or-test-key")
	t.Setenv(EnvModelNested, "openai/gpt-6")
	t.Setenv(EnvRulesKeeper, "google/gemini-3-pro")
	t.Setenv(EnvWorldKeeper, "sonnet")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		RulesKeeperAgent: "google/gemini-3-pro", CharacterCreatorAgent: "openai/gpt-6", WorldKeeperAgent: ModelSonnet5,
	} {
		got, err := cfg.ModelForAgent(name)
		if err != nil || got != want {
			t.Errorf("%s model = %q, %v; want %q", name, got, err, want)
		}
	}
	if _, err := cfg.ModelForAgent("unknown-agent"); err == nil {
		t.Fatal("unknown agents must not inherit the default silently")
	}
}

func TestModelOverridesCLIBeatsEnvironment(t *testing.T) {
	clearModelEnv(t)
	t.Setenv(EnvAPIKey, "or-test-key")
	t.Setenv(EnvModelDM, "sonnet")
	t.Setenv(EnvModelNested, "openai/gpt-6")
	t.Setenv(EnvRulesKeeper, "google/gemini-3-pro")
	t.Setenv(EnvWorldKeeper, "google/gemini-3-pro")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	updated, err := cfg.WithModelOverrides(ModelOverrides{
		DM: "x-ai/grok-4.6", Nested: "anthropic/claude-haiku-4.5",
		Agents: map[string]string{RulesKeeperAgent: "openai/gpt-6"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ModelDM != "x-ai/grok-4.6" {
		t.Errorf("CLI DM model = %q", updated.ModelDM)
	}
	if model, _ := updated.ModelForAgent(RulesKeeperAgent); model != "openai/gpt-6" {
		t.Errorf("CLI agent model = %q", model)
	}
	if model, _ := updated.ModelForAgent(CharacterCreatorAgent); model != ModelHaiku45 {
		t.Errorf("CLI nested default = %q", model)
	}
	if model, _ := updated.ModelForAgent(WorldKeeperAgent); model != ModelHaiku45 {
		t.Errorf("CLI nested default did not replace per-agent environment override: %q", model)
	}
	if model, _ := cfg.ModelForAgent(RulesKeeperAgent); model != "google/gemini-3-pro" {
		t.Errorf("original config was mutated: %q", model)
	}
}

func TestExplicitInvalidModelOverridesFail(t *testing.T) {
	clearModelEnv(t)
	t.Setenv(EnvAPIKey, "or-test-key")
	t.Setenv(EnvRulesKeeper, "gpt-6")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("mistyped bare agent model must not become Sonnet")
	}
	t.Setenv(EnvRulesKeeper, "  ")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("whitespace-only agent model must not become Sonnet")
	}
	t.Setenv(EnvRulesKeeper, "")
	t.Setenv(EnvModelDM, " ")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("whitespace-only DM model must not become Sonnet")
	}
	t.Setenv(EnvModelDM, "")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	for _, override := range []ModelOverrides{
		{DM: "invalid-model"},
		{DM: " "},
		{Agents: map[string]string{"scenario-critic": "openai/gpt-6"}},
		{Agents: map[string]string{RulesKeeperAgent: ""}},
	} {
		if _, err := cfg.WithModelOverrides(override); err == nil {
			t.Fatalf("invalid override %+v should fail", override)
		}
	}
}

func TestNewClientRequiresKey(t *testing.T) {
	if _, err := NewOpenRouterClient(Config{}); err == nil {
		t.Fatal("NewOpenRouterClient without API key should fail")
	}
}
