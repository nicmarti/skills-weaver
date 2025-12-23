package spell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Spell represents a spell in BFRPG.
type Spell struct {
	ID             string   `json:"id"`
	NameEN         string   `json:"name_en"`
	NameFR         string   `json:"name_fr"`
	Level          int      `json:"level"`
	Type           string   `json:"type"` // "divine", "arcane", "both"
	Classes        []string `json:"classes,omitempty"`
	AlsoAvailable  []struct {
		Class string `json:"class"`
		Level int    `json:"level"`
	} `json:"also_available,omitempty"`
	Reversible      bool   `json:"reversible"`
	ReverseNameEN   string `json:"reverse_name_en,omitempty"`
	ReverseNameFR   string `json:"reverse_name_fr,omitempty"`
	Range           string `json:"range"`
	RangeReverse    string `json:"range_reverse,omitempty"`
	Duration        string `json:"duration"`
	DurationReverse string `json:"duration_reverse,omitempty"`
	DescriptionEN   string `json:"description_en"`
	DescriptionFR   string `json:"description_fr"`
	Save            string `json:"save,omitempty"`
	Healing         string `json:"healing,omitempty"`
	Damage          string `json:"damage,omitempty"`
}

// SpellLists contains spell lists by class and level.
type SpellLists struct {
	Divine map[string][]string `json:"divine"`
	Arcane map[string][]string `json:"arcane"`
}

// SpellData holds all spell data from JSON.
type SpellData struct {
	Spells     []Spell    `json:"spells"`
	SpellLists SpellLists `json:"spell_lists"`
}

// SpellBook manages spell data.
type SpellBook struct {
	data    *SpellData
	dataDir string
}

// NewSpellBook creates a new spell book from the data directory.
func NewSpellBook(dataDir string) (*SpellBook, error) {
	path := filepath.Join(dataDir, "spells.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spells.json: %w", err)
	}

	var spellData SpellData
	if err := json.Unmarshal(data, &spellData); err != nil {
		return nil, fmt.Errorf("parsing spells.json: %w", err)
	}

	return &SpellBook{
		data:    &spellData,
		dataDir: dataDir,
	}, nil
}

// GetSpell returns a spell by ID.
func (sb *SpellBook) GetSpell(id string) (*Spell, error) {
	id = strings.ToLower(id)
	for i := range sb.data.Spells {
		if strings.ToLower(sb.data.Spells[i].ID) == id {
			return &sb.data.Spells[i], nil
		}
	}
	return nil, fmt.Errorf("spell not found: %s", id)
}

