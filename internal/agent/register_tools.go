package agent

import (
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/dmtools"
	"dungeons/internal/equipment"
	"dungeons/internal/locations"
	"dungeons/internal/monster"
	"dungeons/internal/names"
	"dungeons/internal/skills"
	"dungeons/internal/spell"
)

// registerAllTools registers all available tools in the registry.
func registerAllTools(registry *ToolRegistry, dataDir string, adv *adventure.Adventure, agentManager *AgentManager) error {
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

	// Register foreshadowing tools
	registry.Register(dmtools.NewPlantForeshadowTool(adv))
	registry.Register(dmtools.NewResolveForeshadowTool(adv))
	registry.Register(dmtools.NewListForeshadowsTool(adv))
	registry.Register(dmtools.NewGetStaleForeshadowsTool(adv))

	// Register character info tools
	registry.Register(dmtools.NewGetPartyInfoTool(adv))
	registry.Register(dmtools.NewGetCharacterInfoTool(adv))

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

	// Register equipment lookup tool
	equipmentCatalog, err := equipment.NewCatalog(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create equipment catalog: %w", err)
	}
	registry.Register(dmtools.NewGetEquipmentTool(equipmentCatalog))

	// Register spell lookup tool
	spellManager, err := spell.NewManagerFromDataDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create spell manager: %w", err)
	}
	registry.Register(dmtools.NewGetSpellTool(spellManager))

	// Register encounter tools (uses existing bestiary)
	bestiary, err := monster.NewBestiary(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create bestiary for encounters: %w", err)
	}
	registry.Register(dmtools.NewGenerateEncounterTool(bestiary))
	registry.Register(dmtools.NewRollMonsterHPTool(bestiary))

	// Register inventory management tools
	registry.Register(dmtools.NewAddItemTool(adv))
	registry.Register(dmtools.NewRemoveItemTool(adv))

	// Register name generation tools
	nameGenerator, err := names.NewGenerator(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create name generator: %w", err)
	}
	registry.Register(dmtools.NewGenerateNameTool(nameGenerator))

	// Register location name generation tool
	locationGenerator, err := locations.NewGenerator(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create location generator: %w", err)
	}
	registry.Register(dmtools.NewGenerateLocationNameTool(locationGenerator))

	// Register agent invocation tool (requires agentManager to be passed)
	if agentManager != nil {
		registry.Register(dmtools.NewInvokeAgentTool(agentManager))
	}

	// Register skill invocation tool
	skillRegistry, err := skills.NewRegistry()
	if err != nil {
		// Log warning but don't fail - skills are optional enhancements
		fmt.Printf("Warning: Skills not available: %v\n", err)
	} else {
		registry.Register(dmtools.NewInvokeSkillTool(skillRegistry))
	}

	return nil
}
