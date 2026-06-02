package web

import (
	"strings"
	"testing"
	"unicode/utf8"

	"dungeons/internal/character"
	"dungeons/internal/data"
	"dungeons/internal/npcmanager"
)

// TestWizardCoherencePromptSectionUTF8 ensures the description truncation is
// rune-aware: a long accented (multi-byte) description must not be split into
// invalid UTF-8 in the generated prompt section.
func TestWizardCoherencePromptSectionUTF8(t *testing.T) {
	c := WizardCoherence{
		RecentAdventures: []RecentAdventure{
			// One ASCII char shifts the byte offsets so the 240-byte cut lands
			// in the MIDDLE of a 2-byte rune (otherwise it'd hit a boundary).
			{Name: "Test", Description: "a" + strings.Repeat("é", 300)},
		},
	}
	out := c.PromptSection()
	if !utf8.ValidString(out) {
		t.Errorf("PromptSection produced invalid UTF-8 after truncation")
	}
	if !strings.Contains(out, "…") {
		t.Errorf("expected the long description to be truncated with an ellipsis")
	}
}

// TestNpcImportanceFromRole verifies the campaign-plan role -> importance
// mapping, including legacy English aliases (normalized to French).
func TestNpcImportanceFromRole(t *testing.T) {
	cases := map[string]npcmanager.ImportanceLevel{
		"antagoniste":      npcmanager.ImportanceKey,
		"antagonist":       npcmanager.ImportanceKey, // legacy EN
		"donneur_de_quete": npcmanager.ImportanceKey,
		"quest_giver":      npcmanager.ImportanceKey, // legacy EN
		"allie":            npcmanager.ImportanceRecurring,
		"ally":             npcmanager.ImportanceRecurring, // legacy EN
		"rival":            npcmanager.ImportanceRecurring,
		"informateur":      npcmanager.ImportanceMentioned,
		"inconnu":          npcmanager.ImportanceMentioned, // unknown -> default
	}
	for role, want := range cases {
		if got := npcImportanceFromRole(role); got != want {
			t.Errorf("npcImportanceFromRole(%q) = %q, want %q", role, got, want)
		}
	}
}

// testGameData builds a minimal in-memory GameData covering one spellcasting
// class, one martial class and one species. No disk access required.
func testGameData() *data.GameData {
	return &data.GameData{
		Species: map[string]*data.Species{
			"human": {ID: "human", Name: "Humain", Speed: 30},
			"halfling": {ID: "halfling", Name: "Halfelin", Speed: 25},
		},
		Classes: map[string]*data.Class{
			"wizard":  {ID: "wizard", Name: "Magicien", HitDieSides: 6, SpellcastingAbility: "intelligence"},
			"fighter": {ID: "fighter", Name: "Guerrier", HitDieSides: 10},
		},
	}
}

func findSkill(sheet CharacterSheetData, name string) (CharacterSheetSkill, bool) {
	for _, sk := range sheet.Skills {
		if sk.Name == name {
			return sk, true
		}
	}
	return CharacterSheetSkill{}, false
}

func findSlot(sheet CharacterSheetData, level int) (CharacterSheetSpellSlot, bool) {
	for _, sl := range sheet.SpellSlots {
		if sl.Level == level {
			return sl, true
		}
	}
	return CharacterSheetSpellSlot{}, false
}

// Skill total = ability modifier, plus proficiency bonus only when proficient.
// This addition is computed in buildCharacterSheetData, not in the character package.
func TestBuildCharacterSheetData_SkillModifiers(t *testing.T) {
	char := &character.Character{
		Name:             "Test",
		Species:          "human",
		Class:            "fighter",
		Level:            5,
		ProficiencyBonus: 3,
		Modifiers:        character.AbilityScores{Strength: 3, Dexterity: 1},
		Skills:           map[string]bool{"athletics": true}, // proficient, strength-based
	}

	sheet := buildCharacterSheetData(char, testGameData(), nil)

	athletics, ok := findSkill(sheet, "Athlétisme")
	if !ok {
		t.Fatal("Athlétisme skill missing from sheet")
	}
	if !athletics.IsProficient {
		t.Error("Athlétisme should be proficient")
	}
	// 3 (STR) + 3 (proficiency) = 6
	if athletics.Modifier != 6 {
		t.Errorf("proficient skill modifier = %d, want 6", athletics.Modifier)
	}
	if athletics.ModifierStr != "+6" {
		t.Errorf("proficient skill modifier str = %q, want %q", athletics.ModifierStr, "+6")
	}

	// acrobatics: dexterity-based, not proficient → no proficiency bonus
	acro, ok := findSkill(sheet, "Acrobaties")
	if !ok {
		t.Fatal("Acrobaties skill missing from sheet")
	}
	if acro.IsProficient {
		t.Error("Acrobaties should not be proficient")
	}
	if acro.Modifier != 1 {
		t.Errorf("non-proficient skill modifier = %d, want 1 (DEX only)", acro.Modifier)
	}
}

