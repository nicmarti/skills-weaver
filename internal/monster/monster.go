package monster

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"dungeons/internal/dice"
)

// Abilities represents a monster's ability scores (D&D 5e).
type Abilities struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// Attack represents a monster's attack.
type Attack struct {
	Name       string `json:"name"`
	NameFR     string `json:"name_fr"`
	Bonus      int    `json:"bonus"`        // Attack bonus (to-hit modifier)
	Damage     string `json:"damage"`       // Damage dice (e.g., "1d8+2")
	DamageAvg  int    `json:"damage_avg"`   // Average damage
	DamageType string `json:"damage_type,omitempty"` // slashing, piercing, bludgeoning, etc. (D&D 5e)
	Special    string `json:"special,omitempty"` // Special effects
}

// Monster represents a creature in the game.
type Monster struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	NameFR        string   `json:"name_fr"`
	Type          string   `json:"type"`
	Size          string   `json:"size"`

	// Legacy BFRPG fields (deprecated - D&D 5e only)
	HitDice       string   `json:"hit_dice,omitempty"`       // Deprecated: use hit_points_avg
	HitPointsAvg  int      `json:"hit_points_avg"`
	SaveAs        string   `json:"save_as,omitempty"`        // Deprecated: D&D 5e uses proficiency bonus
	Morale        int      `json:"morale,omitempty"`         // Deprecated: not used in D&D 5e

	// D&D 5e fields
	ChallengeRating   string    `json:"challenge_rating,omitempty"` // "0", "1/8", "1/4", "1/2", "1", "2", etc.
	ProficiencyBonus  int       `json:"proficiency_bonus,omitempty"` // +2 to +9
	Abilities         *Abilities `json:"abilities,omitempty"` // Ability scores (D&D 5e)

	// Common fields
	ArmorClass    int      `json:"armor_class"`
	Attacks       []Attack `json:"attacks"`
	Movement      int      `json:"movement"`
	MovementFly   int      `json:"movement_fly,omitempty"`
	TreasureType  string   `json:"treasure_type"`
	XP            int      `json:"xp"`
	Special       []string `json:"special"`
	DescriptionFR string   `json:"description_fr"`
}

// GetCRValue converts the Challenge Rating string to a float64 (D&D 5e).
// CR can be: "0", "1/8", "1/4", "1/2", "1", "2", "3", etc.
func (m *Monster) GetCRValue() float64 {
	switch m.ChallengeRating {
	case "0":
		return 0
	case "1/8":
		return 0.125
	case "1/4":
		return 0.25
	case "1/2":
		return 0.5
	default:
		val, err := strconv.ParseFloat(m.ChallengeRating, 64)
		if err != nil {
			return 0
		}
		return val
	}
}

// GetCRXP returns the XP value for a given CR (D&D 5e).
func GetCRXP(cr string) int {
	xpByCR := map[string]int{
		"0":    10,
		"1/8":  25,
		"1/4":  50,
		"1/2":  100,
		"1":    200,
		"2":    450,
		"3":    700,
		"4":    1100,
		"5":    1800,
		"6":    2300,
		"7":    2900,
		"8":    3900,
		"9":    5000,
		"10":   5900,
		"11":   7200,
		"12":   8400,
		"13":   10000,
		"14":   11500,
		"15":   13000,
		"16":   15000,
		"17":   18000,
		"18":   20000,
		"19":   22000,
		"20":   25000,
		"21":   33000,
		"22":   41000,
		"23":   50000,
		"24":   62000,
		"25":   75000,
		"26":   90000,
		"27":   105000,
		"28":   120000,
		"29":   135000,
		"30":   155000,
	}
	if xp, ok := xpByCR[cr]; ok {
		return xp
	}
	return 0
}

// IsBFRPG is deprecated. All monsters now use D&D 5e format.
// Kept for backward compatibility.
func (m *Monster) IsBFRPG() bool {
	return false
}

// IsDnD5e returns true (all monsters are D&D 5e format).
func (m *Monster) IsDnD5e() bool {
	return true
}

