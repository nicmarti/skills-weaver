// Package data provides loading and access to D&D 5e game data.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Species represents a playable species (race) in D&D 5e.
type Species struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	NameEN           string         `json:"name_en"`
	Description      string         `json:"description"`
	Size             string         `json:"size"`             // "Small" or "Medium"
	Speed            int            `json:"speed"`            // Usually 25 or 30 feet
	Languages        []string       `json:"languages"`        // At least Common + 1-2 others
	SpecialTraits    []string       `json:"special_traits"`   // Narrative descriptions
	AbilityModifiers map[string]int `json:"ability_modifiers"` // +2 DEX, +1 CON, etc.
}

// Skill represents one of the 18 D&D 5e skills.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameEN      string `json:"name_en"`
	Ability     string `json:"ability"`     // "strength", "dexterity", etc.
	Description string `json:"description"` // Brief explanation
}

// Class represents a playable class in D&D 5e.
type Class struct {
	ID                      string         `json:"id"`
	Name                    string         `json:"name"`
	NameEN                  string         `json:"name_en"`
	HitDie                  string         `json:"hit_die"`                     // "d6", "d8", "d10", "d12"
	HitDieSides             int            `json:"hit_die_sides"`               // 6, 8, 10, 12
	PrimaryAbility          string         `json:"primary_ability"`             // "strength", "charisma", etc.
	SavingThrowProfs        []string       `json:"saving_throw_proficiencies"`  // 2 abilities
	SkillProfs              []string       `json:"skill_proficiencies"`         // Available skills
	SkillChoiceCount        int            `json:"skill_choice_count"`          // How many to choose
	ProficiencyBonus        map[string]int `json:"proficiency_bonus"`           // By level: "1": 2, "5": 3, etc.
	SpellcastingAbility     string         `json:"spellcasting_ability"`        // Empty if non-caster
	StartingEquipment       string         `json:"starting_equipment,omitempty"` // Brief description
}

// Weapon represents a weapon item.
type Weapon struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	NameEN     string   `json:"name_en"`
	Damage     string   `json:"damage"`
	Weight     float64  `json:"weight"`
	Cost       float64  `json:"cost"`
	Type       string   `json:"type"`
	Properties []string `json:"properties"`
	Range      string   `json:"range,omitempty"`
}

// Armor represents an armor item.
type Armor struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	NameEN  string  `json:"name_en"`
	ACBonus int     `json:"ac_bonus"`
	Weight  float64 `json:"weight"`
	Cost    float64 `json:"cost"`
	Type    string  `json:"type"`
}

// Gear represents adventuring gear.
type Gear struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	NameEN string  `json:"name_en"`
	Cost   float64 `json:"cost"`
	Weight float64 `json:"weight"`
}

// Spell5e represents a D&D 5e spell.
type Spell5e struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	NameEN        string   `json:"name_en"`
	Level         int      `json:"level"`         // 0-9 (0 = cantrip)
	School        string   `json:"school"`        // "evocation", "abjuration", etc.
	CastingTime   string   `json:"casting_time"`  // "1 action", "1 bonus action", etc.
	Range         string   `json:"range"`         // "9 m", "Contact", "Personnelle"
	Components    []string `json:"components"`    // ["V", "S", "M"]
	Material      string   `json:"material,omitempty"` // Required if M in components
	Duration      string   `json:"duration"`      // "Instantanée", "1 heure", etc.
	Concentration bool     `json:"concentration"` // True if requires concentration
	Ritual        bool     `json:"ritual"`        // True if can be cast as ritual
	Classes       []string `json:"classes"`       // Class IDs that can cast
	DescriptionFR string   `json:"description_fr"`
	DescriptionEN string   `json:"description_en,omitempty"`
	Upcast        string   `json:"upcast,omitempty"`  // Effect when cast at higher level
	Damage        string   `json:"damage,omitempty"`  // e.g., "1d4+1"
	Healing       string   `json:"healing,omitempty"` // e.g., "1d8"
	Save          string   `json:"save,omitempty"`    // e.g., "Constitution"
}

