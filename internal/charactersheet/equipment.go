package charactersheet

import (
	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/data"
)

// EquipmentSummary consolidates all equipment
type EquipmentSummary struct {
	Items       []EquipmentItem  // Personal equipment
	SharedItems []EquipmentItem  // Adventure inventory
	SharedGold  int              // Shared gold
	Empty       bool             // True if no equipment
}

// EquipmentItem represents a piece of equipment
type EquipmentItem struct {
	ID       string
	Name     string
	Category string  // "weapon", "armor", "gear", "magic"
	Quantity int
	Weight   float64
	Source   string  // "character", "adventure"
	Details  string  // e.g., "1d8 damage, versatile"
	Icon     string  // Optional icon path
}

// EquipmentExtractor pulls equipment from multiple sources
type EquipmentExtractor struct {
	gameData *data.GameData
}

// NewEquipmentExtractor creates a new equipment extractor
func NewEquipmentExtractor(gd *data.GameData) *EquipmentExtractor {
	return &EquipmentExtractor{
		gameData: gd,
	}
}

// Extract consolidates equipment from character + adventure
func (e *EquipmentExtractor) Extract(c *character.Character, adventureName string) (*EquipmentSummary, error) {
	summary := &EquipmentSummary{}

	// 1. Character equipment (if present)
	if len(c.Equipment) > 0 {
		for _, itemID := range c.Equipment {
			item := e.enrichItem(itemID, "character")
			summary.Items = append(summary.Items, item)
		}
	}

	// 2. Adventure inventory (if specified)
	if adventureName != "" {
		adv, err := adventure.LoadByName("data/adventures", adventureName)
		if err == nil {
			inv, err := adv.LoadInventory()
			if err == nil {
				// Convert adventure inventory items to EquipmentItems
				for _, invItem := range inv.Items {
					item := EquipmentItem{
						ID:       invItem.ID,
						Name:     invItem.Name,
						Category: "gear", // Default category
						Quantity: invItem.Quantity,
						Source:   "adventure",
						Details:  invItem.Description,
					}
					summary.SharedItems = append(summary.SharedItems, item)
				}
				summary.SharedGold = inv.Gold
			}
		}
	}

	// 3. Check if empty
	if len(summary.Items) == 0 && len(summary.SharedItems) == 0 {
		summary.Empty = true
	}

	return summary, nil
}

// enrichItem enriches an item with data from equipment catalog
func (e *EquipmentExtractor) enrichItem(itemID string, source string) EquipmentItem {
	item := EquipmentItem{
		ID:       itemID,
		Name:     itemID, // Fallback to ID
		Category: "gear",
		Quantity: 1,
		Source:   source,
	}

	// Try to enrich from equipment data
	// Note: The equipment package needs to be enhanced to support lookup by ID
	// For now, we'll use the ID as-is

	// TODO: Enhance with real equipment data lookup when equipment package supports it
	// For now, just provide basic categorization
	if e.gameData != nil {
		// Future: Look up weapon/armor stats from gameData
		item.Details = "" // Will be populated when equipment lookup is available
	}

	return item
}
