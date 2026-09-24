package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dungeons/internal/llm"
)

// readOnlyStub is a read-only tool allowed by the rules-keeper policy.
type readOnlyStub struct {
	calls int
}

func (s *readOnlyStub) Name() string        { return "get_party_info" }
func (s *readOnlyStub) Description() string { return "Returns a summary of the party." }
func (s *readOnlyStub) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (s *readOnlyStub) Execute(params map[string]interface{}) (interface{}, error) {
	s.calls++
	return map[string]interface{}{"party": "Bob (Fighter, lvl 8, HP 30/64)"}, nil
}

// setupNestedManager builds an AgentManager on a fake client with personas in
// a temp dir. The personas deliberately declare `model: haiku` so tests can
// prove the persona model is metadata only and never overrides the config.
func setupNestedManager(t *testing.T, cfg llm.Config, client llm.Client) (*AgentManager, string) {
	t.Helper()

	tmpDir := t.TempDir()
	personaDir := filepath.Join(tmpDir, "agents")
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		t.Fatal(err)
	}
	personas := map[string]string{
		"rules-keeper":      "---\nname: rules-keeper\nmodel: haiku\n---\nYou are a D&D 5e rules expert.",
		"character-creator": "---\nname: character-creator\nmodel: haiku\n---\nYou help create D&D characters.",
		"world-keeper":      "---\nname: world-keeper\nmodel: haiku\n---\nYou maintain world consistency.",
	}
	for name, body := range personas {
		if err := os.WriteFile(filepath.Join(personaDir, name+".md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	adventureCtx := createTestAdventureContext(t, tmpDir)
	personaLoader := NewPersonaLoaderWithPaths([]string{personaDir})
	am := NewAgentManager(client, cfg, adventureCtx, nil, nil, personaLoader)
	return am, tmpDir
}

// TestNestedAgents_ToolLoopWithFilteredTools proves the neutral nested tool
// loop: the filtered read-only registry is offered to the model, call IDs and
// JSON results survive, and tool requests demand parameter support.
func TestNestedAgents_ToolLoopWithFilteredTools(t *testing.T) {
	fake := newNestedFakeClient()
	fake.Script = []llm.Response{
		{
			FinishReason: llm.FinishToolCalls,
			Message: llm.AssistantToolCallMessage("", []llm.ToolCall{
				{ID: "call-1", Name: "get_party_info", Arguments: json.RawMessage(`{}`)},
			}),
			Model: "anthropic/claude-sonnet-5",
		},
		{
			FinishReason: llm.FinishStop,
			Message:      llm.AssistantMessage("The party is healthy."),
			Model:        "anthropic/claude-sonnet-5",
			Usage:        llm.Usage{PromptTokens: 80, CompletionTokens: 30, TotalTokens: 110, Cost: 0.002, Present: true},
		},
	}
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)

	// Wire a main registry so rules-keeper gets its policy-filtered tools.
	advCtx := am.adventureCtx
	registry := NewToolRegistry(advCtx)
	stub := &readOnlyStub{}
	registry.Register(stub)
	am.SetMainToolRegistry(registry)

	resp, err := am.InvokeAgent("rules-keeper", "How is the party?", "", 1)
	if err != nil {
		t.Fatalf("InvokeAgent: %v", err)
	}
	if resp != "The party is healthy." {
		t.Errorf("response = %q", resp)
	}
	if stub.calls != 1 {
		t.Errorf("stub executed %d times, want 1", stub.calls)
	}

	state, _ := am.GetNestedAgentState("rules-keeper")
	msgs := state.conversationCtx.NeutralMessages()
	if len(msgs) != 4 { // user, assistant+tool call, tool result, final
		t.Fatalf("conversation = %+v", msgs)
	}
	if len(msgs[1].ToolCalls) != 1 || msgs[1].ToolCalls[0].ID != "call-1" {
		t.Fatalf("assistant tool call not preserved: %+v", msgs[1])
	}
	if msgs[2].Role != llm.RoleTool || msgs[2].ToolResults[0].ToolCallID != "call-1" {
		t.Fatalf("tool result ID mismatch: %+v", msgs[2])
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(msgs[2].ToolResults[0].Content), &parsed); err != nil || parsed["party"] == nil {
		t.Fatalf("tool result is not the tool's JSON: %s", msgs[2].ToolResults[0].Content)
	}
	if msgs[3].Text != "The party is healthy." {
		t.Fatalf("final message = %+v", msgs[3])
	}

	// Tool requests must require endpoint parameter support.
	firstReq := fake.requests()[0]
	if len(firstReq.Tools) == 0 {
		t.Error("filtered tools were not offered to the model")
	}
	if !firstReq.RequireParameters {
		t.Error("tool request must set provider.require_parameters")
	}
	secondReq := fake.requests()[1]
	if len(secondReq.Tools) == 0 || !secondReq.RequireParameters {
		t.Error("follow-up request must keep tools and parameter requirements")
	}

	// Metrics: usage aggregated across both calls, routed model recorded.
	m := state.metrics
	if m.TotalInputTokens != 80 || m.TotalOutputTokens != 30 {
		t.Errorf("token metrics = in:%d out:%d", m.TotalInputTokens, m.TotalOutputTokens)
	}
	if m.RoutedModel != "anthropic/claude-sonnet-5" {
		t.Errorf("RoutedModel = %q", m.RoutedModel)
	}
	if m.TotalCost != 0.002 {
		t.Errorf("TotalCost = %v", m.TotalCost)
	}
}

// TestNestedAgents_BoundedByPolicyIterations proves the loop stops at the
// agent's policy iteration limit instead of looping forever.
func TestNestedAgents_BoundedByPolicyIterations(t *testing.T) {
	fake := newNestedFakeClient()
	alwaysTool := llm.Response{
		FinishReason: llm.FinishToolCalls,
		Message: llm.AssistantToolCallMessage("", []llm.ToolCall{
			{ID: "call-x", Name: "get_party_info", Arguments: json.RawMessage(`{}`)},
		}),
	}
	fake.Script = make([]llm.Response, 20)
	for i := range fake.Script {
		fake.Script[i] = alwaysTool
	}
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)

	registry := NewToolRegistry(am.adventureCtx)
	registry.Register(&readOnlyStub{})
	am.SetMainToolRegistry(registry)

	_, err := am.InvokeAgent("rules-keeper", "loop forever", "", 1)
	if err == nil || !strings.Contains(err.Error(), "iterations") {
		t.Fatalf("expected bounded-loop error, got %v", err)
	}

	// rules-keeper policy allows 8 iterations.
	if got := fake.callCount(); got != 8 {
		t.Errorf("provider called %d times, want 8 (policy MaxIterations)", got)
	}
}

