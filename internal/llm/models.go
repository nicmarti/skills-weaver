package llm

import (
	"fmt"
	"strings"
)

// Concrete OpenRouter model IDs pinned per docs/openrouter-migration-plan.md.
// Mutable "~latest" aliases are not used as defaults so game behavior stays
// reproducible; explicit user model overrides may deliberately select them.
const (
	ModelHaiku45 = "anthropic/claude-haiku-4.5"
	ModelSonnet5 = "anthropic/claude-sonnet-5"
	ModelOpus5   = "anthropic/claude-opus-5"
)

// OpenRouter defaults: Sonnet 5 for every role. Alternative models remain
// selectable explicitly; these defaults do not affect the legacy Anthropic
// runtime until the agent loops are migrated to llm.Client.
const (
	DefaultModel         = ModelSonnet5
	DefaultModelDM       = ModelSonnet5
	DefaultModelNested   = ModelSonnet5
	DefaultModelFast     = ModelSonnet5
	DefaultModelCampaign = ModelSonnet5
	DefaultModelAdvisor  = ModelSonnet5
)

const (
	RulesKeeperAgent      = "rules-keeper"
	CharacterCreatorAgent = "character-creator"
	WorldKeeperAgent      = "world-keeper"
)

var nestedAgentNames = []string{RulesKeeperAgent, CharacterCreatorAgent, WorldKeeperAgent}

// IsNestedAgent recognizes the three actual nested-agent personas.
func IsNestedAgent(name string) bool {
	switch name {
	case RulesKeeperAgent, CharacterCreatorAgent, WorldKeeperAgent:
		return true
	default:
		return false
	}
}

// ParseModelChoice validates an explicit configuration/CLI model selection.
// Aliases remain convenient, but provider-qualified OpenRouter IDs are not
// restricted to Claude. Catalog availability and model capabilities are
// checked at request time, not guessed from the ID here.
func ParseModelChoice(input string) (string, error) {
	model := strings.TrimSpace(input)
	switch strings.ToLower(model) {
	case "sonnet":
		return ModelSonnet5, nil
	case "haiku":
		return ModelHaiku45, nil
	case "opus":
		return ModelOpus5, nil
	}
	if model == "" || strings.ContainsAny(model, " \t\r\n") {
		return "", fmt.Errorf("invalid OpenRouter model %q: expected a provider/model ID or sonnet, haiku, opus", input)
	}
	id := strings.TrimPrefix(model, "~") // Allow deliberate mutable ~latest overrides.
	provider, slug, ok := strings.Cut(id, "/")
	if !ok || provider == "" || slug == "" || strings.HasPrefix(slug, "/") || strings.HasSuffix(slug, "/") {
		return "", fmt.Errorf("invalid OpenRouter model %q: expected a provider/model ID or sonnet, haiku, opus", input)
	}
	return model, nil
}

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