// EncounterEntry represents a monster entry in an encounter table.
type EncounterEntry struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Weight int    `json:"weight"`
}

// EncounterTable represents a table for generating random encounters.
type EncounterTable struct {
	Description string           `json:"description"`
	Monsters    []EncounterEntry `json:"monsters"`
}

// MonstersData holds all monster data from JSON.
type MonstersData struct {
	Monsters        []Monster                  `json:"monsters"`
	EncounterTables map[string]EncounterTable `json:"encounter_tables"`
}

// MonsterInstance represents a specific monster with rolled HP.
type MonsterInstance struct {
	Monster   *Monster
	HitPoints int
	MaxHP     int
}

// EncounterResult represents a generated encounter.
type EncounterResult struct {
	TableName   string
	Description string
	Monsters    []MonsterInstance
	TotalXP     int
}

// Bestiary manages monster data and encounter generation.
type Bestiary struct {
	data    *MonstersData
	rng     *rand.Rand
	roller  *dice.Roller
	dataDir string
}

// NewBestiary creates a new bestiary from the data directory.
// Loads D&D 5e monsters from multiple JSON files in data/5e/:
// - monsters.json (beasts, undead, etc.)
// - humanoids.json (guards, bandits, cultists, etc.)
func NewBestiary(dataDir string) (*Bestiary, error) {
	// List of JSON files to load
	files := []string{
		filepath.Join(dataDir, "5e", "monsters.json"),
		filepath.Join(dataDir, "5e", "humanoids.json"),
	}

	// Merged data
	var allData MonstersData
	allData.Monsters = []Monster{}
	allData.EncounterTables = make(map[string]EncounterTable)

	// Load each file
	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			// Skip missing files (humanoids.json may not exist in older installations)
			continue
		}

		var fileData MonstersData
		if err := json.Unmarshal(data, &fileData); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", filepath.Base(filePath), err)
		}

		// Merge monsters
		allData.Monsters = append(allData.Monsters, fileData.Monsters...)

		// Merge encounter tables
		for name, table := range fileData.EncounterTables {
			allData.EncounterTables[name] = table
		}
	}

	// Ensure at least one file was loaded
	if len(allData.Monsters) == 0 {
		return nil, fmt.Errorf("no monsters found in data/5e/ directory")
	}

	return &Bestiary{
		data:    &allData,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		roller:  dice.New(),
		dataDir: dataDir,
	}, nil
}

// GetMonster returns a monster by ID.
func (b *Bestiary) GetMonster(id string) (*Monster, error) {
	id = strings.ToLower(id)
	for i := range b.data.Monsters {
		if strings.ToLower(b.data.Monsters[i].ID) == id {
			return &b.data.Monsters[i], nil
		}
	}
	return nil, fmt.Errorf("monster not found: %s", id)
}

// SearchMonsters searches monsters by name or type.
func (b *Bestiary) SearchMonsters(query string) []*Monster {
	query = strings.ToLower(query)
	var results []*Monster

	for i := range b.data.Monsters {
		m := &b.data.Monsters[i]
		if strings.Contains(strings.ToLower(m.Name), query) ||
			strings.Contains(strings.ToLower(m.NameFR), query) ||
			strings.Contains(strings.ToLower(m.Type), query) ||
			strings.Contains(strings.ToLower(m.ID), query) {
			results = append(results, m)
		}
	}

	return results
}

// ListByType returns all monsters of a given type.
func (b *Bestiary) ListByType(monsterType string) []*Monster {
	monsterType = strings.ToLower(monsterType)
	var results []*Monster

	for i := range b.data.Monsters {
		if strings.ToLower(b.data.Monsters[i].Type) == monsterType {
			results = append(results, &b.data.Monsters[i])
		}
	}

	// Sort by XP
	sort.Slice(results, func(i, j int) bool {
		return results[i].XP < results[j].XP
	})

	return results
}

