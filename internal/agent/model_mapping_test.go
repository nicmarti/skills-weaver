package agent

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestMapPersonaModelToAnthropic(t *testing.T) {
	tests := []struct {
		name        string
		personaModel string
		want        anthropic.Model
	}{
		{"sonnet lowercase", "sonnet", anthropic.ModelClaudeSonnet4_5},
		{"SONNET uppercase", "SONNET", anthropic.ModelClaudeSonnet4_5},
		{"Sonnet mixed case", "Sonnet", anthropic.ModelClaudeSonnet4_5},
		{"haiku lowercase", "haiku", anthropic.ModelClaudeHaiku4_5},
		{"HAIKU uppercase", "HAIKU", anthropic.ModelClaudeHaiku4_5},
		{"opus lowercase", "opus", anthropic.ModelClaudeOpus4_5},
		{"OPUS uppercase", "OPUS", anthropic.ModelClaudeOpus4_5},
		{"empty string defaults to sonnet", "", DefaultNestedAgentModel},
		{"unknown model defaults to sonnet", "gpt-4", DefaultNestedAgentModel},
		{"whitespace is trimmed", "  sonnet  ", anthropic.ModelClaudeSonnet4_5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapPersonaModelToAnthropic(tt.personaModel)
			if got != tt.want {
				t.Errorf("MapPersonaModelToAnthropic(%q) = %v, want %v", tt.personaModel, got, tt.want)
			}
		})
	}
}

func TestGetModelDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		model anthropic.Model
		want  string
	}{
		{"sonnet 4.5", anthropic.ModelClaudeSonnet4_5, "claude-sonnet-4-5"},
		{"sonnet 4.5 dated", anthropic.ModelClaudeSonnet4_5_20250929, "claude-sonnet-4-5"},
		{"haiku 4.5", anthropic.ModelClaudeHaiku4_5, "claude-haiku-4-5"},
		{"haiku 4.5 dated", anthropic.ModelClaudeHaiku4_5_20251001, "claude-haiku-4-5"},
		{"opus 4.5", anthropic.ModelClaudeOpus4_5, "claude-opus-4-5"},
		{"opus 4.5 dated", anthropic.ModelClaudeOpus4_5_20251101, "claude-opus-4-5"},
		{"unknown model returns string", anthropic.Model("claude-unknown"), "claude-unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetModelDisplayName(tt.model)
			if got != tt.want {
				t.Errorf("GetModelDisplayName(%v) = %q, want %q", tt.model, got, tt.want)
			}
		})
	}
}

func TestDefaultNestedAgentModel(t *testing.T) {
	// Verify default is Sonnet 4.5
	if DefaultNestedAgentModel != anthropic.ModelClaudeSonnet4_5 {
		t.Errorf("DefaultNestedAgentModel = %v, want %v",
			DefaultNestedAgentModel, anthropic.ModelClaudeSonnet4_5)
	}
}
