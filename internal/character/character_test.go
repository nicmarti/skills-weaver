package character

import (
	"os"
	"path/filepath"
	"testing"

	"dungeons/internal/data"
)

func getTestDataDir() string {
	return filepath.Join("..", "..", "data")
}

func loadTestGameData(t *testing.T) *data.GameData {
	gd, err := data.Load(getTestDataDir())
	if err != nil {
		t.Fatalf("Failed to load game data: %v", err)
	}
	return gd
}

func TestNew(t *testing.T) {
	c := New("Aldric", "human", "fighter")

	if c.Name != "Aldric" {
		t.Errorf("Name = %q, want %q", c.Name, "Aldric")
	}
	if c.Race != "human" {
		t.Errorf("Race = %q, want %q", c.Race, "human")
	}
	if c.Class != "fighter" {
		t.Errorf("Class = %q, want %q", c.Class, "fighter")
	}
	if c.Level != 1 {
		t.Errorf("Level = %d, want 1", c.Level)
	}
	if c.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestGenerateAbilitiesStandard(t *testing.T) {
	c := New("Test", "human", "fighter")
	results := c.GenerateAbilities(MethodStandard)

	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// Check all abilities are in valid range (3-18)
	abilities := []int{
		c.Abilities.Strength,
		c.Abilities.Intelligence,
		c.Abilities.Wisdom,
		c.Abilities.Dexterity,
		c.Abilities.Constitution,
		c.Abilities.Charisma,
	}

	for i, score := range abilities {
		if score < 3 || score > 18 {
			t.Errorf("Ability %d = %d, want 3-18", i, score)
		}
	}
}

func TestGenerateAbilitiesClassic(t *testing.T) {
	c := New("Test", "human", "fighter")
	results := c.GenerateAbilities(MethodClassic)

	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// Classic method uses 3d6, so range is still 3-18
	abilities := []int{
		c.Abilities.Strength,
		c.Abilities.Intelligence,
		c.Abilities.Wisdom,
		c.Abilities.Dexterity,
		c.Abilities.Constitution,
		c.Abilities.Charisma,
	}

	for i, score := range abilities {
		if score < 3 || score > 18 {
			t.Errorf("Ability %d = %d, want 3-18", i, score)
		}
	}
}

func TestApplyRacialModifiers(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		race     string
		checkDEX int // Expected change to DEX
		checkCON int // Expected change to CON
	}{
		{"human", 0, 0},      // No modifiers
		{"elf", 1, -1},       // +1 DEX, -1 CON
		{"dwarf", 0, 1},      // +1 CON, -1 CHA
		{"halfling", 1, 0},   // +1 DEX, -1 STR
	}

	for _, tt := range tests {
		t.Run(tt.race, func(t *testing.T) {
			c := New("Test", tt.race, "fighter")
			// Set known base values
			c.Abilities.Strength = 10
			c.Abilities.Intelligence = 10
			c.Abilities.Wisdom = 10
			c.Abilities.Dexterity = 10
			c.Abilities.Constitution = 10
			c.Abilities.Charisma = 10

			err := c.ApplyRacialModifiers(gd)
			if err != nil {
				t.Fatalf("ApplyRacialModifiers() error = %v", err)
			}

			if c.Abilities.Dexterity != 10+tt.checkDEX {
				t.Errorf("DEX = %d, want %d", c.Abilities.Dexterity, 10+tt.checkDEX)
			}
			if c.Abilities.Constitution != 10+tt.checkCON {
				t.Errorf("CON = %d, want %d", c.Abilities.Constitution, 10+tt.checkCON)
			}
		})
	}
}

func TestCalculateModifiers(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{3, -3},
		{4, -2},
		{5, -2},
		{6, -1},
		{7, -1},
		{8, -1},
		{9, 0},
		{10, 0},
		{11, 0},
		{12, 0},
		{13, 1},
		{14, 1},
		{15, 1},
		{16, 2},
		{17, 2},
		{18, 3},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.score)), func(t *testing.T) {
			c := New("Test", "human", "fighter")
			c.Abilities.Strength = tt.score
			c.CalculateModifiers()

			if c.Modifiers.Strength != tt.want {
				t.Errorf("Modifier for %d = %d, want %d", tt.score, c.Modifiers.Strength, tt.want)
			}
		})
	}
}

func TestRollHitPoints(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		class   string
		maxHP   int // Max possible HP at level 1
	}{
		{"fighter", 8},    // d8
		{"cleric", 6},     // d6
		{"magic-user", 4}, // d4
		{"thief", 4},      // d4
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			c := New("Test", "human", tt.class)
			c.Modifiers.Constitution = 0 // No CON modifier

			err := c.RollHitPoints(gd)
			if err != nil {
				t.Fatalf("RollHitPoints() error = %v", err)
			}

			// At level 1, HP should be max die + CON mod
			if c.HitPoints != tt.maxHP {
				t.Errorf("HP = %d, want %d", c.HitPoints, tt.maxHP)
			}
			if c.MaxHitPoints != tt.maxHP {
				t.Errorf("MaxHP = %d, want %d", c.MaxHitPoints, tt.maxHP)
			}
		})
	}
}