// StartingEquipment represents class-specific starting equipment options.
type StartingEquipment struct {
	Required      []string   `json:"required"`
	WeaponChoices [][]string `json:"weapon_choices"`
	ArmorChoices  []string   `json:"armor_choices"`
}

// GameData holds all loaded game data.
type GameData struct {
	dataDir           string
	Species           map[string]*Species
	Classes           map[string]*Class
	Skills            map[string]*Skill
	Weapons           map[string]*Weapon
	Armor             map[string]*Armor
	Gear              map[string]*Gear
	Spells5e          map[string]*Spell5e
	StartingEquipment map[string]*StartingEquipment
}

// speciesFile represents the JSON structure for species.
type speciesFile struct {
	Species []Species `json:"species"`
}

// skillsFile represents the JSON structure for skills.
type skillsFile struct {
	Skills []Skill `json:"skills"`
}

// classesFile represents the JSON structure for classes.
type classesFile struct {
	Classes []Class `json:"classes"`
}

// equipmentFile represents the JSON structure for equipment.
type equipmentFile struct {
	Weapons           []Weapon                     `json:"weapons"`
	Armor             []Armor                      `json:"armor"`
	AdventuringGear   []Gear                       `json:"adventuring_gear"`
	StartingEquipment map[string]StartingEquipment `json:"starting_equipment"`
	Ammunition        []Gear                       `json:"ammunition"`
}

// spellsFile represents the JSON structure for D&D 5e spells.
type spellsFile struct {
	Spells []Spell5e `json:"spells"`
}

// Load reads all game data from JSON files in the specified directory.
// If dataDir is empty, it defaults to "./data".
func Load(dataDir string) (*GameData, error) {
	if dataDir == "" {
		dataDir = "data"
	}

	gd := &GameData{
		dataDir:           dataDir,
		Species:           make(map[string]*Species),
		Classes:           make(map[string]*Class),
		Skills:            make(map[string]*Skill),
		Weapons:           make(map[string]*Weapon),
		Armor:             make(map[string]*Armor),
		Gear:              make(map[string]*Gear),
		Spells5e:          make(map[string]*Spell5e),
		StartingEquipment: make(map[string]*StartingEquipment),
	}

	// Load species (D&D 5e races)
	if err := gd.loadSpecies(); err != nil {
		return nil, fmt.Errorf("loading species: %w", err)
	}

	// Load skills (D&D 5e 18 skills)
	if err := gd.loadSkills(); err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}

	// Load classes
	if err := gd.loadClasses(); err != nil {
		return nil, fmt.Errorf("loading classes: %w", err)
	}

	// Load equipment
	if err := gd.loadEquipment(); err != nil {
		return nil, fmt.Errorf("loading equipment: %w", err)
	}

	// Load spells (D&D 5e)
	if err := gd.loadSpells(); err != nil {
		return nil, fmt.Errorf("loading spells: %w", err)
	}

	return gd, nil
}

func (gd *GameData) loadSpecies() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "5e", "species.json"))
	if err != nil {
		return err
	}

	var sf speciesFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}

	for i := range sf.Species {
		species := &sf.Species[i]
		gd.Species[species.ID] = species
	}

	return nil
}

func (gd *GameData) loadSkills() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "5e", "skills.json"))
	if err != nil {
		return err
	}

	var sf skillsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}

	for i := range sf.Skills {
		skill := &sf.Skills[i]
		gd.Skills[skill.ID] = skill
	}

	return nil
}

func (gd *GameData) loadClasses() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "5e", "classes.json"))
	if err != nil {
		return err
	}

	var cf classesFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return err
	}

	for i := range cf.Classes {
		class := &cf.Classes[i]
		gd.Classes[class.ID] = class
	}

	return nil
}

