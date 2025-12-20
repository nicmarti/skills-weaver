package names

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenderNames holds first and last names for a gender.
type GenderNames struct {
	First []string `json:"first"`
	Last  []string `json:"last"`
}

// RaceNames holds names for both genders.
type RaceNames struct {
	Male   GenderNames `json:"male"`
	Female GenderNames `json:"female"`
}

// NPCNames holds names for different NPC types.
type NPCNames struct {
	Innkeeper []string `json:"innkeeper"`
	Merchant  []string `json:"merchant"`
	Guard     []string `json:"guard"`
	Noble     []string `json:"noble"`
	Wizard    []string `json:"wizard"`
	Villain   []string `json:"villain"`
}

// NamesData holds all name data.
type NamesData struct {
	Dwarf    RaceNames `json:"dwarf"`
	Elf      RaceNames `json:"elf"`
	Halfling RaceNames `json:"halfling"`
	Human    RaceNames `json:"human"`
	NPC      NPCNames  `json:"npc"`
}

// Generator generates random names.
type Generator struct {
	data *NamesData
	rng  *rand.Rand
}

// NewGenerator creates a new name generator.
func NewGenerator(dataDir string) (*Generator, error) {
	path := filepath.Join(dataDir, "names.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading names.json: %w", err)
	}

	var names NamesData
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("parsing names.json: %w", err)
	}

	return &Generator{
		data: &names,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// GenerateName generates a random full name for a race and gender.
func (g *Generator) GenerateName(race, gender string) (string, error) {
	race = strings.ToLower(race)
	gender = strings.ToLower(gender)

	var raceNames *RaceNames
	switch race {
	case "dwarf", "nain":
		raceNames = &g.data.Dwarf
	case "elf", "elfe":
		raceNames = &g.data.Elf
	case "halfling", "halfelin":
		raceNames = &g.data.Halfling
	case "human", "humain":
		raceNames = &g.data.Human
	default:
		return "", fmt.Errorf("unknown race: %s", race)
	}

	var genderNames *GenderNames
	switch gender {
	case "male", "m", "masculin", "homme":
		genderNames = &raceNames.Male
	case "female", "f", "feminin", "femme":
		genderNames = &raceNames.Female
	default:
		// Random gender
		if g.rng.Intn(2) == 0 {
			genderNames = &raceNames.Male
		} else {
			genderNames = &raceNames.Female
		}
	}

	if len(genderNames.First) == 0 || len(genderNames.Last) == 0 {
		return "", fmt.Errorf("no names available for %s %s", race, gender)
	}

	first := genderNames.First[g.rng.Intn(len(genderNames.First))]
	last := genderNames.Last[g.rng.Intn(len(genderNames.Last))]

	return fmt.Sprintf("%s %s", first, last), nil
}

// GenerateFirstName generates only a first name.
func (g *Generator) GenerateFirstName(race, gender string) (string, error) {
	race = strings.ToLower(race)
	gender = strings.ToLower(gender)

	var raceNames *RaceNames
	switch race {
	case "dwarf", "nain":
		raceNames = &g.data.Dwarf
	case "elf", "elfe":
		raceNames = &g.data.Elf
	case "halfling", "halfelin":
		raceNames = &g.data.Halfling
	case "human", "humain":
		raceNames = &g.data.Human
	default:
		return "", fmt.Errorf("unknown race: %s", race)
	}

	var genderNames *GenderNames
	switch gender {
	case "male", "m", "masculin", "homme":
		genderNames = &raceNames.Male
	case "female", "f", "feminin", "femme":
		genderNames = &raceNames.Female
	default:
		if g.rng.Intn(2) == 0 {
			genderNames = &raceNames.Male
		} else {
			genderNames = &raceNames.Female
		}
	}

	if len(genderNames.First) == 0 {
		return "", fmt.Errorf("no first names available for %s %s", race, gender)
	}

	return genderNames.First[g.rng.Intn(len(genderNames.First))], nil
}

// GenerateNPCName generates a name for a specific NPC type.
func (g *Generator) GenerateNPCName(npcType string) (string, error) {
	npcType = strings.ToLower(npcType)

	var names []string
	switch npcType {
	case "innkeeper", "tavernier", "aubergiste":
		names = g.data.NPC.Innkeeper
	case "merchant", "marchand", "commercant":
		names = g.data.NPC.Merchant
	case "guard", "garde", "soldat":
		names = g.data.NPC.Guard
	case "noble", "seigneur", "dame":
		names = g.data.NPC.Noble
	case "wizard", "mage", "sorcier", "magicien":
		names = g.data.NPC.Wizard
	case "villain", "vilain", "mechant", "ennemi":
		names = g.data.NPC.Villain
	default:
		return "", fmt.Errorf("unknown NPC type: %s", npcType)
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no names available for NPC type: %s", npcType)
	}

	return names[g.rng.Intn(len(names))], nil
}

// GenerateMultiple generates multiple names.
func (g *Generator) GenerateMultiple(race, gender string, count int) ([]string, error) {
	names := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(names) < count {
		name, err := g.GenerateName(race, gender)
		if err != nil {
			return nil, err
		}

		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}

		// Prevent infinite loop if not enough unique names
		if len(seen) > count*10 {
			break
		}
	}

	return names, nil
}

// GetAvailableRaces returns the list of available races.
func (g *Generator) GetAvailableRaces() []string {
	return []string{"dwarf", "elf", "halfling", "human"}
}

// GetAvailableNPCTypes returns the list of available NPC types.
func (g *Generator) GetAvailableNPCTypes() []string {
	return []string{"innkeeper", "merchant", "guard", "noble", "wizard", "villain"}
}