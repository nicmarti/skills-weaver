package dmtools

import (
	"fmt"
	"strings"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/npc"
	"dungeons/internal/npcmanager"
	"dungeons/internal/treasure"
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
func NewLogEventTool(adv *adventure.Adventure) *SimpleTool {
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

			// Call real persistence
			if err := adv.LogEvent(eventType, content); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			timestamp := time.Now()
			return map[string]interface{}{
				"success": true,
				"display": fmt.Sprintf("✓ Logged [%s at %s]: %s", eventType, timestamp.Format("15:04"), content),
			}, nil
		},
	}
}

// NewAddGoldTool creates a tool to modify party gold.
func NewAddGoldTool(adv *adventure.Adventure) *SimpleTool {
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

			// Load inventory
			inv, err := adv.LoadInventory()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load inventory: %v", err),
				}, nil
			}

			// Modify gold
			inv.Gold += amount

			// Save inventory
			if err := adv.SaveInventory(inv); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save inventory: %v", err),
				}, nil
			}

			// Log event
			logContent := fmt.Sprintf("%+d po", amount)
			if reason != "" {
				logContent += fmt.Sprintf(" (%s)", reason)
			}
			adv.LogEvent("expense", logContent)

			verb := "ajouté"
			displayAmount := amount
			if amount < 0 {
				verb = "retiré"
				displayAmount = -amount
			}

			display := fmt.Sprintf("✓ %s %d po (total: %d po)", verb, displayAmount, inv.Gold)
			if reason != "" {
				display += fmt.Sprintf(" - %s", reason)
			}

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}
}

// NewGetInventoryTool creates a tool to get the party's inventory.
func NewGetInventoryTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_inventory",
		description: "Get the party's current shared inventory (gold and items)",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Load real inventory
			inv, err := adv.LoadInventory()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load inventory: %v", err),
				}, nil
			}

			// Format display
			display := fmt.Sprintf("Or: %d po", inv.Gold)
			if len(inv.Items) > 0 {
				display += "\nObjets:"
				for _, item := range inv.Items {
					if item.Quantity > 1 {
						display += fmt.Sprintf("\n- %s x%d", item.Name, item.Quantity)
					} else {
						display += fmt.Sprintf("\n- %s", item.Name)
					}
				}
			} else {
				display += "\nAucun objet"
			}

			return map[string]interface{}{
				"success": true,
				"gold":    inv.Gold,
				"items":   inv.Items,
				"display": display,
			}, nil
		},
	}
}

// NewGenerateTreasureTool creates a tool to generate treasure.
func NewGenerateTreasureTool(dataDir string) (*SimpleTool, error) {
	// Create treasure generator
	gen, err := treasure.NewGenerator(dataDir)
	if err != nil {
		return nil, fmt.Errorf("creating treasure generator: %w", err)
	}

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
			treasureType := strings.ToUpper(params["treasure_type"].(string))

			// Generate real treasure
			generated, err := gen.GenerateTreasure(treasureType)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			// Format display
			display := fmt.Sprintf("Trésor type %s (valeur totale: %d po)\n", treasureType, generated.TotalValueGP)

			// Coins
			if len(generated.Coins) > 0 {
				display += "\nPièces:"
				for _, coin := range generated.Coins {
					display += fmt.Sprintf("\n- %d %s (%d po)", coin.Amount, coin.NameFR, coin.ValueGP)
				}
			}

			// Gems
			if len(generated.Gems) > 0 {
				display += fmt.Sprintf("\n\nGemmes (%d):", len(generated.Gems))
				for _, gem := range generated.Gems {
					display += fmt.Sprintf("\n- %s (%d po)", gem.Name, gem.Value)
				}
			}

			// Jewelry
			if len(generated.Jewelry) > 0 {
				display += fmt.Sprintf("\n\nBijoux (%d):", len(generated.Jewelry))
				for _, jewel := range generated.Jewelry {
					display += fmt.Sprintf("\n- %s (%d po)", jewel.Name, jewel.Value)
				}
			}

			// Magic items
			if len(generated.MagicItems) > 0 {
				display += fmt.Sprintf("\n\nObjets magiques (%d):", len(generated.MagicItems))
				for _, item := range generated.MagicItems {
					display += fmt.Sprintf("\n- %s", item.Name)
				}
			}

			return map[string]interface{}{
				"success":  true,
				"treasure": generated,
				"display":  display,
			}, nil
		},
	}, nil
}

// NewGenerateNPCTool creates a tool to generate NPCs with automatic persistence.
func NewGenerateNPCTool(dataDir string, adv *adventure.Adventure) (*SimpleTool, error) {
	// Create NPC generator
	gen, err := npc.NewGenerator(dataDir)
	if err != nil {
		return nil, fmt.Errorf("creating NPC generator: %w", err)
	}

	return &SimpleTool{
		name:        "generate_npc",
		description: "Generate a complete NPC with name, appearance, personality, motivation, and secret. Automatically saves to adventure for future reference. Use 'name' parameter to create an NPC with a specific name (useful for officializing existing narrative NPCs).",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "NPC name (optional). If specified, uses this name instead of generating a random one. Useful for officializing NPCs already mentioned in narrative.",
				},
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
				"occupation": map[string]interface{}{
					"type":        "string",
					"description": "NPC occupation. Can be: (1) a category like 'commoner', 'skilled', 'authority', 'underworld', 'religious', 'adventurer' for random selection, OR (2) a specific occupation like 'aubergiste', 'marchand', 'garde de ville', 'voleur', 'prêtre', etc. (optional, defaults to random)",
				},
				"attitude": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"friendly", "neutral", "unfriendly", "hostile"},
					"description": "NPC attitude (optional, defaults to neutral)",
				},
				"context": map[string]interface{}{
					"type":        "string",
					"description": "Context of encounter (location, situation). Example: 'Tavern du Voile Écarlate, asking about rumors'",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Build options from parameters
			var opts []npc.Option
			if name, ok := params["name"].(string); ok && name != "" {
				opts = append(opts, npc.WithName(name))
			}
			if race, ok := params["race"].(string); ok && race != "" {
				opts = append(opts, npc.WithRace(race))
			}
			if gender, ok := params["gender"].(string); ok && gender != "" {
				opts = append(opts, npc.WithGender(gender))
			}
			if occupation, ok := params["occupation"].(string); ok && occupation != "" {
				opts = append(opts, npc.WithOccupationType(occupation))
			}
			if attitude, ok := params["attitude"].(string); ok && attitude != "" {
				opts = append(opts, npc.WithAttitude(attitude))
			}

			// Generate real NPC
			generated, err := gen.Generate(opts...)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			// Extract context (optional)
			context := ""
			if ctx, ok := params["context"].(string); ok {
				context = ctx
			}

			// Use ToShortDescription for compact display
			display := generated.ToShortDescription()

			// Auto-save to adventure's npcs-generated.json
			npcMgr := npcmanager.NewManager(adv.BasePath())
			sessionNum, _ := npcMgr.GetCurrentSessionNumber()

			// Note: World-keeper validation will be added in next step
			worldKeeperNotes := "Pending world-keeper validation"

			record, err := npcMgr.AddNPC(sessionNum, generated, context, worldKeeperNotes)
			if err != nil {
				// Log error but don't fail the generation
				fmt.Printf("Warning: Failed to save NPC to adventure: %v\n", err)
			} else {
				display += fmt.Sprintf(" [ID: %s, saved to session_%d]", record.ID, sessionNum)
			}

			return map[string]interface{}{
				"success": true,
				"npc":     generated,
				"display": display,
			}, nil
		},
	}, nil
}
