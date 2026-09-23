package llm

import "testing"

func TestResolveModel(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty defaults to sonnet", "", ModelSonnet5},
		{"sonnet alias", "sonnet", ModelSonnet5},
		{"SONNET uppercase", "SONNET", ModelSonnet5},
		{"whitespace trimmed", "  sonnet  ", ModelSonnet5},
		{"haiku alias", "haiku", ModelHaiku45},
		{"HAIKU uppercase", "HAIKU", ModelHaiku45},
		{"opus alias", "opus", ModelOpus5},
		{"full sonnet id passes through", "anthropic/claude-sonnet-5", ModelSonnet5},
		{"full opus id passes through", "anthropic/claude-opus-5", ModelOpus5},
		{"full haiku id passes through", "anthropic/claude-haiku-4.5", ModelHaiku45},
		{"other provider id passes through", "openai/gpt-5.2", "openai/gpt-5.2"},
		{"unknown bare word defaults", "gpt-4", DefaultModel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveModel(tc.input); got != tc.want {
				t.Errorf("ResolveModel(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseModelChoice(t *testing.T) {
	for input, want := range map[string]string{
		"sonnet":                          ModelSonnet5,
		"opus":                            ModelOpus5,
		"haiku":                           ModelHaiku45,
		"openai/gpt-6":                    "openai/gpt-6",
		"google/gemini-3-pro":             "google/gemini-3-pro",
		"openrouter/auto":                 "openrouter/auto",
		"~anthropic/claude-sonnet-latest": "~anthropic/claude-sonnet-latest",
		"openai/gpt-6:nitro":              "openai/gpt-6:nitro",
	} {
		got, err := ParseModelChoice(input)
		if err != nil || got != want {
			t.Errorf("ParseModelChoice(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "gpt-6", "anthropic/", "/model", "~invalid", "openai//model", "openai/my model"} {
		if _, err := ParseModelChoice(input); err == nil {
			t.Errorf("ParseModelChoice(%q) should reject malformed override", input)
		}
	}
}

func TestDisplayName(t *testing.T) {
	if got := DisplayName(ModelSonnet5); got != "Claude Sonnet 5" {
		t.Errorf("DisplayName(Sonnet5) = %q", got)
	}
	if got := DisplayName(ModelHaiku45); got != "Claude Haiku 4.5" {
		t.Errorf("DisplayName(Haiku45) = %q", got)
	}
	if got := DisplayName(ModelOpus5); got != "Claude Opus 5" {
		t.Errorf("DisplayName(Opus5) = %q", got)
	}
	if got := DisplayName("custom/model"); got != "custom/model" {
		t.Errorf("DisplayName(unknown) = %q, want passthrough", got)
	}
}

func TestSelectableModels(t *testing.T) {
	opts := SelectableModels()
	if len(opts) != 3 {
		t.Fatalf("SelectableModels() len = %d, want 3", len(opts))
	}
	want := []string{ModelHaiku45, ModelSonnet5, ModelOpus5}
	for i, o := range opts {
		if o.ID != want[i] {
			t.Errorf("SelectableModels()[%d].ID = %q, want %q", i, o.ID, want[i])
		}
	}
}

func TestDefaultModels(t *testing.T) {
	cfg := DefaultModels()
	if cfg.ModelDM != ModelSonnet5 {
		t.Errorf("ModelDM = %q, want %q", cfg.ModelDM, ModelSonnet5)
	}
	if cfg.ModelFast != ModelSonnet5 {
		t.Errorf("ModelFast = %q, want %q", cfg.ModelFast, ModelSonnet5)
	}
	if cfg.ModelNested != ModelSonnet5 {
		t.Errorf("ModelNested = %q, want %q", cfg.ModelNested, ModelSonnet5)
	}
	if cfg.ModelCampaign != ModelSonnet5 {
		t.Errorf("ModelCampaign = %q, want %q", cfg.ModelCampaign, ModelSonnet5)
	}
	if cfg.ModelAdvisor != ModelSonnet5 {
		t.Errorf("ModelAdvisor = %q, want %q", cfg.ModelAdvisor, ModelSonnet5)
	}
}
