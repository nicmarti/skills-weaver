package agent

import (
	"dungeons/internal/dmtools"
	"dungeons/internal/skills"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestToolDefinitionsPreserveCompleteSchemaAndOrder(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"items": map[string]interface{}{
				"type": "array", "minItems": 1,
				"items": map[string]interface{}{
					"type": "object", "additionalProperties": false,
					"properties": map[string]interface{}{
						"kind": map[string]interface{}{"type": "string", "enum": []string{"sword", "shield"}},
					},
				},
			},
		},
		"required": []interface{}{"items"}, "additionalProperties": false,
		"minProperties": 1, "$defs": map[string]interface{}{"tag": map[string]interface{}{"type": "string"}},
	}
	registry := NewToolRegistry(nil)
	registry.Register(newMockTool("zeta"))
	registry.Register(&mockTool{name: "alpha", description: "Full schema", schema: schema})
	definitions, err := registry.ToolDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != 2 || definitions[0].Name != "alpha" || definitions[1].Name != "zeta" || !reflect.DeepEqual(definitions[0].Parameters, schema) {
		t.Fatalf("lost order or schema fields: %+v", definitions)
	}
	if definitions[0].Description != "Full schema" {
		t.Fatal("tool description was dropped")
	}
}

func TestToolDefinitionsRejectMalformedSchemas(t *testing.T) {
	cases := []struct {
		name   string
		schema map[string]interface{}
	}{
		{"missing schema", nil},
		{"wrong root type", map[string]interface{}{"type": "array", "properties": map[string]interface{}{}}},
		{"missing properties", map[string]interface{}{"type": "object"}},
		{"non-string required", map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []interface{}{123}}},
		{"non-json schema", map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "callback": func() {}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewToolRegistry(nil)
			registry.Register(&mockTool{name: "broken", schema: tc.schema})
			if _, err := registry.ToolDefinitions(); err == nil {
				t.Fatal("invalid schema must fail before it reaches a provider")
			}
		})
	}
}

