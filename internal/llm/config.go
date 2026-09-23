package llm

import (
	"fmt"
	"os"
	"strings"
)

// Environment variables read by LoadConfig. Only the API key is required.
const (
	EnvAPIKey           = "OPENROUTER_API_KEY"
	EnvModelDM          = "OPENROUTER_MODEL_DM"
	EnvModelNested      = "OPENROUTER_MODEL_NESTED"
	EnvRulesKeeper      = "OPENROUTER_MODEL_RULES_KEEPER"
	EnvCharacterCreator = "OPENROUTER_MODEL_CHARACTER_CREATOR"
	EnvWorldKeeper      = "OPENROUTER_MODEL_WORLD_KEEPER"
	EnvModelFast        = "OPENROUTER_MODEL_FAST"
	EnvModelCampaign    = "OPENROUTER_MODEL_CAMPAIGN"
	EnvHTTPReferer      = "OPENROUTER_HTTP_REFERER"
	EnvAppName          = "OPENROUTER_APP_NAME"
)

// Config holds the process-wide OpenRouter configuration. One shared client is
// built from it per process; no other package reads provider environment
// variables directly.
type Config struct {
	APIKey string

	// Model overrides accept aliases ("sonnet", "haiku", "opus") or full
	// OpenRouter model IDs. Empty selects the plan default.
	ModelDM     string
	ModelNested string
	// AgentModels contains named overrides for the nested agents only.
	AgentModels   map[string]string
	ModelFast     string
	ModelCampaign string

	// Optional OpenRouter attribution metadata.
	HTTPReferer string
	AppName     string
}

// LoadConfig builds the process-wide configuration from the environment.
// ANTHROPIC_API_KEY is intentionally NOT accepted as a fallback: an Anthropic
// key cannot authenticate against OpenRouter.
func LoadConfig() (Config, error) {
	cfg := Config{
		APIKey:      strings.TrimSpace(os.Getenv(EnvAPIKey)),
		ModelDM:     os.Getenv(EnvModelDM),
		ModelNested: os.Getenv(EnvModelNested),
		AgentModels: map[string]string{
			RulesKeeperAgent:      os.Getenv(EnvRulesKeeper),
			CharacterCreatorAgent: os.Getenv(EnvCharacterCreator),
			WorldKeeperAgent:      os.Getenv(EnvWorldKeeper),
		},
		ModelFast:     os.Getenv(EnvModelFast),
		ModelCampaign: os.Getenv(EnvModelCampaign),
		HTTPReferer:   strings.TrimSpace(os.Getenv(EnvHTTPReferer)),
		AppName:       strings.TrimSpace(os.Getenv(EnvAppName)),
	}

	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("%s environment variable not set (set it in your .envrc file or export it)", EnvAPIKey)
	}

	return cfg.normalizeModels()
}

// resolveOrDefault validates explicit overrides instead of silently replacing
// mistyped IDs with Sonnet. An unset override selects the role default.
func resolveOrDefault(input, def, label string) (string, error) {
	if input == "" {
		return def, nil
	}
	model, err := ParseModelChoice(input)
	if err != nil {
		return "", fmt.Errorf("%s: %w", label, err)
	}
	return model, nil
}

// ModelOverrides are future sw-dm flag values, applied after environment
// variables. The CLI cannot use them for live gameplay until Phases 4/5 wire
// both the DM and nested agents to llm.Client.
type ModelOverrides struct {
	DM     string
	Nested string
	Agents map[string]string
}

// WithModelOverrides gives explicit CLI choices precedence over environment
// values without altering the source config or falling back on invalid IDs.
func (cfg Config) WithModelOverrides(overrides ModelOverrides) (Config, error) {
	if overrides.DM != "" {
		cfg.ModelDM = overrides.DM
	}
	agents := make(map[string]string, len(cfg.AgentModels)+len(overrides.Agents))
	if overrides.Nested != "" {
		// An explicit CLI nested default overrides per-agent environment values;
		// a more specific CLI --agent-model can override it again below.
		cfg.ModelNested = overrides.Nested
	} else {
		for name, model := range cfg.AgentModels {
			agents[name] = model
		}
	}
	for name, model := range overrides.Agents {
		if !IsNestedAgent(name) || strings.TrimSpace(model) == "" {
			return Config{}, fmt.Errorf("invalid model override for agent %q", name)
		}
		agents[name] = model
	}
	cfg.AgentModels = agents
	return cfg.normalizeModels()
}

func (cfg Config) normalizeModels() (Config, error) {
	var err error
	for _, role := range []struct {
		value *string
		def   string
		name  string
	}{
		{&cfg.ModelDM, DefaultModelDM, EnvModelDM},
		{&cfg.ModelNested, DefaultModelNested, EnvModelNested},
		{&cfg.ModelFast, DefaultModelFast, EnvModelFast},
		{&cfg.ModelCampaign, DefaultModelCampaign, EnvModelCampaign},
	} {
		*role.value, err = resolveOrDefault(*role.value, role.def, role.name)
		if err != nil {
			return Config{}, err
		}
	}
	if cfg.AgentModels == nil {
		cfg.AgentModels = make(map[string]string)
	}
	for name := range cfg.AgentModels {
		if !IsNestedAgent(name) {
			return Config{}, fmt.Errorf("unknown nested agent %q", name)
		}
	}
	for _, name := range nestedAgentNames {
		if raw := cfg.AgentModels[name]; raw != "" {
			cfg.AgentModels[name], err = resolveOrDefault(raw, DefaultModelNested, name)
			if err != nil {
				return Config{}, err
			}
		} else {
			delete(cfg.AgentModels, name)
		}
	}
	return cfg, nil
}

// ModelForAgent resolves a nested agent's model after applying role defaults.
func (cfg Config) ModelForAgent(name string) (string, error) {
	if !IsNestedAgent(name) {
		return "", fmt.Errorf("unknown nested agent %q", name)
	}
	if model := cfg.AgentModels[name]; model != "" {
		return ParseModelChoice(model)
	}
	return resolveOrDefault(cfg.ModelNested, DefaultModelNested, EnvModelNested)
}

// DefaultModels returns the plan's concrete production defaults.
func DefaultModels() Config {
	return Config{
		ModelDM:       DefaultModelDM,
		ModelNested:   DefaultModelNested,
		AgentModels:   make(map[string]string),
		ModelFast:     DefaultModelFast,
		ModelCampaign: DefaultModelCampaign,
	}
}

// LegacyAnthropicKey returns the pre-migration key for runtime paths that have
// not been migrated to llm.Client yet (nested agents until Phase 5, utilities
// until Phase 7). It must disappear with Phase 12.
func (cfg Config) LegacyAnthropicKey() string {
	return strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
}
