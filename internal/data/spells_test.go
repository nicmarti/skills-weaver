package data

import (
	"testing"
)

func TestLoadSpells(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(gd.Spells5e) == 0 {
		t.Error("No spells loaded")
	}

	t.Logf("Loaded %d spells", len(gd.Spells5e))
}

func TestGetSpell5e(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test getting a specific spell
	spell, ok := gd.GetSpell5e("aide")
	if !ok {
		t.Fatal("Spell 'aide' not found")
	}

	if spell.Name != "Aide" {
		t.Errorf("Spell name = %q, want 'Aide'", spell.Name)
	}
	if spell.Level != 2 {
		t.Errorf("Spell level = %d, want 2", spell.Level)
	}
	if spell.School != "abjuration" {
		t.Errorf("Spell school = %q, want 'abjuration'", spell.School)
	}
}

func TestListCantrips(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test wizard cantrips
	cantrips := gd.ListCantrips("wizard")
	if len(cantrips) == 0 {
		t.Error("No wizard cantrips found")
	}

	// Verify all are level 0
	for _, spell := range cantrips {
		if spell.Level != 0 {
			t.Errorf("Cantrip %q has level %d, want 0", spell.Name, spell.Level)
		}
		// Verify wizard is in classes
		hasWizard := false
		for _, class := range spell.Classes {
			if class == "wizard" {
				hasWizard = true
				break
			}
		}
		if !hasWizard {
			t.Errorf("Cantrip %q doesn't have wizard in classes", spell.Name)
		}
	}

	t.Logf("Wizard has %d cantrips", len(cantrips))
}

func TestListSpellsByClass(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test level 1 wizard spells
	level1 := gd.ListSpellsByClass("wizard", 1)
	if len(level1) == 0 {
		t.Error("No level 1 wizard spells found")
	}

	// Verify all are level 1 and have wizard
	for _, spell := range level1 {
		if spell.Level != 1 {
			t.Errorf("Spell %q has level %d, want 1", spell.Name, spell.Level)
		}
		hasWizard := false
		for _, class := range spell.Classes {
			if class == "wizard" {
				hasWizard = true
				break
			}
		}
		if !hasWizard {
			t.Errorf("Spell %q doesn't have wizard in classes", spell.Name)
		}
	}

	t.Logf("Wizard has %d level 1 spells", len(level1))
}

func TestListSpellsBySchool(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	schools := []string{"abjuration", "conjuration", "divination", "enchantment", "evocation", "illusion", "necromancy", "transmutation"}

	for _, school := range schools {
		spells := gd.ListSpellsBySchool(school)
		if len(spells) == 0 {
			t.Errorf("No spells found for school %q", school)
		}
		t.Logf("School %s: %d spells", school, len(spells))
	}
}

func TestSearchSpells(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Search for "lumière" (light)
	results := gd.SearchSpells("lumière")
	if len(results) == 0 {
		t.Error("No results for 'lumière'")
	}

	t.Logf("Found %d spells matching 'lumière'", len(results))
}

func TestSpellComponents(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Count spells with material components
	materialCount := 0
	withMaterialDesc := 0

	for _, spell := range gd.Spells5e {
		hasM := false
		for _, comp := range spell.Components {
			if comp == "M" {
				hasM = true
				materialCount++
				break
			}
		}

		if hasM && spell.Material != "" {
			withMaterialDesc++
			if withMaterialDesc <= 3 {
				// Log first 3 examples
				t.Logf("Spell %q material: %s", spell.Name, spell.Material)
			}
		}
	}

	t.Logf("Found %d spells with M component, %d with material description", materialCount, withMaterialDesc)

	if materialCount == 0 {
		t.Error("No spells with M component found")
	}
}

func TestSpellConcentration(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	concentrationCount := 0
	for _, spell := range gd.Spells5e {
		if spell.Concentration {
			concentrationCount++
		}
	}

	t.Logf("Found %d spells requiring concentration", concentrationCount)
	if concentrationCount == 0 {
		t.Error("No concentration spells found")
	}
}

func TestSpellRitual(t *testing.T) {
	dataDir := getTestDataDir(t)
	gd, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	ritualCount := 0
	for _, spell := range gd.Spells5e {
		if spell.Ritual {
			ritualCount++
		}
	}

	t.Logf("Found %d ritual spells", ritualCount)
	if ritualCount == 0 {
		t.Error("No ritual spells found")
	}
}
