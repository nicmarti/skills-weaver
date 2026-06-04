package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

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

func TestResolveAdvisorConfig(t *testing.T) {
	meta := &PersonaMetadata{Advisor: "opus-4.7", AdvisorMaxUses: 3}

	t.Run("disabled when flag off", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "")
		if _, _, _, ok := resolveAdvisorConfig(meta, anthropic.ModelClaudeSonnet4_6); ok {
			t.Error("expected advisor disabled when SW_ADVISOR_ENABLED is unset")
		}
	})

	t.Run("enabled with valid pair", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "1")
		model, maxUses, caching, ok := resolveAdvisorConfig(meta, anthropic.ModelClaudeSonnet4_6)
		if !ok {
			t.Fatal("expected advisor enabled")
		}
		if model != anthropic.ModelClaudeOpus4_7 {
			t.Errorf("advisor model = %v, want Opus 4.7", model)
		}
		if maxUses != 3 {
			t.Errorf("maxUses = %d, want 3", maxUses)
		}
		if caching != "" {
			t.Errorf("caching = %q, want off (unset)", caching)
		}
	})

	t.Run("caching parsed", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "1")
		m := &PersonaMetadata{Advisor: "opus-4.7", AdvisorCaching: "5m"}
		_, _, caching, ok := resolveAdvisorConfig(m, anthropic.ModelClaudeSonnet4_6)
		if !ok || caching != anthropic.BetaCacheControlEphemeralTTLTTL5m {
			t.Errorf("caching = %q (ok=%v), want 5m", caching, ok)
		}
	})

	t.Run("default max uses", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "true")
		m := &PersonaMetadata{Advisor: "opus-4.7"}
		_, maxUses, _, ok := resolveAdvisorConfig(m, anthropic.ModelClaudeSonnet4_6)
		if !ok || maxUses != defaultAdvisorMaxUses {
			t.Errorf("maxUses = %d (ok=%v), want %d", maxUses, ok, defaultAdvisorMaxUses)
		}
	})

	t.Run("disabled when no advisor field", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "1")
		if _, _, _, ok := resolveAdvisorConfig(&PersonaMetadata{}, anthropic.ModelClaudeSonnet4_6); ok {
			t.Error("expected advisor disabled when no advisor configured")
		}
	})

	t.Run("disabled when unrecognized advisor model", func(t *testing.T) {
		t.Setenv("SW_ADVISOR_ENABLED", "1")
		if _, _, _, ok := resolveAdvisorConfig(&PersonaMetadata{Advisor: "gpt-9"}, anthropic.ModelClaudeSonnet4_6); ok {
			t.Error("expected advisor disabled for unrecognized model")
		}
	})
}

