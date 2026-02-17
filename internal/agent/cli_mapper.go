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
	case "update_location":
		// No CLI equivalent - internal adventure operation (updates game state)
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
	case "get_party_info":
		return "./sw-adventure party \"<adventure>\""
	case "get_character_info":
		return mapGetCharacterInfo(params)
	case "get_equipment":
		return mapGetEquipment(params)
	case "get_spell":
		return mapGetSpell(params)
	case "generate_encounter":
		return mapGenerateEncounter(params)
	case "roll_monster_hp":
		return mapRollMonsterHP(params)
	case "add_item":
		return mapAddItem(params)
	case "remove_item":
		return mapRemoveItem(params)
	case "generate_name":
		return mapGenerateName(params)
	case "generate_location_name":
		return mapGenerateLocationName(params)
	case "invoke_agent":
		// No CLI equivalent - internal agent-to-agent communication
		return ""
	case "invoke_skill":
		return mapInvokeSkill(params)
	case "add_xp":
		return mapAddXP(params)
	case "update_hp":
		return mapUpdateHP(params)
	case "use_spell_slot":
		return mapUseSpellSlot(params)
	case "update_character_stat":
		return mapUpdateCharacterStat(params)
	case "long_rest":
		return mapLongRest(params)
	case "create_character":
		return mapCreateCharacter(params)
	case "update_time":
		return mapUpdateTime(params)
	case "set_flag":
		return mapSetFlag(params)
	case "add_quest":
		return mapAddQuest(params)
	case "complete_quest":
		return mapCompleteQuest(params)
	case "set_variable":
		return mapSetVariable(params)
	case "get_state":
		return mapGetState(params)
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

	if name, ok := params["name"].(string); ok && name != "" {
		parts = append(parts, fmt.Sprintf("--name=\"%s\"", name))
	}
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
	// Handle direct prompt-based generation (new schema)
	if prompt, ok := params["prompt"].(string); ok && prompt != "" {
		style := ""
		if s, ok := params["style"].(string); ok && s != "" {
			style = fmt.Sprintf(" --style=%s", s)
		}
		// Note: The tool uses seedream model by default
		return fmt.Sprintf("./sw-image custom \"%s\"%s --model=seedream", prompt, style)
	}

	// Handle type-based generation (legacy schema)
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

	if style, ok := params["style"].(string); ok && style != "" {
		parts = append(parts, fmt.Sprintf("--style=%s", style))
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

	// Add image_size if specified (only relevant when generate_image is true)
	if imageSize, ok := params["image_size"].(string); ok && imageSize != "" {
		parts = append(parts, fmt.Sprintf("--image-size=%s", imageSize))
	}

	return strings.Join(parts, " ")
}

func mapGetCharacterInfo(params map[string]interface{}) string {
	name, ok := params["name"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("./sw-character show \"%s\"", name)
}

func mapGetEquipment(params map[string]interface{}) string {
	if itemID, ok := params["item_id"].(string); ok && itemID != "" {
		return fmt.Sprintf("./sw-equipment show %s", itemID)
	}
	if search, ok := params["search"].(string); ok && search != "" {
		return fmt.Sprintf("./sw-equipment search \"%s\"", search)
	}
	return "./sw-equipment"
}

func mapGetSpell(params map[string]interface{}) string {
	if spellID, ok := params["spell_id"].(string); ok && spellID != "" {
		return fmt.Sprintf("./sw-spell show %s", spellID)
	}
	if search, ok := params["search"].(string); ok && search != "" {
		return fmt.Sprintf("./sw-spell search \"%s\"", search)
	}
	var parts []string
	parts = append(parts, "./sw-spell list")
	if class, ok := params["class"].(string); ok && class != "" {
		parts = append(parts, fmt.Sprintf("--class=%s", class))
	}
	if level, ok := params["level"].(float64); ok {
		parts = append(parts, fmt.Sprintf("--level=%.0f", level))
	}
	return strings.Join(parts, " ")
}

func mapGenerateEncounter(params map[string]interface{}) string {
	if table, ok := params["table"].(string); ok && table != "" {
		return fmt.Sprintf("./sw-monster encounter %s", table)
	}
	if level, ok := params["level"].(float64); ok {
		return fmt.Sprintf("./sw-monster encounter --level=%.0f", level)
	}
	return "./sw-monster encounter"
}

func mapRollMonsterHP(params map[string]interface{}) string {
	monsterID, ok := params["monster_id"].(string)
	if !ok {
		return ""
	}
	count := 1
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}
	return fmt.Sprintf("./sw-monster roll %s --count=%d", monsterID, count)
}

