package dmtools

import (
	"fmt"

	"dungeons/internal/adventure"
)

// NewAddItemTool creates a tool to add items to the party inventory.
func NewAddItemTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "add_item",
		description: "Add an item to the party's shared inventory. Use this when the party finds loot, buys equipment, or receives gifts. The item is automatically logged to the journal.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "The item name (e.g., 'Potion de soin', 'Épée +1', 'Corde 15m').",
				},
				"quantity": map[string]interface{}{
					"type":        "integer",
					"description": "Number of items to add (default: 1).",
					"minimum":     1,
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional description of the item (e.g., 'Guérit 1d8+1 PV').",
				},
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Where the item came from (e.g., 'Trésor gobelin', 'Achat à Cordova').",
				},
			},
			"required": []string{"name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "Item name is required",
				}, nil
			}

			quantity := 1
			if q, ok := params["quantity"].(float64); ok && q > 0 {
				quantity = int(q)
			}

			description := ""
			if d, ok := params["description"].(string); ok {
				description = d
			}

			source := "Aventure"
			if s, ok := params["source"].(string); ok && s != "" {
				source = s
			}

			// Add item to inventory
			err := adv.AddItem(name, quantity, description, source)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to add item: %v", err),
				}, nil
			}

			// Get updated inventory for confirmation
			inv, _ := adv.LoadInventory()

			return map[string]interface{}{
				"success":  true,
				"added":    name,
				"quantity": quantity,
				"message":  fmt.Sprintf("Ajouté %d× %s à l'inventaire", quantity, name),
				"inventory_gold": inv.Gold,
				"inventory_items": len(inv.Items),
			}, nil
		},
	}
}

// NewRemoveItemTool creates a tool to remove items from the party inventory.
func NewRemoveItemTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "remove_item",
		description: "Remove an item from the party's shared inventory. Use this when items are consumed, sold, lost, or given away.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "The exact item name to remove (case-sensitive).",
				},
				"quantity": map[string]interface{}{
					"type":        "integer",
					"description": "Number of items to remove (default: 1).",
					"minimum":     1,
				},
			},
			"required": []string{"name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "Item name is required",
				}, nil
			}

			quantity := 1
			if q, ok := params["quantity"].(float64); ok && q > 0 {
				quantity = int(q)
			}

			// Remove item from inventory
			err := adv.RemoveItem(name, quantity)
			if err != nil {
				// Get inventory to show available items
				inv, _ := adv.LoadInventory()
				available := make([]string, 0, len(inv.Items))
				for _, item := range inv.Items {
					available = append(available, fmt.Sprintf("%s (×%d)", item.Name, item.Quantity))
				}

				return map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("Failed to remove item: %v", err),
					"available": available,
				}, nil
			}

			// Get updated inventory for confirmation
			inv, _ := adv.LoadInventory()

			return map[string]interface{}{
				"success":  true,
				"removed":  name,
				"quantity": quantity,
				"message":  fmt.Sprintf("Retiré %d× %s de l'inventaire", quantity, name),
				"inventory_gold": inv.Gold,
				"inventory_items": len(inv.Items),
			}, nil
		},
	}
}