func (gd *GameData) loadEquipment() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "equipment.json"))
	if err != nil {
		return err
	}

	var ef equipmentFile
	if err := json.Unmarshal(data, &ef); err != nil {
		return err
	}

	for i := range ef.Weapons {
		w := &ef.Weapons[i]
		gd.Weapons[w.ID] = w
	}

	for i := range ef.Armor {
		a := &ef.Armor[i]
		gd.Armor[a.ID] = a
	}

	for i := range ef.AdventuringGear {
		g := &ef.AdventuringGear[i]
		gd.Gear[g.ID] = g
	}

	for i := range ef.Ammunition {
		g := &ef.Ammunition[i]
		gd.Gear[g.ID] = g
	}

	for classID, se := range ef.StartingEquipment {
		seCopy := se
		gd.StartingEquipment[classID] = &seCopy
	}

	return nil
}

// GetSpecies returns a species by ID.
func (gd *GameData) GetSpecies(id string) (*Species, bool) {
	s, ok := gd.Species[id]
	return s, ok
}

// GetSkill returns a skill by ID.
func (gd *GameData) GetSkill(id string) (*Skill, bool) {
	s, ok := gd.Skills[id]
	return s, ok
}

// GetClass returns a class by ID.
func (gd *GameData) GetClass(id string) (*Class, bool) {
	c, ok := gd.Classes[id]
	return c, ok
}

// GetWeapon returns a weapon by ID.
func (gd *GameData) GetWeapon(id string) (*Weapon, bool) {
	w, ok := gd.Weapons[id]
	return w, ok
}

// GetArmor returns an armor by ID.
func (gd *GameData) GetArmor(id string) (*Armor, bool) {
	a, ok := gd.Armor[id]
	return a, ok
}

// ListSpecies returns all available species.
func (gd *GameData) ListSpecies() []*Species {
	species := make([]*Species, 0, len(gd.Species))
	for _, s := range gd.Species {
		species = append(species, s)
	}
	return species
}

// ListSkills returns all available skills.
func (gd *GameData) ListSkills() []*Skill {
	skills := make([]*Skill, 0, len(gd.Skills))
	for _, s := range gd.Skills {
		skills = append(skills, s)
	}
	return skills
}

// ListClasses returns all available classes.
func (gd *GameData) ListClasses() []*Class {
	classes := make([]*Class, 0, len(gd.Classes))
	for _, c := range gd.Classes {
		classes = append(classes, c)
	}
	return classes
}

// AbilityModifier calculates the D&D 5e ability modifier.
// Formula: (ability_score - 10) ÷ 2 (rounded down)
func AbilityModifier(score int) int {
	return (score - 10) / 2
}

// ProficiencyBonusByLevel returns the proficiency bonus for a given character level.
func ProficiencyBonusByLevel(level int) int {
	switch {
	case level >= 17:
		return 6
	case level >= 13:
		return 5
	case level >= 9:
		return 4
	case level >= 5:
		return 3
	default:
		return 2
	}
}

// ValidationError represents a data validation issue.
type ValidationError struct {
	File     string `json:"file"`
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" or "warning"
}

// ValidateGameData performs comprehensive validation of all game data.
// It checks cross-references between files and reports inconsistencies.
func ValidateGameData(gd *GameData) []ValidationError {
	var errors []ValidationError

	// Validate starting equipment references
	errors = append(errors, validateStartingEquipment(gd)...)

	return errors
}

// validateSpeciesClassRefs is removed - D&D 5e has no species/class restrictions.

// isValidClassRef checks if a class ID is valid.
// Supports multi-class notation like "fighter/magic-user".
func (gd *GameData) isValidClassRef(classID string) bool {
	// Direct match
	if _, ok := gd.Classes[classID]; ok {
		return true
	}

	// Check for multi-class (e.g., "fighter/magic-user")
	if strings.Contains(classID, "/") {
		parts := strings.Split(classID, "/")
		for _, part := range parts {
			if _, ok := gd.Classes[part]; !ok {
				return false
			}
		}
		return true
	}

	return false
}

