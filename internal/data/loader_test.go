package data

import (
	"path/filepath"
	"testing"
)

func getTestDataDir(t *testing.T) string {
	// Go up from internal/data to project root, then into data/
	return filepath.Join("..", "..", "data")
}

func TestLoad(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if gd == nil {
		t.Fatal("Load() returned nil")
	}
}

func TestLoadRaces(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expectedRaces := []string{"human", "elf", "dwarf", "halfling"}
	for _, raceID := range expectedRaces {
		race, ok := gd.GetRace(raceID)
		if !ok {
			t.Errorf("Race %q not found", raceID)
			continue
		}
		if race.Name == "" {
			t.Errorf("Race %q has empty name", raceID)
		}
	}

	if len(gd.Races) != 4 {
		t.Errorf("Expected 4 races, got %d", len(gd.Races))
	}
}

func TestLoadClasses(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expectedClasses := []string{"fighter", "cleric", "magic-user", "thief"}
	for _, classID := range expectedClasses {
		class, ok := gd.GetClass(classID)
		if !ok {
			t.Errorf("Class %q not found", classID)
			continue
		}
		if class.Name == "" {
			t.Errorf("Class %q has empty name", classID)
		}
		if class.HitDieSides <= 0 {
			t.Errorf("Class %q has invalid hit die sides: %d", classID, class.HitDieSides)
		}
	}

	if len(gd.Classes) != 4 {
		t.Errorf("Expected 4 classes, got %d", len(gd.Classes))
	}
}

func TestLoadEquipment(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check weapons
	weapon, ok := gd.GetWeapon("longsword")
	if !ok {
		t.Error("Weapon 'longsword' not found")
	} else {
		if weapon.Damage != "1d8" {
			t.Errorf("Longsword damage = %q, want 1d8", weapon.Damage)
		}
	}

	// Check armor
	armor, ok := gd.GetArmor("chainmail")
	if !ok {
		t.Error("Armor 'chainmail' not found")
	} else {
		if armor.ACBonus != 4 {
			t.Errorf("Chainmail AC bonus = %d, want 4", armor.ACBonus)
		}
	}
}

func TestCanPlayClass(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		race    string
		class   string
		allowed bool
	}{
		{"human", "fighter", true},
		{"human", "cleric", true},
		{"human", "magic-user", true},
		{"human", "thief", true},
		{"elf", "fighter", true},
		{"elf", "magic-user", true},
		{"elf", "cleric", false}, // Elves cannot be clerics
		{"dwarf", "fighter", true},
		{"dwarf", "cleric", true},
		{"dwarf", "magic-user", false}, // Dwarves cannot be magic-users
		{"halfling", "fighter", true},
		{"halfling", "thief", true},
		{"halfling", "cleric", false}, // Halflings cannot be clerics
		{"halfling", "magic-user", false}, // Halflings cannot be magic-users
	}

	for _, tt := range tests {
		t.Run(tt.race+"/"+tt.class, func(t *testing.T) {
			got := gd.CanPlayClass(tt.race, tt.class)
			if got != tt.allowed {
				t.Errorf("CanPlayClass(%q, %q) = %v, want %v", tt.race, tt.class, got, tt.allowed)
			}
		})
	}
}

func TestGetLevelLimit(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		race  string
		class string
		limit int
	}{
		{"human", "fighter", 0},   // Unlimited
		{"elf", "fighter", 6},     // Limited to 6
		{"elf", "magic-user", 9},  // Limited to 9
		{"dwarf", "fighter", 7},   // Limited to 7
		{"halfling", "fighter", 4}, // Limited to 4
	}

	for _, tt := range tests {
		t.Run(tt.race+"/"+tt.class, func(t *testing.T) {
			got := gd.GetLevelLimit(tt.race, tt.class)
			if got != tt.limit {
				t.Errorf("GetLevelLimit(%q, %q) = %d, want %d", tt.race, tt.class, got, tt.limit)
			}
		})
	}
}

func TestRaceAbilityModifiers(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Elf: +1 DEX, -1 CON
	elf, _ := gd.GetRace("elf")
	if elf.AbilityModifiers["dexterity"] != 1 {
		t.Errorf("Elf DEX modifier = %d, want 1", elf.AbilityModifiers["dexterity"])
	}
	if elf.AbilityModifiers["constitution"] != -1 {
		t.Errorf("Elf CON modifier = %d, want -1", elf.AbilityModifiers["constitution"])
	}

	// Dwarf: +1 CON, -1 CHA
	dwarf, _ := gd.GetRace("dwarf")
	if dwarf.AbilityModifiers["constitution"] != 1 {
		t.Errorf("Dwarf CON modifier = %d, want 1", dwarf.AbilityModifiers["constitution"])
	}
	if dwarf.AbilityModifiers["charisma"] != -1 {
		t.Errorf("Dwarf CHA modifier = %d, want -1", dwarf.AbilityModifiers["charisma"])
	}

	// Halfling: +1 DEX, -1 STR
	halfling, _ := gd.GetRace("halfling")
	if halfling.AbilityModifiers["dexterity"] != 1 {
		t.Errorf("Halfling DEX modifier = %d, want 1", halfling.AbilityModifiers["dexterity"])
	}
	if halfling.AbilityModifiers["strength"] != -1 {
		t.Errorf("Halfling STR modifier = %d, want -1", halfling.AbilityModifiers["strength"])
	}

	// Human: no modifiers
	human, _ := gd.GetRace("human")
	if len(human.AbilityModifiers) != 0 {
		t.Errorf("Human should have no ability modifiers, got %v", human.AbilityModifiers)
	}
}

func TestClassHitDice(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		class    string
		hitDie   string
		hitSides int
	}{
		{"fighter", "d8", 8},
		{"cleric", "d6", 6},
		{"magic-user", "d4", 4},
		{"thief", "d4", 4},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			class, ok := gd.GetClass(tt.class)
			if !ok {
				t.Fatalf("Class %q not found", tt.class)
			}
			if class.HitDie != tt.hitDie {
				t.Errorf("HitDie = %q, want %q", class.HitDie, tt.hitDie)
			}
			if class.HitDieSides != tt.hitSides {
				t.Errorf("HitDieSides = %d, want %d", class.HitDieSides, tt.hitSides)
			}
		})
	}
}

func TestListRaces(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	races := gd.ListRaces()
	if len(races) != 4 {
		t.Errorf("ListRaces() returned %d races, want 4", len(races))
	}
}

func TestListClasses(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	classes := gd.ListClasses()
	if len(classes) != 4 {
		t.Errorf("ListClasses() returned %d classes, want 4", len(classes))
	}
}
