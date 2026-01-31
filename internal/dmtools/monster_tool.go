package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/monster"
)

// MonsterTool provides monster-related functionality.
type MonsterTool struct {
	bestiary *monster.Bestiary
}

// NewMonsterTool creates a new monster tool.
func NewMonsterTool(dataDir string) (*MonsterTool, error) {
	bestiary, err := monster.NewBestiary(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load bestiary: %w", err)
	}

	return &MonsterTool{
		bestiary: bestiary,
	}, nil
}

// Name returns the tool name.
func (t *MonsterTool) Name() string {
	return "get_monster"
}

// Description returns the tool description.
func (t *MonsterTool) Description() string {
	return "Get complete stats for a specific monster by ID (e.g., 'goblin', 'orc', 'dragon_red_adult'). Returns AC, HD, attacks, saves, morale, treasure type, and special abilities."
}

// InputSchema returns the JSON schema for tool input.
func (t *MonsterTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"monster_id": map[string]interface{}{
				"type":        "string",
				"description": "Monster ID in snake_case (e.g., 'goblin', 'orc', 'skeleton', 'dragon_red_adult')",
			},
		},
		"required": []string{"monster_id"},
	}
}

// Execute executes the tool with the given parameters.
func (t *MonsterTool) Execute(params map[string]interface{}) (interface{}, error) {
	monsterID, ok := params["monster_id"].(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"error":   "monster_id parameter is required",
		}, nil
	}

	// Get monster using public method
	found, err := t.bestiary.GetMonster(monsterID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Monster not found: %s", monsterID),
		}, nil
	}

	// Format attacks
	attacksInfo := []string{}
	for _, atk := range found.Attacks {
		atkStr := fmt.Sprintf("%s +%d (%s dmg)", atk.NameFR, atk.Bonus, atk.Damage)
		if atk.DamageType != "" {
			atkStr += fmt.Sprintf(" [%s]", atk.DamageType)
		}
		attacksInfo = append(attacksInfo, atkStr)
	}

	// D&D 5e format display
	display := fmt.Sprintf(`%s (%s)
CA: %d | CR: %s | PV: %d (moy.) | Mvt: %d'
Bonus maîtrise: +%d | XP: %d
Attaques: %s
Type trésor: %s`,
		found.NameFR,
		found.Type,
		found.ArmorClass,
		found.ChallengeRating,
		found.HitPointsAvg,
		found.Movement,
		found.ProficiencyBonus,
		found.XP,
		strings.Join(attacksInfo, ", "),
		found.TreasureType,
	)

	if len(found.Special) > 0 {
		display += fmt.Sprintf("\nSpécial: %s", strings.Join(found.Special, ", "))
	}

	return map[string]interface{}{
		"success": true,
		"monster": found,
		"display": display,
	}, nil
}
