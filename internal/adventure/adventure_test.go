package adventure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Test slugify function (pure function - easy to test)
// =============================================================================

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic lowercase",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "uppercase to lowercase",
			input:    "HELLO WORLD",
			expected: "hello-world",
		},
		{
			name:     "french accents - é",
			input:    "L'épée sacrée",
			expected: "lepee-sacree",
		},
		{
			name:     "french accents - è",
			input:    "Très belle",
			expected: "tres-belle",
		},
		{
			name:     "french accents - ê",
			input:    "Forêt enchantée",
			expected: "foret-enchantee",
		},
		{
			name:     "french accents - à",
			input:    "À la taverne",
			expected: "a-la-taverne",
		},
		{
			name:     "french accents - ù",
			input:    "Où est-ce",
			expected: "ou-est-ce",
		},
		{
			name:     "french accents - î",
			input:    "L'île mystérieuse",
			expected: "lile-mysterieuse",
		},
		{
			name:     "french accents - ô",
			input:    "Le chateau",
			expected: "le-chateau",
		},
		{
			name:     "apostrophe removal",
			input:    "L'aventure d'Aldric",
			expected: "laventure-daldric",
		},
		{
			name:     "double quotes removal",
			input:    "La \"Crypte\" des Ombres",
			expected: "la-crypte-des-ombres",
		},
		{
			name:     "multiple spaces",
			input:    "La   Crypte    des  Ombres",
			expected: "la---crypte----des--ombres",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  La Crypte des Ombres  ",
			expected: "--la-crypte-des-ombres--",
		},
		{
			name:     "mixed french characters",
			input:    "La Crypte des Ombres",
			expected: "la-crypte-des-ombres",
		},
		{
			name:     "complex french title",
			input:    "L'Épée du Chateau Enchante",
			expected: "lepee-du-chateau-enchante",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "Adventure",
			expected: "adventure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Test New function
// =============================================================================

func TestNew(t *testing.T) {
	name := "La Crypte des Ombres"
	description := "Une aventure mystérieuse"

	adv := New(name, description)

	// Check fields are set correctly
	if adv.Name != name {
		t.Errorf("Name = %q, want %q", adv.Name, name)
	}
	if adv.Description != description {
		t.Errorf("Description = %q, want %q", adv.Description, description)
	}
	if adv.Slug != "la-crypte-des-ombres" {
		t.Errorf("Slug = %q, want %q", adv.Slug, "la-crypte-des-ombres")
	}
	if adv.Status != StatusActive {
		t.Errorf("Status = %q, want %q", adv.Status, StatusActive)
	}
	if adv.SessionCount != 0 {
		t.Errorf("SessionCount = %d, want 0", adv.SessionCount)
	}

	// Check ID is generated
	if adv.ID == "" {
		t.Error("ID should not be empty")
	}

	// Check timestamps are set
	if adv.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if adv.LastPlayed.IsZero() {
		t.Error("LastPlayed should be set")
	}

	// Check timestamps are close to now (within 1 second)
	now := time.Now()
	if now.Sub(adv.CreatedAt) > time.Second {
		t.Error("CreatedAt should be close to now")
	}
	if now.Sub(adv.LastPlayed) > time.Second {
		t.Error("LastPlayed should be close to now")
	}
}

// =============================================================================
// Test SetBasePath and BasePath
// =============================================================================

func TestSetBasePathAndBasePath(t *testing.T) {
	adv := New("Test Adventure", "Test description")

	// Initially basePath should be empty
	if adv.BasePath() != "" {
		t.Errorf("BasePath() = %q, want empty string", adv.BasePath())
	}

	// Set basePath
	testPath := "/tmp/adventures/test-adventure"
	adv.SetBasePath(testPath)

	if adv.BasePath() != testPath {
		t.Errorf("BasePath() = %q, want %q", adv.BasePath(), testPath)
	}
}

// =============================================================================
// Test Save and Load
// =============================================================================

func TestSaveAndLoad(t *testing.T) {
	baseDir := t.TempDir()

	// Create an adventure
	original := New("Test Adventure", "This is a test adventure")
	original.Status = StatusActive
	original.SessionCount = 5

	// Save it
	err := original.Save(baseDir)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check directory structure was created
	adventurePath := filepath.Join(baseDir, original.Slug)
	if _, err := os.Stat(adventurePath); os.IsNotExist(err) {
		t.Errorf("Adventure directory was not created: %s", adventurePath)
	}

	charactersPath := filepath.Join(adventurePath, "characters")
	if _, err := os.Stat(charactersPath); os.IsNotExist(err) {
		t.Errorf("Characters directory was not created: %s", charactersPath)
	}

	jsonPath := filepath.Join(adventurePath, "adventure.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("adventure.json was not created: %s", jsonPath)
	}

	// Load it back
	loaded, err := Load(adventurePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare fields
	if loaded.ID != original.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, original.ID)
	}
	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Slug != original.Slug {
		t.Errorf("Slug = %q, want %q", loaded.Slug, original.Slug)
	}
	if loaded.Description != original.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, original.Description)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status = %q, want %q", loaded.Status, original.Status)
	}
	if loaded.SessionCount != original.SessionCount {
		t.Errorf("SessionCount = %d, want %d", loaded.SessionCount, original.SessionCount)
	}
	if loaded.BasePath() != adventurePath {
		t.Errorf("BasePath() = %q, want %q", loaded.BasePath(), adventurePath)
	}
}

