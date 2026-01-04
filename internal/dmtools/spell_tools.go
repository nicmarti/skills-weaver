package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/spell"
)

// NewGetSpellTool creates a tool to look up spell information.
func NewGetSpellTool(spellBook *spell.SpellBook) *SimpleTool {
	return &SimpleTool{
		name:        "get_spell",
		description: "Look up spell details (range, duration, effects, reversible forms). Use this when a player casts a spell to verify effects, or when planning magical encounters.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"spell_id": map[string]interface{}{
					"type":        "string",
					"description": "The spell ID or name to look up (e.g., 'magic_missile', 'cure_light_wounds', 'sleep'). Case-insensitive.",
				},
				"search": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Search term to find spells by partial name match.",
				},
				"class": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Filter by class ('cleric' or 'magic-user').",
					"enum":        []string{"cleric", "magic-user"},
				},
				"level": map[string]interface{}{
					"type":        "integer",
					"description": "Optional: Filter by spell level (1-6).",
					"minimum":     1,
					"maximum":     6,
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			spellID, hasID := params["spell_id"].(string)
			searchTerm, hasSearch := params["search"].(string)
			class, hasClass := params["class"].(string)
			level, hasLevel := params["level"].(float64)

			// If spell_id is provided, look up specific spell
			if hasID && spellID != "" {
				s, err := spellBook.GetSpell(spellID)
				if err != nil {
					// Try to find by search
					results := spellBook.SearchSpells(spellID)
					if len(results) > 0 {
						suggestions := make([]string, 0, len(results))
						for _, r := range results {
							suggestions = append(suggestions, r.ToShortDescription())
						}
						return map[string]interface{}{
							"success":     false,
							"error":       fmt.Sprintf("Spell '%s' not found", spellID),
							"suggestions": suggestions,
						}, nil
					}
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Spell '%s' not found", spellID),
					}, nil
				}

				return map[string]interface{}{
					"success": true,
					"spell":   formatSpell(s),
					"display": s.ToMarkdown(),
				}, nil
			}

			// If search term is provided, search for spells
			if hasSearch && searchTerm != "" {
				results := spellBook.SearchSpells(searchTerm)
				if len(results) == 0 {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("No spells found matching '%s'", searchTerm),
					}, nil
				}

				spells := make([]map[string]interface{}, 0, len(results))
				for _, s := range results {
					spells = append(spells, formatSpell(s))
				}

				return map[string]interface{}{
					"success": true,
					"count":   len(spells),
					"spells":  spells,
					"display": formatSpellList(results),
				}, nil
			}

			// If class and/or level is provided, list spells
			if hasClass || hasLevel {
				var results []*spell.Spell

				if hasClass && hasLevel {
					results = spellBook.ListByClassAndLevel(class, int(level))
				} else if hasClass {
					results = spellBook.ListByClass(class)
				} else if hasLevel {
					results = spellBook.ListByLevel(int(level))
				}

				if len(results) == 0 {
					return map[string]interface{}{
						"success": false,
						"error":   "No spells found with the given criteria",
					}, nil
				}

				spells := make([]map[string]interface{}, 0, len(results))
				for _, s := range results {
					spells = append(spells, formatSpell(s))
				}

				return map[string]interface{}{
					"success": true,
					"count":   len(spells),
					"spells":  spells,
					"display": formatSpellList(results),
				}, nil
			}

			return map[string]interface{}{
				"success": false,
				"error":   "Provide 'spell_id', 'search', or filter by 'class' and/or 'level'",
			}, nil
		},
	}
}

// formatSpell converts a spell to a map.
func formatSpell(s *spell.Spell) map[string]interface{} {
	result := map[string]interface{}{
		"id":          s.ID,
		"name_fr":     s.NameFR,
		"name_en":     s.NameEN,
		"level":       s.Level,
		"type":        s.Type,
		"range":       s.Range,
		"duration":    s.Duration,
		"description": s.DescriptionFR,
	}

	if s.Save != "" {
		result["save"] = s.Save
	}
	if s.Healing != "" {
		result["healing"] = s.Healing
	}
	if s.Damage != "" {
		result["damage"] = s.Damage
	}
	if s.Reversible {
		result["reversible"] = true
		result["reverse_name_fr"] = s.ReverseNameFR
		result["reverse_name_en"] = s.ReverseNameEN
	}

	return result
}

// formatSpellList formats a list of spells for display.
func formatSpellList(spells []*spell.Spell) string {
	var sb strings.Builder
	sb.WriteString("## Liste des sorts\n\n")

	currentLevel := -1
	for _, s := range spells {
		if s.Level != currentLevel {
			currentLevel = s.Level
			sb.WriteString(fmt.Sprintf("\n### Niveau %d\n\n", currentLevel))
		}
		sb.WriteString("- ")
		sb.WriteString(s.ToShortDescription())
		sb.WriteString("\n")
	}

	return sb.String()
}
