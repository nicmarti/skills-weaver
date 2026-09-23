package agent

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// The Anthropic beta Advisor execution path was removed in Phase 5 of the
// OpenRouter migration (internal/agent/advisor.go): nested agents now run on
// the neutral llm.Client, and the native openrouter:advisor server tool lands
// in Phase 6 behind SW_ADVISOR_ENABLED. These model-mapping tests stay because
// model_mapping.go survives until Phase 12 (it is still referenced by legacy
// state files and the historical-metrics documentation).

func TestMapAdvisorModelToAnthropic(t *testing.T) {
	cases := []struct {
		in        string
		wantOK    bool
		wantModel anthropic.Model
	}{
		{"opus-4.7", true, anthropic.ModelClaudeOpus4_7},
		{"opus4.7", true, anthropic.ModelClaudeOpus4_7},
		{"opus-4-7", true, anthropic.ModelClaudeOpus4_7},
		{"opus", true, anthropic.ModelClaudeOpus4_7},
		{"OPUS-4.7", true, anthropic.ModelClaudeOpus4_7},
		{"  opus-4.7 ", true, anthropic.ModelClaudeOpus4_7},
		{"opus-4.8", true, anthropic.ModelClaudeOpus4_8},
		{"opus4.8", true, anthropic.ModelClaudeOpus4_8},
		{"opus-4-8", true, anthropic.ModelClaudeOpus4_8},
		{"", false, ""},
		{"sonnet", false, ""},
		{"haiku", false, ""},
		{"opus-4.6", false, ""},
	}
	for _, c := range cases {
		got, ok := MapAdvisorModelToAnthropic(c.in)
		if ok != c.wantOK {
			t.Errorf("MapAdvisorModelToAnthropic(%q) ok=%v, want %v", c.in, ok, c.wantOK)
		}
		if c.wantOK && got != c.wantModel {
			t.Errorf("MapAdvisorModelToAnthropic(%q) = %v, want %v", c.in, got, c.wantModel)
		}
	}
}

func TestIsValidAdvisorPair(t *testing.T) {
	valid := []struct{ exec, adv anthropic.Model }{
		{anthropic.ModelClaudeSonnet4_6, anthropic.ModelClaudeOpus4_7},
		{anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeOpus4_7},
		{anthropic.ModelClaudeOpus4_6, anthropic.ModelClaudeOpus4_7},
		{anthropic.ModelClaudeOpus4_7, anthropic.ModelClaudeOpus4_7},
		{anthropic.ModelClaudeOpus4_8, anthropic.ModelClaudeOpus4_8}, // 4.8 executor advised by 4.8
		{anthropic.ModelClaudeSonnet4_6, anthropic.ModelClaudeOpus4_8},
	}
	for _, p := range valid {
		if !IsValidAdvisorPair(p.exec, p.adv) {
			t.Errorf("IsValidAdvisorPair(%v, %v) = false, want true", p.exec, p.adv)
		}
	}

	invalid := []struct{ exec, adv anthropic.Model }{
		{anthropic.ModelClaudeSonnet4_6, anthropic.ModelClaudeSonnet4_6}, // advisor must be Opus 4.7/4.8
		{anthropic.ModelClaudeSonnet4_6, anthropic.ModelClaudeOpus4_6},   // 4.6 not a valid advisor
		{anthropic.ModelClaudeOpus4_8, anthropic.ModelClaudeOpus4_7},     // advisor less capable than executor
	}
	for _, p := range invalid {
		if IsValidAdvisorPair(p.exec, p.adv) {
			t.Errorf("IsValidAdvisorPair(%v, %v) = true, want false", p.exec, p.adv)
		}
	}
}