func TestLoad_NonExistentDirectory(t *testing.T) {
	_, err := Load("/nonexistent/path")
	if err == nil {
		t.Error("Load() should return error for non-existent path")
	}
}

// =============================================================================
// Test LoadByName
// =============================================================================

func TestLoadByName(t *testing.T) {
	baseDir := t.TempDir()

	// Create and save an adventure
	adv := New("La Crypte des Ombres", "Test")
	if err := adv.Save(baseDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	tests := []struct {
		name      string
		searchFor string
		wantErr   bool
	}{
		{
			name:      "find by exact name",
			searchFor: "La Crypte des Ombres",
			wantErr:   false,
		},
		{
			name:      "find by slug",
			searchFor: "la-crypte-des-ombres",
			wantErr:   false,
		},
		{
			name:      "find by different case",
			searchFor: "LA CRYPTE DES OMBRES",
			wantErr:   false,
		},
		{
			name:      "not found",
			searchFor: "Non-existent Adventure",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaded, err := LoadByName(baseDir, tt.searchFor)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if loaded.Name != adv.Name {
					t.Errorf("LoadByName() Name = %q, want %q", loaded.Name, adv.Name)
				}
			}
		})
	}
}

func TestLoadByName_EmptyDirectory(t *testing.T) {
	baseDir := t.TempDir()

	_, err := LoadByName(baseDir, "Non-existent")
	if err == nil {
		t.Error("LoadByName() should return error for non-existent adventure")
	}
}

// =============================================================================
// Test ListAdventures
// =============================================================================

func TestListAdventures(t *testing.T) {
	baseDir := t.TempDir()

	// Test empty directory
	adventures, err := ListAdventures(baseDir)
	if err != nil {
		t.Fatalf("ListAdventures() error = %v", err)
	}
	if len(adventures) != 0 {
		t.Errorf("ListAdventures() returned %d adventures, want 0", len(adventures))
	}

	// Create multiple adventures
	adv1 := New("Adventure One", "First")
	adv2 := New("Adventure Two", "Second")
	adv3 := New("Adventure Three", "Third")

	if err := adv1.Save(baseDir); err != nil {
		t.Fatalf("Save() adv1 error = %v", err)
	}
	if err := adv2.Save(baseDir); err != nil {
		t.Fatalf("Save() adv2 error = %v", err)
	}
	if err := adv3.Save(baseDir); err != nil {
		t.Fatalf("Save() adv3 error = %v", err)
	}

	// List adventures
	adventures, err = ListAdventures(baseDir)
	if err != nil {
		t.Fatalf("ListAdventures() error = %v", err)
	}
	if len(adventures) != 3 {
		t.Errorf("ListAdventures() returned %d adventures, want 3", len(adventures))
	}

	// Verify names are present
	names := make(map[string]bool)
	for _, a := range adventures {
		names[a.Name] = true
	}
	if !names["Adventure One"] || !names["Adventure Two"] || !names["Adventure Three"] {
		t.Error("ListAdventures() missing expected adventure names")
	}
}

