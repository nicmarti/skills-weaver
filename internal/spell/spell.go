package spell

import (
	"dungeons/internal/data"
	"fmt"
	"sort"
	"strings"
)

// Manager manages spell queries and provides convenience methods for D&D 5e spells.
// It wraps data.GameData to provide spell-specific functionality.
type Manager struct {
	gameData *data.GameData
}

// NewManager creates a new spell manager from GameData.
func NewManager(gd *data.GameData) *Manager {
	return &Manager{gameData: gd}
}

// NewManagerFromDataDir creates a new spell manager by loading GameData from a directory.
func NewManagerFromDataDir(dataDir string) (*Manager, error) {
	gd, err := data.Load(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading game data: %w", err)
	}
	return &Manager{gameData: gd}, nil
}

// GetSpell returns a spell by ID.
func (m *Manager) GetSpell(id string) (*data.Spell5e, error) {
	spell, ok := m.gameData.GetSpell5e(id)
	if !ok {
		return nil, fmt.Errorf("spell not found: %s", id)
	}
	return spell, nil
}

// ListAllSpells returns all spells sorted by level then name.
func (m *Manager) ListAllSpells() []*data.Spell5e {
	results := make([]*data.Spell5e, 0, len(m.gameData.Spells5e))
	for _, spell := range m.gameData.Spells5e {
		results = append(results, spell)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].Name < results[j].Name
	})
	return results
}

// ListByClass returns spells for a specific class, sorted by level then name.
func (m *Manager) ListByClass(classID string) []*data.Spell5e {
	classID = normalizeClassID(classID)
	results := m.gameData.ListSpellsByClass(classID, -1)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].Name < results[j].Name
	})
	return results
}