func TestRollHitPointsWithCON(t *testing.T) {
	gd := loadTestGameData(t)

	c := New("Test", "human", "fighter")
	c.Modifiers.Constitution = 2 // +2 CON modifier

	err := c.RollHitPoints(gd)
	if err != nil {
		t.Fatalf("RollHitPoints() error = %v", err)
	}

	// Fighter d8 + 2 CON = 10 HP
	if c.HitPoints != 10 {
		t.Errorf("HP = %d, want 10", c.HitPoints)
	}
}

func TestValidate(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		name    string
		race    string
		class   string
		wantErr bool
	}{
		{"Valid human fighter", "human", "fighter", false},
		{"Valid elf magic-user", "elf", "magic-user", false},
		{"Valid dwarf cleric", "dwarf", "cleric", false},
		{"Invalid elf cleric", "elf", "cleric", true},      // Elves can't be clerics
		{"Invalid dwarf magic-user", "dwarf", "magic-user", true}, // Dwarves can't be magic-users
		{"Invalid halfling cleric", "halfling", "cleric", true},
		{"Unknown race", "orc", "fighter", true},
		{"Unknown class", "human", "paladin", true},
		{"Empty name", "", "human", true}, // Will be tested separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charName := "Test"
			if tt.name == "Empty name" {
				charName = ""
			}
			c := New(charName, tt.race, tt.class)
			err := c.Validate(gd)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	gd := loadTestGameData(t)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "character_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and save character
	c := New("Aldric", "human", "fighter")
	c.GenerateAbilities(MethodStandard)
	c.ApplyRacialModifiers(gd)
	c.CalculateModifiers()
	c.RollHitPoints(gd)
	c.RollStartingGold(gd)

	err = c.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load character
	loaded, err := Load(filepath.Join(tmpDir, "aldric.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify
	if loaded.Name != c.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, c.Name)
	}
	if loaded.Race != c.Race {
		t.Errorf("Race = %q, want %q", loaded.Race, c.Race)
	}
	if loaded.HitPoints != c.HitPoints {
		t.Errorf("HitPoints = %d, want %d", loaded.HitPoints, c.HitPoints)
	}
}

func TestListCharacters(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "character_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple characters
	names := []string{"Aldric", "Lyra", "Gorim"}
	for _, name := range names {
		c := New(name, "human", "fighter")
		c.Save(tmpDir)
	}

	// List
	characters, err := ListCharacters(tmpDir)
	if err != nil {
		t.Fatalf("ListCharacters() error = %v", err)
	}

	if len(characters) != 3 {
		t.Errorf("ListCharacters() returned %d, want 3", len(characters))
	}
}

func TestDelete(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "character_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create and save character
	c := New("ToDelete", "human", "fighter")
	c.Save(tmpDir)

	// Verify it exists
	_, err = Load(filepath.Join(tmpDir, "todelete.json"))
	if err != nil {
		t.Fatalf("Character should exist before delete")
	}

	// Delete
	err = Delete(tmpDir, "ToDelete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = Load(filepath.Join(tmpDir, "todelete.json"))
	if err == nil {
		t.Error("Character should not exist after delete")
	}
}

func TestToMarkdown(t *testing.T) {
	gd := loadTestGameData(t)

	c := New("Aldric", "human", "fighter")
	c.Abilities = AbilityScores{
		Strength:     16,
		Intelligence: 10,
		Wisdom:       12,
		Dexterity:    14,
		Constitution: 15,
		Charisma:     9,
	}
	c.CalculateModifiers()
	c.HitPoints = 9
	c.MaxHitPoints = 9
	c.ArmorClass = 15
	c.Gold = 120

	md := c.ToMarkdown(gd)

	// Check it contains expected content
	if md == "" {
		t.Error("ToMarkdown() returned empty string")
	}

	// Should contain character name
	if !contains(md, "Aldric") {
		t.Error("Markdown should contain character name")
	}

	// Should contain race and class
	if !contains(md, "Humain") || !contains(md, "Guerrier") {
		t.Error("Markdown should contain race and class")
	}
}

func TestRollStartingGold(t *testing.T) {
	gd := loadTestGameData(t)

	c := New("Test", "human", "fighter")
	err := c.RollStartingGold(gd)
	if err != nil {
		t.Fatalf("RollStartingGold() error = %v", err)
	}

	// 3d6 * 10 = 30 to 180
	if c.Gold < 30 || c.Gold > 180 {
		t.Errorf("Gold = %d, want 30-180", c.Gold)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
