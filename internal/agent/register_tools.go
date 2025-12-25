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

	// Register NPC tool with adventure context for persistence
	npcTool, err := dmtools.NewGenerateNPCTool(dataDir, adv)
	if err != nil {
		return fmt.Errorf("failed to create NPC tool: %w", err)
	}
	registry.Register(npcTool)

	// Register session management tools (MUST be registered for proper session tracking)
	registry.Register(dmtools.NewStartSessionTool(adv))
	registry.Register(dmtools.NewEndSessionTool(adv))
	registry.Register(dmtools.NewGetSessionInfoTool(adv))

	// Register adventure tools - now passing Adventure object for real persistence
	registry.Register(dmtools.NewLogEventTool(adv))
	registry.Register(dmtools.NewAddGoldTool(adv))
	registry.Register(dmtools.NewGetInventoryTool(adv))

	// Register NPC management tools
	registry.Register(dmtools.NewUpdateNPCImportanceTool(adv))
	registry.Register(dmtools.NewGetNPCHistoryTool(adv))

	// Register image generation tool
	imageTool, err := dmtools.NewGenerateImageTool(adv.BasePath())
	if err != nil {
		// Log warning but don't fail if FAL_KEY is not set
		fmt.Printf("Warning: Image generation tool not available: %v\n", err)
	} else {
		registry.Register(imageTool)
	}

	// Register map generation tool
	mapTool, err := dmtools.NewGenerateMapTool(dataDir, adv.BasePath())
	if err != nil {
		// Log warning but don't fail if ANTHROPIC_API_KEY is not set
		fmt.Printf("Warning: Map generation tool not available: %v\n", err)
	} else {
		registry.Register(mapTool)
	}

	return nil
}