// validateStartingEquipment checks that all starting equipment references valid items.
func validateStartingEquipment(gd *GameData) []ValidationError {
	var errors []ValidationError

	for classID, se := range gd.StartingEquipment {
		// Check required items
		for _, itemID := range se.Required {
			if !gd.itemExists(itemID) {
				errors = append(errors, ValidationError{
					File:     "equipment.json",
					Field:    fmt.Sprintf("starting_equipment[%s].required", classID),
					Message:  fmt.Sprintf("references non-existent item '%s'", itemID),
					Severity: "error",
				})
			}
		}

		// Check weapon choices
		for i, choices := range se.WeaponChoices {
			for _, itemID := range choices {
				if !gd.itemExists(itemID) {
					errors = append(errors, ValidationError{
						File:     "equipment.json",
						Field:    fmt.Sprintf("starting_equipment[%s].weapon_choices[%d]", classID, i),
						Message:  fmt.Sprintf("references non-existent weapon '%s'", itemID),
						Severity: "error",
					})
				}
			}
		}

		// Check armor choices
		for _, itemID := range se.ArmorChoices {
			if !gd.itemExists(itemID) {
				errors = append(errors, ValidationError{
					File:     "equipment.json",
					Field:    fmt.Sprintf("starting_equipment[%s].armor_choices", classID),
					Message:  fmt.Sprintf("references non-existent armor '%s'", itemID),
					Severity: "error",
				})
			}
		}
	}

	return errors
}

// itemExists checks if an item ID exists in weapons, armor, or gear.
func (gd *GameData) itemExists(id string) bool {
	if _, ok := gd.Weapons[id]; ok {
		return true
	}
	if _, ok := gd.Armor[id]; ok {
		return true
	}
	if _, ok := gd.Gear[id]; ok {
		return true
	}
	return false
}

// loadSpells loads D&D 5e spells from data/5e/spells.json.
func (gd *GameData) loadSpells() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "5e", "spells.json"))
	if err != nil {
		return err
	}

	var sf spellsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}

	for i := range sf.Spells {
		spell := &sf.Spells[i]
		gd.Spells5e[spell.ID] = spell
	}

	return nil
}

// GetSpell5e returns a D&D 5e spell by ID.
func (gd *GameData) GetSpell5e(id string) (*Spell5e, bool) {
	spell, ok := gd.Spells5e[id]
	return spell, ok
}

// ListSpellsByClass returns all spells available to a specific class and spell level.
// If level is -1, returns all spells for the class regardless of level.
func (gd *GameData) ListSpellsByClass(classID string, level int) []*Spell5e {
	spells := make([]*Spell5e, 0)
	for _, spell := range gd.Spells5e {
		// Check if this class can cast this spell
		hasClass := false
		for _, c := range spell.Classes {
			if c == classID {
				hasClass = true
				break
			}
		}
		if !hasClass {
			continue
		}

		// Check level if specified
		if level >= 0 && spell.Level != level {
			continue
		}

		spells = append(spells, spell)
	}
	return spells
}

// ListCantrips returns all cantrips (level 0 spells) for a specific class.
func (gd *GameData) ListCantrips(classID string) []*Spell5e {
	return gd.ListSpellsByClass(classID, 0)
}

// ListSpellsBySchool returns all spells of a specific school of magic.
func (gd *GameData) ListSpellsBySchool(school string) []*Spell5e {
	spells := make([]*Spell5e, 0)
	for _, spell := range gd.Spells5e {
		if strings.EqualFold(spell.School, school) {
			spells = append(spells, spell)
		}
	}
	return spells
}

// SearchSpells searches for spells by name (French or English).
func (gd *GameData) SearchSpells(query string) []*Spell5e {
	query = strings.ToLower(query)
	spells := make([]*Spell5e, 0)
	for _, spell := range gd.Spells5e {
		if strings.Contains(strings.ToLower(spell.Name), query) ||
			strings.Contains(strings.ToLower(spell.NameEN), query) ||
			strings.Contains(strings.ToLower(spell.ID), query) {
			spells = append(spells, spell)
		}
	}
	return spells
}

// ListAllSpells returns all spells.
func (gd *GameData) ListAllSpells() []*Spell5e {
	spells := make([]*Spell5e, 0, len(gd.Spells5e))
	for _, spell := range gd.Spells5e {
		spells = append(spells, spell)
	}
	return spells
}