func mapAddItem(params map[string]interface{}) string {
	name, ok := params["name"].(string)
	if !ok {
		return ""
	}
	quantity := 1
	if q, ok := params["quantity"].(float64); ok {
		quantity = int(q)
	}
	return fmt.Sprintf("./sw-adventure add-item \"<adventure>\" \"%s\" %d", name, quantity)
}

func mapRemoveItem(params map[string]interface{}) string {
	name, ok := params["name"].(string)
	if !ok {
		return ""
	}
	quantity := 1
	if q, ok := params["quantity"].(float64); ok {
		quantity = int(q)
	}
	return fmt.Sprintf("./sw-adventure remove-item \"<adventure>\" \"%s\" %d", name, quantity)
}

func mapGenerateName(params map[string]interface{}) string {
	if npcType, ok := params["npc_type"].(string); ok && npcType != "" {
		return fmt.Sprintf("./sw-names npc %s", npcType)
	}
	if race, ok := params["race"].(string); ok && race != "" {
		var parts []string
		parts = append(parts, fmt.Sprintf("./sw-names generate %s", race))
		if gender, ok := params["gender"].(string); ok && gender != "" {
			parts = append(parts, fmt.Sprintf("--gender=%s", gender))
		}
		if count, ok := params["count"].(float64); ok && count > 1 {
			parts = append(parts, fmt.Sprintf("--count=%.0f", count))
		}
		return strings.Join(parts, " ")
	}
	return "./sw-names"
}

func mapGenerateLocationName(params map[string]interface{}) string {
	locationType, ok1 := params["type"].(string)
	kingdom, ok2 := params["kingdom"].(string)
	if !ok1 || !ok2 {
		return ""
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("./sw-location-names %s --kingdom=%s", locationType, kingdom))
	if count, ok := params["count"].(float64); ok && count > 1 {
		parts = append(parts, fmt.Sprintf("--count=%.0f", count))
	}
	return strings.Join(parts, " ")
}

func mapInvokeSkill(params map[string]interface{}) string {
	// invoke_skill already contains the exact CLI command
	command, ok := params["command"].(string)
	if !ok {
		return ""
	}
	return command
}

func mapAddXP(params map[string]interface{}) string {
	amount, ok := params["amount"].(float64)
	if !ok {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("./sw-adventure add-xp \"<adventure>\" %.0f", amount))

	if name, ok := params["character_name"].(string); ok && name != "" {
		parts = append(parts, fmt.Sprintf("--character=\"%s\"", name))
	}
	if reason, ok := params["reason"].(string); ok && reason != "" {
		parts = append(parts, fmt.Sprintf("--reason=\"%s\"", reason))
	}

	return strings.Join(parts, " ")
}

func mapUpdateHP(params map[string]interface{}) string {
	name, ok := params["character_name"].(string)
	if !ok {
		return ""
	}
	amount, ok := params["amount"].(float64)
	if !ok {
		return ""
	}

	// No direct CLI equivalent - this is an internal adventure operation
	// that modifies character JSON files directly
	// Return a descriptive pseudo-command for logging purposes
	reason := ""
	if r, ok := params["reason"].(string); ok && r != "" {
		reason = fmt.Sprintf(" --reason=\"%s\"", r)
	}
	return fmt.Sprintf("# update_hp \"%s\" %.0f%s (internal operation - modifies character JSON)", name, amount, reason)
}