func TestParseAdvisorCaching(t *testing.T) {
	cases := map[string]anthropic.BetaCacheControlEphemeralTTL{
		"5m":    anthropic.BetaCacheControlEphemeralTTLTTL5m,
		"1h":    anthropic.BetaCacheControlEphemeralTTLTTL1h,
		" 5M ":  anthropic.BetaCacheControlEphemeralTTLTTL5m,
		"":      "",
		"off":   "",
		"false": "",
		"10m":   "",
	}
	for in, want := range cases {
		if got := parseAdvisorCaching(in); got != want {
			t.Errorf("parseAdvisorCaching(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestAdvisorMetricsPersistence verifies advisor metrics survive a save/load
// round-trip through agent-states.json.
func TestAdvisorMetricsPersistence(t *testing.T) {
	t.Setenv("SW_ADVISOR_ENABLED", "1")

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	persona := "---\nname: world-keeper\nmodel: sonnet\nadvisor: opus-4.7\n---\n\nYou maintain world consistency."
	if err := os.WriteFile(filepath.Join(personaDir, "world-keeper.md"), []byte(persona), 0644); err != nil {
		t.Fatal(err)
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	logger, _ := NewLogger(tmpDir)
	mockClient := NewMockAnthropicClient()
	mockClient.GetMockMessagesService().SimulateAdvisorText = "advice"
	clientFactory := func(apiKey string) anthropicClient { return mockClient }

	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)
	if _, err := am.InvokeAgentSilent("world-keeper", "Brief.", 1); err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}

	statePath := filepath.Join(tmpDir, "agent-states.json")
	if err := am.SaveAgentStates(statePath); err != nil {
		t.Fatalf("SaveAgentStates: %v", err)
	}

	am2 := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)
	if err := am2.LoadAgentStates(statePath); err != nil {
		t.Fatalf("LoadAgentStates: %v", err)
	}
	state, ok := am2.GetNestedAgentState("world-keeper")
	if !ok {
		t.Fatal("world-keeper not restored")
	}
	if state.metrics.AdvisorCalls != 1 {
		t.Errorf("restored AdvisorCalls = %d, want 1", state.metrics.AdvisorCalls)
	}
	if state.metrics.AdvisorOutputTokens != 150 {
		t.Errorf("restored AdvisorOutputTokens = %d, want 150", state.metrics.AdvisorOutputTokens)
	}
	if state.metrics.AdvisorCacheCreationTokens != 1000 {
		t.Errorf("restored AdvisorCacheCreationTokens = %d, want 1000", state.metrics.AdvisorCacheCreationTokens)
	}
	if state.metrics.AdvisorCacheReadTokens != 500 {
		t.Errorf("restored AdvisorCacheReadTokens = %d, want 500", state.metrics.AdvisorCacheReadTokens)
	}
	if state.metrics.AdvisorModelUsed != "claude-opus-4-7" {
		t.Errorf("restored AdvisorModelUsed = %q, want claude-opus-4-7", state.metrics.AdvisorModelUsed)
	}
}

func TestToBetaMessages(t *testing.T) {
	std := []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock("carte du monde"),
			anthropic.NewImageBlockBase64("image/png", "ZmFrZQ=="),
		),
		anthropic.NewAssistantMessage(
			anthropic.NewTextBlock("ok"),
			anthropic.NewToolUseBlock("tu_1", map[string]any{"x": 1}, "get_party_info"),
		),
		anthropic.NewUserMessage(
			anthropic.NewToolResultBlock("tu_1", `{"hp":30}`, false),
		),
	}
	beta, err := toBetaMessages(std)
	if err != nil {
		t.Fatalf("toBetaMessages: %v", err)
	}
	if len(beta) != 3 {
		t.Fatalf("got %d beta messages, want 3", len(beta))
	}
}

// TestInvokeAgentSilent_AdvisorRouting verifies that an advisor-enabled persona
// is routed through the beta Messages API and that advisor advice + token
// metrics are captured.
func TestInvokeAgentSilent_AdvisorRouting(t *testing.T) {
	t.Setenv("SW_ADVISOR_ENABLED", "1")

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	persona := `---
name: world-keeper
description: World consistency guardian
model: sonnet
advisor: opus-4.7
advisor_max_uses: 2
advisor_caching: 5m
---

You maintain world consistency.`
	if err := os.WriteFile(filepath.Join(personaDir, "world-keeper.md"), []byte(persona), 0644); err != nil {
		t.Fatal(err)
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	logger, _ := NewLogger(tmpDir)

	mockClient := NewMockAnthropicClient()
	mockClient.GetMockMessagesService().SimulateAdvisorText = "Open cold on the city's wrongness; fold the fifth actor into the betrayer."
	clientFactory := func(apiKey string) anthropicClient { return mockClient }

	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	resp, err := am.InvokeAgentSilent("world-keeper", "Brief me for the Shasseth session.", 1)
	if err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty briefing")
	}

	svc := mockClient.GetMockMessagesService()
	if svc.BetaCallCount != 1 {
		t.Errorf("BetaCallCount = %d, want 1 (advisor path should use beta API)", svc.BetaCallCount)
	}

	// Verify the beta request carried the advisor tool with the configured model.
	if svc.LastBetaParams == nil {
		t.Fatal("expected LastBetaParams to be set")
	}
	if len(svc.LastBetaParams.Tools) != 1 || svc.LastBetaParams.Tools[0].OfAdvisorTool20260301 == nil {
		t.Fatalf("expected advisor tool in beta request, got %+v", svc.LastBetaParams.Tools)
	}
	if got := svc.LastBetaParams.Tools[0].OfAdvisorTool20260301.Model; got != anthropic.ModelClaudeOpus4_7 {
		t.Errorf("advisor tool model = %v, want Opus 4.7", got)
	}
	if got := svc.LastBetaParams.Tools[0].OfAdvisorTool20260301.Caching.TTL; got != anthropic.BetaCacheControlEphemeralTTLTTL5m {
		t.Errorf("advisor caching TTL = %q, want 5m", got)
	}

	state, ok := am.GetNestedAgentState("world-keeper")
	if !ok {
		t.Fatal("world-keeper state not tracked")
	}
	if state.metrics.AdvisorCalls != 1 {
		t.Errorf("AdvisorCalls = %d, want 1", state.metrics.AdvisorCalls)
	}
	if state.metrics.AdvisorOutputTokens != 150 {
		t.Errorf("AdvisorOutputTokens = %d, want 150 (from mock iterations)", state.metrics.AdvisorOutputTokens)
	}
	if state.metrics.AdvisorModelUsed != "claude-opus-4-7" {
		t.Errorf("AdvisorModelUsed = %q, want claude-opus-4-7", state.metrics.AdvisorModelUsed)
	}
	// Executor tokens come from the two "message" iterations (10 + 40 output).
	if state.metrics.TotalOutputTokens != 50 {
		t.Errorf("executor TotalOutputTokens = %d, want 50", state.metrics.TotalOutputTokens)
	}
}

