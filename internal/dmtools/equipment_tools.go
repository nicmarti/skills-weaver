package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/equipment"
)

// NewGetEquipmentTool creates a tool to look up equipment stats.
func NewGetEquipmentTool(catalog *equipment.Catalog) *SimpleTool {
	return &SimpleTool{
		name:        "get_equipment",
		description: "Look up equipment stats (weapons, armor, gear). Use this to check damage dice, AC bonus, cost, weight, and properties of any item. Essential during combat or shopping scenes.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"item_id": map[string]interface{}{
					"type":        "string",
					"description": "The item ID or name to look up (e.g., 'longsword', 'chainmail', 'rope'). Case-insensitive, supports French or English names.",
				},
				"search": map[string]interface{}{
					"type":        "string",
					"description": "Optional: Search term to find items by partial name match. Use when you don't know the exact item ID.",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			itemID, hasID := params["item_id"].(string)
			searchTerm, hasSearch := params["search"].(string)

			// If search term is provided, search for items
			if hasSearch && searchTerm != "" {
				results := catalog.SearchItems(searchTerm)
				if len(results) == 0 {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("No items found matching '%s'", searchTerm),
					}, nil
				}

				// Format results
				items := make([]map[string]interface{}, 0, len(results))
				for _, item := range results {
					items = append(items, formatEquipmentItem(item))
				}

				return map[string]interface{}{
					"success": true,
					"count":   len(items),
					"items":   items,
					"display": formatSearchResults(results),
				}, nil
			}

			// If item_id is provided, look up specific item
			if hasID && itemID != "" {
				item, itemType, err := catalog.GetItem(itemID)
				if err != nil {
					// Try to provide suggestions
					suggestions := catalog.SearchItems(itemID)
					if len(suggestions) > 0 {
						suggestionNames := make([]string, 0, len(suggestions))
						for _, s := range suggestions {
							suggestionNames = append(suggestionNames, equipment.ItemToShortDescription(s))
						}
						return map[string]interface{}{
							"success":     false,
							"error":       fmt.Sprintf("Item '%s' not found", itemID),
							"suggestions": suggestionNames,
						}, nil
					}
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Item '%s' not found", itemID),
					}, nil
				}

				return map[string]interface{}{
					"success":   true,
					"item_type": itemType,
					"item":      formatEquipmentItem(item),
					"display":   equipment.ItemToMarkdown(item),
				}, nil
			}

			return map[string]interface{}{
				"success": false,
				"error":   "Either 'item_id' or 'search' parameter is required",
			}, nil
		},
	}
}

// formatEquipmentItem converts any equipment item to a map.
func formatEquipmentItem(item interface{}) map[string]interface{} {
	switch v := item.(type) {
	case *equipment.Weapon:
		result := map[string]interface{}{
			"type":   "weapon",
			"id":     v.ID,
			"name":   v.Name,
			"damage": v.Damage,
			"cost":   v.Cost,
			"weight": v.Weight,
		}
		if len(v.Properties) > 0 {
			result["properties"] = v.Properties
		}
		if v.Range != "" {
			result["range"] = v.Range
		}
		return result

	case *equipment.Armor:
		return map[string]interface{}{
			"type":     "armor",
			"id":       v.ID,
			"name":     v.Name,
			"ac_bonus": v.ACBonus,
			"cost":     v.Cost,
			"weight":   v.Weight,
			"category": v.Type,
		}

	case *equipment.Gear:
		return map[string]interface{}{
			"type":   "gear",
			"id":     v.ID,
			"name":   v.Name,
			"cost":   v.Cost,
			"weight": v.Weight,
		}

	case *equipment.Ammunition:
		return map[string]interface{}{
			"type":   "ammunition",
			"id":     v.ID,
			"name":   v.Name,
			"cost":   v.Cost,
			"weight": v.Weight,
		}

	default:
		return map[string]interface{}{}
	}
}

// formatSearchResults formats search results for display.
func formatSearchResults(items []interface{}) string {
	var sb strings.Builder
	sb.WriteString("## Résultats de recherche\n\n")

	for _, item := range items {
		sb.WriteString("- ")
		sb.WriteString(equipment.ItemToShortDescription(item))
		sb.WriteString("\n")
	}

	return sb.String()
}
