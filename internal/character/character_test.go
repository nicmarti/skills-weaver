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
	if c.Species != "human" {
		t.Errorf("Species = %q, want %q", c.Species, "human")
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
		species  string
		checkDEX int // Expected change to DEX (D&D 5e)
		checkCON int // Expected change to CON (D&D 5e)
	}{
		{"human", 0, 0},      // No modifiers (variant human)
		{"elf", 2, 0},        // +2 DEX (D&D 5e, no penalties)
		{"dwarf", 0, 2},      // +2 CON (D&D 5e, no penalties)
		{"halfling", 2, 0},   // +2 DEX (D&D 5e, no penalties)
	}

	for _, tt := range tests {
		t.Run(tt.species, func(t *testing.T) {
			c := New("Test", tt.species, "fighter")
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
		// D&D 5e formula: (score - 10) / 2 (integer division in Go)
		{3, -3},  // (3-10)/2 = -7/2 = -3
		{4, -3},  // (4-10)/2 = -6/2 = -3
		{5, -2},  // (5-10)/2 = -5/2 = -2
		{6, -2},  // (6-10)/2 = -4/2 = -2
		{7, -1},  // (7-10)/2 = -3/2 = -1
		{8, -1},  // (8-10)/2 = -2/2 = -1
		{9, 0},   // (9-10)/2 = -1/2 = 0
		{10, 0},  // (10-10)/2 = 0/2 = 0
		{11, 0},  // (11-10)/2 = 1/2 = 0
		{12, 1},  // (12-10)/2 = 2/2 = 1
		{13, 1},  // (13-10)/2 = 3/2 = 1
		{14, 2},  // (14-10)/2 = 4/2 = 2
		{15, 2},  // (15-10)/2 = 5/2 = 2
		{16, 3},  // (16-10)/2 = 6/2 = 3
		{17, 3},  // (17-10)/2 = 7/2 = 3
		{18, 4},  // (18-10)/2 = 8/2 = 4
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

func TestRollHitPointsMaxHP(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		class  string
		maxHP  int // Max possible HP at level 1 (D&D 5e)
	}{
		{"fighter", 10},  // d10
		{"cleric", 8},    // d8
		{"wizard", 6},    // d6
		{"rogue", 8},     // d8
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			c := New("Test", "human", tt.class)
			c.Modifiers.Constitution = 0 // No CON modifier

			err := c.RollHitPoints(gd, true) // maxHP = true
			if err != nil {
				t.Fatalf("RollHitPoints() error = %v", err)
			}

			// With maxHP=true, HP should be max die + CON mod
			if c.HitPoints != tt.maxHP {
				t.Errorf("HP = %d, want %d", c.HitPoints, tt.maxHP)
			}
			if c.MaxHitPoints != tt.maxHP {
				t.Errorf("MaxHP = %d, want %d", c.MaxHitPoints, tt.maxHP)
			}
		})
	}
}

func TestRollHitPointsRandomRoll(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		class string
		minHP int // Minimum HP (1 on die + 0 CON)
		maxHP int // Maximum HP (max die + 0 CON)
	}{
		{"fighter", 1, 10},  // d10: 1-10 (D&D 5e)
		{"cleric", 1, 8},    // d8: 1-8 (D&D 5e)
		{"wizard", 1, 6},    // d6: 1-6 (D&D 5e)
		{"rogue", 1, 8},     // d8: 1-8 (D&D 5e)
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			// Run multiple times to test randomness
			for i := 0; i < 20; i++ {
				c := New("Test", "human", tt.class)
				c.Modifiers.Constitution = 0 // No CON modifier

				err := c.RollHitPoints(gd, false) // maxHP = false (random)
				if err != nil {
					t.Fatalf("RollHitPoints() error = %v", err)
				}

				// With maxHP=false, HP should be within die range
				if c.HitPoints < tt.minHP || c.HitPoints > tt.maxHP {
					t.Errorf("HP = %d, want %d-%d", c.HitPoints, tt.minHP, tt.maxHP)
				}
				if c.MaxHitPoints != c.HitPoints {
					t.Errorf("MaxHP = %d should equal HP = %d", c.MaxHitPoints, c.HitPoints)
				}
			}
		})
	}
}

func TestRollHitPointsWithCON(t *testing.T) {
	gd := loadTestGameData(t)

	c := New("Test", "human", "fighter")
	c.Modifiers.Constitution = 2 // +2 CON modifier

	err := c.RollHitPoints(gd, true) // maxHP = true
	if err != nil {
		t.Fatalf("RollHitPoints() error = %v", err)
	}

	// Fighter d10 + 2 CON = 12 HP (D&D 5e)
	if c.HitPoints != 12 {
		t.Errorf("HP = %d, want 12", c.HitPoints)
	}
}

func TestRollHitPointsMinimumOneHP(t *testing.T) {
	gd := loadTestGameData(t)

	// Wizard with very low CON (D&D 5e)
	c := New("Test", "human", "wizard")
	c.Modifiers.Constitution = -3 // -3 CON modifier

	err := c.RollHitPoints(gd, true) // maxHP = true
	if err != nil {
		t.Fatalf("RollHitPoints() error = %v", err)
	}

	// d6 (6) + (-3) = 3 HP (D&D 5e)
	if c.HitPoints != 3 {
		t.Errorf("HP = %d, want 3", c.HitPoints)
	}
}