// TestInvokeAgentSilent_SingleCallNoTools proves the silent path is one
// neutral call without client tools, and that new OpenRouter metrics persist
// on the agent state.
func TestInvokeAgentSilent_SingleCallNoTools(t *testing.T) {
	fake := newNestedFakeClient()
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)

	resp, err := am.InvokeAgentSilent("rules-keeper", "Secret briefing question.", 1)
	if err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}
	if resp == "" {
		t.Fatal("expected non-empty response")
	}
	if got := fake.callCount(); got != 1 {
		t.Errorf("provider called %d times, want 1 (single silent call)", got)
	}
	last := fake.requests()[0]
	if len(last.Tools) != 0 {
		t.Error("silent invocation must not offer client tools")
	}
	if last.RequireParameters {
		t.Error("silent invocation must not require parameters (no tools)")
	}

	state, _ := am.GetNestedAgentState("rules-keeper")
	if state.metrics.RequestedModel != llm.DefaultModelNested {
		t.Errorf("RequestedModel = %q, want %q", state.metrics.RequestedModel, llm.DefaultModelNested)
	}
	if state.metrics.RoutedModel == "" || state.metrics.TotalCost <= 0 {
		t.Errorf("OpenRouter metrics not recorded: routed=%q cost=%v",
			state.metrics.RoutedModel, state.metrics.TotalCost)
	}
}

