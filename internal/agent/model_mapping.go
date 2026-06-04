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
		return anthropic.ModelClaudeOpus4_8
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
// executor. The current SDK exposes Opus 4.7 and Opus 4.8 as advisor models.
// Supported values (case-insensitive): "opus-4.8"/"opus4.8" → Opus 4.8;
// "opus-4.7"/"opus4.7"/"opus" → Opus 4.7 (kept as the default for backward
// compatibility with personas that simply declare "opus").
// Returns ok=false if the string is empty or not a recognized advisor model.
func MapAdvisorModelToAnthropic(advisorModel string) (anthropic.Model, bool) {
	switch strings.ToLower(strings.TrimSpace(advisorModel)) {
	case "opus-4.8", "opus4.8", "opus-4-8":
		return anthropic.ModelClaudeOpus4_8, true
	case "opus-4.7", "opus4.7", "opus-4-7", "opus":
		return anthropic.ModelClaudeOpus4_7, true
	default:
		return "", false
	}
}

// IsValidAdvisorPair reports whether the given executor/advisor models form a
// valid pair for the Advisor tool. The advisor must be at least as capable as
// the executor. Per the Advisor tool docs, valid executors are Haiku 4.5,
// Sonnet 4.6, and Opus 4.6/4.7/4.8; the advisor must be Opus 4.7 or Opus 4.8.
func IsValidAdvisorPair(executor, advisor anthropic.Model) bool {
	advisorOK := advisor == anthropic.ModelClaudeOpus4_7 || advisor == anthropic.ModelClaudeOpus4_8
	if !advisorOK {
		return false
	}
	// Advisor must be at least as capable as the executor: an Opus 4.8 executor
	// can only be advised by Opus 4.8, not by the less-capable Opus 4.7.
	if executor == anthropic.ModelClaudeOpus4_8 && advisor != anthropic.ModelClaudeOpus4_8 {
		return false
	}
	switch executor {
	case anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeHaiku4_5_20251001,
		anthropic.ModelClaudeSonnet4_6,
		anthropic.ModelClaudeOpus4_6, anthropic.ModelClaudeOpus4_7, anthropic.ModelClaudeOpus4_8:
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
	case anthropic.ModelClaudeOpus4_8:
		return "claude-opus-4-8"
	case anthropic.ModelClaudeOpus4_7:
		return "claude-opus-4-7"
	case anthropic.ModelClaudeOpus4_5, anthropic.ModelClaudeOpus4_5_20251101:
		return "claude-opus-4-5"
	default:
		return string(model)
	}
}
