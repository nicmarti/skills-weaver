// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// DefaultNestedAgentModel is the model used for nested agents when no model is specified in persona.
// Using Sonnet 4.5 as the default for nested agents since they need to handle complex reasoning.
var DefaultNestedAgentModel = anthropic.ModelClaudeSonnet4_5

// MapPersonaModelToAnthropic converts a persona model string to an Anthropic SDK model constant.
// Supported values: "sonnet", "haiku", "opus" (case-insensitive).
// Defaults to Sonnet 4.5 if the model string is not recognized.
func MapPersonaModelToAnthropic(personaModel string) anthropic.Model {
	switch strings.ToLower(strings.TrimSpace(personaModel)) {
	case "sonnet":
		return anthropic.ModelClaudeSonnet4_5
	case "haiku":
		return anthropic.ModelClaudeHaiku4_5
	case "opus":
		return anthropic.ModelClaudeOpus4_5
	case "":
		// Empty string = use default
		return DefaultNestedAgentModel
	default:
		// Unknown model = use default
		return DefaultNestedAgentModel
	}
}

// GetModelDisplayName returns a human-readable name for an Anthropic model.
func GetModelDisplayName(model anthropic.Model) string {
	switch model {
	case anthropic.ModelClaudeSonnet4_5, anthropic.ModelClaudeSonnet4_5_20250929:
		return "claude-sonnet-4-5"
	case anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeHaiku4_5_20251001:
		return "claude-haiku-4-5"
	case anthropic.ModelClaudeOpus4_5, anthropic.ModelClaudeOpus4_5_20251101:
		return "claude-opus-4-5"
	default:
		return string(model)
	}
}