// A spellcasting class with slots is flagged as spellcaster, and each slot's
// Available is computed as Total - Used.
func TestBuildCharacterSheetData_SpellcasterSlots(t *testing.T) {
	char := &character.Character{
		Name:           "Mage",
		Species:        "human",
		Class:          "wizard",
		Level:          3,
		SpellSaveDC:    13,
		SpellSlots:     map[int]int{1: 4, 2: 2},
		SpellSlotsUsed: map[int]int{1: 1},
	}

	sheet := buildCharacterSheetData(char, testGameData(), nil)

	if !sheet.IsSpellcaster {
		t.Fatal("wizard with slots should be a spellcaster")
	}
	if sheet.SpellcastingAbility != "Intelligence" {
		t.Errorf("spellcasting ability = %q, want %q", sheet.SpellcastingAbility, "Intelligence")
	}

	lvl1, ok := findSlot(sheet, 1)
	if !ok {
		t.Fatal("level 1 spell slot missing")
	}
	if lvl1.Available != 3 { // 4 total - 1 used
		t.Errorf("level 1 available = %d, want 3", lvl1.Available)
	}

	lvl2, ok := findSlot(sheet, 2)
	if !ok {
		t.Fatal("level 2 spell slot missing")
	}
	if lvl2.Available != 2 { // 2 total - 0 used
		t.Errorf("level 2 available = %d, want 2", lvl2.Available)
	}
}

// A non-casting class never reports as a spellcaster.
func TestBuildCharacterSheetData_NonCaster(t *testing.T) {
	char := &character.Character{
		Name:    "Brute",
		Species: "human",
		Class:   "fighter",
		Level:   3,
	}

	sheet := buildCharacterSheetData(char, testGameData(), nil)

	if sheet.IsSpellcaster {
		t.Error("fighter should not be a spellcaster")
	}
}

// Known species/class pull their display name, speed and hit dice from GameData.
func TestBuildCharacterSheetData_KnownSpeciesClass(t *testing.T) {
	char := &character.Character{
		Name:    "Bilbon",
		Species: "halfling",
		Class:   "wizard",
		Level:   2,
	}

	sheet := buildCharacterSheetData(char, testGameData(), nil)

	if sheet.SpeciesName != "Halfelin" {
		t.Errorf("species name = %q, want %q", sheet.SpeciesName, "Halfelin")
	}
	if sheet.Speed != 25 {
		t.Errorf("speed = %d, want 25", sheet.Speed)
	}
	if sheet.ClassName != "Magicien" {
		t.Errorf("class name = %q, want %q", sheet.ClassName, "Magicien")
	}
	if sheet.HitDice != "2d6" { // level 2, d6 hit die
		t.Errorf("hit dice = %q, want %q", sheet.HitDice, "2d6")
	}
}

// Unknown species/class fall back to defaults: speed 30, d8 hit die, raw IDs as names.
func TestBuildCharacterSheetData_UnknownSpeciesClassFallback(t *testing.T) {
	char := &character.Character{
		Name:    "Inconnu",
		Species: "dragonborn", // not in testGameData
		Class:   "artificer",  // not in testGameData
		Level:   4,
	}

	sheet := buildCharacterSheetData(char, testGameData(), nil)

	if sheet.SpeciesName != "dragonborn" {
		t.Errorf("unknown species name = %q, want raw id %q", sheet.SpeciesName, "dragonborn")
	}
	if sheet.Speed != 30 {
		t.Errorf("unknown species speed = %d, want default 30", sheet.Speed)
	}
	if sheet.ClassName != "artificer" {
		t.Errorf("unknown class name = %q, want raw id %q", sheet.ClassName, "artificer")
	}
	if sheet.HitDice != "4d8" { // level 4, default d8
		t.Errorf("unknown class hit dice = %q, want default %q", sheet.HitDice, "4d8")
	}
}
