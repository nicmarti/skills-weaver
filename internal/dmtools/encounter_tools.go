package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/monster"
)

// NewGenerateEncounterTool creates a tool to generate random encounters.
func NewGenerateEncounterTool(bestiary *monster.Bestiary) *SimpleTool {
	return &SimpleTool{
		name:        "generate_encounter",
		description: "Generate a random encounter with monsters. Use this to create balanced combat encounters for the party. Returns monsters with rolled HP, ready for combat.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"table": map[string]interface{}{
					"type":        "string",
					"description": "Encounter table name (e.g., 'dungeon_level_1', 'dungeon_level_2', 'forest', 'undead_crypt'). Use 'list' to see available tables.",
				},
				"level": map[string]interface{}{
					"type":        "integer",
					"description": "Alternative: Party level (1-10). System will select appropriate encounter table.",
					"minimum":     1,
					"maximum":     10,
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			tableName, hasTable := params["table"].(string)
			partyLevel, hasLevel := params["level"].(float64)

			// Special case: list available tables
			if hasTable && tableName == "list" {
				tables := bestiary.GetEncounterTables()
				return map[string]interface{}{
					"success": true,
					"tables":  tables,
					"display": formatTableList(tables),
				}, nil
			}

			var result *monster.EncounterResult
			var err error

			if hasTable && tableName != "" {
				result, err = bestiary.GenerateEncounter(tableName)
			} else if hasLevel {
				result, err = bestiary.GenerateEncounterByLevel(int(partyLevel))
			} else {
				return map[string]interface{}{
					"success": false,
					"error":   "Provide 'table' name or 'level' for party",
					"tables":  bestiary.GetEncounterTables(),
				}, nil
			}

			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"tables":  bestiary.GetEncounterTables(),
				}, nil
			}

			// Format encounter for response
			monsters := make([]map[string]interface{}, 0, len(result.Monsters))
			for i, inst := range result.Monsters {
				monsters = append(monsters, map[string]interface{}{
					"index":        i + 1,
					"name":         inst.Monster.NameFR,
					"id":           inst.Monster.ID,
					"hp":           inst.HitPoints,
					"max_hp":       inst.MaxHP,
					"ac":           inst.Monster.ArmorClass,
					"attacks":      formatAttacks(inst.Monster.Attacks),
					"xp":           inst.Monster.XP,
					"morale":       inst.Monster.Morale,
					"treasure_type": inst.Monster.TreasureType,
				})
			}

			return map[string]interface{}{
				"success":     true,
				"table":       result.TableName,
				"description": result.Description,
				"monsters":    monsters,
				"total_xp":    result.TotalXP,
				"display":     result.ToMarkdown(),
			}, nil
		},
	}
}

// NewRollMonsterHPTool creates a tool to create monster instances with rolled HP.
func NewRollMonsterHPTool(bestiary *monster.Bestiary) *SimpleTool {
	return &SimpleTool{
		name:        "roll_monster_hp",
		description: "Create monster instances with individually rolled HP. Use this when you need specific monsters for combat, not a random encounter.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"monster_id": map[string]interface{}{
					"type":        "string",
					"description": "The monster ID (e.g., 'goblin', 'orc', 'skeleton').",
				},
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of monsters to create (default: 1).",
					"minimum":     1,
					"maximum":     20,
				},
			},
			"required": []string{"monster_id"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			monsterID, ok := params["monster_id"].(string)
			if !ok || monsterID == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "monster_id is required",
				}, nil
			}

			count := 1
			if c, ok := params["count"].(float64); ok && c > 0 {
				count = int(c)
				if count > 20 {
					count = 20
				}
			}

			// Get the monster template
			m, err := bestiary.GetMonster(monsterID)
			if err != nil {
				// Try to find suggestions
				suggestions := bestiary.SearchMonsters(monsterID)
				if len(suggestions) > 0 {
					suggestionNames := make([]string, 0, len(suggestions))
					for _, s := range suggestions {
						suggestionNames = append(suggestionNames, s.ToShortDescription())
					}
					return map[string]interface{}{
						"success":     false,
						"error":       fmt.Sprintf("Monster '%s' not found", monsterID),
						"suggestions": suggestionNames,
					}, nil
				}
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Monster '%s' not found", monsterID),
				}, nil
			}

			// Create instances
			instances := make([]map[string]interface{}, 0, count)
			totalXP := 0
			hpList := make([]int, 0, count)

			for i := 0; i < count; i++ {
				inst := bestiary.CreateInstance(m)
				hpList = append(hpList, inst.HitPoints)
				totalXP += m.XP

				instances = append(instances, map[string]interface{}{
					"index":  i + 1,
					"hp":     inst.HitPoints,
					"max_hp": inst.MaxHP,
				})
			}

			return map[string]interface{}{
				"success": true,
				"monster": map[string]interface{}{
					"id":            m.ID,
					"name":          m.NameFR,
					"ac":            m.ArmorClass,
					"hit_dice":      m.HitDice,
					"attacks":       formatAttacks(m.Attacks),
					"special":       m.Special,
					"morale":        m.Morale,
					"treasure_type": m.TreasureType,
					"xp":            m.XP,
				},
				"instances": instances,
				"total_xp":  totalXP,
				"display":   formatMonsterInstances(m, hpList),
			}, nil
		},
	}
}

// formatAttacks formats monster attacks for the response.
func formatAttacks(attacks []monster.Attack) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(attacks))
	for _, atk := range attacks {
		a := map[string]interface{}{
			"name":   atk.NameFR,
			"bonus":  atk.Bonus,
			"damage": atk.Damage,
		}
		if atk.Special != "" {
			a["special"] = atk.Special
		}
		result = append(result, a)
	}
	return result
}

// formatTableList formats encounter tables for display.
func formatTableList(tables []string) string {
	var sb strings.Builder
	sb.WriteString("## Tables de rencontres disponibles\n\n")
	for _, t := range tables {
		sb.WriteString(fmt.Sprintf("- `%s`\n", t))
	}
	return sb.String()
}

// formatMonsterInstances formats monster instances for display.
func formatMonsterInstances(m *monster.Monster, hpList []int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s × %d\n\n", m.NameFR, len(hpList)))
	sb.WriteString(fmt.Sprintf("**CA** %d | **DV** %s | **Moral** %d | **Trésor** %s\n\n", m.ArmorClass, m.HitDice, m.Morale, m.TreasureType))

	// Attacks
	sb.WriteString("**Attaques** :\n")
	for _, atk := range m.Attacks {
		sb.WriteString(fmt.Sprintf("- %s : +%d, %s", atk.NameFR, atk.Bonus, atk.Damage))
		if atk.Special != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", atk.Special))
		}
		sb.WriteString("\n")
	}

	// HP list
	sb.WriteString("\n**Points de Vie** :\n")
	for i, hp := range hpList {
		sb.WriteString(fmt.Sprintf("- #%d : %d PV\n", i+1, hp))
	}

	sb.WriteString(fmt.Sprintf("\n**XP Total** : %d\n", m.XP*len(hpList)))

	return sb.String()
}