// ListAll returns all monsters.
func (b *Bestiary) ListAll() []*Monster {
	results := make([]*Monster, len(b.data.Monsters))
	for i := range b.data.Monsters {
		results[i] = &b.data.Monsters[i]
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// ListByCR returns all monsters of a given Challenge Rating (D&D 5e).
func (b *Bestiary) ListByCR(cr string) []*Monster {
	var results []*Monster

	for i := range b.data.Monsters {
		if b.data.Monsters[i].ChallengeRating == cr {
			results = append(results, &b.data.Monsters[i])
		}
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// GetTypes returns all unique monster types.
func (b *Bestiary) GetTypes() []string {
	typeSet := make(map[string]bool)
	for _, m := range b.data.Monsters {
		typeSet[m.Type] = true
	}

	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// GetEncounterTables returns the list of available encounter tables.
func (b *Bestiary) GetEncounterTables() []string {
	tables := make([]string, 0, len(b.data.EncounterTables))
	for name := range b.data.EncounterTables {
		tables = append(tables, name)
	}
	sort.Strings(tables)
	return tables
}

// RollHP rolls hit points for a monster.
func (b *Bestiary) RollHP(m *Monster) int {
	result, err := b.roller.Roll(m.HitDice)
	if err != nil {
		return m.HitPointsAvg
	}
	if result.Total < 1 {
		return 1
	}
	return result.Total
}

// CreateInstance creates a monster instance with rolled HP.
func (b *Bestiary) CreateInstance(m *Monster) MonsterInstance {
	hp := b.RollHP(m)
	return MonsterInstance{
		Monster:   m,
		HitPoints: hp,
		MaxHP:     hp,
	}
}

// GenerateEncounter generates a random encounter from a table.
func (b *Bestiary) GenerateEncounter(tableName string) (*EncounterResult, error) {
	table, ok := b.data.EncounterTables[tableName]
	if !ok {
		return nil, fmt.Errorf("encounter table not found: %s", tableName)
	}

	// Calculate total weight
	totalWeight := 0
	for _, entry := range table.Monsters {
		totalWeight += entry.Weight
	}

	// Select a monster type based on weight
	roll := b.rng.Intn(totalWeight)
	var selectedEntry EncounterEntry
	for _, entry := range table.Monsters {
		roll -= entry.Weight
		if roll < 0 {
			selectedEntry = entry
			break
		}
	}

	// Get the monster
	monster, err := b.GetMonster(selectedEntry.ID)
	if err != nil {
		return nil, err
	}

	// Roll for number of monsters
	numResult, err := b.roller.Roll(selectedEntry.Number)
	var count int
	if err != nil {
		count = 1
	} else {
		count = numResult.Total
	}
	if count < 1 {
		count = 1
	}

	// Create instances
	instances := make([]MonsterInstance, count)
	totalXP := 0
	for i := 0; i < count; i++ {
		instances[i] = b.CreateInstance(monster)
		totalXP += monster.XP
	}

	return &EncounterResult{
		TableName:   tableName,
		Description: table.Description,
		Monsters:    instances,
		TotalXP:     totalXP,
	}, nil
}

// GenerateEncounterByLevel generates an encounter suitable for a party level.
func (b *Bestiary) GenerateEncounterByLevel(partyLevel int) (*EncounterResult, error) {
	var tableName string
	switch {
	case partyLevel <= 2:
		tableName = "dungeon_level_1"
	case partyLevel <= 4:
		tableName = "dungeon_level_2"
	case partyLevel <= 6:
		tableName = "dungeon_level_3"
	default:
		tableName = "dungeon_level_4"
	}
	return b.GenerateEncounter(tableName)
}

// ToMarkdown returns a formatted markdown description of a monster.
func (m *Monster) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", m.NameFR, m.Name))

	// Basic stats
	sb.WriteString(fmt.Sprintf("**Type** : %s | **Taille** : %s\n\n", m.Type, m.Size))

	// D&D 5e: Show abilities if present
	if m.Abilities != nil {
		sb.WriteString("### Caractéristiques\n\n")
		sb.WriteString("| FOR | DEX | CON | INT | SAG | CHA |\n")
		sb.WriteString("|-----|-----|-----|-----|-----|-----|\n")
		sb.WriteString(fmt.Sprintf("| %d | %d | %d | %d | %d | %d |\n\n",
			m.Abilities.Strength, m.Abilities.Dexterity, m.Abilities.Constitution,
			m.Abilities.Intelligence, m.Abilities.Wisdom, m.Abilities.Charisma))
	}

	// Combat stats (D&D 5e format)
	sb.WriteString("### Statistiques de Combat\n\n")
	sb.WriteString("| Stat | Valeur |\n")
	sb.WriteString("|------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Challenge Rating** | %s (XP %d) |\n", m.ChallengeRating, m.XP))
	sb.WriteString(fmt.Sprintf("| **Bonus de Maîtrise** | +%d |\n", m.ProficiencyBonus))
	sb.WriteString(fmt.Sprintf("| **Points de Vie** | %d (moyenne) |\n", m.HitPointsAvg))
	sb.WriteString(fmt.Sprintf("| **Classe d'Armure** | %d |\n", m.ArmorClass))
	sb.WriteString(fmt.Sprintf("| **Mouvement** | %d", m.Movement))
	if m.MovementFly > 0 {
		sb.WriteString(fmt.Sprintf(" (vol %d)", m.MovementFly))
	}
	sb.WriteString(" |\n")
	sb.WriteString(fmt.Sprintf("| **Trésor** | %s |\n", m.TreasureType))

	// Attacks
	sb.WriteString("\n### Attaques\n\n")
	for _, atk := range m.Attacks {
		sb.WriteString(fmt.Sprintf("- **%s** : +%d, %s", atk.NameFR, atk.Bonus, atk.Damage))
		if atk.DamageType != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", atk.DamageType))
		}
		if atk.Special != "" {
			sb.WriteString(fmt.Sprintf(" - %s", atk.Special))
		}
		sb.WriteString("\n")
	}

	// Special abilities
	if len(m.Special) > 0 {
		sb.WriteString("\n### Capacités Spéciales\n\n")
		for _, s := range m.Special {
			sb.WriteString(fmt.Sprintf("- %s\n", s))
		}
	}

	// Description
	sb.WriteString("\n### Description\n\n")
	sb.WriteString(m.DescriptionFR + "\n")

	return sb.String()
}

// ToShortDescription returns a one-line description (D&D 5e format).
func (m *Monster) ToShortDescription() string {
	return fmt.Sprintf("%s (%s) - CA %d, CR %s (%d PV), XP %d",
		m.NameFR, m.Type, m.ArmorClass, m.ChallengeRating, m.HitPointsAvg, m.XP)
}

// ToJSON returns the monster as JSON string.
func (m *Monster) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown returns a formatted description of an encounter.
func (e *EncounterResult) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Rencontre : %s\n\n", e.Description))

	// Group monsters by type
	monsterCounts := make(map[string][]MonsterInstance)
	for _, inst := range e.Monsters {
		monsterCounts[inst.Monster.ID] = append(monsterCounts[inst.Monster.ID], inst)
	}

	sb.WriteString("### Monstres\n\n")
	for _, instances := range monsterCounts {
		m := instances[0].Monster
		sb.WriteString(fmt.Sprintf("**%s** x%d\n", m.NameFR, len(instances)))
		sb.WriteString(fmt.Sprintf("- CA %d, ", m.ArmorClass))

		// List individual HP
		hps := make([]string, len(instances))
		for i, inst := range instances {
			hps[i] = strconv.Itoa(inst.HitPoints)
		}
		sb.WriteString(fmt.Sprintf("PV : %s\n", strings.Join(hps, ", ")))

		// Attacks
		for _, atk := range m.Attacks {
			sb.WriteString(fmt.Sprintf("- %s : +%d, %s", atk.NameFR, atk.Bonus, atk.Damage))
			if atk.Special != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", atk.Special))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("**XP Total** : %d\n", e.TotalXP))

	return sb.String()
}

// ToJSON returns the encounter as JSON string.
func (e *EncounterResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
