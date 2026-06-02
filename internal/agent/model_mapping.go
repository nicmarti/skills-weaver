// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// DefaultNestedAgentModel is the model used for nested agents when no model is specified in persona.
// Using Sonnet 4.6 as the default for nested agents since they need to handle complex reasoning.
var DefaultNestedAgentModel = anthropic.ModelClaudeSonnet4_6

// MapPersonaModelToAnthropic converts a persona model string to an Anthropic SDK model constant.
// Supported values: "sonnet", "haiku", "opus" (case-insensitive).
// Defaults to Sonnet 4.6 if the model string is not recognized.
func MapPersonaModelToAnthropic(personaModel string) anthropic.Model {
	switch strings.ToLower(strings.TrimSpace(personaModel)) {
	case "sonnet":
		return anthropic.ModelClaudeSonnet4_6
	case "haiku":
		return anthropic.ModelClaudeHaiku4_5
	case "opus":
		return anthropic.ModelClaudeOpus4_6
	case "":
		// Empty string = use default
		return DefaultNestedAgentModel
	default:
		// Unknown model = use default
		return DefaultNestedAgentModel
	}
}

// MapAdvisorModelToAnthropic converts a persona "advisor" model string to an
// Anthropic SDK model constant suitable for the Advisor tool.
//
// The Advisor tool (beta) requires the advisor to be at least as capable as the
// executor. With the current SDK only Opus 4.7 is available as an advisor model.
// Supported values (case-insensitive): "opus-4.7", "opus4.7", "opus".
// Returns ok=false if the string is empty or not a recognized advisor model.
func MapAdvisorModelToAnthropic(advisorModel string) (anthropic.Model, bool) {
	switch strings.ToLower(strings.TrimSpace(advisorModel)) {
	case "opus-4.7", "opus4.7", "opus-4-7", "opus":
		return anthropic.ModelClaudeOpus4_7, true
	default:
		return "", false
	}
}

// IsValidAdvisorPair reports whether the given executor/advisor models form a
// valid pair for the Advisor tool. The advisor must be at least as capable as
// the executor. Per the Advisor tool docs, valid executors are Haiku 4.5,
// Sonnet 4.6, and Opus 4.6/4.7; the advisor must be Opus 4.7 (or 4.8, not yet
// in this SDK).
func IsValidAdvisorPair(executor, advisor anthropic.Model) bool {
	advisorOK := advisor == anthropic.ModelClaudeOpus4_7
	if !advisorOK {
		return false
	}
	switch executor {
	case anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeHaiku4_5_20251001,
		anthropic.ModelClaudeSonnet4_6,
		anthropic.ModelClaudeOpus4_6, anthropic.ModelClaudeOpus4_7:
		return true
	default:
		return false
	}
}

// GetModelDisplayName returns a human-readable name for an Anthropic model.
func GetModelDisplayName(model anthropic.Model) string {
	switch model {
	case anthropic.ModelClaudeSonnet4_6:
		return "claude-sonnet-4-6"
	case anthropic.ModelClaudeSonnet4_5, anthropic.ModelClaudeSonnet4_5_20250929:
		return "claude-sonnet-4-5"
	case anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeHaiku4_5_20251001:
		return "claude-haiku-4-5"
	case anthropic.ModelClaudeOpus4_6:
		return "claude-opus-4-6"
	case anthropic.ModelClaudeOpus4_7:
		return "claude-opus-4-7"
	case anthropic.ModelClaudeOpus4_5, anthropic.ModelClaudeOpus4_5_20251101:
		return "claude-opus-4-5"
	default:
		return string(model)
	}
}
