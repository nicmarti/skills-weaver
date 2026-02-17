package dmtools

import (
	"fmt"
	"path/filepath"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
)

// D&D 5e XP thresholds by level
// Level X requires xpThresholds[X] total XP
var xpThresholds = map[int]int{
	1:  0,
	2:  300,
	3:  900,
	4:  2700,
	5:  6500,
	6:  14000,
	7:  23000,
	8:  34000,
	9:  48000,
	10: 64000,
	11: 85000,
	12: 100000,
	13: 120000,
	14: 140000,
	15: 165000,
	16: 195000,
	17: 225000,
	18: 265000,
	19: 305000,
	20: 355000,
}

// LevelForXP returns the level for a given XP total.
func LevelForXP(xp int) int {
	level := 1
	for lvl := 2; lvl <= 20; lvl++ {
		if xp >= xpThresholds[lvl] {
			level = lvl
		} else {
			break
		}
	}
	return level
}

// XPForLevel returns the XP required to reach a given level.
func XPForLevel(level int) int {
	if level < 1 {
		return 0
	}
	if level > 20 {
		return xpThresholds[20]
	}
	return xpThresholds[level]
}

// XPToNextLevel returns the XP needed to reach the next level from current XP.
// Returns 0 if already at max level (20).
func XPToNextLevel(currentXP int) int {
	currentLevel := LevelForXP(currentXP)
	if currentLevel >= 20 {
		return 0
	}
	nextLevelXP := xpThresholds[currentLevel+1]
	return nextLevelXP - currentXP
}

// NewAddXPTool creates a tool to award XP to party characters.
func NewAddXPTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "add_xp",
		description: "Award experience points (XP) to party characters. Can award to all characters (default) or a specific character. Automatically detects level up and saves characters. Logs to journal with type 'xp'.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "integer",
					"description": "XP amount to award (positive number)",
				},
				"character_name": map[string]interface{}{
					"type":        "string",
					"description": "Optional: specific character name. If omitted, XP is awarded to all party members.",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Optional: reason for XP award (e.g., 'Combat orcs', 'Quest completed'). Used in journal log.",
				},
			},
			"required": []string{"amount"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Extract amount (required)
			amountFloat, ok := params["amount"].(float64)
			if !ok {
				return map[string]interface{}{
					"success": false,
					"error":   "amount is required and must be a number",
				}, nil
			}
			amount := int(amountFloat)
			if amount <= 0 {
				return map[string]interface{}{
					"success": false,
					"error":   "amount must be a positive number",
				}, nil
			}

			// Extract optional parameters
			characterName := ""
			if name, ok := params["character_name"].(string); ok {
				characterName = name
			}

			reason := ""
			if r, ok := params["reason"].(string); ok {
				reason = r
			}

			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load characters: %v", err),
				}, nil
			}

			if len(characters) == 0 {
				return map[string]interface{}{
					"success": false,
					"error":   "No characters in party",
				}, nil
			}

			// Filter characters if specific name provided
			var targetChars []*character.Character
			if characterName != "" {
				nameLower := strings.ToLower(characterName)
				for _, char := range characters {
					if strings.ToLower(char.Name) == nameLower {
						targetChars = append(targetChars, char)
						break
					}
				}
				if len(targetChars) == 0 {
					available := []string{}
					for _, char := range characters {
						available = append(available, char.Name)
					}
					return map[string]interface{}{
						"success":   false,
						"error":     fmt.Sprintf("Character '%s' not found in party", characterName),
						"available": available,
					}, nil
				}
			} else {
				targetChars = characters
			}

			// Award XP and track level ups
			results := []map[string]interface{}{}
			levelUps := []string{}

			for _, char := range targetChars {
				oldLevel := char.Level
				oldXP := char.XP

				// Add XP
				char.XP += amount
				newLevel := LevelForXP(char.XP)

				// Check for level up
				leveledUp := false
				if newLevel > oldLevel {
					char.Level = newLevel
					char.CalculateProficiencyBonus()
					leveledUp = true
					levelUps = append(levelUps, fmt.Sprintf("%s (niveau %d → %d)", char.Name, oldLevel, newLevel))
				}

				// Save character
				charDir := filepath.Join(adv.BasePath(), "characters")
				if err := char.Save(charDir); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Failed to save character %s: %v", char.Name, err),
					}, nil
				}

				// Track result
				xpToNext := XPToNextLevel(char.XP)
				result := map[string]interface{}{
					"name":       char.Name,
					"xp_added":   amount,
					"total_xp":   char.XP,
					"level":      char.Level,
					"xp_to_next": xpToNext,
					"leveled_up": leveledUp,
				}
				if leveledUp {
					result["old_level"] = oldLevel
					result["old_xp"] = oldXP
				}
				results = append(results, result)
			}

			// Log to journal
			logContent := fmt.Sprintf("+%d XP", amount)
			if reason != "" {
				logContent += fmt.Sprintf(" (%s)", reason)
			}
			if characterName != "" {
				logContent = fmt.Sprintf("%s: %s", characterName, logContent)
			} else {
				logContent = fmt.Sprintf("Groupe: %s", logContent)
			}
			if len(levelUps) > 0 {
				logContent += fmt.Sprintf(" - LEVEL UP: %s", strings.Join(levelUps, ", "))
			}
			adv.LogEvent("xp", logContent)

			// Build display string
			display := buildXPDisplay(results, amount, reason, levelUps)

			return map[string]interface{}{
				"success":   true,
				"results":   results,
				"level_ups": levelUps,
				"display":   display,
			}, nil
		},
	}
}

// buildXPDisplay formats the XP award result for display.
func buildXPDisplay(results []map[string]interface{}, _ int, reason string, levelUps []string) string {
	var sb strings.Builder

	// Header
	if reason != "" {
		sb.WriteString(fmt.Sprintf("## XP Awarded (%s)\n\n", reason))
	} else {
		sb.WriteString("## XP Awarded\n\n")
	}

	// Table
	sb.WriteString("| Personnage | XP Ajouté | Total XP | Niveau | Prochain |\n")
	sb.WriteString("|------------|-----------|----------|--------|----------|\n")

	for _, r := range results {
		levelStr := fmt.Sprintf("%d", r["level"].(int))
		if r["leveled_up"].(bool) {
			levelStr += " **LEVEL UP!**"
		}

		xpToNext := r["xp_to_next"].(int)
		nextStr := fmt.Sprintf("%d", xpToNext)
		if xpToNext == 0 {
			nextStr = "MAX"
		}

		sb.WriteString(fmt.Sprintf("| %s | +%d | %d | %s | %s |\n",
			r["name"].(string),
			r["xp_added"].(int),
			r["total_xp"].(int),
			levelStr,
			nextStr,
		))
	}

	// Level up section
	if len(levelUps) > 0 {
		sb.WriteString("\n### Level Up!\n\n")
		for _, lu := range levelUps {
			sb.WriteString(fmt.Sprintf("**%s**\n", lu))
		}
		sb.WriteString("\n_Consultez le rules-keeper pour les bénéfices de niveau (dés de vie, compétences, sorts, capacités de classe)._\n")
	}

	return sb.String()
}