// TestInvokeAgentSilent_NoAdvisorWhenFlagOff confirms the standard (non-beta)
// path is used when the feature flag is off, even if the persona declares an advisor.
func TestInvokeAgentSilent_NoAdvisorWhenFlagOff(t *testing.T) {
	t.Setenv("SW_ADVISOR_ENABLED", "")

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	persona := `---
name: world-keeper
model: sonnet
advisor: opus-4.7
---

You maintain world consistency.`
	if err := os.WriteFile(filepath.Join(personaDir, "world-keeper.md"), []byte(persona), 0644); err != nil {
		t.Fatal(err)
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	logger, _ := NewLogger(tmpDir)
	mockClient := NewMockAnthropicClient()
	clientFactory := func(apiKey string) anthropicClient { return mockClient }
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	if _, err := am.InvokeAgentSilent("world-keeper", "Brief me.", 1); err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}

	svc := mockClient.GetMockMessagesService()
	if svc.BetaCallCount != 0 {
		t.Errorf("BetaCallCount = %d, want 0 (flag off → standard path)", svc.BetaCallCount)
	}
	if svc.CallCount != 1 {
		t.Errorf("CallCount = %d, want 1 (standard path)", svc.CallCount)
	}
}

// TestInvokeAgent_AdvisorRouting verifies that the non-silent InvokeAgent tool
// loop routes advisor-enabled agents through the beta API with the advisor tool.
func TestInvokeAgent_AdvisorRouting(t *testing.T) {
	t.Setenv("SW_ADVISOR_ENABLED", "1")

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	persona := `---
name: world-keeper
model: sonnet
advisor: opus-4.7
---

You maintain world consistency.`
	if err := os.WriteFile(filepath.Join(personaDir, "world-keeper.md"), []byte(persona), 0644); err != nil {
		t.Fatal(err)
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	logger, _ := NewLogger(tmpDir)
	mockClient := NewMockAnthropicClient()
	mockClient.GetMockMessagesService().SimulateAdvisorText = "Cadre la cohérence avant tout."
	clientFactory := func(apiKey string) anthropicClient { return mockClient }
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	resp, err := am.InvokeAgent("world-keeper", "Vérifie la cohérence de Shasseth.", "", 1)
	if err != nil {
		t.Fatalf("InvokeAgent: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty response")
	}

	svc := mockClient.GetMockMessagesService()
	if svc.BetaCallCount == 0 {
		t.Errorf("expected beta API to be used for advisor-enabled agent, BetaCallCount=0")
	}
	// Beta request must include the advisor tool (alongside any read-only tools).
	if svc.LastBetaParams == nil {
		t.Fatal("expected LastBetaParams to be set")
	}
	var hasAdvisorTool bool
	for _, tool := range svc.LastBetaParams.Tools {
		if tool.OfAdvisorTool20260301 != nil {
			hasAdvisorTool = true
		}
	}
	if !hasAdvisorTool {
		t.Error("expected advisor tool in beta request tools")
	}

	state, _ := am.GetNestedAgentState("world-keeper")
	if state.metrics.AdvisorCalls != 1 {
		t.Errorf("AdvisorCalls = %d, want 1", state.metrics.AdvisorCalls)
	}
}

// TestInvokeAgentSilent_AdvisorRealAPI exercises the full production advisor
// path against the live beta Messages API. Gated: requires ANTHROPIC_API_KEY
// and RUN_REAL_API_TESTS=1. Uses the real world-keeper persona on disk.
func TestInvokeAgentSilent_AdvisorRealAPI(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" || os.Getenv("RUN_REAL_API_TESTS") == "" {
		t.Skip("Skipping real advisor API test: set ANTHROPIC_API_KEY and RUN_REAL_API_TESTS=1 to enable")
	}
	t.Setenv("SW_ADVISOR_ENABLED", "1")

	tmpDir := t.TempDir()
	adventureCtx := createTestAdventureContext(t, tmpDir)
	// Use the real persona dir so the on-disk world-keeper.md (with advisor) loads.
	personaLoader := NewPersonaLoaderWithPaths([]string{"../../core_agents/agents"})
	logger, _ := NewLogger(tmpDir)
	am := NewAgentManager(os.Getenv("ANTHROPIC_API_KEY"), adventureCtx, logger, nil, personaLoader)

	resp, err := am.InvokeAgentSilent("world-keeper",
		"Brief stratégique: les PJ épuisés arrivent à la cité de Shasseth où l'antagoniste Vaskir prépare un rituel. Deux foreshadows critiques à payer. Comment ouvrir la session?", 1)
	if err != nil {
		t.Fatalf("real advisor API call failed: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty briefing from real API")
	}
	t.Logf("Briefing (%d chars): %s", len(resp), resp)

	state, _ := am.GetNestedAgentState("world-keeper")
	t.Logf("Executor tokens: in=%d out=%d | Advisor calls=%d tokens in=%d out=%d (model=%s)",
		state.metrics.TotalInputTokens, state.metrics.TotalOutputTokens,
		state.metrics.AdvisorCalls, state.metrics.AdvisorInputTokens,
		state.metrics.AdvisorOutputTokens, state.metrics.AdvisorModelUsed)

	if state.metrics.AdvisorCalls == 0 {
		t.Log("WARNING: advisor was not consulted by the executor this run")
	}
}

