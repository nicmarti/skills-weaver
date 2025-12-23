package dmtools

import (
	"fmt"
	"time"
)

// SimpleTool is a basic tool implementation.
type SimpleTool struct {
	name        string
	description string
	schema      map[string]interface{}
	execute     func(map[string]interface{}) (interface{}, error)
}

func (t *SimpleTool) Name() string        { return t.name }
func (t *SimpleTool) Description() string { return t.description }
func (t *SimpleTool) InputSchema() map[string]interface{} { return t.schema }
func (t *SimpleTool) Execute(params map[string]interface{}) (interface{}, error) {
	return t.execute(params)
}

// NewLogEventTool creates a tool to log events to the adventure journal.
func NewLogEventTool(adventurePath string) *SimpleTool {
	return &SimpleTool{
		name:        "log_event",
		description: "Log an important event to the adventure journal (combat, discovery, loot, quest, npc encounter, location, etc.)",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"event_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"combat", "loot", "story", "note", "quest", "npc", "location", "discovery"},
					"description": "Type of event",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Event description in French",
				},
			},
			"required": []string{"event_type", "content"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			eventType := params["event_type"].(string)
			content := params["content"].(string)

			// Simplified: just return success for now
			timestamp := time.Now()

			return map[string]interface{}{
				"success": true,
				"display": fmt.Sprintf("✓ Logged [%s at %s]: %s", eventType, timestamp.Format("15:04"), content),
			}, nil
		},
	}
}

// NewAddGoldTool creates a tool to modify party gold.
func NewAddGoldTool(adventurePath string) *SimpleTool {
	return &SimpleTool{
		name:        "add_gold",
		description: "Add or remove gold from the party's shared inventory (use negative numbers to remove)",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"amount": map[string]interface{}{
					"type":        "number",
					"description": "Amount of gold to add (negative to remove)",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Reason for the gold change",
				},
			},
			"required": []string{"amount"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			amount := int(params["amount"].(float64))
			reason := ""
			if r, ok := params["reason"].(string); ok {
				reason = r
			}

			verb := "ajouté"
			displayAmount := amount
			if amount < 0 {
				verb = "retiré"
				displayAmount = -amount
			}

			display := fmt.Sprintf("✓ %s %d po", verb, displayAmount)
			if reason != "" {
				display += fmt.Sprintf(" (%s)", reason)
			}

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}
}

// NewGetInventoryTool creates a tool to get the party's inventory.
func NewGetInventoryTool(adventurePath string) *SimpleTool {
	return &SimpleTool{
		name:        "get_inventory",
		description: "Get the party's current shared inventory (gold and items)",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Simplified: return mock data for now
			display := "Or: 1293 po\nObjets: Potions de soin x4"

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}
}

// NewGenerateTreasureTool creates a tool to generate treasure.
func NewGenerateTreasureTool(dataDir string) (*SimpleTool, error) {
	return &SimpleTool{
		name:        "generate_treasure",
		description: "Generate a treasure hoard according to BFRPG treasure types (A-U). Each monster type has an associated treasure type.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"treasure_type": map[string]interface{}{
					"type":        "string",
					"description": "Treasure type code (A-U). Common: R (goblin), L (orc), H (ogre), A (dragon)",
				},
			},
			"required": []string{"treasure_type"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			treasureType := params["treasure_type"].(string)

			// Simplified: return mock treasure
			display := fmt.Sprintf("Trésor type %s:\n- Pièces: 100 po\n- 2 gemmes", treasureType)

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}, nil
}

// NewGenerateNPCTool creates a tool to generate NPCs.
func NewGenerateNPCTool(dataDir string) (*SimpleTool, error) {
	return &SimpleTool{
		name:        "generate_npc",
		description: "Generate a complete NPC with name, appearance, personality, motivation, and secret",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"race": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"human", "elf", "dwarf", "halfling"},
					"description": "NPC race (optional, defaults to random)",
				},
				"gender": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"m", "f"},
					"description": "NPC gender (optional, defaults to random)",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Simplified: return mock NPC
			display := "Garrick le Marchand\nRace: Humain | Sexe: Masculin\nApparence: Grand, barbe grise\nPersonnalité: Rusé mais honnête"

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}, nil
}
