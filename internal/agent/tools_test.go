package agent

import (
	"testing"
)

// mockTool is a simple test tool implementation
type mockTool struct {
	name        string
	description string
	schema      map[string]interface{}
}

func (t *mockTool) Name() string                                        { return t.name }
func (t *mockTool) Description() string                                 { return t.description }
func (t *mockTool) InputSchema() map[string]interface{}                 { return t.schema }
func (t *mockTool) Execute(params map[string]interface{}) (interface{}, error) { return nil, nil }

func newMockTool(name string) *mockTool {
	return &mockTool{
		name:        name,
		description: "Mock tool: " + name,
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

func TestToolRegistry_CreateFilteredRegistry(t *testing.T) {
	// Create a registry with several tools
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	registry.Register(newMockTool("roll_dice"))
	registry.Register(newMockTool("get_monster"))
	registry.Register(newMockTool("invoke_agent"))
	registry.Register(newMockTool("log_event"))
	registry.Register(newMockTool("get_spell"))

	tests := []struct {
		name       string
		allowed    []string
		forbidden  []string
		wantCount  int
		wantTools  []string
		dontWant   []string
	}{
		{
			name:       "filter by allowed list only",
			allowed:    []string{"roll_dice", "get_monster"},
			forbidden:  []string{},
			wantCount:  2,
			wantTools:  []string{"roll_dice", "get_monster"},
			dontWant:   []string{"invoke_agent", "log_event", "get_spell"},
		},
		{
			name:       "forbidden takes precedence",
			allowed:    []string{"roll_dice", "get_monster", "invoke_agent"},
			forbidden:  []string{"invoke_agent"},
			wantCount:  2,
			wantTools:  []string{"roll_dice", "get_monster"},
			dontWant:   []string{"invoke_agent"},
		},
		{
			name:       "empty allowed list with forbidden",
			allowed:    []string{},
			forbidden:  []string{"invoke_agent", "log_event"},
			wantCount:  3,
			wantTools:  []string{"roll_dice", "get_monster", "get_spell"},
			dontWant:   []string{"invoke_agent", "log_event"},
		},
		{
			name:       "empty both lists returns all",
			allowed:    []string{},
			forbidden:  []string{},
			wantCount:  5,
			wantTools:  []string{"roll_dice", "get_monster", "invoke_agent", "log_event", "get_spell"},
			dontWant:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := registry.CreateFilteredRegistry(tt.allowed, tt.forbidden)

			// Check count
			if filtered.Count() != tt.wantCount {
				t.Errorf("CreateFilteredRegistry() count = %d, want %d", filtered.Count(), tt.wantCount)
			}

			// Check wanted tools are present
			for _, toolName := range tt.wantTools {
				if _, exists := filtered.Get(toolName); !exists {
					t.Errorf("CreateFilteredRegistry() should include %q", toolName)
				}
			}

			// Check unwanted tools are absent
			for _, toolName := range tt.dontWant {
				if _, exists := filtered.Get(toolName); exists {
					t.Errorf("CreateFilteredRegistry() should NOT include %q", toolName)
				}
			}
		})
	}
}

func TestToolRegistry_Count(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}

	if registry.Count() != 0 {
		t.Errorf("Count() of empty registry = %d, want 0", registry.Count())
	}

	registry.Register(newMockTool("tool1"))
	if registry.Count() != 1 {
		t.Errorf("Count() after 1 registration = %d, want 1", registry.Count())
	}

	registry.Register(newMockTool("tool2"))
	registry.Register(newMockTool("tool3"))
	if registry.Count() != 3 {
		t.Errorf("Count() after 3 registrations = %d, want 3", registry.Count())
	}
}

func TestToolRegistry_Names(t *testing.T) {
	registry := &ToolRegistry{tools: make(map[string]Tool)}
	registry.Register(newMockTool("alpha"))
	registry.Register(newMockTool("beta"))
	registry.Register(newMockTool("gamma"))

	names := registry.Names()

	if len(names) != 3 {
		t.Errorf("Names() returned %d items, want 3", len(names))
	}

	// Check all names are present (order doesn't matter)
	expected := map[string]bool{"alpha": true, "beta": true, "gamma": true}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("Names() returned unexpected name: %q", name)
		}
		delete(expected, name)
	}
	if len(expected) > 0 {
		t.Errorf("Names() missing names: %v", expected)
	}
}