func mapUseSpellSlot(params map[string]interface{}) string {
	name, ok := params["character_name"].(string)
	if !ok {
		return ""
	}
	level, ok := params["spell_level"].(float64)
	if !ok {
		return ""
	}

	// No direct CLI equivalent - this is an internal adventure operation
	// Return a descriptive pseudo-command for logging purposes
	spellName := ""
	if s, ok := params["spell_name"].(string); ok && s != "" {
		spellName = fmt.Sprintf(" --spell=\"%s\"", s)
	}
	return fmt.Sprintf("# use_spell_slot \"%s\" level=%.0f%s (internal operation - modifies character JSON)", name, level, spellName)
}

func mapUpdateTime(params map[string]interface{}) string {
	// No direct CLI equivalent - this is an internal adventure operation
	day := ""
	if d, ok := params["day"].(float64); ok {
		day = fmt.Sprintf("day=%.0f ", d)
	}
	hour := ""
	if h, ok := params["hour"].(float64); ok {
		hour = fmt.Sprintf("hour=%.0f ", h)
	}
	minute := ""
	if m, ok := params["minute"].(float64); ok {
		minute = fmt.Sprintf("minute=%.0f", m)
	}
	return fmt.Sprintf("# update_time %s%s%s(internal operation - modifies state.json)", day, hour, minute)
}

func mapSetFlag(params map[string]interface{}) string {
	flag, ok := params["flag"].(string)
	if !ok {
		return ""
	}
	value := "true"
	if v, ok := params["value"].(bool); ok && !v {
		value = "false"
	}
	return fmt.Sprintf("# set_flag \"%s\" value=%s (internal operation - modifies state.json)", flag, value)
}

func mapAddQuest(params map[string]interface{}) string {
	name, ok := params["name"].(string)
	if !ok {
		return ""
	}
	desc := ""
	if d, ok := params["description"].(string); ok && d != "" {
		desc = fmt.Sprintf(" --description=\"%s\"", d)
	}
	return fmt.Sprintf("# add_quest \"%s\"%s (internal operation - modifies state.json)", name, desc)
}

func mapCompleteQuest(params map[string]interface{}) string {
	name, ok := params["quest_name"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("# complete_quest \"%s\" (internal operation - modifies state.json)", name)
}

func mapSetVariable(params map[string]interface{}) string {
	key, ok := params["key"].(string)
	if !ok {
		return ""
	}
	value := ""
	if v, ok := params["value"].(string); ok {
		value = fmt.Sprintf(" value=\"%s\"", v)
	}
	return fmt.Sprintf("# set_variable \"%s\"%s (internal operation - modifies state.json)", key, value)
}

func mapGetState(params map[string]interface{}) string {
	// No parameters for get_state
	return "# get_state (internal operation - reads state.json)"
}

func mapCreateCharacter(params map[string]interface{}) string {
	name, _ := params["name"].(string)
	species, _ := params["species"].(string)
	class, _ := params["class"].(string)
	return fmt.Sprintf("# create_character \"%s\" species=%s class=%s (internal operation - creates character JSON + updates party)", name, species, class)
}

func mapUpdateCharacterStat(params map[string]interface{}) string {
	name, ok := params["character_name"].(string)
	if !ok {
		return ""
	}
	stat, ok := params["stat"].(string)
	if !ok {
		return ""
	}
	value, ok := params["value"].(float64)
	if !ok {
		return ""
	}
	reason := ""
	if r, ok := params["reason"].(string); ok && r != "" {
		reason = fmt.Sprintf(" --reason=\"%s\"", r)
	}
	return fmt.Sprintf("# update_character_stat \"%s\" %s=%.0f%s (internal operation - modifies character JSON)", name, stat, value, reason)
}

func mapLongRest(params map[string]interface{}) string {
	if name, ok := params["character_name"].(string); ok && name != "" {
		return fmt.Sprintf("# long_rest \"%s\" (internal operation - restores HP, spell slots, hit dice)", name)
	}
	return "# long_rest (internal operation - restores all characters)"
}