func TestRollHitPointsRandomWithLowCON(t *testing.T) {
	gd := loadTestGameData(t)

	// Wizard with low CON, random roll (D&D 5e)
	// d6 (1-6) + (-3) could result in -2 to +3, should clamp to 1
	for i := 0; i < 20; i++ {
		c := New("Test", "human", "wizard")
		c.Modifiers.Constitution = -3

		err := c.RollHitPoints(gd, false)
		if err != nil {
			t.Fatalf("RollHitPoints() error = %v", err)
		}

		// Should never be below 1
		if c.HitPoints < 1 {
			t.Errorf("HP = %d, minimum should be 1", c.HitPoints)
		}
	}
}

func TestValidate(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		name    string
		species string
		class   string
		wantErr bool
	}{
		// D&D 5e: No species/class restrictions - all combinations valid
		{"Valid human fighter", "human", "fighter", false},
		{"Valid elf wizard", "elf", "wizard", false},
		{"Valid dwarf cleric", "dwarf", "cleric", false},
		{"Valid elf cleric", "elf", "cleric", false},      // Valid in D&D 5e
		{"Valid dwarf wizard", "dwarf", "wizard", false}, // Valid in D&D 5e
		{"Valid halfling cleric", "halfling", "cleric", false}, // Valid in D&D 5e
		{"Valid orc fighter", "orc", "fighter", false},    // Orc is a valid species in D&D 5e
		{"Valid human paladin", "human", "paladin", false}, // Paladin exists in D&D 5e
		{"Unknown species", "drow", "fighter", true},      // Drow not in our 9 species
		{"Unknown class", "human", "artificer", true},     // Artificer not in our 12 classes
		{"Empty name", "", "human", true}, // Will be tested separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charName := "Test"
			if tt.name == "Empty name" {
				charName = ""
			}
			c := New(charName, tt.species, tt.class)
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
	c.RollHitPoints(gd, true) // maxHP = true
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
	if loaded.Species != c.Species {
		t.Errorf("Species = %q, want %q", loaded.Species, c.Species)
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

func TestCalculateArmorClass(t *testing.T) {
	gd := loadTestGameData(t)

	tests := []struct {
		name      string
		dexMod    int
		equipment []string
		wantAC    int
	}{
		{
			name:      "Unarmored DEX 10 (mod 0)",
			dexMod:    0,
			equipment: []string{},
			wantAC:    10, // Base AC (D&D 5e)
		},
		{
			name:      "Unarmored DEX 18 (mod +3)",
			dexMod:    3,
			equipment: []string{},
			wantAC:    13, // 10 + 3 (D&D 5e)
		},
		{
			name:      "Unarmored DEX 3 (mod -3)",
			dexMod:    -3,
			equipment: []string{},
			wantAC:    7, // 10 - 3 (D&D 5e)
		},
		{
			name:      "Leather armor DEX 10",
			dexMod:    0,
			equipment: []string{"leather"},
			wantAC:    12, // 10 + 0 + 2 (D&D 5e)
		},
		{
			name:      "Chainmail DEX 10",
			dexMod:    0,
			equipment: []string{"chainmail"},
			wantAC:    14, // 10 + 0 + 4 (D&D 5e)
		},
		{
			name:      "Plate mail DEX 10",
			dexMod:    0,
			equipment: []string{"plate"},
			wantAC:    16, // 10 + 0 + 6 (D&D 5e)
		},
		{
			name:      "Plate mail + shield DEX 10",
			dexMod:    0,
			equipment: []string{"plate", "shield"},
			wantAC:    17, // 10 + 0 + 6 + 1 (D&D 5e)
		},
		{
			name:      "Leather + shield DEX 14 (mod +1)",
			dexMod:    1,
			equipment: []string{"leather", "shield"},
			wantAC:    14, // 10 + 1 + 2 + 1 (D&D 5e)
		},
		{
			name:      "Plate + shield DEX 16 (mod +2)",
			dexMod:    2,
			equipment: []string{"plate", "shield"},
			wantAC:    19, // 10 + 2 + 6 + 1 (D&D 5e)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New("Test", "human", "fighter")
			c.Modifiers.Dexterity = tt.dexMod
			c.Equipment = tt.equipment

			c.CalculateArmorClass(gd)

			if c.ArmorClass != tt.wantAC {
				t.Errorf("CalculateArmorClass() = %d, want %d", c.ArmorClass, tt.wantAC)
			}
		})
	}
}

func TestCalculateArmorClassWithNonArmorItems(t *testing.T) {
	gd := loadTestGameData(t)

	// Non-armor items should not affect AC
	c := New("Test", "human", "fighter")
	c.Modifiers.Dexterity = 0
	c.Equipment = []string{"longsword", "backpack", "rope_50ft", "leather"}

	c.CalculateArmorClass(gd)

	// Should only count leather armor (10 + 0 + 2 = 12) [D&D 5e]
	if c.ArmorClass != 12 {
		t.Errorf("CalculateArmorClass() with mixed items = %d, want 12", c.ArmorClass)
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