// fakeTool is a minimal read-only Tool for exercising the combined
// advisor+tools beta request against the real API.
type fakeTool struct{}

func (fakeTool) Name() string        { return "get_party_info" }
func (fakeTool) Description() string { return "Returns a summary of the party." }
func (fakeTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}
func (fakeTool) Execute(params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"party": "Bob (Fighter, lvl 8, HP 30/64)"}, nil
}

// TestInvokeAgent_AdvisorWithToolsRealAPI exercises the combined advisor +
// client-side tools beta request against the live API. Gated.
func TestInvokeAgent_AdvisorWithToolsRealAPI(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" || os.Getenv("RUN_REAL_API_TESTS") == "" {
		t.Skip("Skipping real advisor+tools API test: set ANTHROPIC_API_KEY and RUN_REAL_API_TESTS=1 to enable")
	}
	t.Setenv("SW_ADVISOR_ENABLED", "1")

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	persona := "---\nname: world-keeper\ntools: [get_party_info]\nmodel: sonnet\nadvisor: opus-4.7\n---\n\nYou maintain world consistency. Use get_party_info if you need the party state."
	if err := os.WriteFile(filepath.Join(personaDir, "world-keeper.md"), []byte(persona), 0644); err != nil {
		t.Fatal(err)
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	logger, _ := NewLogger(tmpDir)
	am := NewAgentManager(os.Getenv("ANTHROPIC_API_KEY"), adventureCtx, logger, nil, personaLoader)

	// Wire a main registry containing the allowed read-only tool so the nested
	// agent's filtered registry is non-empty (combined advisor+tools path).
	reg := NewToolRegistry(adventureCtx)
	reg.Register(fakeTool{})
	am.SetMainToolRegistry(reg)

	resp, err := am.InvokeAgent("world-keeper",
		"Combien de PV a le groupe et est-ce cohérent pour le niveau 8 ? Vérifie via tes outils si besoin.", "", 1)
	if err != nil {
		t.Fatalf("real advisor+tools API call failed: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty response")
	}
	state, _ := am.GetNestedAgentState("world-keeper")
	t.Logf("resp=%q | exec in=%d out=%d | advisor calls=%d in=%d out=%d (%s)",
		resp, state.metrics.TotalInputTokens, state.metrics.TotalOutputTokens,
		state.metrics.AdvisorCalls, state.metrics.AdvisorInputTokens,
		state.metrics.AdvisorOutputTokens, state.metrics.AdvisorModelUsed)
}

func TestAdvisorFeatureEnabled(t *testing.T) {
	for _, on := range []string{"1", "true", "TRUE", "yes", "on"} {
		t.Setenv("SW_ADVISOR_ENABLED", on)
		if !advisorFeatureEnabled() {
			t.Errorf("advisorFeatureEnabled() = false for %q, want true", on)
		}
	}
	for _, off := range []string{"", "0", "false", "no", "off", "garbage"} {
		t.Setenv("SW_ADVISOR_ENABLED", off)
		if advisorFeatureEnabled() {
			t.Errorf("advisorFeatureEnabled() = true for %q, want false", off)
		}
	}
}
