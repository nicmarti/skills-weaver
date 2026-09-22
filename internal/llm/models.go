package llm

import "strings"

// Concrete OpenRouter model IDs pinned per docs/openrouter-migration-plan.md.
// Mutable "~latest" aliases are intentionally not used in production so that
// game behavior stays reproducible across sessions.
const (
	ModelHaiku45 = "anthropic/claude-haiku-4.5"
	ModelSonnet5 = "anthropic/claude-sonnet-5"
	ModelOpus5   = "anthropic/claude-opus-5"
)

// Role defaults per the approved model catalog (2026-09-22):
//   - Main DM / nested agents / campaign generation: Sonnet 5
//   - Fast utility generation: Haiku 4.5
//   - World Keeper advisor: Opus 5
const (
	DefaultModel         = ModelSonnet5
	DefaultModelDM       = ModelSonnet5
	DefaultModelFast     = ModelHaiku45
	DefaultModelCampaign = ModelSonnet5
	DefaultModelAdvisor  = ModelOpus5
)

// ResolveModel maps persona/configuration aliases and full OpenRouter model
// IDs to a concrete model ID. Accepted aliases: "", "sonnet", "haiku", "opus".
// Any input containing "/" is treated as a full model ID and passed through
// unchanged. Unknown bare words fall back to the default (Sonnet 5), matching
// the previous persona-mapping behavior.
func ResolveModel(input string) string {
	trimmed := strings.TrimSpace(input)
	switch strings.ToLower(trimmed) {
	case "", "sonnet":
		return DefaultModel
	case "haiku":
		return ModelHaiku45
	case "opus":
		return ModelOpus5
	}
	if strings.Contains(trimmed, "/") {
		return trimmed
	}
	return DefaultModel
}

// DisplayName returns a human-readable label for a model ID.
func DisplayName(model string) string {
	switch model {
	case ModelHaiku45:
		return "Claude Haiku 4.5"
	case ModelSonnet5:
		return "Claude Sonnet 5"
	case ModelOpus5:
		return "Claude Opus 5"
	default:
		return model
	}
}

// ModelOption is one selectable entry for the web model selector.
type ModelOption struct {
	ID    string
	Label string
	Tier  string // "fast", "balanced", "premium"
}

// SelectableModels returns the models exposed to end users, in selector order.
func SelectableModels() []ModelOption {
	return []ModelOption{
		{ID: ModelHaiku45, Label: DisplayName(ModelHaiku45), Tier: "fast"},
		{ID: ModelSonnet5, Label: DisplayName(ModelSonnet5), Tier: "balanced"},
		{ID: ModelOpus5, Label: DisplayName(ModelOpus5), Tier: "premium"},
	}
}
