package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/data"
	"dungeons/internal/spell"
)

// NewGetSpellTool creates a tool to look up spell information for D&D 5e.
func NewGetSpellTool(manager *spell.Manager) *SimpleTool {
	return &SimpleTool{
		name:        "get_spell",
		description: "Look up D&D 5e spell details (level, school, components, duration, concentration, ritual). Use when players cast spells or when planning magical encounters.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"spell_id": map[string]interface{}{
					"type":        "string",
					"description": "The spell ID or name to look up (e.g., 'projectile_magique', 'soin_des_blessures', 'sommeil'). Case-insensitive.",
				},
				"search": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Search term to find spells by partial name match.",
				},
				"class": map[string]interface{}{
					"type": "string",
					"description": "Optional: Filter by class. Valid classes: wizard, sorcerer, cleric, druid, bard, warlock, paladin, ranger, fighter, rogue.",
					"enum": []string{
						"wizard", "magicien",
						"sorcerer", "ensorceleur",
						"cleric", "clerc",
						"druid", "druide",
						"bard", "barde",
						"warlock", "occultiste",
						"paladin",
						"ranger", "rôdeur",
						"fighter", "guerrier",
						"rogue", "roublard",
					},
				},
				"level": map[string]interface{}{
					"type":        "integer",
					"description": "Optional: Filter by spell level (0-9, where 0 = cantrips).",
					"minimum":     0,
					"maximum":     9,
				},
				"school": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Filter by magic school.",
					"enum":        []string{"abjuration", "conjuration", "divination", "enchantment", "evocation", "illusion", "necromancy", "transmutation"},
				},
				"concentration": map[string]interface{}{
					"type":        "boolean",
					"description": "Optional: Filter concentration spells only.",
				},
				"ritual": map[string]interface{}{
					"type":        "boolean",
					"description": "Optional: Filter ritual spells only.",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			spellID, hasID := params["spell_id"].(string)
			searchTerm, hasSearch := params["search"].(string)
			class, hasClass := params["class"].(string)
			level, hasLevel := params["level"].(float64)
			school, hasSchool := params["school"].(string)
			concentration, hasConcentration := params["concentration"].(bool)
			ritual, hasRitual := params["ritual"].(bool)

			// If spell_id is provided, look up specific spell
			if hasID && spellID != "" {
				s, err := manager.GetSpell(spellID)
				if err != nil {
					// Try to find by search
					results := manager.SearchSpells(spellID)
					if len(results) > 0 {
						suggestions := make([]string, 0, len(results))
						for _, r := range results {
							suggestions = append(suggestions, spell.ToShortDescription(r))
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
					"spell":   formatSpell5e(s),
					"display": spell.ToMarkdown(s),
				}, nil
			}

			// If search term is provided, search for spells
			if hasSearch && searchTerm != "" {
				results := manager.SearchSpells(searchTerm)
				if len(results) == 0 {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("No spells found matching '%s'", searchTerm),
					}, nil
				}

				spells := make([]map[string]interface{}, 0, len(results))
				for _, s := range results {
					spells = append(spells, formatSpell5e(s))
				}

				return map[string]interface{}{
					"success": true,
					"count":   len(spells),
					"spells":  spells,
					"display": formatSpellList5e(results),
				}, nil
			}

			// If filters are provided, list spells
			if hasSchool {
				results := manager.ListBySchool(school)
				return formatResults(results)
			}

			if hasConcentration && concentration {
				results := manager.ListConcentration()
				return formatResults(results)
			}

			if hasRitual && ritual {
				results := manager.ListRituals()
				return formatResults(results)
			}

			if hasClass && hasLevel {
				results := manager.ListByClassAndLevel(class, int(level))
				return formatResults(results)
			}

			if hasClass {
				results := manager.ListByClass(class)
				return formatResults(results)
			}

			if hasLevel {
				results := manager.ListByLevel(int(level))
				return formatResults(results)
			}

			return map[string]interface{}{
				"success": false,
				"error":   "Provide 'spell_id', 'search', or filter by 'class', 'level', 'school', 'concentration', or 'ritual'",
			}, nil
		},
	}
}

// formatResults formats a list of spells into the response.
func formatResults(results []*data.Spell5e) (interface{}, error) {
	if len(results) == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "No spells found with the given criteria",
		}, nil
	}

	spells := make([]map[string]interface{}, 0, len(results))
	for _, s := range results {
		spells = append(spells, formatSpell5e(s))
	}

	return map[string]interface{}{
		"success": true,
		"count":   len(spells),
		"spells":  spells,
		"display": formatSpellList5e(results),
	}, nil
}

// formatSpell5e converts a D&D 5e spell to a map.
func formatSpell5e(s *data.Spell5e) map[string]interface{}{
	result := map[string]interface{}{
		"id":           s.ID,
		"name":         s.Name,
		"level":        s.Level,
		"school":       s.School,
		"casting_time": s.CastingTime,
		"range":        s.Range,
		"components":   s.Components,
		"duration":     s.Duration,
		"concentration": s.Concentration,
		"ritual":       s.Ritual,
		"classes":      s.Classes,
		"description":  s.DescriptionFR,
	}

	if s.Material != "" {
		result["material"] = s.Material
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
	if s.Upcast != "" {
		result["upcast"] = s.Upcast
	}

	return result
}

// formatSpellList5e formats a list of D&D 5e spells for display.
func formatSpellList5e(spells []*data.Spell5e) string {
	var sb strings.Builder
	sb.WriteString("## Liste des sorts\n\n")

	currentLevel := -1
	for _, s := range spells {
		if s.Level != currentLevel {
			currentLevel = s.Level
			levelName := fmt.Sprintf("Niveau %d", currentLevel)
			if currentLevel == 0 {
				levelName = "Cantrips"
			}
			sb.WriteString(fmt.Sprintf("\n### %s\n\n", levelName))
		}
		sb.WriteString("- ")
		sb.WriteString(spell.ToShortDescription(s))
		sb.WriteString("\n")
	}

	return sb.String()
}
