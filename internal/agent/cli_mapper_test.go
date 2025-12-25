package agent

import (
	"testing"
)

func TestToolToCLICommand_RollDice(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "simple d20",
			params:   map[string]interface{}{"notation": "d20"},
			expected: "./sw-dice roll d20",
		},
		{
			name:     "complex notation",
			params:   map[string]interface{}{"notation": "2d6+3"},
			expected: "./sw-dice roll 2d6+3",
		},
		{
			name:     "keep highest",
			params:   map[string]interface{}{"notation": "4d6kh3"},
			expected: "./sw-dice roll 4d6kh3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand("roll_dice", tt.params)
			if result != tt.expected {
				t.Errorf("ToolToCLICommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolToCLICommand_GetMonster(t *testing.T) {
	params := map[string]interface{}{
		"monster_id": "goblin",
	}
	expected := "./sw-monster show goblin"

	result := ToolToCLICommand("get_monster", params)
	if result != expected {
		t.Errorf("ToolToCLICommand() = %v, want %v", result, expected)
	}
}

func TestToolToCLICommand_GenerateTreasure(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "lowercase type",
			params:   map[string]interface{}{"treasure_type": "r"},
			expected: "./sw-treasure generate R",
		},
		{
			name:     "uppercase type",
			params:   map[string]interface{}{"treasure_type": "A"},
			expected: "./sw-treasure generate A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand("generate_treasure", tt.params)
			if result != tt.expected {
				t.Errorf("ToolToCLICommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolToCLICommand_GenerateNPC(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "full parameters",
			params:   map[string]interface{}{"race": "human", "gender": "f", "occupation": "aubergiste", "attitude": "friendly"},
			expected: "./sw-npc generate --race=human --gender=f --occupation=aubergiste --attitude=friendly",
		},
		{
			name:     "minimal parameters",
			params:   map[string]interface{}{},
			expected: "./sw-npc generate",
		},
		{
			name:     "partial parameters",
			params:   map[string]interface{}{"race": "dwarf", "occupation": "skilled"},
			expected: "./sw-npc generate --race=dwarf --occupation=skilled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand("generate_npc", tt.params)
			if result != tt.expected {
				t.Errorf("ToolToCLICommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolToCLICommand_GenerateImage(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name: "character image",
			params: map[string]interface{}{
				"type":           "character",
				"character_name": "Aldric",
				"style":          "epic",
			},
			expected: "./sw-image character \"Aldric\" --style=epic",
		},
		{
			name: "npc image",
			params: map[string]interface{}{
				"type":       "npc",
				"race":       "dwarf",
				"gender":     "m",
				"occupation": "forgeron",
			},
			expected: "./sw-image npc --race=dwarf --gender=m --occupation=forgeron",
		},
		{
			name: "scene image",
			params: map[string]interface{}{
				"type":        "scene",
				"description": "Combat contre des gobelins",
				"scene_type":  "battle",
			},
			expected: "./sw-image scene \"Combat contre des gobelins\" --type=battle",
		},
		{
			name: "monster image",
			params: map[string]interface{}{
				"type":         "monster",
				"monster_type": "dragon",
			},
			expected: "./sw-image monster dragon",
		},
		{
			name: "location image",
			params: map[string]interface{}{
				"type":          "location",
				"location_type": "dungeon",
				"name":          "La Crypte",
			},
			expected: "./sw-image location dungeon \"La Crypte\"",
		},
		{
			name: "item image",
			params: map[string]interface{}{
				"type":        "item",
				"item_type":   "weapon",
				"description": "épée flamboyante",
			},
			expected: "./sw-image item weapon \"épée flamboyante\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand("generate_image", tt.params)
			if result != tt.expected {
				t.Errorf("ToolToCLICommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolToCLICommand_GenerateMap(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name: "city map basic",
			params: map[string]interface{}{
				"type": "city",
				"name": "Cordova",
			},
			expected: "./sw-map generate city \"Cordova\"",
		},
		{
			name: "city map with kingdom",
			params: map[string]interface{}{
				"type":    "city",
				"name":    "Cordova",
				"kingdom": "valdorine",
			},
			expected: "./sw-map generate city \"Cordova\" --kingdom=valdorine",
		},
		{
			name: "city map with features array",
			params: map[string]interface{}{
				"type":     "city",
				"name":     "Cordova",
				"features": []interface{}{"Port", "Marché", "Taverne"},
			},
			expected: "./sw-map generate city \"Cordova\" --features=\"Port,Marché,Taverne\"",
		},
		{
			name: "city map with features string",
			params: map[string]interface{}{
				"type":     "city",
				"name":     "Cordova",
				"features": "Port,Marché",
			},
			expected: "./sw-map generate city \"Cordova\" --features=\"Port,Marché\"",
		},
		{
			name: "city map with image generation",
			params: map[string]interface{}{
				"type":           "city",
				"name":           "Cordova",
				"generate_image": true,
			},
			expected: "./sw-map generate city \"Cordova\" --generate-image",
		},
		{
			name: "region map",
			params: map[string]interface{}{
				"type":  "region",
				"name":  "Côte Occidentale",
				"scale": "large",
			},
			expected: "./sw-map generate region \"Côte Occidentale\" --scale=large",
		},
		{
			name: "dungeon map",
			params: map[string]interface{}{
				"type":  "dungeon",
				"name":  "La Crypte",
				"level": float64(1),
			},
			expected: "./sw-map generate dungeon \"La Crypte\" --level=1",
		},
		{
			name: "tactical map",
			params: map[string]interface{}{
				"type":    "tactical",
				"name":    "Embuscade",
				"terrain": "forêt",
				"scene":   "Combat en forêt",
			},
			expected: "./sw-map generate tactical \"Embuscade\" --terrain=forêt --scene=\"Combat en forêt\"",
		},
		{
			name: "complete example",
			params: map[string]interface{}{
				"type":           "city",
				"name":           "Port-Sombre",
				"kingdom":        "valdorine",
				"features":       []interface{}{"Port", "Docks", "Marché aux poissons"},
				"generate_image": true,
			},
			expected: "./sw-map generate city \"Port-Sombre\" --kingdom=valdorine --features=\"Port,Docks,Marché aux poissons\" --generate-image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand("generate_map", tt.params)
			if result != tt.expected {
				t.Errorf("ToolToCLICommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolToCLICommand_NoEquivalent(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		params   map[string]interface{}
	}{
		{
			name:     "log_event has no CLI",
			toolName: "log_event",
			params:   map[string]interface{}{"event_type": "combat", "content": "test"},
		},
		{
			name:     "update_npc_importance has no CLI",
			toolName: "update_npc_importance",
			params:   map[string]interface{}{"npc_name": "Test", "importance": "key"},
		},
		{
			name:     "get_npc_history has no CLI",
			toolName: "get_npc_history",
			params:   map[string]interface{}{"npc_name": "Test"},
		},
		{
			name:     "unknown tool",
			toolName: "unknown_tool",
			params:   map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolToCLICommand(tt.toolName, tt.params)
			if result != "" {
				t.Errorf("ToolToCLICommand() = %v, want empty string", result)
			}
		})
	}
}
