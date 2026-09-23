package llm

import (
	"encoding/json"
	"reflect"
	"testing"
)

func complexToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"equipment": map[string]interface{}{
				"type": "array", "minItems": 1,
				"items": map[string]interface{}{
					"type": "object", "additionalProperties": false,
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string", "enum": []string{"shield", "sword"}},
					},
				},
			},
		},
		"required": []interface{}{"equipment"}, "additionalProperties": false,
		"minProperties": 1,
	}
}

func decodeToolSchema(t *testing.T, wire []byte, root string) (map[string]interface{}, map[string]interface{}) {
	t.Helper()
	var payload map[string]interface{}
	if err := json.Unmarshal(wire, &payload); err != nil {
		t.Fatal(err)
	}
	tool := payload
	if root != "" {
		tool = payload[root].(map[string]interface{})
	}
	return tool, tool["input_schema"].(map[string]interface{})
}

func TestLegacyToolAdaptersRetainCompleteSchema(t *testing.T) {
	definition := ToolDefinition{Name: "equip_party", Description: "Prepare the party", Parameters: complexToolSchema()}
	standard, err := LegacyAnthropicToolParams([]ToolDefinition{definition})
	if err != nil {
		t.Fatal(err)
	}
	beta, err := LegacyAnthropicBetaToolParams([]ToolDefinition{definition})
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(definition.Parameters)
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]interface{}
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Fatal(err)
	}
	for name, tool := range map[string]interface{}{"standard": standard[0], "beta": beta[0]} {
		wire, err := json.Marshal(tool)
		if err != nil {
			t.Fatal(err)
		}
		_, schema := decodeToolSchema(t, wire, "")
		if !reflect.DeepEqual(schema, expected) {
			t.Errorf("%s tool dropped JSON Schema constraints: got %v want %v", name, schema, expected)
		}
	}
}

func TestChatFunctionsRetainCompleteSchemaWithoutStrictMode(t *testing.T) {
	definition := ToolDefinition{Name: "equip_party", Description: "Prepare the party", Parameters: complexToolSchema()}
	request, err := buildChatRequest(Request{
		Model: ModelSonnet5, Messages: []Message{UserMessage("Equip the party")},
		Tools: []ToolDefinition{definition},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(wire, &payload); err != nil {
		t.Fatal(err)
	}
	tools := payload["tools"].([]interface{})
	if len(tools) != 1 || tools[0].(map[string]interface{})["type"] != "function" {
		t.Fatalf("expected one Chat function tool: %v", tools)
	}
	function := tools[0].(map[string]interface{})["function"].(map[string]interface{})
	if _, present := function["strict"]; present {
		t.Fatal("strict schema mode must not be enabled during initial migration")
	}
	schema := function["parameters"].(map[string]interface{})
	want, _ := json.Marshal(definition.Parameters)
	var expected map[string]interface{}
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(schema, expected) {
		t.Fatalf("Chat function dropped JSON Schema constraints: got %v want %v", schema, expected)
	}
}
