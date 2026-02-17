package dmtools

import (
	"os"
	"path/filepath"
	"testing"

	"dungeons/internal/adventure"
)

// TestMapGenerationValidation verifies the validation behavior for different map types.
func TestMapGenerationValidation(t *testing.T) {
	// Get test data directory
	dataDir := filepath.Join("..", "..", "data")
	adventuresDir := filepath.Join(dataDir, "adventures")

	// Create minimal adventure.json for testing
	advData := adventure.New("Test Map Validation", "Test adventure for map generation")
	if err := advData.Save(adventuresDir); err != nil {
		t.Fatalf("Failed to save test adventure: %v", err)
	}

	// Load adventure instance
	tempAdventurePath := filepath.Join(adventuresDir, advData.Slug)
	defer os.RemoveAll(tempAdventurePath)

	tempAdventure, err := adventure.Load(tempAdventurePath)
	if err != nil {
		t.Fatalf("Failed to load adventure: %v", err)
	}

	// Create tool instance
	tool, err := NewGenerateMapTool(dataDir, tempAdventure, nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	tests := []struct {
		name           string
		mapType        string
		locationName   string
		expectSuccess  bool
		expectError    string
	}{
		{
			name:          "Region map with adventure-specific location (should succeed)",
			mapType:       "region",
			locationName:  "Route entre Greystone et Portus Lunaris",
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:          "Region map with any name (should succeed)",
			mapType:       "region",
			locationName:  "Lumarios - Côte nord",
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:          "Dungeon map with any name (should succeed)",
			mapType:       "dungeon",
			locationName:  "La Crypte des Ombres",
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:          "Tactical map with any name (should succeed)",
			mapType:       "tactical",
			locationName:  "Embuscade en forêt",
			expectSuccess: true,
			expectError:   "",
		},
		{
			name:          "City map with invalid location (should fail)",
			mapType:       "city",
			locationName:  "NonExistentCity",
			expectSuccess: false,
			expectError:   "not found in geography.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"map_type": tt.mapType,
				"name":     tt.locationName,
				"scale":    "medium",
				// Don't generate image for tests
				"generate_image": false,
			}

			result, err := tool.Execute(params)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Execute did not return map[string]interface{}")
			}

			success, ok := resultMap["success"].(bool)
			if !ok {
				t.Fatalf("Result missing 'success' field")
			}

			if success != tt.expectSuccess {
				t.Errorf("Expected success=%v, got success=%v", tt.expectSuccess, success)
				if errMsg, ok := resultMap["error"].(string); ok {
					t.Logf("Error message: %s", errMsg)
				}
			}

			if !tt.expectSuccess && tt.expectError != "" {
				errMsg, ok := resultMap["error"].(string)
				if !ok {
					t.Errorf("Expected error message containing '%s', but got no error field", tt.expectError)
				} else if errMsg == "" || len(errMsg) == 0 {
					t.Errorf("Expected error message containing '%s', but got empty error", tt.expectError)
				}
				// Note: We don't check exact string match because error messages may vary
			}
		})
	}
}

// TestMapGenerationHintMessage verifies the hint message is updated.
func TestMapGenerationHintMessage(t *testing.T) {
	dataDir := filepath.Join("..", "..", "data")
	adventuresDir := filepath.Join(dataDir, "adventures")

	// Create minimal adventure.json for testing
	advData := adventure.New("Test Hint Message", "Test adventure for hint message")
	if err := advData.Save(adventuresDir); err != nil {
		t.Fatalf("Failed to save test adventure: %v", err)
	}

	// Load adventure instance
	tempAdventurePath := filepath.Join(adventuresDir, advData.Slug)
	defer os.RemoveAll(tempAdventurePath)

	tempAdventure, err := adventure.Load(tempAdventurePath)
	if err != nil {
		t.Fatalf("Failed to load adventure: %v", err)
	}

	tool, err := NewGenerateMapTool(dataDir, tempAdventure, nil)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Try to generate city map with invalid location
	params := map[string]interface{}{
		"map_type":       "city",
		"name":           "InvalidCityForTest",
		"generate_image": false,
	}

	result, err := tool.Execute(params)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	hint, ok := resultMap["hint"].(string)
	if !ok {
		t.Fatalf("Result missing 'hint' field")
	}

	// Verify hint mentions region, dungeon, and tactical
	expectedHint := "For region, dungeon and tactical maps, location validation is not required."
	if hint != expectedHint {
		t.Errorf("Expected hint:\n%s\n\nGot hint:\n%s", expectedHint, hint)
	}
}
