package agent

import (
	"fmt"
	"strings"
)

// ToolToCLICommand converts a tool call to its equivalent CLI command.
// Returns empty string if the tool has no CLI equivalent.
func ToolToCLICommand(toolName string, params map[string]interface{}) string {
	switch toolName {
	case "roll_dice":
		return mapRollDice(params)
	case "get_monster":
		return mapGetMonster(params)
	case "log_event":
		// No CLI equivalent - internal adventure operation
		return ""
	case "add_gold":
		return mapAddGold(params)
	case "get_inventory":
		return mapGetInventory(params)
	case "generate_treasure":
		return mapGenerateTreasure(params)
	case "generate_npc":
		return mapGenerateNPC(params)
	case "generate_image":
		return mapGenerateImage(params)
	case "generate_map":
		return mapGenerateMap(params)
	case "update_npc_importance":
		// No CLI equivalent - internal adventure operation
		return ""
	case "get_npc_history":
		// No CLI equivalent - internal adventure operation
		return ""
	default:
		return ""
	}
}

func mapRollDice(params map[string]interface{}) string {
	notation, ok := params["notation"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("./sw-dice roll %s", notation)
}

func mapGetMonster(params map[string]interface{}) string {
	monsterID, ok := params["monster_id"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("./sw-monster show %s", monsterID)
}

func mapAddGold(params map[string]interface{}) string {
	// This would require knowing the adventure name, which we don't have in params
	// Return a generic command
	amount, ok := params["amount"].(float64)
	if !ok {
		return ""
	}
	reason := ""
	if r, ok := params["reason"].(string); ok {
		reason = fmt.Sprintf(" \"%s\"", r)
	}
	return fmt.Sprintf("./sw-adventure add-gold \"<adventure>\" %.0f%s", amount, reason)
}

func mapGetInventory(params map[string]interface{}) string {
	return "./sw-adventure inventory \"<adventure>\""
}

func mapGenerateTreasure(params map[string]interface{}) string {
	treasureType, ok := params["treasure_type"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("./sw-treasure generate %s", strings.ToUpper(treasureType))
}

func mapGenerateNPC(params map[string]interface{}) string {
	var parts []string
	parts = append(parts, "./sw-npc generate")

	if race, ok := params["race"].(string); ok && race != "" {
		parts = append(parts, fmt.Sprintf("--race=%s", race))
	}
	if gender, ok := params["gender"].(string); ok && gender != "" {
		parts = append(parts, fmt.Sprintf("--gender=%s", gender))
	}
	if occupation, ok := params["occupation"].(string); ok && occupation != "" {
		parts = append(parts, fmt.Sprintf("--occupation=%s", occupation))
	}
	if attitude, ok := params["attitude"].(string); ok && attitude != "" {
		parts = append(parts, fmt.Sprintf("--attitude=%s", attitude))
	}

	return strings.Join(parts, " ")
}

func mapGenerateImage(params map[string]interface{}) string {
	imageType, ok := params["type"].(string)
	if !ok {
		return ""
	}

	switch imageType {
	case "character":
		characterName, ok := params["character_name"].(string)
		if !ok {
			return ""
		}
		style := ""
		if s, ok := params["style"].(string); ok {
			style = fmt.Sprintf(" --style=%s", s)
		}
		return fmt.Sprintf("./sw-image character \"%s\"%s", characterName, style)

	case "npc":
		var parts []string
		parts = append(parts, "./sw-image npc")
		if race, ok := params["race"].(string); ok {
			parts = append(parts, fmt.Sprintf("--race=%s", race))
		}
		if gender, ok := params["gender"].(string); ok {
			parts = append(parts, fmt.Sprintf("--gender=%s", gender))
		}
		if occupation, ok := params["occupation"].(string); ok {
			parts = append(parts, fmt.Sprintf("--occupation=%s", occupation))
		}
		return strings.Join(parts, " ")

	case "scene":
		description, ok := params["description"].(string)
		if !ok {
			return ""
		}
		sceneType := ""
		if st, ok := params["scene_type"].(string); ok {
			sceneType = fmt.Sprintf(" --type=%s", st)
		}
		return fmt.Sprintf("./sw-image scene \"%s\"%s", description, sceneType)

	case "monster":
		monsterType, ok := params["monster_type"].(string)
		if !ok {
			return ""
		}
		return fmt.Sprintf("./sw-image monster %s", monsterType)

	case "location":
		locationType, ok := params["location_type"].(string)
		if !ok {
			return ""
		}
		name, ok := params["name"].(string)
		if !ok {
			return ""
		}
		return fmt.Sprintf("./sw-image location %s \"%s\"", locationType, name)

	case "item":
		itemType, ok := params["item_type"].(string)
		if !ok {
			return ""
		}
		description, ok := params["description"].(string)
		if !ok {
			return ""
		}
		return fmt.Sprintf("./sw-image item %s \"%s\"", itemType, description)

	default:
		return ""
	}
}

func mapGenerateMap(params map[string]interface{}) string {
	mapType, ok := params["type"].(string)
	if !ok {
		return ""
	}

	name, ok := params["name"].(string)
	if !ok {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("./sw-map generate %s \"%s\"", mapType, name))

	// Add optional parameters
	if kingdom, ok := params["kingdom"].(string); ok && kingdom != "" {
		parts = append(parts, fmt.Sprintf("--kingdom=%s", kingdom))
	}

	if scale, ok := params["scale"].(string); ok && scale != "" {
		parts = append(parts, fmt.Sprintf("--scale=%s", scale))
	}

	if level, ok := params["level"].(float64); ok {
		parts = append(parts, fmt.Sprintf("--level=%.0f", level))
	}

	if terrain, ok := params["terrain"].(string); ok && terrain != "" {
		parts = append(parts, fmt.Sprintf("--terrain=%s", terrain))
	}

	if scene, ok := params["scene"].(string); ok && scene != "" {
		parts = append(parts, fmt.Sprintf("--scene=\"%s\"", scene))
	}

	// Check if features is present and format it
	if features, ok := params["features"]; ok {
		switch v := features.(type) {
		case []interface{}:
			// Array of features
			var featureStrs []string
			for _, f := range v {
				if fStr, ok := f.(string); ok {
					featureStrs = append(featureStrs, fStr)
				}
			}
			if len(featureStrs) > 0 {
				parts = append(parts, fmt.Sprintf("--features=\"%s\"", strings.Join(featureStrs, ",")))
			}
		case string:
			// Single string
			parts = append(parts, fmt.Sprintf("--features=\"%s\"", v))
		}
	}

	// Check if generate_image flag is present
	if generateImage, ok := params["generate_image"].(bool); ok && generateImage {
		parts = append(parts, "--generate-image")
	}

	return strings.Join(parts, " ")
}
