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

// Attack represents a monster's attack.
type Attack struct {
	Name      string `json:"name"`
	NameFR    string `json:"name_fr"`
	Bonus     int    `json:"bonus"`
	Damage    string `json:"damage"`
	DamageAvg int    `json:"damage_avg"`
	Special   string `json:"special,omitempty"`
}

// Monster represents a creature in the game.
type Monster struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	NameFR        string   `json:"name_fr"`
	Type          string   `json:"type"`
	Size          string   `json:"size"`
	HitDice       string   `json:"hit_dice"`
	HitPointsAvg  int      `json:"hit_points_avg"`
	ArmorClass    int      `json:"armor_class"`
	Attacks       []Attack `json:"attacks"`
	Movement      int      `json:"movement"`
	MovementFly   int      `json:"movement_fly,omitempty"`
	SaveAs        string   `json:"save_as"`
	Morale        int      `json:"morale"`
	TreasureType  string   `json:"treasure_type"`
	XP            int      `json:"xp"`
	Special       []string `json:"special"`
	DescriptionFR string   `json:"description_fr"`
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
func NewBestiary(dataDir string) (*Bestiary, error) {
	path := filepath.Join(dataDir, "monsters.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading monsters.json: %w", err)
	}

	var monstersData MonstersData
	if err := json.Unmarshal(data, &monstersData); err != nil {
		return nil, fmt.Errorf("parsing monsters.json: %w", err)
	}

	return &Bestiary{
		data:    &monstersData,
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

	// Combat stats
	sb.WriteString("### Statistiques de Combat\n\n")
	sb.WriteString(fmt.Sprintf("| Stat | Valeur |\n"))
	sb.WriteString(fmt.Sprintf("|------|--------|\n"))
	sb.WriteString(fmt.Sprintf("| **Dés de Vie** | %s (moy. %d PV) |\n", m.HitDice, m.HitPointsAvg))
	sb.WriteString(fmt.Sprintf("| **Classe d'Armure** | %d |\n", m.ArmorClass))
	sb.WriteString(fmt.Sprintf("| **Mouvement** | %d", m.Movement))
	if m.MovementFly > 0 {
		sb.WriteString(fmt.Sprintf(" (vol %d)", m.MovementFly))
	}
	sb.WriteString(" |\n")
	sb.WriteString(fmt.Sprintf("| **Sauvegarde** | %s |\n", m.SaveAs))
	sb.WriteString(fmt.Sprintf("| **Moral** | %d |\n", m.Morale))
	sb.WriteString(fmt.Sprintf("| **Trésor** | %s |\n", m.TreasureType))
	sb.WriteString(fmt.Sprintf("| **XP** | %d |\n", m.XP))

	// Attacks
	sb.WriteString("\n### Attaques\n\n")
	for _, atk := range m.Attacks {
		sb.WriteString(fmt.Sprintf("- **%s** : +%d, %s", atk.NameFR, atk.Bonus, atk.Damage))
		if atk.Special != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", atk.Special))
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

// ToShortDescription returns a one-line description.
func (m *Monster) ToShortDescription() string {
	return fmt.Sprintf("%s (%s) - CA %d, DV %s (%d PV), XP %d",
		m.NameFR, m.Type, m.ArmorClass, m.HitDice, m.HitPointsAvg, m.XP)
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
