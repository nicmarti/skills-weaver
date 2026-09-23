package llm

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// LegacyAnthropicToolParams is a temporary wire adapter for the unmigrated DM
// loop. Phase 4 removes it when the main agent calls Client directly.
func LegacyAnthropicToolParams(definitions []ToolDefinition) ([]anthropic.ToolUnionParam, error) {
	tools := make([]anthropic.ToolUnionParam, 0, len(definitions))
	for _, def := range definitions {
		properties, required, extras, err := legacySchemaParts(def.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool %q: %w", def.Name, err)
		}
		tool := anthropic.ToolParam{
			Name: def.Name, Description: param.NewOpt(def.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object", Properties: properties, Required: required, ExtraFields: extras,
			},
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &tool})
	}
	return tools, nil
}

// LegacyAnthropicBetaToolParams uses the same schema extraction for the old
// Advisor path. Phase 6 removes this and the beta-specific tool conversion.
func LegacyAnthropicBetaToolParams(definitions []ToolDefinition) ([]anthropic.BetaToolUnionParam, error) {
	tools := make([]anthropic.BetaToolUnionParam, 0, len(definitions))
	for _, def := range definitions {
		properties, required, extras, err := legacySchemaParts(def.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool %q: %w", def.Name, err)
		}
		tool := anthropic.BetaToolParam{
			Name: def.Name, Description: param.NewOpt(def.Description),
			InputSchema: anthropic.BetaToolInputSchemaParam{
				Type: "object", Properties: properties, Required: required, ExtraFields: extras,
			},
		}
		tools = append(tools, anthropic.BetaToolUnionParam{OfTool: &tool})
	}
	return tools, nil
}

// Schema extensions that are absent from the old SDK's typed structure remain
// in ExtraFields, preserving complete JSON Schema constraints on the wire.
func legacySchemaParts(schema map[string]interface{}) (interface{}, []string, map[string]interface{}, error) {
	if schema["type"] != "object" {
		return nil, nil, nil, fmt.Errorf("expected object input schema")
	}
	properties := schema["properties"]
	var required []string
	switch values := schema["required"].(type) {
	case nil:
	case []string:
		required = values
	case []interface{}:
		for _, value := range values {
			name, ok := value.(string)
			if !ok {
				return nil, nil, nil, fmt.Errorf("non-string required property")
			}
			required = append(required, name)
		}
	default:
		return nil, nil, nil, fmt.Errorf("invalid required list")
	}
	extras := make(map[string]interface{}, len(schema))
	for key, value := range schema {
		if key != "type" && key != "properties" && key != "required" {
			extras[key] = value
		}
	}
	return properties, required, extras, nil
}