// TestNestedAgents_ErrorHandling proves provider failures surface as
// AgentError without committing partial history.
func TestNestedAgents_ErrorHandling(t *testing.T) {
	fake := newNestedFakeClient()
	fake.SimulateError = true
	fake.ErrorMessage = "provider exploded"
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)

	_, err := am.InvokeAgent("rules-keeper", "question", "", 1)
	var agentErr *AgentError
	if err == nil {
		t.Fatal("expected error from failing provider")
	}
	if e, ok := err.(*AgentError); ok {
		agentErr = e
	}
	if agentErr == nil {
		t.Fatalf("expected *AgentError, got %T: %v", err, err)
	}
	if agentErr.AgentName != "rules-keeper" {
		t.Errorf("AgentError.AgentName = %q", agentErr.AgentName)
	}
	if !strings.Contains(agentErr.Error(), "provider exploded") {
		t.Errorf("error lost the provider cause: %v", agentErr)
	}
	state, _ := am.GetNestedAgentState("rules-keeper")
	for _, msg := range state.conversationCtx.NeutralMessages() {
		if msg.Role == llm.RoleAssistant {
			t.Fatal("partial assistant history committed after provider failure")
		}
	}
}

// TestNestedAgents_ModelResolutionFromConfig proves each nested model comes
// from llm.Config (per-agent override > nested default > Sonnet 5) and the
// persona `model:` field is metadata only.
func TestNestedAgents_ModelResolutionFromConfig(t *testing.T) {
	fake := newNestedFakeClient()
	// Personas declare model: haiku; the config must win.
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)

	rk, err := am.getOrCreateNestedAgent("rules-keeper")
	if err != nil {
		t.Fatal(err)
	}
	if rk.model != llm.DefaultModelNested {
		t.Errorf("rules-keeper model = %q, want nested default %q (persona 'model: haiku' must be ignored)",
			rk.model, llm.DefaultModelNested)
	}

	// Per-agent override is honored and never silently replaced.
	cfg := llm.DefaultModels()
	cfg.AgentModels[llm.WorldKeeperAgent] = "google/gemini-3-pro"
	fake2 := newNestedFakeClient()
	am2, _ := setupNestedManager(t, cfg, fake2)
	wk, err := am2.getOrCreateNestedAgent("world-keeper")
	if err != nil {
		t.Fatal(err)
	}
	if wk.model != "google/gemini-3-pro" {
		t.Errorf("world-keeper model = %q, want the explicit override", wk.model)
	}
}

// TestWorldKeeper_ImageSentToVisionModel_TextFallbackOtherwise proves the
// World Keeper map image reaches vision-capable models unchanged, while
// non-vision models get the text-only-map fallback (no invalid image payload,
// no silent model swap).
func TestWorldKeeper_ImageSentToVisionModel_TextFallbackOtherwise(t *testing.T) {
	withWorldResources := func(am *AgentManager) {
		am.worldResources = &WorldResources{
			MapDescription:    "Les Quatre Royaumes...",
			MapImageBase64:    "QUJD",
			MapImageMediaType: "image/png",
		}
	}

	// Vision-capable default model: the image is sent.
	fake := newNestedFakeClient()
	am, _ := setupNestedManager(t, llm.DefaultModels(), fake)
	withWorldResources(am)
	if _, err := am.InvokeAgentSilent("world-keeper", "brief", 1); err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}
	req := fake.requests()[0]
	hasImage := false
	for _, msg := range req.Messages {
		for _, img := range msg.Images {
			if img.Base64 == "QUJD" {
				hasImage = true
			}
		}
	}
	if !hasImage {
		t.Fatal("world map image missing from request to a vision-capable model")
	}

	// Non-vision model: images stripped, text preserved.
	cfg := llm.DefaultModels()
	cfg.AgentModels[llm.WorldKeeperAgent] = "openai/gpt-6"
	fake2 := newNestedFakeClient()
	am2, _ := setupNestedManager(t, cfg, fake2)
	withWorldResources(am2)
	if _, err := am2.InvokeAgentSilent("world-keeper", "brief", 1); err != nil {
		t.Fatalf("InvokeAgentSilent (fallback): %v", err)
	}
	req2 := fake2.requests()[0]
	for _, msg := range req2.Messages {
		if len(msg.Images) > 0 {
			t.Fatalf("image payload sent to non-vision model: %+v", msg.Images)
		}
	}
	// The textual map exchange (map message text) is preserved.
	foundMapText := false
	for _, msg := range req2.Messages {
		if strings.Contains(msg.Text, "carte du monde") {
			foundMapText = true
		}
	}
	if !foundMapText {
		t.Error("text-only fallback dropped the map text entirely")
	}
	// The model was never silently swapped.
	if req2.Model != "openai/gpt-6" {
		t.Errorf("request model = %q, want the user-selected non-vision model", req2.Model)
	}
}