func TestListAdventures_NonExistentDirectory(t *testing.T) {
	adventures, err := ListAdventures("/nonexistent/path")
	if err != nil {
		t.Fatalf("ListAdventures() should not error for non-existent dir, got %v", err)
	}
	if len(adventures) != 0 {
		t.Errorf("ListAdventures() returned %d adventures, want 0", len(adventures))
	}
}

func TestListAdventures_SkipsInvalidDirectories(t *testing.T) {
	baseDir := t.TempDir()

	// Create a valid adventure
	adv := New("Valid Adventure", "Test")
	if err := adv.Save(baseDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create an invalid directory (no adventure.json)
	invalidPath := filepath.Join(baseDir, "invalid-adventure")
	if err := os.MkdirAll(invalidPath, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Create a file (not a directory)
	filePath := filepath.Join(baseDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// List should return only the valid adventure
	adventures, err := ListAdventures(baseDir)
	if err != nil {
		t.Fatalf("ListAdventures() error = %v", err)
	}
	if len(adventures) != 1 {
		t.Errorf("ListAdventures() returned %d adventures, want 1", len(adventures))
	}
	if adventures[0].Name != "Valid Adventure" {
		t.Errorf("ListAdventures() Name = %q, want %q", adventures[0].Name, "Valid Adventure")
	}
}

// =============================================================================
// Test Delete
// =============================================================================

func TestDelete(t *testing.T) {
	baseDir := t.TempDir()

	// Create and save an adventure
	adv := New("To Delete", "This will be deleted")
	if err := adv.Save(baseDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	adventurePath := filepath.Join(baseDir, adv.Slug)

	// Verify it exists
	if _, err := os.Stat(adventurePath); os.IsNotExist(err) {
		t.Fatal("Adventure directory should exist before deletion")
	}

	// Delete it
	err := Delete(baseDir, "To Delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(adventurePath); !os.IsNotExist(err) {
		t.Error("Adventure directory should not exist after deletion")
	}
}

func TestDelete_NonExistent(t *testing.T) {
	baseDir := t.TempDir()

	err := Delete(baseDir, "Non-existent Adventure")
	if err == nil {
		t.Error("Delete() should return error for non-existent adventure")
	}
}

// =============================================================================
// Test UpdateLastPlayed
// =============================================================================

func TestUpdateLastPlayed(t *testing.T) {
	adv := New("Test", "Test")

	original := adv.LastPlayed

	// Wait a tiny bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	adv.UpdateLastPlayed()

	if !adv.LastPlayed.After(original) {
		t.Error("UpdateLastPlayed() should update LastPlayed to a later time")
	}

	// Should be close to now
	now := time.Now()
	if now.Sub(adv.LastPlayed) > time.Second {
		t.Error("UpdateLastPlayed() should set LastPlayed close to now")
	}
}

// =============================================================================
// Test ToMarkdown
// =============================================================================

func TestToMarkdown(t *testing.T) {
	adv := New("Test Adventure", "A test description")
	adv.Status = StatusActive
	adv.SessionCount = 3

	markdown := adv.ToMarkdown()

	// Check required sections are present
	if !strings.Contains(markdown, "# Test Adventure") {
		t.Error("ToMarkdown() should contain adventure name as header")
	}
	if !strings.Contains(markdown, "A test description") {
		t.Error("ToMarkdown() should contain description")
	}
	if !strings.Contains(markdown, "## Informations") {
		t.Error("ToMarkdown() should contain Informations section")
	}
	if !strings.Contains(markdown, "active") {
		t.Error("ToMarkdown() should contain status")
	}
	if !strings.Contains(markdown, "3") {
		t.Error("ToMarkdown() should contain session count")
	}
}

func TestToMarkdown_NoDescription(t *testing.T) {
	adv := New("Test Adventure", "")

	markdown := adv.ToMarkdown()

	// Should still work without description
	if !strings.Contains(markdown, "# Test Adventure") {
		t.Error("ToMarkdown() should contain adventure name")
	}
	if !strings.Contains(markdown, "## Informations") {
		t.Error("ToMarkdown() should contain Informations section")
	}
}

// =============================================================================
// Test LoadState and SaveState
// =============================================================================

func TestLoadState_DefaultState(t *testing.T) {
	baseDir := t.TempDir()

	adv := New("Test", "Test")
	if err := adv.Save(baseDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load state when no state.json exists
	state, err := adv.LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	// Check default values
	if state.CurrentLocation != "Début de l'aventure" {
		t.Errorf("CurrentLocation = %q, want %q", state.CurrentLocation, "Début de l'aventure")
	}
	if state.Time.Day != 1 || state.Time.Hour != 8 || state.Time.Minute != 0 {
		t.Errorf("Time = {%d, %d, %d}, want {1, 8, 0}", state.Time.Day, state.Time.Hour, state.Time.Minute)
	}
	if len(state.Quests) != 0 {
		t.Errorf("Quests length = %d, want 0", len(state.Quests))
	}
	if state.Flags == nil {
		t.Error("Flags should be initialized")
	}
	if state.Variables == nil {
		t.Error("Variables should be initialized")
	}
}

func TestSaveStateAndLoadState(t *testing.T) {
	baseDir := t.TempDir()

	adv := New("Test", "Test")
	if err := adv.Save(baseDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create a custom state
	originalState := &GameState{
		CurrentLocation: "The Tavern",
		Time:            GameTime{Day: 3, Hour: 14, Minute: 30},
		Quests: []Quest{
			{
				ID:          "quest-1",
				Name:        "Find the sword",
				Description: "Locate the legendary sword",
				Status:      "active",
				AddedAt:     "2024-01-01T10:00:00Z",
			},
		},
		Flags: map[string]bool{
			"tavern_visited": true,
			"met_wizard":     false,
		},
		Variables: map[string]string{
			"hero_name": "Aldric",
			"gold":      "100",
		},
	}

	// Save the state
	err := adv.SaveState(originalState)
	if err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Load it back
	loadedState, err := adv.LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	// Compare
	if loadedState.CurrentLocation != originalState.CurrentLocation {
		t.Errorf("CurrentLocation = %q, want %q", loadedState.CurrentLocation, originalState.CurrentLocation)
	}
	if loadedState.Time != originalState.Time {
		t.Errorf("Time = %+v, want %+v", loadedState.Time, originalState.Time)
	}
	if len(loadedState.Quests) != len(originalState.Quests) {
		t.Errorf("Quests length = %d, want %d", len(loadedState.Quests), len(originalState.Quests))
	}
	if len(loadedState.Quests) > 0 {
		if loadedState.Quests[0].Name != originalState.Quests[0].Name {
			t.Errorf("Quest Name = %q, want %q", loadedState.Quests[0].Name, originalState.Quests[0].Name)
		}
	}
	if loadedState.Flags["tavern_visited"] != true {
		t.Error("Flags['tavern_visited'] should be true")
	}
	if loadedState.Variables["hero_name"] != "Aldric" {
		t.Errorf("Variables['hero_name'] = %q, want %q", loadedState.Variables["hero_name"], "Aldric")
	}
}
