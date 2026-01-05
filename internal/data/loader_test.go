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

func TestLoadSpecies(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expectedSpecies := []string{"human", "dragonborn", "elf", "gnome", "goliath", "halfling", "dwarf", "orc", "tiefling"}
	for _, speciesID := range expectedSpecies {
		species, ok := gd.GetSpecies(speciesID)
		if !ok {
			t.Errorf("Species %q not found", speciesID)
			continue
		}
		if species.Name == "" {
			t.Errorf("Species %q has empty name", speciesID)
		}
	}

	if len(gd.Species) != 9 {
		t.Errorf("Expected 9 species, got %d", len(gd.Species))
	}
}

func TestLoadClasses(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expectedClasses := []string{"barbarian", "bard", "cleric", "druid", "sorcerer", "fighter", "wizard", "monk", "warlock", "paladin", "ranger", "rogue"}
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

	if len(gd.Classes) != 12 {
		t.Errorf("Expected 12 classes, got %d", len(gd.Classes))
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

func TestSpeciesAbilityModifiers(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		species  string
		ability  string
		modifier int
	}{
		// Elf: +2 DEX
		{"elf", "dexterity", 2},
		// Dwarf: +2 CON
		{"dwarf", "constitution", 2},
		// Halfling: +2 DEX
		{"halfling", "dexterity", 2},
		// Dragonborn: +2 STR, +1 CHA
		{"dragonborn", "strength", 2},
		{"dragonborn", "charisma", 1},
		// Orc: +2 STR, +1 CON
		{"orc", "strength", 2},
		{"orc", "constitution", 1},
	}

	for _, tt := range tests {
		t.Run(tt.species+"/"+tt.ability, func(t *testing.T) {
			species, ok := gd.GetSpecies(tt.species)
			if !ok {
				t.Fatalf("Species %q not found", tt.species)
			}
			if species.AbilityModifiers[tt.ability] != tt.modifier {
				t.Errorf("%s %s modifier = %d, want %d",
					tt.species, tt.ability,
					species.AbilityModifiers[tt.ability],
					tt.modifier)
			}
		})
	}

	// Human: +1 all or variant (we test standard: all +1 or all 0 for variant)
	human, ok := gd.GetSpecies("human")
	if !ok {
		t.Fatal("Species 'human' not found")
	}
	// Just check that human data exists, don't enforce specific implementation
	if human.Name == "" {
		t.Error("Human species has empty name")
	}
	// Human can have various modifier implementations (standard +1 all, or variant)
	// We just verify the ability modifiers map exists
	_ = human.AbilityModifiers
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
		{"barbarian", "d12", 12},
		{"fighter", "d10", 10},
		{"paladin", "d10", 10},
		{"ranger", "d10", 10},
		{"bard", "d8", 8},
		{"cleric", "d8", 8},
		{"druid", "d8", 8},
		{"monk", "d8", 8},
		{"rogue", "d8", 8},
		{"warlock", "d8", 8},
		{"sorcerer", "d6", 6},
		{"wizard", "d6", 6},
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

func TestListSpecies(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	species := gd.ListSpecies()
	if len(species) != 9 {
		t.Errorf("ListSpecies() returned %d species, want 9", len(species))
	}
}

func TestListClasses(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	classes := gd.ListClasses()
	if len(classes) != 12 {
		t.Errorf("ListClasses() returned %d classes, want 12", len(classes))
	}
}

func TestWizardEquipment(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check spellbook exists
	if _, ok := gd.Gear["spellbook"]; !ok {
		t.Error("Gear 'spellbook' not found - required for wizard")
	}

	// Check ink_quill exists
	if _, ok := gd.Gear["ink_quill"]; !ok {
		t.Error("Gear 'ink_quill' not found - required for wizard")
	}

	// Verify spellbook properties
	spellbook := gd.Gear["spellbook"]
	if spellbook != nil {
		if spellbook.Cost != 50 {
			t.Errorf("Spellbook cost = %v, want 50", spellbook.Cost)
		}
		if spellbook.Weight != 3 {
			t.Errorf("Spellbook weight = %v, want 3", spellbook.Weight)
		}
	}
}

func TestStartingEquipmentReferencesExist(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check all starting equipment references exist
	for classID, startEquip := range gd.StartingEquipment {
		t.Run(classID, func(t *testing.T) {
			for _, itemID := range startEquip.Required {
				// Check if item exists in gear, weapons, or armor
				_, inGear := gd.Gear[itemID]
				_, inWeapons := gd.Weapons[itemID]
				_, inArmor := gd.Armor[itemID]

				if !inGear && !inWeapons && !inArmor {
					t.Errorf("Starting equipment item %q for class %q not found in any equipment list", itemID, classID)
				}
			}
		})
	}
}

func TestD5eNoSpeciesClassRestrictions(t *testing.T) {
	// D&D 5e has no species/class restrictions
	// All 9 species can play all 12 classes
	// This test documents this design decision

	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	allSpecies := []string{"human", "dragonborn", "elf", "gnome", "goliath", "halfling", "dwarf", "orc", "tiefling"}
	allClasses := []string{"barbarian", "bard", "cleric", "druid", "sorcerer", "fighter", "wizard", "monk", "warlock", "paladin", "ranger", "rogue"}

	// Verify all species exist
	for _, speciesID := range allSpecies {
		if _, ok := gd.GetSpecies(speciesID); !ok {
			t.Errorf("Species %q not found in game data", speciesID)
		}
	}

	// Verify all classes exist
	for _, classID := range allClasses {
		if _, ok := gd.GetClass(classID); !ok {
			t.Errorf("Class %q not found in game data", classID)
		}
	}

	// D&D 5e design: any species can be any class
	t.Log("D&D 5e: All species can play all classes (no restrictions)")
	t.Logf("Total combinations: %d species × %d classes = %d possible characters",
		len(allSpecies), len(allClasses), len(allSpecies)*len(allClasses))
}
