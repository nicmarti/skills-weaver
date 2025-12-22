// Package character provides character creation and management for BFRPG.
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

// Character represents a player character in BFRPG.
type Character struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Race         string        `json:"race"`
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
	// Spell system fields
	KnownSpells    []string       `json:"known_spells,omitempty"`    // Spell IDs known by the character
	PreparedSpells []string       `json:"prepared_spells,omitempty"` // Spell IDs prepared for the day
	SpellSlots     map[int]int          `json:"spell_slots,omitempty"`     // Available spell slots by level
	SpellSlotsUsed map[int]int          `json:"spell_slots_used,omitempty"` // Used spell slots by level
	Appearance     *CharacterAppearance `json:"appearance,omitempty"`       // Visual description for image generation
	CreatedAt      time.Time            `json:"created_at"`
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
func New(name, race, class string) *Character {
	return &Character{
		ID:        uuid.New().String(),
		Name:      name,
		Race:      race,
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

// ApplyRacialModifiers applies racial ability modifiers from game data.
func (c *Character) ApplyRacialModifiers(gd *data.GameData) error {
	race, ok := gd.GetRace(c.Race)
	if !ok {
		return fmt.Errorf("unknown race: %s", c.Race)
	}

	for ability, mod := range race.AbilityModifiers {
		switch ability {
		case "strength":
			c.Abilities.Strength += mod
		case "intelligence":
			c.Abilities.Intelligence += mod
		case "wisdom":
			c.Abilities.Wisdom += mod
		case "dexterity":
			c.Abilities.Dexterity += mod
		case "constitution":
			c.Abilities.Constitution += mod
		case "charisma":
			c.Abilities.Charisma += mod
		}
	}

	// Clamp values to 3-18 range
	c.Abilities.Strength = clamp(c.Abilities.Strength, 3, 18)
	c.Abilities.Intelligence = clamp(c.Abilities.Intelligence, 3, 18)
	c.Abilities.Wisdom = clamp(c.Abilities.Wisdom, 3, 18)
	c.Abilities.Dexterity = clamp(c.Abilities.Dexterity, 3, 18)
	c.Abilities.Constitution = clamp(c.Abilities.Constitution, 3, 18)
	c.Abilities.Charisma = clamp(c.Abilities.Charisma, 3, 18)

	return nil
}

// CalculateModifiers computes ability modifiers according to BFRPG rules.
func (c *Character) CalculateModifiers() {
	c.Modifiers.Strength = abilityModifier(c.Abilities.Strength)
	c.Modifiers.Intelligence = abilityModifier(c.Abilities.Intelligence)
	c.Modifiers.Wisdom = abilityModifier(c.Abilities.Wisdom)
	c.Modifiers.Dexterity = abilityModifier(c.Abilities.Dexterity)
	c.Modifiers.Constitution = abilityModifier(c.Abilities.Constitution)
	c.Modifiers.Charisma = abilityModifier(c.Abilities.Charisma)
}

// abilityModifier returns the BFRPG modifier for an ability score.
func abilityModifier(score int) int {
	switch {
	case score <= 3:
		return -3
	case score <= 5:
		return -2
	case score <= 8:
		return -1
	case score <= 12:
		return 0
	case score <= 15:
		return +1
	case score <= 17:
		return +2
	default:
		return +3
	}
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

// RollStartingGold rolls starting gold based on class.
func (c *Character) RollStartingGold(gd *data.GameData) error {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return fmt.Errorf("unknown class: %s", c.Class)
	}

	roller := dice.New()

	// Parse starting gold expression (e.g., "3d6x10" or "2d6x10")
	goldExpr := class.StartingGold
	if goldExpr == "" {
		goldExpr = "3d6x10"
	}

	// Handle "3d6x10" format
	goldExpr = strings.Replace(goldExpr, "x10", "", 1)
	result, err := roller.Roll(goldExpr)
	if err != nil {
		return fmt.Errorf("rolling starting gold: %w", err)
	}

	c.Gold = result.Total * 10

	return nil
}

// CalculateArmorClass computes AC (base 11 in BFRPG ascending AC).
func (c *Character) CalculateArmorClass(gd *data.GameData) {
	// Base AC in BFRPG (ascending) is 11
	// AC = 11 + DEX modifier + armor bonus + shield bonus
	baseAC := 11
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

	// Check if class has spells
	if class.SpellsPerLevel == nil || len(class.SpellsPerLevel) == 0 {
		return false
	}

	// Get spell slots for current level
	levelKey := fmt.Sprintf("%d", c.Level)
	slots, ok := class.SpellsPerLevel[levelKey]
	if !ok || len(slots) == 0 {
		// No spells at this level (e.g., Cleric level 1)
		c.SpellSlots = nil
		c.SpellSlotsUsed = nil
		return true // Still a spellcaster, just no slots yet
	}

	// Initialize spell slots by spell level
	c.SpellSlots = make(map[int]int)
	c.SpellSlotsUsed = make(map[int]int)

	for spellLevel, count := range slots {
		c.SpellSlots[spellLevel+1] = count // spellLevel is 0-indexed, spell levels are 1-indexed
		c.SpellSlotsUsed[spellLevel+1] = 0
	}

	return true
}

// CanCastSpells returns true if the character's class can cast spells.
func (c *Character) CanCastSpells(gd *data.GameData) bool {
	class, ok := gd.GetClass(c.Class)
	if !ok {
		return false
	}
	return class.SpellsPerLevel != nil && len(class.SpellsPerLevel) > 0
}

// GetSpellType returns the spell type for the character's class.
// Returns "arcane" for Magic-User, "divine" for Cleric, or "" for non-casters.
func (c *Character) GetSpellType(gd *data.GameData) string {
	switch c.Class {
	case "magic-user":
		return "arcane"
	case "cleric":
		return "divine"
	default:
		return ""
	}
}

// Validate checks if the character's race/class combination is valid.
func (c *Character) Validate(gd *data.GameData) error {
	if c.Name == "" {
		return fmt.Errorf("character name is required")
	}

	race, ok := gd.GetRace(c.Race)
	if !ok {
		return fmt.Errorf("unknown race: %s", c.Race)
	}

	if _, ok := gd.GetClass(c.Class); !ok {
		return fmt.Errorf("unknown class: %s", c.Class)
	}

	// Check if race can play this class
	allowed := false
	for _, allowedClass := range race.AllowedClasses {
		if allowedClass == c.Class {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("%s cannot be a %s", race.Name, c.Class)
	}

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
		return fmt.Sprintf("%s, %s %s", c.Name, c.Race, c.Class)
	}

	a := c.Appearance
	parts := []string{c.Name}

	// Age and race
	if a.Age > 0 {
		parts = append(parts, fmt.Sprintf("%d-year-old %s", a.Age, c.Race))
	} else {
		parts = append(parts, c.Race)
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
		return fmt.Sprintf("%s the %s %s", c.Name, c.Race, c.Class)
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
			c.Name, c.Race, c.Class, strings.Join(equipment, ", "))
	}

	return fmt.Sprintf("%s (%s %s)", c.Name, c.Race, c.Class)
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
	race, _ := gd.GetRace(c.Race)
	class, _ := gd.GetClass(c.Class)

	raceName := c.Race
	className := c.Class
	if race != nil {
		raceName = race.Name
	}
	if class != nil {
		className = class.Name
	}

	sb.WriteString(fmt.Sprintf("# %s\n", c.Name))
	sb.WriteString(fmt.Sprintf("**%s %s, Niveau %d**\n\n", raceName, className, c.Level))

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

	if class != nil {
		if ab, ok := class.AttackBonus["1"]; ok {
			sb.WriteString(fmt.Sprintf("- **Bonus d'attaque** : +%d\n", ab))
		}
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