// ListByClassAndLevel returns spells for a class at a specific level.
func (m *Manager) ListByClassAndLevel(classID string, level int) []*data.Spell5e {
	classID = normalizeClassID(classID)
	results := m.gameData.ListSpellsByClass(classID, level)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// ListByLevel returns all spells of a specific level.
func (m *Manager) ListByLevel(level int) []*data.Spell5e {
	var results []*data.Spell5e
	for _, spell := range m.gameData.Spells5e {
		if spell.Level == level {
			results = append(results, spell)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// ListCantrips returns cantrips (level 0 spells) for a specific class.
func (m *Manager) ListCantrips(classID string) []*data.Spell5e {
	classID = normalizeClassID(classID)
	return m.gameData.ListCantrips(classID)
}

// ListBySchool returns all spells of a specific school.
func (m *Manager) ListBySchool(school string) []*data.Spell5e {
	return m.gameData.ListSpellsBySchool(school)
}

// ListConcentration returns all spells requiring concentration.
func (m *Manager) ListConcentration() []*data.Spell5e {
	var results []*data.Spell5e
	for _, spell := range m.gameData.Spells5e {
		if spell.Concentration {
			results = append(results, spell)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].Name < results[j].Name
	})
	return results
}

// ListRituals returns all ritual spells.
func (m *Manager) ListRituals() []*data.Spell5e {
	var results []*data.Spell5e
	for _, spell := range m.gameData.Spells5e {
		if spell.Ritual {
			results = append(results, spell)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Level != results[j].Level {
			return results[i].Level < results[j].Level
		}
		return results[i].Name < results[j].Name
	})
	return results
}

// SearchSpells searches spells by name (FR or EN) or ID.
func (m *Manager) SearchSpells(query string) []*data.Spell5e {
	return m.gameData.SearchSpells(query)
}

// GetSpellCount returns the total number of spells.
func (m *Manager) GetSpellCount() int {
	return len(m.gameData.Spells5e)
}

// GetAvailableLevels returns levels that have spells (0-9).
func (m *Manager) GetAvailableLevels() []int {
	levelMap := make(map[int]bool)
	for _, s := range m.gameData.Spells5e {
		levelMap[s.Level] = true
	}
	levels := make([]int, 0, len(levelMap))
	for l := range levelMap {
		levels = append(levels, l)
	}
	sort.Ints(levels)
	return levels
}

// GetAvailableSchools returns all magic schools present in the spells.
func (m *Manager) GetAvailableSchools() []string {
	schoolMap := make(map[string]bool)
	for _, s := range m.gameData.Spells5e {
		schoolMap[s.School] = true
	}
	schools := make([]string, 0, len(schoolMap))
	for school := range schoolMap {
		schools = append(schools, school)
	}
	sort.Strings(schools)
	return schools
}

// ToMarkdown returns a formatted markdown description of a spell.
func ToMarkdown(s *data.Spell5e) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s\n\n", s.Name))

	// Level and school
	levelStr := fmt.Sprintf("Niveau %d", s.Level)
	if s.Level == 0 {
		levelStr = "Cantrip"
	}
	sb.WriteString(fmt.Sprintf("**%s** | **École** : %s\n\n", levelStr, GetSchoolLabel(s.School)))

	// Stats table
	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Temps d'incantation** | %s |\n", s.CastingTime))
	sb.WriteString(fmt.Sprintf("| **Portée** | %s |\n", s.Range))

	// Components
	compStr := strings.Join(s.Components, ", ")
	if s.Material != "" {
		compStr += fmt.Sprintf(" (%s)", s.Material)
	}
	sb.WriteString(fmt.Sprintf("| **Composantes** | %s |\n", compStr))

	sb.WriteString(fmt.Sprintf("| **Durée** | %s |\n", s.Duration))

	if s.Concentration {
		sb.WriteString("| **Concentration** | Oui |\n")
	}
	if s.Ritual {
		sb.WriteString("| **Rituel** | Oui |\n")
	}

	// Classes
	classesStr := ""
	for i, class := range s.Classes {
		if i > 0 {
			classesStr += ", "
		}
		classesStr += GetClassLabel(class)
	}
	sb.WriteString(fmt.Sprintf("| **Classes** | %s |\n", classesStr))

	if s.Save != "" {
		sb.WriteString(fmt.Sprintf("| **Sauvegarde** | %s |\n", s.Save))
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

	if s.Upcast != "" {
		sb.WriteString("\n### Aux niveaux supérieurs\n\n")
		sb.WriteString(s.Upcast)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ToShortDescription returns a one-line description of a spell.
func ToShortDescription(s *data.Spell5e) string {
	concentration := ""
	if s.Concentration {
		concentration = " (C)"
	}
	ritual := ""
	if s.Ritual {
		ritual = " (R)"
	}

	levelStr := fmt.Sprintf("N%d", s.Level)
	if s.Level == 0 {
		levelStr = "Cantrip"
	}

	return fmt.Sprintf("%s%s%s [%s %s] - %s, %s",
		s.Name, concentration, ritual, levelStr, GetSchoolLabel(s.School), s.Range, s.Duration)
}

// GetClassLabel returns a human-readable class label in French.
func GetClassLabel(classID string) string {
	// Normalize first
	classID = normalizeClassID(classID)

	labels := map[string]string{
		"barbarian": "Barbare",
		"bard":      "Barde",
		"cleric":    "Clerc",
		"druid":     "Druide",
		"fighter":   "Guerrier",
		"monk":      "Moine",
		"paladin":   "Paladin",
		"ranger":    "Rôdeur",
		"rogue":     "Roublard",
		"sorcerer":  "Ensorceleur",
		"warlock":   "Occultiste",
		"wizard":    "Magicien",
	}

	if label, ok := labels[classID]; ok {
		return label
	}
	return classID
}

// GetSchoolLabel returns a human-readable school label in French.
func GetSchoolLabel(school string) string {
	labels := map[string]string{
		"abjuration":    "Abjuration",
		"conjuration":   "Invocation",
		"divination":    "Divination",
		"enchantment":   "Enchantement",
		"evocation":     "Évocation",
		"illusion":      "Illusion",
		"necromancy":    "Nécromancie",
		"transmutation": "Transmutation",
	}

	if label, ok := labels[school]; ok {
		return label
	}
	return school
}

// normalizeClassID normalizes class names to their canonical IDs.
func normalizeClassID(class string) string {
	class = strings.ToLower(class)

	// French to English mappings
	aliases := map[string]string{
		"clerc":       "cleric",
		"magicien":    "wizard",
		"mage":        "wizard",
		"magic-user":  "wizard",
		"guerrier":    "fighter",
		"roublard":    "rogue",
		"voleur":      "rogue",
		"ensorceleur": "sorcerer",
		"occultiste":  "warlock",
		"rôdeur":      "ranger",
		"rodeur":      "ranger",
		"druide":      "druid",
	}

	if normalized, ok := aliases[class]; ok {
		return normalized
	}

	return class
}
