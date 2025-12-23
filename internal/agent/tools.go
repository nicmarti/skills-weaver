package agent

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
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
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ToAnthropicTools converts the registry to Anthropic API tool format.
func (tr *ToolRegistry) ToAnthropicTools() []anthropic.ToolParam {
	tools := make([]anthropic.ToolParam, 0, len(tr.tools))

	for _, tool := range tr.tools {
		schema := tool.InputSchema()

		// Extract properties and required fields from schema map
		properties := schema["properties"]
		required := []string{}
		if req, ok := schema["required"].([]string); ok {
			required = req
		} else if req, ok := schema["required"].([]interface{}); ok {
			for _, r := range req {
				if str, ok := r.(string); ok {
					required = append(required, str)
				}
			}
		}

		tools = append(tools, anthropic.ToolParam{
			Name:        tool.Name(),
			Description: param.NewOpt(tool.Description()),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	return tools
}

// ToAnthropicToolsParam converts the registry to ToolUnionParam format for API calls.
func (tr *ToolRegistry) ToAnthropicToolsParam() []anthropic.ToolUnionParam {
	toolParams := tr.ToAnthropicTools()
	tools := make([]anthropic.ToolUnionParam, len(toolParams))
	for i, toolParam := range toolParams {
		tools[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
	}
	return tools
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
