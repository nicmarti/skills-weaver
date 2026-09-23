package main

import (
	"testing"

	"dungeons/internal/llm"
)

func TestParseModelFlags_None(t *testing.T) {
	overrides, err := parseModelFlags([]string{})
	if err != nil {
		t.Fatalf("parseModelFlags: %v", err)
	}
	if overrides.DM != "" || overrides.Nested != "" || len(overrides.Agents) != 0 {
		t.Errorf("expected empty overrides, got %+v", overrides)
	}
}

func TestParseModelFlags_AllRoles(t *testing.T) {
	overrides, err := parseModelFlags([]string{
		"--model", "anthropic/claude-sonnet-5",
		"--nested-model", "haiku",
		"--agent-model", "rules-keeper=openai/gpt-5-mini",
		"--agent-model", "world-keeper=opus",
	})
	if err != nil {
		t.Fatalf("parseModelFlags: %v", err)
	}
	if overrides.DM != "anthropic/claude-sonnet-5" {
		t.Errorf("DM = %q", overrides.DM)
	}
	if overrides.Nested != "haiku" {
		t.Errorf("Nested = %q", overrides.Nested)
	}
	if overrides.Agents["rules-keeper"] != "openai/gpt-5-mini" {
		t.Errorf("rules-keeper = %q", overrides.Agents["rules-keeper"])
	}
	if overrides.Agents["world-keeper"] != "opus" {
		t.Errorf("world-keeper = %q", overrides.Agents["world-keeper"])
	}
	if _, ok := overrides.Agents["character-creator"]; ok {
		t.Error("character-creator should not be set")
	}
}

func TestParseModelFlags_RejectInvalid(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"empty model", []string{"--model", "  "}},
		{"empty nested", []string{"--nested-model", ""}},
		{"unknown agent", []string{"--agent-model", "bard=sonnet"}},
		{"missing name", []string{"--agent-model", "=sonnet"}},
		{"missing model", []string{"--agent-model", "rules-keeper="}},
		{"missing separator", []string{"--agent-model", "rules-keepersonnet"}},
		{"unknown flag", []string{"--temperature", "0.5"}},
		{"positional arg", []string{"adventure"}},
	}
	for _, tc := range cases {
		if _, err := parseModelFlags(tc.args); err == nil {
			t.Errorf("%s: expected error, got none", tc.name)
		}
	}
}

func TestParseModelFlags_OverridesApplyAfterEnv(t *testing.T) {
	overrides, err := parseModelFlags([]string{"--model", "sonnet"})
	if err != nil {
		t.Fatalf("parseModelFlags: %v", err)
	}
	cfg, err := llm.DefaultModels().WithModelOverrides(overrides)
	if err != nil {
		t.Fatalf("WithModelOverrides: %v", err)
	}
	if cfg.ModelDM != llm.ModelSonnet5 {
		t.Errorf("ModelDM = %q, want %q", cfg.ModelDM, llm.ModelSonnet5)
	}
	// Aliases resolve and invalid IDs are rejected rather than defaulted.
	cfg, err = llm.DefaultModels().WithModelOverrides(llm.ModelOverrides{DM: "not-a-model"})
	if err == nil {
		t.Error("invalid model ID should be rejected, not defaulted")
	}
	if cfg.ModelDM != "" {
		t.Errorf("failed override must not produce a config, got %q", cfg.ModelDM)
	}
}