// ListAllSpells returns all spells sorted by level then name.
func (sb *SpellBook) ListAllSpells() []*Spell {
	results := make([]*Spell, len(sb.data.Spells))
	for i := range sb.data.Spells {
		results[i] = &sb.data.Spells[i]
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// ListByClass returns spells for a specific class.
func (sb *SpellBook) ListByClass(class string) []*Spell {
	class = strings.ToLower(class)
	var results []*Spell

	// Normalize class name
	var spellType string
	switch class {
	case "cleric", "clerc":
		spellType = "divine"
	case "magic-user", "magicien", "mage", "wizard":
		spellType = "arcane"
	default:
		return results
	}

	for i := range sb.data.Spells {
		s := &sb.data.Spells[i]
		if s.Type == spellType || s.Type == "both" {
			results = append(results, s)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// ListByClassAndLevel returns spells for a class at a specific level.
func (sb *SpellBook) ListByClassAndLevel(class string, level int) []*Spell {
	class = strings.ToLower(class)
	var results []*Spell

	// Get spell list for class and level
	var spellIDs []string
	switch class {
	case "cleric", "clerc":
		spellIDs = sb.data.SpellLists.Divine[fmt.Sprintf("%d", level)]
	case "magic-user", "magicien", "mage", "wizard":
		spellIDs = sb.data.SpellLists.Arcane[fmt.Sprintf("%d", level)]
	default:
		return results
	}

	// Get spells by ID
	for _, id := range spellIDs {
		if spell, err := sb.GetSpell(id); err == nil {
			results = append(results, spell)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// ListByLevel returns all spells of a specific level.
func (sb *SpellBook) ListByLevel(level int) []*Spell {
	var results []*Spell
	for i := range sb.data.Spells {
		if sb.data.Spells[i].Level == level {
			results = append(results, &sb.data.Spells[i])
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// ListReversible returns all reversible spells.
func (sb *SpellBook) ListReversible() []*Spell {
	var results []*Spell
	for i := range sb.data.Spells {
		if sb.data.Spells[i].Reversible {
			results = append(results, &sb.data.Spells[i])
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// SearchSpells searches spells by name (FR or EN) or ID.
func (sb *SpellBook) SearchSpells(query string) []*Spell {
	query = strings.ToLower(query)
	var results []*Spell

	for i := range sb.data.Spells {
		s := &sb.data.Spells[i]
		if strings.Contains(strings.ToLower(s.NameFR), query) ||
			strings.Contains(strings.ToLower(s.NameEN), query) ||
			strings.Contains(strings.ToLower(s.ID), query) {
			results = append(results, s)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].NameFR < results[j].NameFR
	})
	return results
}

// GetSpellCount returns the total number of spells.
func (sb *SpellBook) GetSpellCount() int {
	return len(sb.data.Spells)
}

// GetAvailableLevels returns levels that have spells.
func (sb *SpellBook) GetAvailableLevels() []int {
	levelMap := make(map[int]bool)
	for _, s := range sb.data.Spells {
		levelMap[s.Level] = true
	}
	levels := make([]int, 0, len(levelMap))
	for l := range levelMap {
		levels = append(levels, l)
	}
	sort.Ints(levels)
	return levels
}

// ToMarkdown returns a formatted markdown description of a spell.
func (s *Spell) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", s.NameFR, s.NameEN))

	// Type and level
	typeStr := ""
	switch s.Type {
	case "divine":
		typeStr = "Divin (Clerc)"
	case "arcane":
		typeStr = "Arcanique (Magicien)"
	case "both":
		typeStr = "Divin et Arcanique"
	}
	sb.WriteString(fmt.Sprintf("**Type** : %s | **Niveau** : %d\n\n", typeStr, s.Level))

	// Stats table
	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Portée** | %s |\n", s.Range))
	sb.WriteString(fmt.Sprintf("| **Durée** | %s |\n", s.Duration))
	if s.Save != "" {
		sb.WriteString(fmt.Sprintf("| **Sauvegarde** | %s |\n", s.Save))
	}
	if s.Reversible {
		sb.WriteString(fmt.Sprintf("| **Réversible** | Oui (%s) |\n", s.ReverseNameFR))
	}
	if s.Healing != "" {
		sb.WriteString(fmt.Sprintf("| **Soins** | %s |\n", s.Healing))
	}
	if s.Damage != "" {
		sb.WriteString(fmt.Sprintf("| **Dégâts** | %s |\n", s.Damage))
	}

	sb.WriteString("\n### Description\n\n")
	sb.WriteString(s.DescriptionFR)
	sb.WriteString("\n")

	return sb.String()
}

// ToShortDescription returns a one-line description of a spell.
func (s *Spell) ToShortDescription() string {
	reversible := ""
	if s.Reversible {
		reversible = " *"
	}

	typeStr := ""
	switch s.Type {
	case "divine":
		typeStr = "Div"
	case "arcane":
		typeStr = "Arc"
	case "both":
		typeStr = "D/A"
	}

	return fmt.Sprintf("%s%s [N%d %s] - %s, %s", s.NameFR, reversible, s.Level, typeStr, s.Range, s.Duration)
}

// ToJSON returns the spell as JSON string.
func (s *Spell) ToJSON() (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetClassLabel returns a human-readable class label.
func GetClassLabel(class string) string {
	switch strings.ToLower(class) {
	case "cleric", "clerc":
		return "Clerc"
	case "magic-user", "magicien", "mage", "wizard":
		return "Magicien"
	default:
		return class
	}
}

// GetTypeLabel returns a human-readable type label in French.
func GetTypeLabel(spellType string) string {
	switch spellType {
	case "divine":
		return "Divin"
	case "arcane":
		return "Arcanique"
	case "both":
		return "Divin et Arcanique"
	default:
		return spellType
	}
}