// TestNestedAgents_NewMetricsPersistRoundTrip proves the new OpenRouter
// metric fields survive save/load while the historical Advisor fields remain
// untouched.
func TestNestedAgents_NewMetricsPersistRoundTrip(t *testing.T) {
	fake := newNestedFakeClient()
	am, tmpDir := setupNestedManager(t, llm.DefaultModels(), fake)

	if _, err := am.InvokeAgentSilent("rules-keeper", "question", 1); err != nil {
		t.Fatalf("InvokeAgentSilent: %v", err)
	}

	statePath := filepath.Join(tmpDir, "agent-states.json")
	if err := am.SaveAgentStates(statePath); err != nil {
		t.Fatalf("SaveAgentStates: %v", err)
	}

	fake2 := newNestedFakeClient()
	am2, _ := setupNestedManager(t, llm.DefaultModels(), fake2)
	if err := am2.LoadAgentStates(statePath); err != nil {
		t.Fatalf("LoadAgentStates: %v", err)
	}
	state, ok := am2.GetNestedAgentState("rules-keeper")
	if !ok {
		t.Fatal("rules-keeper not restored")
	}
	if state.metrics.RequestedModel != llm.DefaultModelNested {
		t.Errorf("restored RequestedModel = %q", state.metrics.RequestedModel)
	}
	if state.metrics.RoutedModel == "" {
		t.Error("restored RoutedModel lost")
	}
	if state.metrics.TotalCost <= 0 {
		t.Errorf("restored TotalCost = %v, want the recorded aggregate cost", state.metrics.TotalCost)
	}
	if state.metrics.CachedTokens <= 0 || state.metrics.ReasoningTokens <= 0 {
		t.Errorf("restored cache/reasoning tokens lost: cached=%d reasoning=%d",
			state.metrics.CachedTokens, state.metrics.ReasoningTokens)
	}
}

// TestNestedAgents_AdvisorFieldsUntouched proves the historical Advisor
// metrics survive the new save format unchanged (readability requirement).
func TestNestedAgents_AdvisorFieldsUntouched(t *testing.T) {
	fake := newNestedFakeClient()
	am, tmpDir := setupNestedManager(t, llm.DefaultModels(), fake)

	rk, err := am.getOrCreateNestedAgent("rules-keeper")
	if err != nil {
		t.Fatal(err)
	}
	// Simulate pre-migration historical values loaded from disk.
	rk.metrics.AdvisorCalls = 2
	rk.metrics.AdvisorModelUsed = "claude-opus-4-7"
	rk.metrics.AdvisorCacheReadTokens = 120

	statePath := filepath.Join(tmpDir, "agent-states.json")
	if err := am.SaveAgentStates(statePath); err != nil {
		t.Fatalf("SaveAgentStates: %v", err)
	}

	fake2 := newNestedFakeClient()
	am2, _ := setupNestedManager(t, llm.DefaultModels(), fake2)
	if err := am2.LoadAgentStates(statePath); err != nil {
		t.Fatalf("LoadAgentStates: %v", err)
	}
	state, _ := am2.GetNestedAgentState("rules-keeper")
	if state.metrics.AdvisorCalls != 2 || state.metrics.AdvisorModelUsed != "claude-opus-4-7" || state.metrics.AdvisorCacheReadTokens != 120 {
		t.Errorf("historical advisor metrics changed: %+v", state.metrics)
	}
}
