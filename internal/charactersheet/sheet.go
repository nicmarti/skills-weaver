package charactersheet

import (
	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/data"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SheetGenerator handles character sheet generation
type SheetGenerator struct {
	gameData        *data.GameData
	bioGenerator    *BiographyGenerator
	equipExtractor  *EquipmentExtractor
	templateManager *TemplateManager
}

// SheetOptions configures sheet generation
type SheetOptions struct {
	CharacterName    string
	Adventure        string
	IncludeBiography bool
	RefreshBio       bool
	IncludePortrait  bool
	OutputPath       string
	GenerateBanner   bool
	GenerateIcons    bool
}

// Sheet represents a complete character sheet
type Sheet struct {
	Character    *character.Character
	Biography    *Biography
	Equipment    *EquipmentSummary
	Adventure    *AdventureContext
	RaceName     string
	ClassName    string
	Gold         int
	ClassBanner  string
	GeneratedAt  string
}

// AdventureContext provides adventure-specific data
type AdventureContext struct {
	Name         string
	SharedGold   int
	SharedItems  []EquipmentItem
	Party        []string
	SessionCount int
	LastPlayed   time.Time
}

// NewSheetGenerator creates a new sheet generator
func NewSheetGenerator(dataDir string) (*SheetGenerator, error) {
	gd, err := data.Load(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load game data: %w", err)
	}

	return &SheetGenerator{
		gameData:        gd,
		bioGenerator:    NewBiographyGenerator(),
		equipExtractor:  NewEquipmentExtractor(gd),
		templateManager: NewTemplateManager(),
	}, nil
}

// Generate creates a character sheet
func (g *SheetGenerator) Generate(opts SheetOptions) (*Sheet, error) {
	// 1. Load character
	charPath := filepath.Join("data", "characters", sanitizeFilename(opts.CharacterName)+".json")
	c, err := character.Load(charPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %w", err)
	}

	sheet := &Sheet{
		Character:   c,
		RaceName:    getRaceName(c.Species),
		ClassName:   getClassName(c.Class),
		Gold:        c.Gold,
		GeneratedAt: time.Now().Format("2 January 2006 à 15:04"),
	}

	// 2. Load or generate biography
	if opts.IncludeBiography {
		bioPath := filepath.Join("data", "characters", sanitizeFilename(opts.CharacterName)+"_bio.json")

		if opts.RefreshBio || !fileExists(bioPath) {
			// Generate new biography
			bio, err := g.bioGenerator.Generate(c, opts.Adventure)
			if err != nil {
				return nil, fmt.Errorf("failed to generate biography: %w", err)
			}

			// Save biography cache
			if err := bio.Save(filepath.Join("data", "characters")); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save biography cache: %v\n", err)
			}

			sheet.Biography = bio
		} else {
			// Load cached biography
			bio, err := LoadBiography(opts.CharacterName, filepath.Join("data", "characters"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load biography cache: %v\n", err)
			} else {
				sheet.Biography = bio
			}
		}
	}

	// 3. Extract equipment
	equipSummary, err := g.equipExtractor.Extract(c, opts.Adventure)
	if err != nil {
		return nil, fmt.Errorf("failed to extract equipment: %w", err)
	}
	sheet.Equipment = equipSummary

	// 4. Load adventure context if specified
	if opts.Adventure != "" {
		advCtx, err := LoadAdventureContext(opts.Adventure)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load adventure context: %v\n", err)
		} else {
			sheet.Adventure = advCtx
			sheet.Gold = advCtx.SharedGold // Use shared gold if in adventure
		}
	}

	return sheet, nil
}

// RenderHTML generates HTML from sheet
func (g *SheetGenerator) RenderHTML(sheet *Sheet) (string, error) {
	return g.templateManager.RenderTemplate(sheet)
}

// Save writes HTML to file
func (g *SheetGenerator) Save(html string, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Helper functions

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getRaceName(race string) string {
	names := map[string]string{
		"human":    "Humain",
		"elf":      "Elfe",
		"dwarf":    "Nain",
		"halfling": "Halfelin",
	}
	if name, ok := names[race]; ok {
		return name
	}
	return race
}

func getClassName(class string) string {
	names := map[string]string{
		"fighter":    "Guerrier",
		"cleric":     "Clerc",
		"magic-user": "Magicien",
		"thief":      "Voleur",
	}
	if name, ok := names[class]; ok {
		return name
	}
	return class
}

// LoadAdventureContext loads adventure data
func LoadAdventureContext(adventureName string) (*AdventureContext, error) {
	ctx := &AdventureContext{
		Name: adventureName,
	}

	// Load adventure
	adv, err := adventure.LoadByName("data/adventures", adventureName)
	if err != nil {
		return ctx, fmt.Errorf("loading adventure: %w", err)
	}

	// Load shared inventory to get gold
	inv, err := adv.LoadInventory()
	if err == nil {
		ctx.SharedGold = inv.Gold
	}

	// Load party information
	party, err := adv.LoadParty()
	if err == nil {
		ctx.Party = party.Characters
	}

	return ctx, nil
}
