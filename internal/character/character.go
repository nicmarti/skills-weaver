// Package character provides character creation and management for D&D 5e.
package character

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dungeons/internal/data"
	"dungeons/internal/dice"

	"github.com/google/uuid"
)

// AbilityScores represents the six ability scores in BFRPG order.
type AbilityScores struct {
	Strength     int `json:"strength"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Charisma     int `json:"charisma"`
}

// Character represents a player character in D&D 5e.
type Character struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Species      string        `json:"species"`      // Replaces "Race" in D&D 5e
	Class        string        `json:"class"`
	Level        int           `json:"level"`
	XP           int           `json:"xp"`
	Abilities    AbilityScores `json:"abilities"`
	Modifiers    AbilityScores `json:"modifiers"`
	HitPoints    int           `json:"hit_points"`
	MaxHitPoints int           `json:"max_hit_points"`
	ArmorClass   int           `json:"armor_class"`
	Gold         int           `json:"gold"`
	Equipment    []string      `json:"equipment"`

	// D&D 5e specific fields
	Background       string          `json:"background,omitempty"`        // Character background
	ProficiencyBonus int             `json:"proficiency_bonus"`           // +2 to +6 based on level
	Skills           map[string]bool `json:"skills,omitempty"`            // Proficient skills
	SavingThrowProfs map[string]bool `json:"saving_throw_profs,omitempty"` // Proficient saving throws
	Inspiration      bool            `json:"inspiration,omitempty"`       // Inspiration point
	TempHitPoints    int             `json:"temp_hit_points,omitempty"`   // Temporary HP
	HitDice          int             `json:"hit_dice,omitempty"`          // Remaining hit dice
	MaxHitDice       int             `json:"max_hit_dice,omitempty"`      // Maximum hit dice (= level)

	// Spell system fields
	SpellSaveDC      int        `json:"spell_save_dc,omitempty"`       // DC for spell saves
	SpellAttackBonus int        `json:"spell_attack_bonus,omitempty"`  // Bonus for spell attacks
	KnownSpells      []string   `json:"known_spells,omitempty"`        // Spell IDs known by the character
	PreparedSpells   []string   `json:"prepared_spells,omitempty"`     // Spell IDs prepared for the day
	SpellSlots       map[int]int `json:"spell_slots,omitempty"`        // Available spell slots by level
	SpellSlotsUsed   map[int]int `json:"spell_slots_used,omitempty"`   // Used spell slots by level

	Appearance *CharacterAppearance `json:"appearance,omitempty"` // Visual description for image generation
	CreatedAt  time.Time            `json:"created_at"`
}

// CharacterAppearance stores visual description for image generation.
type CharacterAppearance struct {
	Age                int    `json:"age,omitempty"`
	Gender             string `json:"gender,omitempty"`             // "male", "female", "non-binary"
	Build              string `json:"build,omitempty"`              // "slender", "stocky", "muscular", "average"
	Height             string `json:"height,omitempty"`             // "tall", "average", "short"
	HairColor          string `json:"hair_color,omitempty"`
	HairStyle          string `json:"hair_style,omitempty"`
	EyeColor           string `json:"eye_color,omitempty"`
	SkinTone           string `json:"skin_tone,omitempty"`
	FacialFeature      string `json:"facial_feature,omitempty"`     // "bearded", "clean-shaven", "scarred"
	DistinctiveFeature string `json:"distinctive_feature,omitempty"` // "battle scar", "tattoo", "eye patch"
	ArmorDescription   string `json:"armor_description,omitempty"`   // "plate armor", "leather vest"
	WeaponDescription  string `json:"weapon_description,omitempty"`  // "longsword", "staff with crystal"
	Accessories        string `json:"accessories,omitempty"`         // "shield", "holy symbol", "spell book"
	ReferenceImage     string `json:"reference_image,omitempty"`     // Path to reference image for FLUX PuLID
}

// GenerationMethod specifies how ability scores are generated.
type GenerationMethod string

const (
	// MethodStandard uses 4d6 keep highest 3 for each ability.
	MethodStandard GenerationMethod = "standard"
	// MethodClassic uses 3d6 for each ability.
	MethodClassic GenerationMethod = "classic"
)

// New creates a new character with basic info.
func New(name, species, class string) *Character {
	return &Character{
		ID:        uuid.New().String(),
		Name:      name,
		Species:   species,
		Class:     class,
		Level:     1,
		XP:        0,
		Equipment: []string{},
		CreatedAt: time.Now(),
	}
}

// GenerateAbilities rolls ability scores using the specified method.
func (c *Character) GenerateAbilities(method GenerationMethod) []dice.Result {
	roller := dice.New()

	var results []dice.Result
	if method == MethodClassic {
		results = roller.RollStatsClassic()
	} else {
		results = roller.RollStats()
	}

	// Assign in BFRPG order: STR, INT, WIS, DEX, CON, CHA
	c.Abilities.Strength = results[0].Total
	c.Abilities.Intelligence = results[1].Total
	c.Abilities.Wisdom = results[2].Total
	c.Abilities.Dexterity = results[3].Total
	c.Abilities.Constitution = results[4].Total
	c.Abilities.Charisma = results[5].Total

	return results
}

// ApplyRacialModifiers is deprecated in D&D 5e 2024.
// In D&D 5e 2024, ability score increases come from backgrounds, not species.
// This function is kept for API compatibility but does nothing.
func (c *Character) ApplyRacialModifiers(gd *data.GameData) error {
	// Verify species exists
	species, ok := gd.GetSpecies(c.Species)
	if !ok {
		return fmt.Errorf("unknown species: %s", c.Species)
	}

	// Apply ability modifiers from species (D&D 5e)
	// Examples: Elf +2 DEX, Dwarf +2 CON, Mountain Dwarf +2 STR
	if species.AbilityModifiers != nil {
		if mod, ok := species.AbilityModifiers["strength"]; ok {
			c.Abilities.Strength += mod
		}
		if mod, ok := species.AbilityModifiers["dexterity"]; ok {
			c.Abilities.Dexterity += mod
		}
		if mod, ok := species.AbilityModifiers["constitution"]; ok {
			c.Abilities.Constitution += mod
		}
		if mod, ok := species.AbilityModifiers["intelligence"]; ok {
			c.Abilities.Intelligence += mod
		}
		if mod, ok := species.AbilityModifiers["wisdom"]; ok {
			c.Abilities.Wisdom += mod
		}
		if mod, ok := species.AbilityModifiers["charisma"]; ok {
			c.Abilities.Charisma += mod
		}
	}

	return nil
}

// CalculateModifiers computes ability modifiers according to D&D 5e rules.
// Formula: (ability_score - 10) ÷ 2 (rounded down)
func (c *Character) CalculateModifiers() {
	c.Modifiers.Strength = data.AbilityModifier(c.Abilities.Strength)
	c.Modifiers.Intelligence = data.AbilityModifier(c.Abilities.Intelligence)
	c.Modifiers.Wisdom = data.AbilityModifier(c.Abilities.Wisdom)
	c.Modifiers.Dexterity = data.AbilityModifier(c.Abilities.Dexterity)
	c.Modifiers.Constitution = data.AbilityModifier(c.Abilities.Constitution)
	c.Modifiers.Charisma = data.AbilityModifier(c.Abilities.Charisma)
}

// RollHitPoints calculates hit points for level 1.
//
// Parameters:
//   - maxHP: if true, gives maximum hit die value (popular variant for survivability)
//     if false, rolls the hit die randomly (standard BFRPG rules)
//
// The hit die depends on class:
//   - Fighter: d8 (1-8)
//   - Cleric: d6 (1-6)
//   - Magic-User: d4 (1-4)
//   - Thief: d4 (1-4)
//
// Constitution modifier is always added. Minimum HP is 1.
func (c *Character) RollHitPoints(gd *data.GameData, maxHP bool) error {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return fmt.Errorf("unknown class: %s", c.Class)
	}

	var hp int
	if maxHP {
		// Variant rule: maximum hit die at level 1
		hp = class.HitDieSides
	} else {
		// Standard BFRPG: roll the hit die
		roller := dice.New()
		result, err := roller.Roll(class.HitDie)
		if err != nil {
			return fmt.Errorf("rolling hit die: %w", err)
		}
		hp = result.Total
	}

	// Add Constitution modifier
	hp += c.Modifiers.Constitution

	// Minimum 1 HP
	if hp < 1 {
		hp = 1
	}

	c.HitPoints = hp
	c.MaxHitPoints = hp

	return nil
}

// RollStartingGold rolls starting gold based on class (D&D 5e starting wealth).
func (c *Character) RollStartingGold(gd *data.GameData) error {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return fmt.Errorf("unknown class: %s", c.Class)
	}

	roller := dice.New()

	// D&D 5e starting wealth by class (in gp)
	// Based on PHB starting gold table
	var goldFormula string
	switch class.ID {
	case "barbarian", "fighter", "paladin", "ranger":
		goldFormula = "5d4" // × 10 = 50-200 gp
	case "cleric", "druid", "monk", "rogue":
		goldFormula = "4d4" // × 10 = 40-160 gp
	case "bard", "warlock":
		goldFormula = "5d4" // × 10 = 50-200 gp
	case "sorcerer", "wizard":
		goldFormula = "3d4" // × 10 = 30-120 gp
	default:
		goldFormula = "4d4" // Default
	}

	result, err := roller.Roll(goldFormula)
	if err != nil {
		return fmt.Errorf("rolling starting gold: %w", err)
	}

	c.Gold = result.Total * 10

	return nil
}

// CalculateArmorClass computes AC (base 10 in D&D 5e).
func (c *Character) CalculateArmorClass(gd *data.GameData) {
	// Base AC in D&D 5e is 10 + DEX modifier
	// AC = 10 + DEX modifier + armor bonus + shield bonus
	baseAC := 10
	c.ArmorClass = baseAC + c.Modifiers.Dexterity

	// Add armor bonuses from equipment
	for _, itemID := range c.Equipment {
		if armor, ok := gd.GetArmor(itemID); ok {
			c.ArmorClass += armor.ACBonus
		}
	}
}

// InitializeSpellSlots sets up spell slots based on class and level.
// Returns true if the character is a spellcaster, false otherwise.
func (c *Character) InitializeSpellSlots(gd *data.GameData) bool {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return false
	}

	// Check if class is a spellcaster
	if class.SpellcastingAbility == "" {
		return false
	}

	// Get spell slots from tables based on class and level
	slots := GetSpellSlots(c.Class, c.Level)
	if slots == nil || len(slots) == 0 {
		// Not a caster at this level yet (e.g., Paladin level 1)
		c.SpellSlots = make(map[int]int)
		c.SpellSlotsUsed = make(map[int]int)
		return false
	}

	// Initialize spell slots
	c.SpellSlots = make(map[int]int)
	c.SpellSlotsUsed = make(map[int]int)

	for level, count := range slots {
		c.SpellSlots[level] = count
		c.SpellSlotsUsed[level] = 0
	}

	// Calculate spell save DC and spell attack bonus
	// Formula: 8 + proficiency bonus + spellcasting ability modifier
	abilityMod := c.GetAbilityModifier(class.SpellcastingAbility)
	c.SpellSaveDC = 8 + c.ProficiencyBonus + abilityMod
	c.SpellAttackBonus = c.ProficiencyBonus + abilityMod

	return true
}

// CanCastSpells returns true if the character's class can cast spells.
func (c *Character) CanCastSpells(gd *data.GameData) bool {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return false
	}
	return class.SpellcastingAbility != ""
}

// GetSpellType returns the spell type for the character's class.
// Returns "arcane" or "divine" based on D&D 5e class, or "" for non-casters.
func (c *Character) GetSpellType(gd *data.GameData) string {
	switch c.Class {
	case "wizard", "sorcerer", "warlock", "bard":
		return "arcane"
	case "cleric", "druid", "paladin", "ranger":
		return "divine"
	default:
		return ""
	}
}

// UseSpellSlot consumes a spell slot of the given level.
// Returns an error if no slots are available at that level.
func (c *Character) UseSpellSlot(level int) error {
	if c.SpellSlots == nil {
		return fmt.Errorf("character has no spell slots")
	}

	maxSlots, exists := c.SpellSlots[level]
	if !exists || maxSlots == 0 {
		return fmt.Errorf("no spell slots available at level %d", level)
	}

	if c.SpellSlotsUsed == nil {
		c.SpellSlotsUsed = make(map[int]int)
	}

	used := c.SpellSlotsUsed[level]
	if used >= maxSlots {
		return fmt.Errorf("all level %d spell slots have been used (%d/%d)", level, used, maxSlots)
	}

	c.SpellSlotsUsed[level]++
	return nil
}

// RestoreSpellSlots restores all spell slots (typically after a long rest).
// For Warlock, this should be called after short rest as well.
func (c *Character) RestoreSpellSlots() {
	if c.SpellSlotsUsed != nil {
		for level := range c.SpellSlotsUsed {
			c.SpellSlotsUsed[level] = 0
		}
	}
}

// GetAvailableSlots returns the number of available (unused) spell slots at the given level.
func (c *Character) GetAvailableSlots(level int) int {
	if c.SpellSlots == nil {
		return 0
	}

	maxSlots, exists := c.SpellSlots[level]
	if !exists {
		return 0
	}

	used := 0
	if c.SpellSlotsUsed != nil {
		used = c.SpellSlotsUsed[level]
	}

	available := maxSlots - used
	if available < 0 {
		return 0
	}

	return available
}

// CanCastSpell returns true if the character has at least one spell slot available
// at the given spell level or higher (for upcasting).
func (c *Character) CanCastSpell(spellLevel int) bool {
	if c.SpellSlots == nil {
		return false
	}

	// Check if any slot at this level or higher is available
	for level := spellLevel; level <= 9; level++ {
		if c.GetAvailableSlots(level) > 0 {
			return true
		}
	}

	return false
}

// Validate checks if the character's species/class combination is valid.
// In D&D 5e, all species can play all classes.
func (c *Character) Validate(gd *data.GameData) error {
	if c.Name == "" {
		return fmt.Errorf("character name is required")
	}

	species, ok := gd.GetSpecies(c.Species)
	if !ok {
		return fmt.Errorf("unknown species: %s", c.Species)
	}

	if _, ok := gd.GetClass(c.Class); !ok {
		return fmt.Errorf("unknown class: %s", c.Class)
	}

	// D&D 5e: All species can play all classes
	_ = species // Used for validation

	return nil
}

// Save writes the character to a JSON file.
func (c *Character) Save(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	filename := sanitizeFilename(c.Name) + ".json"
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling character: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// UpdateAppearance updates character appearance.
func (c *Character) UpdateAppearance(appearance CharacterAppearance) {
	c.Appearance = &appearance
}

// GetVisualDescription returns human-readable description.
func (c *Character) GetVisualDescription() string {
	if c.Appearance == nil {
		return fmt.Sprintf("%s, %s %s", c.Name, c.Species, c.Class)
	}

	a := c.Appearance
	parts := []string{c.Name}

	// Age and species
	if a.Age > 0 {
		parts = append(parts, fmt.Sprintf("%d-year-old %s", a.Age, c.Species))
	} else {
		parts = append(parts, c.Species)
	}

	parts = append(parts, c.Class)

	// Physical traits
	if a.Build != "" || a.Height != "" {
		physical := []string{}
		if a.Height != "" {
			physical = append(physical, a.Height)
		}
		if a.Build != "" {
			physical = append(physical, a.Build)
		}
		parts = append(parts, strings.Join(physical, ", "))
	}

	// Distinctive features
	if a.DistinctiveFeature != "" {
		parts = append(parts, fmt.Sprintf("with %s", a.DistinctiveFeature))
	}

	return strings.Join(parts, ", ")
}

// GetImagePromptSnippet returns short reference for image prompts.
func (c *Character) GetImagePromptSnippet() string {
	if c.Appearance == nil {
		return fmt.Sprintf("%s the %s %s", c.Name, c.Species, c.Class)
	}

	// Short form: "Aldric (human fighter, plate armor, longsword)"
	equipment := []string{}
	if c.Appearance.ArmorDescription != "" {
		equipment = append(equipment, c.Appearance.ArmorDescription)
	}
	if c.Appearance.WeaponDescription != "" {
		equipment = append(equipment, c.Appearance.WeaponDescription)
	}

	if len(equipment) > 0 {
		return fmt.Sprintf("%s (%s %s, %s)",
			c.Name, c.Species, c.Class, strings.Join(equipment, ", "))
	}

	return fmt.Sprintf("%s (%s %s)", c.Name, c.Species, c.Class)
}

// Load reads a character from a JSON file.
func Load(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var c Character
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling character: %w", err)
	}

	return &c, nil
}

// ListCharacters returns all characters in a directory.
func ListCharacters(dir string) ([]*Character, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Character{}, nil
		}
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var characters []*Character
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		c, err := Load(path)
		if err != nil {
			continue // Skip invalid files
		}
		characters = append(characters, c)
	}

	return characters, nil
}

// Delete removes a character file.
func Delete(dir, name string) error {
	filename := sanitizeFilename(name) + ".json"
	path := filepath.Join(dir, filename)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("deleting character: %w", err)
	}

	return nil
}

// ToMarkdown generates a readable character sheet.
func (c *Character) ToMarkdown(gd *data.GameData) string {
	var sb strings.Builder

	// Header
	species, _ := gd.GetSpecies(c.Species)
	class, _ := gd.GetClass(c.Class)

	speciesName := c.Species
	className := c.Class
	if species != nil {
		speciesName = species.Name
	}
	if class != nil {
		className = class.Name
	}

	sb.WriteString(fmt.Sprintf("# %s\n", c.Name))
	sb.WriteString(fmt.Sprintf("**%s %s, Niveau %d**\n\n", speciesName, className, c.Level))

	// Abilities
	sb.WriteString("## Caractéristiques\n\n")
	sb.WriteString("| Attribut | Score | Mod |\n")
	sb.WriteString("|----------|-------|-----|\n")
	sb.WriteString(fmt.Sprintf("| Force | %d | %s |\n", c.Abilities.Strength, formatMod(c.Modifiers.Strength)))
	sb.WriteString(fmt.Sprintf("| Intelligence | %d | %s |\n", c.Abilities.Intelligence, formatMod(c.Modifiers.Intelligence)))
	sb.WriteString(fmt.Sprintf("| Sagesse | %d | %s |\n", c.Abilities.Wisdom, formatMod(c.Modifiers.Wisdom)))
	sb.WriteString(fmt.Sprintf("| Dextérité | %d | %s |\n", c.Abilities.Dexterity, formatMod(c.Modifiers.Dexterity)))
	sb.WriteString(fmt.Sprintf("| Constitution | %d | %s |\n", c.Abilities.Constitution, formatMod(c.Modifiers.Constitution)))
	sb.WriteString(fmt.Sprintf("| Charisme | %d | %s |\n", c.Abilities.Charisma, formatMod(c.Modifiers.Charisma)))

	// Combat
	sb.WriteString("\n## Combat\n\n")
	sb.WriteString(fmt.Sprintf("- **Points de vie** : %d/%d\n", c.HitPoints, c.MaxHitPoints))
	sb.WriteString(fmt.Sprintf("- **Classe d'armure** : %d\n", c.ArmorClass))
	sb.WriteString(fmt.Sprintf("- **Bonus de maîtrise** : +%d\n", c.ProficiencyBonus))

	// Attack bonus = proficiency + primary ability modifier
	if class != nil && class.PrimaryAbility != "" {
		var abilityMod int
		switch class.PrimaryAbility {
		case "strength":
			abilityMod = c.Modifiers.Strength
		case "dexterity":
			abilityMod = c.Modifiers.Dexterity
		case "intelligence":
			abilityMod = c.Modifiers.Intelligence
		case "wisdom":
			abilityMod = c.Modifiers.Wisdom
		case "charisma":
			abilityMod = c.Modifiers.Charisma
		}
		attackBonus := c.ProficiencyBonus + abilityMod
		sb.WriteString(fmt.Sprintf("- **Bonus d'attaque** : %s\n", formatMod(attackBonus)))
	}

	// Equipment
	if len(c.Equipment) > 0 {
		sb.WriteString("\n## Équipement\n\n")
		for _, item := range c.Equipment {
			itemName := item
			if weapon, ok := gd.GetWeapon(item); ok {
				itemName = fmt.Sprintf("%s (%s)", weapon.Name, weapon.Damage)
			} else if armor, ok := gd.GetArmor(item); ok {
				itemName = fmt.Sprintf("%s (CA +%d)", armor.Name, armor.ACBonus)
			}
			sb.WriteString(fmt.Sprintf("- %s\n", itemName))
		}
	}

	// Gold
	sb.WriteString(fmt.Sprintf("\n## Or : %d po\n", c.Gold))

	// XP
	sb.WriteString(fmt.Sprintf("\n## Expérience : %d XP\n", c.XP))

	// Spells section (if applicable)
	if c.CanCastSpells(gd) {
		sb.WriteString("\n## Magie\n\n")
		spellType := c.GetSpellType(gd)
		if spellType == "arcane" {
			sb.WriteString("**Type** : Arcanique (Magicien)\n\n")
		} else if spellType == "divine" {
			sb.WriteString("**Type** : Divine (Clerc)\n\n")
		}

		// Spell slots
		if c.SpellSlots != nil && len(c.SpellSlots) > 0 {
			sb.WriteString("### Emplacements de sorts\n\n")
			sb.WriteString("| Niveau | Disponible | Utilisé |\n")
			sb.WriteString("|--------|------------|----------|\n")
			for lvl := 1; lvl <= 6; lvl++ {
				if slots, ok := c.SpellSlots[lvl]; ok && slots > 0 {
					used := 0
					if c.SpellSlotsUsed != nil {
						used = c.SpellSlotsUsed[lvl]
					}
					sb.WriteString(fmt.Sprintf("| %d | %d | %d |\n", lvl, slots, used))
				}
			}
		} else {
			sb.WriteString("*Pas encore d'emplacements de sorts à ce niveau.*\n")
		}

		// Known spells
		if len(c.KnownSpells) > 0 {
			sb.WriteString("\n### Sorts connus\n\n")
			for _, spellID := range c.KnownSpells {
				sb.WriteString(fmt.Sprintf("- %s\n", spellID))
			}
		}

		// Prepared spells
		if len(c.PreparedSpells) > 0 {
			sb.WriteString("\n### Sorts préparés\n\n")
			for _, spellID := range c.PreparedSpells {
				sb.WriteString(fmt.Sprintf("- %s\n", spellID))
			}
		}
	}

	return sb.String()
}

// ToJSON returns the character as a JSON string.
func (c *Character) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CalculateProficiencyBonus calculates the D&D 5e proficiency bonus based on level.
func (c *Character) CalculateProficiencyBonus() {
	c.ProficiencyBonus = data.ProficiencyBonusByLevel(c.Level)
}

// CalculateSpellSaveDC calculates the spell save DC for a spellcaster.
// Formula: 8 + proficiency bonus + spellcasting ability modifier
func (c *Character) CalculateSpellSaveDC(gd *data.GameData) {
	class, ok := gd.GetClass(c.Class)
	if !ok || class.SpellcastingAbility == "" {
		return
	}

	var abilityMod int
	switch class.SpellcastingAbility {
	case "intelligence":
		abilityMod = c.Modifiers.Intelligence
	case "wisdom":
		abilityMod = c.Modifiers.Wisdom
	case "charisma":
		abilityMod = c.Modifiers.Charisma
	}

	c.SpellSaveDC = 8 + c.ProficiencyBonus + abilityMod
	c.SpellAttackBonus = c.ProficiencyBonus + abilityMod
}

// GetAbilityModifier returns the modifier for a specific ability.
func (c *Character) GetAbilityModifier(ability string) int {
	switch ability {
	case "strength":
		return c.Modifiers.Strength
	case "intelligence":
		return c.Modifiers.Intelligence
	case "wisdom":
		return c.Modifiers.Wisdom
	case "dexterity":
		return c.Modifiers.Dexterity
	case "constitution":
		return c.Modifiers.Constitution
	case "charisma":
		return c.Modifiers.Charisma
	default:
		return 0
	}
}

// Helper functions

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func formatMod(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

func sanitizeFilename(name string) string {
	// Replace spaces and special characters
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\"", "")
	return name
}
