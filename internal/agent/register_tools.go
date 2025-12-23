package agent

import (
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/dmtools"
)

// registerAllTools registers all available tools in the registry.
func registerAllTools(registry *ToolRegistry, dataDir string, adv *adventure.Adventure) error {
	// Register dice roller
	registry.Register(dmtools.NewDiceRollerTool())

	// Register monster tool
	monsterTool, err := dmtools.NewMonsterTool(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create monster tool: %w", err)
	}
	registry.Register(monsterTool)

	// Register treasure tool
	treasureTool, err := dmtools.NewGenerateTreasureTool(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create treasure tool: %w", err)
	}
	registry.Register(treasureTool)

	// Register NPC tool
	npcTool, err := dmtools.NewGenerateNPCTool(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create NPC tool: %w", err)
	}
	registry.Register(npcTool)

	// Register adventure tools - now passing Adventure object for real persistence
	registry.Register(dmtools.NewLogEventTool(adv))
	registry.Register(dmtools.NewAddGoldTool(adv))
	registry.Register(dmtools.NewGetInventoryTool(adv))

	// Register image generation tool
	imageTool, err := dmtools.NewGenerateImageTool(adv.BasePath())
	if err != nil {
		// Log warning but don't fail if FAL_KEY is not set
		fmt.Printf("Warning: Image generation tool not available: %v\n", err)
	} else {
		registry.Register(imageTool)
	}

	return nil
}
