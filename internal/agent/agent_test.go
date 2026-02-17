package agent

import (
	"os"
	"path/filepath"
	"testing"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
)

// TestBuildSystemPrompt_MissingFile tests that buildSystemPrompt returns an error
// when the dungeon-master.md file is missing.
func TestBuildSystemPrompt_MissingFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory where dungeon-master.md doesn't exist
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create minimal adventure context
	ctx := &AdventureContext{
		Adventure: &adventure.Adventure{
			Name:        "Test Adventure",
			Description: "Test Description",
		},
		Party: &adventure.Party{
			Characters: []string{},
		},
		Inventory: &adventure.SharedInventory{
			Gold: 100,
		},
		State: &adventure.GameState{
			CurrentLocation: "Test Location",
		},
		Characters:    []*character.Character{},
		RecentJournal: []adventure.JournalEntry{},
	}

	// Create agent with persona loader
	agent := &Agent{
		adventureCtx:  ctx,
		personaLoader: NewPersonaLoader(),
	}

	// Call buildSystemPrompt - should return error
	_, err = agent.buildSystemPrompt()
	if err == nil {
		t.Error("Expected error when dungeon-master.md is missing, got nil")
	}

	// Verify error message mentions the persona
	if err != nil && err.Error() != "" {
		expectedSubstring := "dungeon-master"
		if !contains(err.Error(), expectedSubstring) {
			t.Errorf("Expected error message to contain '%s', got: %v", expectedSubstring, err)
		}
	}
}

// TestBuildSystemPrompt_Success tests that buildSystemPrompt works when the file exists.
func TestBuildSystemPrompt_Success(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude/agents directory structure
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create dungeon-master.md file
	dmFile := filepath.Join(agentsDir, "dungeon-master.md")
	dmContent := "Tu es le Maître du Donjon."
	if err := os.WriteFile(dmFile, []byte(dmContent), 0644); err != nil {
		t.Fatalf("Failed to write dungeon-master.md: %v", err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create minimal adventure context
	ctx := &AdventureContext{
		Adventure: &adventure.Adventure{
			Name:        "Test Adventure",
			Description: "Test Description",
		},
		Party: &adventure.Party{
			Characters: []string{},
		},
		Inventory: &adventure.SharedInventory{
			Gold: 100,
		},
		State: &adventure.GameState{
			CurrentLocation: "Test Location",
		},
		Characters:    []*character.Character{},
		RecentJournal: []adventure.JournalEntry{},
	}

	// Create agent with persona loader
	agent := &Agent{
		adventureCtx:  ctx,
		personaLoader: NewPersonaLoader(),
	}

	// Call buildSystemPrompt - should succeed
	systemPrompt, err := agent.buildSystemPrompt()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify system prompt contains DM content
	if !contains(systemPrompt, dmContent) {
		t.Errorf("Expected system prompt to contain DM content '%s'", dmContent)
	}

	// Verify system prompt contains adventure info
	if !contains(systemPrompt, "Test Adventure") {
		t.Error("Expected system prompt to contain adventure name")
	}
}

// contains is a helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}