func TestProductionRegistryConvertsEveryToolOffline(t *testing.T) {
	// Optional providers are only checked for the presence of environment keys
	// during registration; no tool is executed and no provider is contacted.
	t.Setenv("ANTHROPIC_API_KEY", "tool-schema-test")
	t.Setenv("FAL_KEY", "tool-schema-test")
	am, cleanup := setupAgentManager(t)
	defer cleanup()
	registry := NewToolRegistry(am.adventureCtx)
	if err := registerAllTools(registry, filepath.Join("..", "..", "data"), am.adventureCtx.Adventure, am, nil); err != nil {
		t.Fatal(err)
	}
	// registerAllTools uses cwd-relative defaults; go test runs from internal/agent.
	// Supply the same skill directory explicitly so the optional production
	// invoke_skill definition is checked without changing process cwd.
	skillRegistry, err := skills.NewRegistryWithParser(skills.NewSkillParserWithPaths([]string{
		filepath.Join("..", "..", "core_agents", "skills"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	registry.Register(dmtools.NewInvokeSkillTool(skillRegistry, am.adventureCtx.Adventure.BasePath()))
	definitions, err := registry.ToolDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != registry.Count() || len(definitions) < 30 {
		t.Fatalf("incomplete production registry: definitions=%d registered=%d", len(definitions), registry.Count())
	}
	for _, name := range []string{"roll_dice", "generate_image", "generate_map", "invoke_agent", "invoke_skill", "set_ambient_music"} {
		if _, ok := registry.Get(name); !ok {
			t.Errorf("expected %s in the full production registry", name)
		}
	}
	for _, definition := range definitions {
		if _, err := json.Marshal(definition.Parameters); err != nil {
			t.Errorf("%s: invalid Chat function parameters: %v", definition.Name, err)
		}
	}
	policy := GetPolicyForAgent("world-keeper")
	filtered := registry.CreateFilteredRegistry(policy.GetAllowedToolNames(), policy.ForbiddenTools)
	filteredDefs, err := filtered.ToolDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	if len(filteredDefs) != filtered.Count() || filtered.Count() == 0 {
		t.Fatal("world-keeper lost its read-only tool definitions")
	}
	for _, def := range filteredDefs {
		if _, ok := registry.Get(def.Name); !ok || !policy.IsToolAllowed(def.Name) {
			t.Errorf("unexpected nested agent tool %q", def.Name)
		}
	}
}

type stubTool struct {
	*mockTool
	result interface{}
	err    error
}

func (t *stubTool) Execute(map[string]interface{}) (interface{}, error) {
	return t.result, t.err
}

type silentToolOutput struct{}

func (silentToolOutput) OnTextChunk(string)                              {}
func (silentToolOutput) OnToolStart(string, string)                      {}
func (silentToolOutput) OnToolComplete(string, interface{})              {}
func (silentToolOutput) OnAgentInvocationStart(string)                   {}
func (silentToolOutput) OnAgentInvocationComplete(string, time.Duration) {}
func (silentToolOutput) OnError(error)                                   {}
func (silentToolOutput) OnComplete()                                     {}

func TestMainAndNestedToolResultsAreAlwaysJSON(t *testing.T) {
	unsafeText := "quoted \"value\" \\ path\nnext line"
	cases := []struct {
		name       string
		result     interface{}
		err        error
		wantError  bool
		wantDetail string
	}{
		{"success", map[string]interface{}{"success": true, "display": unsafeText}, nil, false, unsafeText},
		{"reported failure", map[string]interface{}{"success": false, "error": unsafeText}, nil, true, unsafeText},
		{"Go error", nil, errors.New(unsafeText), true, unsafeText},
		{"unserializable result", map[string]interface{}{"bad": func() {}}, nil, true, "not JSON serializable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := &stubTool{mockTool: newMockTool("probe"), result: tc.result, err: tc.err}
			registry := NewToolRegistry(nil)
			registry.Register(tool)
			main := &Agent{toolRegistry: registry, outputHandler: silentToolOutput{}}
			am := &AgentManager{}
			nested := &NestedAgentState{agentName: "rules-keeper", toolRegistry: registry}
			for caller, results := range map[string][]ToolResultMessage{
				"main":   main.executeTools([]ToolUse{{ID: "call-1", Name: "probe"}}),
				"nested": am.executeNestedAgentTools(nested, []ToolUse{{ID: "call-1", Name: "probe"}}),
			} {
				if len(results) != 1 || results[0].IsError != tc.wantError || !json.Valid([]byte(results[0].Content)) {
					t.Fatalf("%s: invalid result %+v", caller, results)
				}
				if !strings.Contains(results[0].Content, `"success"`) && tc.wantError {
					t.Fatalf("%s: missing error status in %s", caller, results[0].Content)
				}
				var decoded map[string]interface{}
				if err := json.Unmarshal([]byte(results[0].Content), &decoded); err != nil {
					t.Fatalf("%s: result decode: %v", caller, err)
				}
				key := "display"
				if tc.wantError {
					key = "error"
				}
				if !strings.Contains(decoded[key].(string), tc.wantDetail) {
					t.Fatalf("%s: expected escaped text to round-trip, got %v", caller, decoded)
				}
			}
		})
	}
}

func TestMissingToolErrorUsesEscapedJSON(t *testing.T) {
	name := "missing\" \\ tool\nnext line"
	registry := NewToolRegistry(nil)
	main := &Agent{toolRegistry: registry, outputHandler: silentToolOutput{}}
	am := &AgentManager{}
	nested := &NestedAgentState{agentName: "rules-keeper", toolRegistry: registry}
	for caller, results := range map[string][]ToolResultMessage{
		"main":   main.executeTools([]ToolUse{{ID: "call-1", Name: name}}),
		"nested": am.executeNestedAgentTools(nested, []ToolUse{{ID: "call-1", Name: name}}),
	} {
		if len(results) != 1 || !results[0].IsError || !json.Valid([]byte(results[0].Content)) {
			t.Fatalf("%s: invalid missing-tool error: %+v", caller, results)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(results[0].Content), &payload); err != nil || !strings.Contains(payload["error"].(string), name) {
			t.Fatalf("%s: missing tool name not preserved: %s (%v)", caller, results[0].Content, err)
		}
	}
}
