package dmtools

import (
	"testing"
)

func TestIsAbilityScore(t *testing.T) {
	tests := []struct {
		stat     string
		expected bool
	}{
		{"strength", true},
		{"dexterity", true},
		{"constitution", true},
		{"intelligence", true},
		{"wisdom", true},
		{"charisma", true},
		{"max_hp", false},
		{"armor_class", false},
		{"spell_save_dc", false},
		{"spell_attack_bonus", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.stat, func(t *testing.T) {
			got := isAbilityScore(tt.stat)
			if got != tt.expected {
				t.Errorf("isAbilityScore(%q) = %v, want %v", tt.stat, got, tt.expected)
			}
		})
	}
}

func TestUpdateCharacterStatToolCreation(t *testing.T) {
	// Test that the tool can be created (nil adventure for schema validation only)
	tool := NewUpdateCharacterStatTool(nil)

	if tool.Name() != "update_character_stat" {
		t.Errorf("expected name 'update_character_stat', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	// Verify required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("expected required field in schema")
	}
	requiredMap := map[string]bool{}
	for _, r := range required {
		requiredMap[r] = true
	}
	if !requiredMap["character_name"] {
		t.Error("character_name should be required")
	}
	if !requiredMap["stat"] {
		t.Error("stat should be required")
	}
	if !requiredMap["value"] {
		t.Error("value should be required")
	}

	// Verify stat enum includes all expected values
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties in schema")
	}
	statProp, ok := props["stat"].(map[string]interface{})
	if !ok {
		t.Fatal("expected stat property")
	}
	enumValues, ok := statProp["enum"].([]string)
	if !ok {
		t.Fatal("expected enum in stat property")
	}
	expectedStats := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma", "max_hp", "armor_class", "spell_save_dc", "spell_attack_bonus"}
	if len(enumValues) != len(expectedStats) {
		t.Errorf("expected %d enum values, got %d", len(expectedStats), len(enumValues))
	}
}

func TestLongRestToolCreation(t *testing.T) {
	tool := NewLongRestTool(nil)

	if tool.Name() != "long_rest" {
		t.Errorf("expected name 'long_rest', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	// long_rest has no required fields (character_name is optional)
	_, hasRequired := schema["required"]
	if hasRequired {
		t.Error("long_rest should not have required fields (character_name is optional)")
	}
}
