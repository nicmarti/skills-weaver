package agent

import (
	"encoding/json"
	"fmt"
	"sort"

	"dungeons/internal/llm"
)

// Tool represents a tool that can be called by the agent.
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(params map[string]interface{}) (interface{}, error)
}

// ToolRegistry manages all available tools.
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates a new tool registry with all available tools.
func NewToolRegistry(adventureCtx *AdventureContext) *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]Tool),
	}

	// Tools will be registered by the agent when initializing
	// This allows us to pass adventure context to tools that need it

	return registry
}

// Register adds a tool to the registry.
func (tr *ToolRegistry) Register(tool Tool) {
	tr.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (tr *ToolRegistry) Get(name string) (Tool, bool) {
	tool, exists := tr.tools[name]
	return tool, exists
}

// GetAll returns all registered tools.
func (tr *ToolRegistry) GetAll() []Tool {
	tools := make([]Tool, 0, len(tr.tools))
	for _, name := range tr.Names() {
		tools = append(tools, tr.tools[name])
	}
	return tools
}

// ToolDefinitions provides complete JSON Schemas to the provider-neutral LLM
// boundary. Never rebuild selected fields: constraints and nested schemas must
// reach the Chat function parameters unchanged.
func (tr *ToolRegistry) ToolDefinitions() ([]llm.ToolDefinition, error) {
	definitions := make([]llm.ToolDefinition, 0, len(tr.tools))
	for _, tool := range tr.GetAll() {
		name := tool.Name()
		if name == "" {
			return nil, fmt.Errorf("tool has an empty name")
		}
		schema := tool.InputSchema()
		if schema == nil || schema["type"] != "object" {
			return nil, fmt.Errorf("tool %q requires an object input schema", name)
		}
		if _, ok := schema["properties"].(map[string]interface{}); !ok {
			return nil, fmt.Errorf("tool %q requires an object of properties", name)
		}
		switch required := schema["required"].(type) {
		case nil, []string:
		case []interface{}:
			for _, value := range required {
				if _, ok := value.(string); !ok {
					return nil, fmt.Errorf("tool %q has a non-string required property", name)
				}
			}
		default:
			return nil, fmt.Errorf("tool %q has an invalid required list", name)
		}
		if _, err := json.Marshal(schema); err != nil {
			return nil, fmt.Errorf("tool %q has an unserializable schema: %w", name, err)
		}
		definitions = append(definitions, llm.ToolDefinition{
			Name: name, Description: tool.Description(), Parameters: schema,
		})
	}
	return definitions, nil
}

// ToolUse represents a tool call from Claude.
type ToolUse struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

// ToolResultMessage represents the result of a tool execution.
type ToolResultMessage struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// String returns a string representation of the tool use.
func (tu ToolUse) String() string {
	return fmt.Sprintf("ToolUse{name=%s, id=%s}", tu.Name, tu.ID)
}

// CreateFilteredRegistry creates a new registry containing only tools that pass the filter.
// A tool is included if:
// 1. Its name is in the allowed list (if allowed is non-empty)
// 2. Its name is NOT in the forbidden list
// If allowed is empty, all tools (except forbidden) are included.
func (tr *ToolRegistry) CreateFilteredRegistry(allowed, forbidden []string) *ToolRegistry {
	filtered := &ToolRegistry{
		tools: make(map[string]Tool),
	}

	// Build lookup sets for efficiency
	allowedSet := make(map[string]bool)
	for _, name := range allowed {
		allowedSet[name] = true
	}

	forbiddenSet := make(map[string]bool)
	for _, name := range forbidden {
		forbiddenSet[name] = true
	}

	for name, tool := range tr.tools {
		// Skip forbidden tools
		if forbiddenSet[name] {
			continue
		}

		// If allowed list is specified, tool must be in it
		if len(allowed) > 0 && !allowedSet[name] {
			continue
		}

		filtered.tools[name] = tool
	}

	return filtered
}

// Count returns the number of tools in the registry.
func (tr *ToolRegistry) Count() int {
	return len(tr.tools)
}

// Names returns a slice of all tool names in the registry.
func (tr *ToolRegistry) Names() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
