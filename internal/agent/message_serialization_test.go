package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dungeons/internal/llm"
)

func TestSerializeImageWithoutPersistingPayload(t *testing.T) {
	for _, text := range []string{"Voici la carte du monde.", ""} {
		msg := llm.UserMessageWithImage(text, llm.ImagePart{
			MediaType: "image/png", Base64: "QUJD", ResourceRef: "world-map",
		})
		stored, err := SerializeMessage(msg)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(stored)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "QUJD") || stored.TextContent == "" || stored.ImageResource != "world-map" {
			t.Fatalf("image reference or placeholder not preserved safely: %s", data)
		}
		loaded, err := DeserializeMessage(stored)
		if err != nil || loaded.Role != llm.RoleUser || len(loaded.Images) != 1 || loaded.Images[0].ResourceRef != "world-map" || loaded.Images[0].Base64 != "" {
			t.Fatalf("image reference round-trip: %+v, %v", loaded, err)
		}
	}
}

func managerWithState(name string, ctx *ConversationContext) *AgentManager {
	return &AgentManager{nestedAgents: map[string]*NestedAgentState{
		name: {agentName: name, conversationCtx: ctx, tokenLimit: 20000, metrics: &AgentMetrics{}},
	}}
}

func TestLegacyAgentStatesFixtureRoundTrip(t *testing.T) {
	fixture := filepath.Join("testdata", "legacy_agent_states.json")
	am := managerWithState("rules-keeper", NewConversationContextWithLimit(20000))
	if err := am.LoadAgentStates(fixture); err != nil {
		t.Fatal(err)
	}
	state := am.nestedAgents["rules-keeper"]
	assertLegacyRestored(t, state)

	path := filepath.Join(t.TempDir(), "agent-states.json")
	if err := am.SaveAgentStates(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved AgentStatesFile
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.SchemaVersion != 2 || saved.Agents["rules-keeper"].ConversationHistory[2].Role != "tool" {
		t.Fatalf("state must save in v2 with tool-role results: %+v", saved)
	}
	if strings.Contains(string(data), `"input": "{`) {
		t.Fatal("tool arguments were saved as an escaped JSON string")
	}
	if !strings.Contains(string(data), "9007199254740993") {
		t.Fatal("tool arguments lost their original JSON number precision")
	}

	am2 := managerWithState("rules-keeper", NewConversationContextWithLimit(20000))
	if err := am2.LoadAgentStates(path); err != nil {
		t.Fatal(err)
	}
	assertLegacyRestored(t, am2.nestedAgents["rules-keeper"])
}

func TestLegacyStateMissingMaxTokensUsesAgentLimit(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "legacy_agent_states.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file AgentStatesFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	file.Agents["rules-keeper"].MaxTokens = 0
	data, err = json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "agent-states.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	am := managerWithState("rules-keeper", NewConversationContextWithLimit(20000))
	if err := am.LoadAgentStates(path); err != nil {
		t.Fatal(err)
	}
	if got := am.nestedAgents["rules-keeper"].conversationCtx.maxTokens; got != 20000 {
		t.Fatalf("missing legacy max_tokens restored a disabled live limit: %d", got)
	}
}

func assertLegacyRestored(t *testing.T, state *NestedAgentState) {
	t.Helper()
	msgs := state.conversationCtx.NeutralMessages()
	if state.invocationCount != 2 || len(msgs) != 5 || msgs[1].Role != llm.RoleAssistant || len(msgs[1].ToolCalls) != 2 || msgs[2].Role != llm.RoleTool || msgs[3].Role != llm.RoleTool || !msgs[3].ToolResults[0].IsError {
		t.Fatalf("legacy multi-tool exchange not restored: messages=%+v invocation=%d", msgs, state.invocationCount)
	}
	var args map[string]json.RawMessage
	if err := json.Unmarshal(msgs[1].ToolCalls[0].Arguments, &args); err != nil || string(args["reference"]) != "9007199254740993" || string(args["name"]) != `"boule de feu"` {
		t.Fatalf("tool arguments lost: %s, %v", msgs[1].ToolCalls[0].Arguments, err)
	}
	if state.metrics.ModelUsed != "claude-sonnet-4-6" || state.metrics.AdvisorModelUsed != "claude-opus-4-7" || state.metrics.AdvisorCalls != 2 || state.metrics.AdvisorCacheReadTokens != 120 || state.metrics.AdvisorCacheCreationTokens != 340 {
		t.Fatalf("historical metrics changed: %+v", state.metrics)
	}
}

func TestIncompleteLegacyExchangeDroppedAtomically(t *testing.T) {
	stored := []SerializableMessage{
		{Role: "user", TextContent: "question"},
		{Role: "assistant", TextContent: "checking", ToolUses: []SerializableToolUse{
			{ID: "a", Name: "first", Input: json.RawMessage(`{}`)},
			{ID: "b", Name: "second", Input: json.RawMessage(`{}`)},
		}},
		{Role: "user", ToolResults: []SerializableToolResult{{ToolUseID: "a", Content: "ok"}}},
		{Role: "assistant", TextContent: "later"},
		{Role: "tool", ToolResults: []SerializableToolResult{{ToolUseID: "orphan", Content: "bad"}}},
	}
	ctx, err := DeserializeConversationContextFromMessages(stored, 20000)
	if err != nil {
		t.Fatal(err)
	}
	msgs := ctx.NeutralMessages()
	if len(msgs) != 2 || msgs[0].Text != "question" || msgs[1].Text != "later" {
		t.Fatalf("incomplete exchange or orphan result leaked into history: %+v", msgs)
	}
}

func TestLegacyCombinedUserTextAndToolResultsStayOutsideExchange(t *testing.T) {
	stored := []SerializableMessage{
		{Role: "assistant", ToolUses: []SerializableToolUse{
			{ID: "a", Name: "first", Input: json.RawMessage(`{}`)},
			{ID: "b", Name: "second", Input: json.RawMessage(`{}`)},
		}},
		{Role: "user", TextContent: "Historical note", ToolResults: []SerializableToolResult{{ToolUseID: "a", Content: "ok"}}},
		{Role: "user", ToolResults: []SerializableToolResult{{ToolUseID: "b", Content: "error", IsError: true}}},
	}
	ctx, err := DeserializeConversationContextFromMessages(stored, 20000)
	if err != nil {
		t.Fatal(err)
	}
	msgs := ctx.NeutralMessages()
	if len(msgs) != 4 || msgs[1].Role != llm.RoleTool || msgs[2].Role != llm.RoleTool || msgs[3].Role != llm.RoleUser || msgs[3].Text != "Historical note" {
		t.Fatalf("legacy text separated a multi-tool exchange: %+v", msgs)
	}
	trimmed, err := SerializeConversationContextWithOptimization(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(trimmed) != 1 || trimmed[0].TextContent != "Historical note" {
		t.Fatalf("newest user turn must not retain half of an older tool exchange: %+v", trimmed)
	}
}

func TestTruncationPreservesMultiToolExchange(t *testing.T) {
	ctx := NewConversationContextWithLimit(120)
	for i := 0; i < 24; i++ {
		ctx.AddUserMessage(strings.Repeat("a", 48))
	}
	ctx.AddAssistantMessageWithTools("checking", []ToolUse{
		{ID: "a", Name: "first", Input: map[string]interface{}{"value": 1}},
		{ID: "b", Name: "second", Input: map[string]interface{}{"value": 2}},
	})
	ctx.AddToolResultMessage(ToolResultMessage{ToolUseID: "a", Content: `{"ok":true}`})
	ctx.AddToolResultMessage(ToolResultMessage{ToolUseID: "b", Content: `{"ok":false}`, IsError: true})
	msgs := ctx.NeutralMessages()
	for i, msg := range msgs {
		if len(msg.ToolCalls) == 2 {
			if i+2 >= len(msgs) || msgs[i+1].ToolResults[0].ToolCallID != "a" || msgs[i+2].ToolResults[0].ToolCallID != "b" {
				t.Fatalf("live truncation split a multi-tool exchange: %+v", msgs)
			}
		}
	}
	stored, err := SerializeConversationContextWithOptimization(ctx, 60)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 3 || stored[0].Role != "assistant" || stored[1].Role != "tool" || stored[2].Role != "tool" {
		t.Fatalf("disk budget split oversized latest exchange: %+v", stored)
	}
}

func TestTruncationRespectsBudgetBeforeTwentyMessages(t *testing.T) {
	ctx := NewConversationContextWithLimit(10)
	ctx.AddUserMessage(strings.Repeat("a", 32))
	ctx.AddUserMessage(strings.Repeat("b", 32))
	msgs := ctx.NeutralMessages()
	if len(msgs) != 1 || msgs[0].Text != strings.Repeat("b", 32) || ctx.tokenEstimate != 8 {
		t.Fatalf("over-budget live context retained old messages: %+v (tokens=%d)", msgs, ctx.tokenEstimate)
	}
	ctx.AddUserMessage(strings.Repeat("c", 120))
	msgs = ctx.NeutralMessages()
	if len(msgs) != 1 || msgs[0].Text != strings.Repeat("c", 120) {
		t.Fatalf("oversized latest turn should remain intact: %+v", msgs)
	}
}

func TestWorldKeeperMapRestoredOnceAfterStateLoad(t *testing.T) {
	fresh := func() *ConversationContext {
		ctx := NewConversationContextWithLimit(20000)
		ctx.AddUserMessageWithImageResource("Voici la carte du monde des Quatre Royaumes. Utilise-la comme référence géographique pour toutes tes validations.", "QUJD", "image/png", "world-map")
		ctx.AddAssistantMessage("J'ai bien reçu la carte du monde des Quatre Royaumes. Je l'utiliserai comme référence pour assurer la cohérence géographique de l'aventure.")
		return ctx
	}
	am := managerWithState("world-keeper", fresh())
	am.worldResources = &WorldResources{MapImageBase64: "QUJD", MapImageMediaType: "image/png"}
	am.nestedAgents["world-keeper"].conversationCtx.AddUserMessage("later question")
	path := filepath.Join(t.TempDir(), "agent-states.json")
	if err := am.SaveAgentStates(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "QUJD") || !strings.Contains(string(data), `"image_resource": "world-map"`) {
		t.Fatal("persisted world map must contain a reference, not base64 bytes")
	}
	am2 := managerWithState("world-keeper", fresh())
	am2.worldResources = am.worldResources
	for i := 0; i < 2; i++ {
		if err := am2.LoadAgentStates(path); err != nil {
			t.Fatal(err)
		}
		msgs := am2.nestedAgents["world-keeper"].conversationCtx.NeutralMessages()
		if len(msgs) != 3 || len(msgs[0].Images) != 1 || msgs[0].Images[0].Base64 != "QUJD" || msgs[2].Text != "later question" {
			t.Fatalf("world map missing or duplicated on restore %d: %+v", i, msgs)
		}
	}
	// An old save has the map text placeholder but no resource-reference field.
	legacyPath := filepath.Join(t.TempDir(), "legacy-map.json")
	legacy := strings.ReplaceAll(string(data), `"image_resource": "world-map",`, "")
	if err := os.WriteFile(legacyPath, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	am3 := managerWithState("world-keeper", fresh())
	am3.worldResources = am.worldResources
	if err := am3.LoadAgentStates(legacyPath); err != nil {
		t.Fatal(err)
	}
	msgs := am3.nestedAgents["world-keeper"].conversationCtx.NeutralMessages()
	if len(msgs) != 3 || len(msgs[0].Images) != 1 || msgs[0].Images[0].Base64 != "QUJD" {
		t.Fatalf("legacy map placeholder was not upgraded: %+v", msgs)
	}
}

func TestWorldKeeperCreationTagsMapResource(t *testing.T) {
	am, cleanup := setupAgentManager(t)
	defer cleanup()
	createTestPersonas(t, am)
	am.worldResources = &WorldResources{MapImageBase64: "QUJD", MapImageMediaType: "image/png"}
	world, err := am.getOrCreateNestedAgent("world-keeper")
	if err != nil {
		t.Fatal(err)
	}
	messages := world.conversationCtx.NeutralMessages()
	if len(messages) != 2 || len(messages[0].Images) != 1 || messages[0].Images[0].ResourceRef != "world-map" {
		t.Fatalf("World Keeper map was not tagged at construction: %+v", messages)
	}
	rules, err := am.getOrCreateNestedAgent("rules-keeper")
	if err != nil {
		t.Fatal(err)
	}
	if len(rules.conversationCtx.NeutralMessages()) != 0 {
		t.Fatal("map resource leaked into another nested agent")
	}
}
