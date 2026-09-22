package llm

import (
	"fmt"
	"os"
	"strings"
)

// Environment variables read by LoadConfig. Only the API key is required.
const (
	EnvAPIKey        = "OPENROUTER_API_KEY"
	EnvModelDM       = "OPENROUTER_MODEL_DM"
	EnvModelFast     = "OPENROUTER_MODEL_FAST"
	EnvModelCampaign = "OPENROUTER_MODEL_CAMPAIGN"
	EnvModelAdvisor  = "OPENROUTER_MODEL_ADVISOR"
	EnvHTTPReferer   = "OPENROUTER_HTTP_REFERER"
	EnvAppName       = "OPENROUTER_APP_NAME"
)

// Config holds the process-wide OpenRouter configuration. One shared client is
// built from it per process; no other package reads provider environment
// variables directly.
type Config struct {
	APIKey string

	// Model overrides accept aliases ("sonnet", "haiku", "opus") or full
	// OpenRouter model IDs. Empty selects the plan default.
	ModelDM       string
	ModelFast     string
	ModelCampaign string
	ModelAdvisor  string

	// Optional OpenRouter attribution metadata.
	HTTPReferer string
	AppName     string
}

// LoadConfig builds the process-wide configuration from the environment.
// ANTHROPIC_API_KEY is intentionally NOT accepted as a fallback: an Anthropic
// key cannot authenticate against OpenRouter.
func LoadConfig() (Config, error) {
	cfg := Config{
		APIKey:        strings.TrimSpace(os.Getenv(EnvAPIKey)),
		ModelDM:       os.Getenv(EnvModelDM),
		ModelFast:     os.Getenv(EnvModelFast),
		ModelCampaign: os.Getenv(EnvModelCampaign),
		ModelAdvisor:  os.Getenv(EnvModelAdvisor),
		HTTPReferer:   strings.TrimSpace(os.Getenv(EnvHTTPReferer)),
		AppName:       strings.TrimSpace(os.Getenv(EnvAppName)),
	}

	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("%s environment variable not set (set it in your .envrc file or export it)", EnvAPIKey)
	}

	cfg.ModelDM = resolveOrDefault(cfg.ModelDM, DefaultModelDM)
	cfg.ModelFast = resolveOrDefault(cfg.ModelFast, DefaultModelFast)
	cfg.ModelCampaign = resolveOrDefault(cfg.ModelCampaign, DefaultModelCampaign)
	cfg.ModelAdvisor = resolveOrDefault(cfg.ModelAdvisor, DefaultModelAdvisor)

	return cfg, nil
}

// resolveOrDefault maps an override to a concrete model ID, falling back to
// the role default when no override is configured.
func resolveOrDefault(input, def string) string {
	if strings.TrimSpace(input) == "" {
		return def
	}
	return ResolveModel(input)
}

// DefaultModels returns the plan's concrete production defaults.
func DefaultModels() Config {
	return Config{
		ModelDM:       DefaultModelDM,
		ModelFast:     DefaultModelFast,
		ModelCampaign: DefaultModelCampaign,
		ModelAdvisor:  DefaultModelAdvisor,
	}
}
