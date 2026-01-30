package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"

	"dungeons/internal/adventure"
	"dungeons/internal/agent"
	"dungeons/internal/character"
	"dungeons/internal/data"
	"dungeons/internal/world"
)

const adventuresDir = "data/adventures"

// handleIndex renders the home page with adventure list.
func (s *Server) handleIndex(c *gin.Context) {
	adventures, err := adventure.ListAdventures(adventuresDir)
	if err != nil {
		adventures = []*adventure.Adventure{}
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":      "SkillsWeaver",
		"Adventures": adventures,
	})
}

// handleAdventuresList returns the adventures list as an HTML partial (for HTMX).
func (s *Server) handleAdventuresList(c *gin.Context) {
	adventures, err := adventure.ListAdventures(adventuresDir)
	if err != nil {
		adventures = []*adventure.Adventure{}
	}

	// Render partial for HTMX request
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":      "SkillsWeaver",
		"Adventures": adventures,
	})
}

// handleCreateAdventure creates a new adventure.
func (s *Server) handleCreateAdventure(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	description := strings.TrimSpace(c.PostForm("description"))
	theme := strings.TrimSpace(c.PostForm("theme"))

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Adventure name is required",
		})
		return
	}

	// Create the adventure
	adv := adventure.New(name, description)
	if err := adv.Save(adventuresDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create adventure: %v", err),
		})
		return
	}

	// Generate campaign plan if theme provided
	if theme != "" && s.apiKey != "" {
		if err := s.generateCampaignPlan(adv, theme); err != nil {
			// Log warning but don't fail adventure creation
			fmt.Printf("Warning: campaign plan generation failed: %v\n", err)
		}
	}

	// Copy global characters if they exist
	if err := s.copyGlobalCharactersToAdventure(adv); err != nil {
		fmt.Printf("Warning: failed to copy global characters: %v\n", err)
	}

	// Redirect to play page
	c.Redirect(http.StatusSeeOther, "/play/"+adv.Slug)
}

// handleGame renders the game page for an adventure.
func (s *Server) handleGame(c *gin.Context) {
	slug := c.Param("slug")

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, slug)
	if err != nil {
		s.renderError(c, http.StatusNotFound, fmt.Sprintf("Adventure not found: %s", slug))
		return
	}

	// Get or create session
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		s.renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	// Reload adventure context to get latest data (characters, inventory, etc.)
	if err := session.AdventureCtx.Reload(); err != nil {
		// Log error but continue with cached data
		fmt.Printf("Warning: failed to reload adventure context: %v\n", err)
	}

	// Build party info
	var partyInfo []gin.H
	for _, charName := range session.AdventureCtx.Party.Characters {
		for _, char := range session.AdventureCtx.Characters {
			if char.Name == charName {
				partyInfo = append(partyInfo, gin.H{
					"Name":    char.Name,
					"Species": char.Species,
					"Class":   char.Class,
					"Level":   char.Level,
				})
				break
			}
		}
	}

	// Check if there's an active game session
	currentSession, _ := adv.GetCurrentSession()
	isSessionActive := currentSession != nil
	activeSessionID := 0
	if currentSession != nil {
		activeSessionID = currentSession.ID
	}

	c.HTML(http.StatusOK, "game.html", gin.H{
		"Title":           adv.Name,
		"Adventure":       adv,
		"Slug":            slug,
		"Party":           partyInfo,
		"Gold":            session.AdventureCtx.Inventory.Gold,
		"CurrentLocation": session.AdventureCtx.State.CurrentLocation,
		"RecentJournal":   session.AdventureCtx.RecentJournal,
		"IsSessionActive": isSessionActive,
		"ActiveSessionID": activeSessionID,
	})
}

// handleMessage processes a message from the user.
func (s *Server) handleMessage(c *gin.Context) {
	slug := c.Param("slug")
	message := strings.TrimSpace(c.PostForm("message"))

	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Get session
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Check if already processing
	if session.IsProcessing() {
		c.JSON(http.StatusConflict, gin.H{"error": "Already processing a message"})
		return
	}

	// Start processing - this returns immediately with the output to read from
	_, err := session.ProcessMessage(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success - client will connect to SSE for results
	c.JSON(http.StatusOK, gin.H{
		"status":  "processing",
		"message": message,
	})
}

// handleStream handles the SSE stream for real-time updates.
func (s *Server) handleStream(c *gin.Context) {
	slug := c.Param("slug")

	// Get session
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Get current output
	output := session.GetCurrentOutput()
	if output == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active output"})
		return
	}

	// Setup SSE
	SetupSSE(c)

	// Stream events
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case event, ok := <-output.Events():
			if !ok {
				// Channel closed, send final event
				WriteSSE(c.Writer, SSEEvent{Event: "done", Data: "{}"})
				c.Writer.Flush()
				return
			}
			WriteSSE(c.Writer, event)
			c.Writer.Flush()
		}
	}
}

// handleCharacters returns the character list for an adventure.
func (s *Server) handleCharacters(c *gin.Context) {
	slug := c.Param("slug")

	// Get session
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Build character info
	var characters []gin.H
	for _, char := range session.AdventureCtx.Characters {
		characters = append(characters, gin.H{
			"Name":       char.Name,
			"Species":    char.Species,
			"Class":      char.Class,
			"Level":      char.Level,
			"HP":         char.HitPoints,
			"MaxHP":      char.MaxHitPoints,
			"AC":         char.ArmorClass,
			"Appearance": char.Appearance,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"characters": characters,
	})
}

// handleAdventureInfo returns updated adventure info (for refreshing UI).
func (s *Server) handleAdventureInfo(c *gin.Context) {
	slug := c.Param("slug")

	// Get session (reload context)
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Reload adventure context to get latest state
	if err := session.AdventureCtx.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return HTML directly for HTMX
	html := fmt.Sprintf(`
<div class="info-item">
    <span class="info-label">Lieu</span>
    <span class="info-value location">%s</span>
</div>
<div class="info-item">
    <span class="info-label">Or</span>
    <span class="info-value gold">%d po</span>
</div>`,
		session.AdventureCtx.State.CurrentLocation,
		session.AdventureCtx.Inventory.Gold)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// handleSessionStatus returns updated session status (for refreshing UI).
func (s *Server) handleSessionStatus(c *gin.Context) {
	slug := c.Param("slug")

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, slug)
	if err != nil {
		c.String(http.StatusNotFound, "Adventure not found")
		return
	}

	// Check if there's an active game session
	currentSession, _ := adv.GetCurrentSession()
	isActive := currentSession != nil

	// Return complete div HTML for HTMX (outerHTML swap)
	var html string
	if isActive {
		html = fmt.Sprintf(`<div class="session-status active"
             id="session-status"
             hx-get="/play/%s/session-status"
             hx-trigger="refreshInfo from:body"
             hx-swap="outerHTML">
    <span class="status-indicator"></span>
    <span class="status-text">Session %d en cours</span>
</div>`, slug, currentSession.ID)
	} else {
		html = fmt.Sprintf(`<div class="session-status inactive"
             id="session-status"
             hx-get="/play/%s/session-status"
             hx-trigger="refreshInfo from:body"
             hx-swap="outerHTML">
    <span class="status-indicator"></span>
    <span class="status-text">Aucune session active</span>
</div>`, slug)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// handleMaps serves map images from data/maps/.
func (s *Server) handleMaps(c *gin.Context) {
	filePath := c.Param("filepath")

	// Construct path to map
	mapPath := filepath.Join("data", "maps", filePath)

	// Security check: ensure path doesn't escape
	cleanPath := filepath.Clean(mapPath)
	if !strings.HasPrefix(cleanPath, filepath.Join("data", "maps")) {
		c.Status(http.StatusForbidden)
		return
	}

	c.File(mapPath)
}

// handleMinimap returns mini-map data for the current location.
func (s *Server) handleMinimap(c *gin.Context) {
	slug := c.Param("slug")

	// Get session
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Get current location from game state
	location := session.AdventureCtx.State.CurrentLocation
	if location == "" {
		location = "Unknown"
	}

	// Resolve map for location
	mapData := s.resolveMapForLocation(location)
	c.JSON(http.StatusOK, mapData)
}

// resolveMapForLocation resolves map information for a given location.
func (s *Server) resolveMapForLocation(location string) gin.H {
	// Load geography data
	geo, err := world.LoadGeography("data")
	if err != nil {
		// Return minimal data if geography can't be loaded
		return gin.H{
			"location":      location,
			"map_available": false,
			"hierarchy":     []string{location},
		}
	}

	// Validate location exists in world
	exists, loc, region, _ := world.ValidateLocationExists(location, geo)
	if !exists {
		// Location not in geography - return basic info
		return gin.H{
			"location":      location,
			"map_available": false,
			"hierarchy":     []string{location},
		}
	}

	// Build hierarchy breadcrumb
	hierarchy := []string{}
	if region != nil && region.Kingdom != "" {
		hierarchy = append(hierarchy, region.Kingdom)
	}
	if region != nil && region.Name != "" && region.Name != location {
		hierarchy = append(hierarchy, region.Name)
	}
	if loc.Type == "city" || loc.Type == "village" || strings.Contains(loc.Type, "capitale") {
		hierarchy = append(hierarchy, loc.Name)
	}

	// Determine map type
	mapType := "region"
	if strings.Contains(loc.Type, "capitale") || strings.Contains(loc.Type, "port") || loc.Type == "city" {
		mapType = "city"
	} else if strings.Contains(loc.Type, "dungeon") || strings.Contains(loc.Type, "crypte") {
		mapType = "dungeon"
	}

	// Check if map file exists
	safeName := strings.ToLower(strings.ReplaceAll(loc.Name, " ", "-"))
	mapFilename := fmt.Sprintf("%s_%s_medium_flux-pro-11.png", safeName, mapType)
	mapPath := filepath.Join("data", "maps", mapFilename)

	mapAvailable := false
	mapURL := ""
	if _, err := os.Stat(mapPath); err == nil {
		mapAvailable = true
		mapURL = fmt.Sprintf("/maps/%s", mapFilename)
	}

	return gin.H{
		"location":      loc.Name,
		"kingdom":       loc.Kingdom,
		"map_available": mapAvailable,
		"map_url":       mapURL,
		"map_type":      mapType,
		"hierarchy":     hierarchy,
	}
}

// GalleryImage represents an image in the gallery.
type GalleryImage struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Session   int    `json:"session,omitempty"`
}

// handleGallery returns the list of available images for the gallery.
func (s *Server) handleGallery(c *gin.Context) {
	slug := c.Param("slug")
	var images []GalleryImage

	// 1. Get session images from data/adventures/<slug>/images/session-N/
	adventureImagesDir := filepath.Join("data", "adventures", slug, "images")
	if entries, err := os.ReadDir(adventureImagesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "session-") {
				sessionNum := 0
				if numStr := strings.TrimPrefix(entry.Name(), "session-"); numStr != "" {
					sessionNum, _ = strconv.Atoi(numStr)
				}

				sessionDir := filepath.Join(adventureImagesDir, entry.Name())
				if files, err := os.ReadDir(sessionDir); err == nil {
					for _, file := range files {
						if !file.IsDir() && isImageFile(file.Name()) {
							title := formatImageTitle(file.Name())
							images = append(images, GalleryImage{
								URL:       fmt.Sprintf("/play/%s/images/%s", slug, filepath.Join(entry.Name(), file.Name())),
								Thumbnail: fmt.Sprintf("/play/%s/images/%s", slug, filepath.Join(entry.Name(), file.Name())),
								Title:     title,
								Category:  "session",
								Session:   sessionNum,
							})
						}
					}
				}
			}
		}
	}

	// 2. Get maps from data/maps/ (only those related to this adventure or global)
	mapsDir := filepath.Join("data", "maps")
	if entries, err := os.ReadDir(mapsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && isImageFile(entry.Name()) {
				title := formatImageTitle(entry.Name())
				images = append(images, GalleryImage{
					URL:       fmt.Sprintf("/maps/%s", entry.Name()),
					Thumbnail: fmt.Sprintf("/maps/%s", entry.Name()),
					Title:     title,
					Category:  "map",
				})
			}
		}
	}

	// Sort: session images first (by session number desc), then maps
	sort.Slice(images, func(i, j int) bool {
		if images[i].Category != images[j].Category {
			return images[i].Category == "session"
		}
		if images[i].Category == "session" {
			return images[i].Session > images[j].Session
		}
		return images[i].Title < images[j].Title
	})

	c.JSON(http.StatusOK, gin.H{
		"images": images,
		"count":  len(images),
	})
}

// isImageFile checks if a file is an image based on extension.
func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".gif"
}

// formatImageTitle creates a human-readable title from filename.
func formatImageTitle(name string) string {
	// Remove extension
	name = strings.TrimSuffix(name, filepath.Ext(name))
	// Remove common suffixes like _flux-pro-11, _schnell
	name = strings.TrimSuffix(name, "_flux-pro-11")
	name = strings.TrimSuffix(name, "_schnell")
	// Replace underscores and hyphens with spaces
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	// Title case
	return strings.Title(name)
}

// CharacterSheetAbility represents an ability score for the template.
type CharacterSheetAbility struct {
	Name        string
	Abbrev      string
	Score       int
	Modifier    int
	ModifierStr string
	IsSaveProf  bool
}

// CharacterSheetSkill represents a skill for the template.
type CharacterSheetSkill struct {
	Name         string
	Ability      string
	AbilityAbbr  string
	IsProficient bool
	Modifier     int
	ModifierStr  string
}

// CharacterSheetSpellSlot represents spell slots at a level.
type CharacterSheetSpellSlot struct {
	Level     int
	Total     int
	Used      int
	Available int
}

// CharacterSheetInventoryItem represents an item in inventory.
type CharacterSheetInventoryItem struct {
	Name     string
	Quantity int
}

// CharacterSheetAppearance represents character appearance.
type CharacterSheetAppearance struct {
	Age                int
	Gender             string
	Build              string
	Height             string
	HairColor          string
	HairStyle          string
	EyeColor           string
	SkinTone           string
	FacialFeature      string
	DistinctiveFeature string
	ArmorDescription   string
	WeaponDescription  string
}

// CharacterSheetData holds all data for the character sheet template.
type CharacterSheetData struct {
	// Identity
	Name        string
	Species     string
	SpeciesName string
	Class       string
	ClassName   string
	Level       int
	XP          int
	Background  string

	// Combat
	HP              int
	MaxHP           int
	AC              int
	Speed           int
	HitDice         string
	Initiative      string
	ProficiencyBonus int

	// Abilities
	Abilities []CharacterSheetAbility

	// Skills (18 D&D 5e skills)
	Skills []CharacterSheetSkill

	// Magic
	IsSpellcaster      bool
	SpellSaveDC        int
	SpellAttackBonus   string
	SpellcastingAbility string
	SpellSlots         []CharacterSheetSpellSlot
	KnownSpells        []string

	// Equipment
	PersonalEquipment []string
	PersonalGold      int
	SharedInventory   []CharacterSheetInventoryItem
	SharedGold        int

	// Appearance
	HasAppearance bool
	Appearance    *CharacterSheetAppearance

	// Biography
	Biography string

	// Portrait image
	HasPortrait  bool
	PortraitURL  string
}

// D&D 5e skills with their associated abilities
var dnd5eSkills = []struct {
	ID      string
	Name    string
	Ability string
	Abbrev  string
}{
	{"acrobatics", "Acrobaties", "dexterity", "DEX"},
	{"animal-handling", "Dressage", "wisdom", "SAG"},
	{"arcana", "Arcanes", "intelligence", "INT"},
	{"athletics", "Athlétisme", "strength", "FOR"},
	{"deception", "Tromperie", "charisma", "CHA"},
	{"history", "Histoire", "intelligence", "INT"},
	{"insight", "Perspicacité", "wisdom", "SAG"},
	{"intimidation", "Intimidation", "charisma", "CHA"},
	{"investigation", "Investigation", "intelligence", "INT"},
	{"medicine", "Médecine", "wisdom", "SAG"},
	{"nature", "Nature", "intelligence", "INT"},
	{"perception", "Perception", "wisdom", "SAG"},
	{"performance", "Représentation", "charisma", "CHA"},
	{"persuasion", "Persuasion", "charisma", "CHA"},
	{"religion", "Religion", "intelligence", "INT"},
	{"sleight-of-hand", "Escamotage", "dexterity", "DEX"},
	{"stealth", "Discrétion", "dexterity", "DEX"},
	{"survival", "Survie", "wisdom", "SAG"},
}

// handleCharacterSheet renders the character sheet partial for a specific character.
func (s *Server) handleCharacterSheet(c *gin.Context) {
	slug := c.Param("slug")
	charName := c.Param("name")

	// Get session
	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		// Try to create session
		var err error
		session, err = s.sessionManager.GetOrCreateSession(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
	}

	// Find character by name (case-insensitive)
	var char *character.Character
	charNameLower := strings.ToLower(charName)
	for _, ch := range session.AdventureCtx.Characters {
		if strings.ToLower(ch.Name) == charNameLower {
			char = ch
			break
		}
	}

	if char == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Character not found: %s", charName)})
		return
	}

	// Load game data for names and class info
	gd, err := data.Load("data")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load game data"})
		return
	}

	// Build character sheet data
	sheetData := buildCharacterSheetData(char, gd, session.AdventureCtx.Inventory)

	c.HTML(http.StatusOK, "character_sheet.html", sheetData)
}

// buildCharacterSheetData constructs the template data from a character.
func buildCharacterSheetData(char *character.Character, gd *data.GameData, inventory *adventure.SharedInventory) CharacterSheetData {
	sheet := CharacterSheetData{
		// Identity
		Name:       char.Name,
		Species:    char.Species,
		Class:      char.Class,
		Level:      char.Level,
		XP:         char.XP,
		Background: char.Background,

		// Combat
		HP:               char.HitPoints,
		MaxHP:            char.MaxHitPoints,
		AC:               char.ArmorClass,
		Speed:            30, // Default D&D 5e speed
		ProficiencyBonus: char.ProficiencyBonus,
	}

	// Get species and class names
	if species, ok := gd.GetSpecies(char.Species); ok {
		sheet.SpeciesName = species.Name
		sheet.Speed = species.Speed
	} else {
		sheet.SpeciesName = char.Species
	}

	classInfo, classOk := gd.GetClass(char.Class)
	if classOk {
		sheet.ClassName = classInfo.Name
		sheet.HitDice = fmt.Sprintf("%dd%d", char.Level, classInfo.HitDieSides)
	} else {
		sheet.ClassName = char.Class
		sheet.HitDice = fmt.Sprintf("%dd8", char.Level) // Default
	}

	// Initiative = DEX modifier
	sheet.Initiative = formatModifier(char.Modifiers.Dexterity)

	// Build abilities
	sheet.Abilities = []CharacterSheetAbility{
		{
			Name:        "Force",
			Abbrev:      "FOR",
			Score:       char.Abilities.Strength,
			Modifier:    char.Modifiers.Strength,
			ModifierStr: formatModifier(char.Modifiers.Strength),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["strength"],
		},
		{
			Name:        "Dextérité",
			Abbrev:      "DEX",
			Score:       char.Abilities.Dexterity,
			Modifier:    char.Modifiers.Dexterity,
			ModifierStr: formatModifier(char.Modifiers.Dexterity),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["dexterity"],
		},
		{
			Name:        "Constitution",
			Abbrev:      "CON",
			Score:       char.Abilities.Constitution,
			Modifier:    char.Modifiers.Constitution,
			ModifierStr: formatModifier(char.Modifiers.Constitution),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["constitution"],
		},
		{
			Name:        "Intelligence",
			Abbrev:      "INT",
			Score:       char.Abilities.Intelligence,
			Modifier:    char.Modifiers.Intelligence,
			ModifierStr: formatModifier(char.Modifiers.Intelligence),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["intelligence"],
		},
		{
			Name:        "Sagesse",
			Abbrev:      "SAG",
			Score:       char.Abilities.Wisdom,
			Modifier:    char.Modifiers.Wisdom,
			ModifierStr: formatModifier(char.Modifiers.Wisdom),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["wisdom"],
		},
		{
			Name:        "Charisme",
			Abbrev:      "CHA",
			Score:       char.Abilities.Charisma,
			Modifier:    char.Modifiers.Charisma,
			ModifierStr: formatModifier(char.Modifiers.Charisma),
			IsSaveProf:  char.SavingThrowProfs != nil && char.SavingThrowProfs["charisma"],
		},
	}

	// Build skills
	sheet.Skills = make([]CharacterSheetSkill, len(dnd5eSkills))
	for i, skill := range dnd5eSkills {
		abilityMod := getAbilityModifier(char, skill.Ability)
		isProficient := char.Skills != nil && char.Skills[skill.ID]

		totalMod := abilityMod
		if isProficient {
			totalMod += char.ProficiencyBonus
		}

		sheet.Skills[i] = CharacterSheetSkill{
			Name:         skill.Name,
			Ability:      skill.Ability,
			AbilityAbbr:  skill.Abbrev,
			IsProficient: isProficient,
			Modifier:     totalMod,
			ModifierStr:  formatModifier(totalMod),
		}
	}

	// Magic section
	if classOk && classInfo.SpellcastingAbility != "" {
		sheet.IsSpellcaster = char.SpellSlots != nil && len(char.SpellSlots) > 0
		if sheet.IsSpellcaster || char.SpellSaveDC > 0 {
			sheet.IsSpellcaster = true
			sheet.SpellSaveDC = char.SpellSaveDC
			sheet.SpellAttackBonus = formatModifier(char.SpellAttackBonus)
			sheet.SpellcastingAbility = getAbilityNameFR(classInfo.SpellcastingAbility)

			// Build spell slots
			for level := 1; level <= 9; level++ {
				if total, ok := char.SpellSlots[level]; ok && total > 0 {
					used := 0
					if char.SpellSlotsUsed != nil {
						used = char.SpellSlotsUsed[level]
					}
					sheet.SpellSlots = append(sheet.SpellSlots, CharacterSheetSpellSlot{
						Level:     level,
						Total:     total,
						Used:      used,
						Available: total - used,
					})
				}
			}

			sheet.KnownSpells = char.KnownSpells
		}
	}

	// Equipment
	sheet.PersonalEquipment = char.Equipment
	sheet.PersonalGold = char.Gold

	// Shared inventory
	if inventory != nil {
		sheet.SharedGold = inventory.Gold
		for _, item := range inventory.Items {
			sheet.SharedInventory = append(sheet.SharedInventory, CharacterSheetInventoryItem{
				Name:     item.Name,
				Quantity: item.Quantity,
			})
		}
	}

	// Appearance
	if char.Appearance != nil {
		sheet.HasAppearance = true
		sheet.Appearance = &CharacterSheetAppearance{
			Age:                char.Appearance.Age,
			Gender:             char.Appearance.Gender,
			Build:              char.Appearance.Build,
			Height:             char.Appearance.Height,
			HairColor:          char.Appearance.HairColor,
			HairStyle:          char.Appearance.HairStyle,
			EyeColor:           char.Appearance.EyeColor,
			SkinTone:           char.Appearance.SkinTone,
			FacialFeature:      char.Appearance.FacialFeature,
			DistinctiveFeature: char.Appearance.DistinctiveFeature,
			ArmorDescription:   char.Appearance.ArmorDescription,
			WeaponDescription:  char.Appearance.WeaponDescription,
		}
	}

	// Check for character portrait image
	charSlug := character.SanitizeFilename(char.Name)
	portraitPath := filepath.Join("data", "characters", charSlug+".png")
	if _, err := os.Stat(portraitPath); err == nil {
		sheet.HasPortrait = true
		sheet.PortraitURL = "/characters/images/" + charSlug + ".png"
	}

	return sheet
}

// formatModifier formats a modifier with + sign for positive values.
func formatModifier(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// getAbilityModifier returns the modifier for a specific ability.
func getAbilityModifier(char *character.Character, ability string) int {
	switch ability {
	case "strength":
		return char.Modifiers.Strength
	case "dexterity":
		return char.Modifiers.Dexterity
	case "constitution":
		return char.Modifiers.Constitution
	case "intelligence":
		return char.Modifiers.Intelligence
	case "wisdom":
		return char.Modifiers.Wisdom
	case "charisma":
		return char.Modifiers.Charisma
	default:
		return 0
	}
}

// getAbilityNameFR returns the French name for an ability.
func getAbilityNameFR(ability string) string {
	switch ability {
	case "strength":
		return "Force"
	case "dexterity":
		return "Dextérité"
	case "constitution":
		return "Constitution"
	case "intelligence":
		return "Intelligence"
	case "wisdom":
		return "Sagesse"
	case "charisma":
		return "Charisme"
	default:
		return ability
	}
}


// generateCampaignPlan generates a 3-act campaign plan using the DM agent.
func (s *Server) generateCampaignPlan(adv *adventure.Adventure, theme string) error {
	if s.apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Load DM persona
	personaLoader := agent.NewPersonaLoader()
	_, dmPersona, err := personaLoader.LoadWithMetadata("dungeon-master")
	if err != nil {
		return fmt.Errorf("failed to load DM persona: %w", err)
	}

	// Build prompt for campaign plan generation
	prompt := fmt.Sprintf(`Generate a complete 3-act D&D 5e campaign plan for this adventure:

**Adventure Name**: %s
**Description**: %s
**Theme**: %s

Create a comprehensive campaign plan with:
1. Campaign title and compelling objective
2. 3 acts with titles, descriptions, key events, and goals
3. Primary antagonist with motivation and arc
4. 1-2 supporting characters
5. 1-2 MacGuffins (artifacts/treasures/knowledge)
6. 3-5 key locations with danger levels
7. 2-3 critical initial foreshadows linked to acts

Target: 8-12 sessions, 3 hours each.

CRITICAL: Return ONLY valid JSON matching this exact structure (no markdown, no explanation):
{
  "version": "1.0.0",
  "metadata": {
    "campaign_title": "Title here",
    "theme": "Theme description",
    "target_duration": {"sessions": 10, "hours_per_session": 3},
    "created_at": "2026-01-29T12:00:00Z",
    "generated_by": "dungeon-master",
    "last_updated": "2026-01-29T12:00:00Z"
  },
  "narrative_structure": {
    "objective": "Main campaign objective",
    "hook": "Opening hook that draws players in",
    "acts": [
      {
        "number": 1,
        "title": "Act 1 Title",
        "description": "What happens in this act",
        "target_sessions": [1,2,3,4],
        "status": "pending",
        "key_events": ["Event 1", "Event 2"],
        "goals": ["Goal 1", "Goal 2"],
        "completion_criteria": {"milestone": "What marks act completion"}
      },
      {
        "number": 2,
        "title": "Act 2 Title",
        "description": "What happens",
        "target_sessions": [5,6,7,8],
        "status": "pending",
        "key_events": ["Event 1"],
        "goals": ["Goal 1"],
        "completion_criteria": {"milestone": "Milestone"}
      },
      {
        "number": 3,
        "title": "Act 3 Title",
        "description": "What happens",
        "target_sessions": [9,10,11,12],
        "status": "pending",
        "key_events": ["Event 1"],
        "goals": ["Goal 1"],
        "completion_criteria": {"milestone": "Milestone"}
      }
    ],
    "climax": {
      "description": "The climactic confrontation",
      "target_session": 12,
      "stakes": "What's at stake if heroes fail"
    },
    "resolution": {
      "success_scenario": "What happens if heroes succeed",
      "failure_scenario": "What happens if heroes fail",
      "epilogue_notes": "How the story concludes"
    }
  },
  "plot_elements": {
    "antagonist": {
      "name": "Antagonist Name",
      "role": "primary",
      "motivation": "Why they do what they do",
      "introduction_session": 2,
      "final_confrontation_session": 12,
      "arc": "How they evolve"
    },
    "secondary_antagonists": [],
    "supporting_characters": [
      {
        "name": "NPC Name",
        "role": "ally",
        "motivation": "Their motivation",
        "arc": "Their character arc",
        "key_sessions": [2,6,10]
      }
    ],
    "macguffins": [
      {
        "name": "Artifact Name",
        "type": "artifact",
        "significance": "Why it matters",
        "introduced_session": 1,
        "resolution": "How it's resolved"
      }
    ],
    "key_locations": [
      {
        "name": "Location Name",
        "kingdom": "kingdom-name",
        "type": "city",
        "role": "Act 1 hub",
        "sessions": [1,2,3],
        "danger_level": "moderate"
      }
    ]
  },
  "foreshadows": {
    "active": [
      {
        "id": "fsh_001",
        "description": "Foreshadow description",
        "planted_at": "2026-01-29T12:00:00Z",
        "planted_session": 0,
        "importance": "critical",
        "category": "villain",
        "tags": ["tag1"],
        "context": "Where/how it appears",
        "related_npcs": [],
        "related_locations": [],
        "linked_to_act": 2,
        "linked_to_plot_point": "climax",
        "target_payoff_session": 10,
        "payoff_type": "revelation"
      }
    ],
    "resolved": [],
    "abandoned": [],
    "next_id": 2
  },
  "progression": {
    "current_act": 1,
    "current_session": 0,
    "completed_plot_points": [],
    "active_threads": [],
    "pending_resolutions": ["fsh_001"]
  },
  "pacing": {
    "sessions_played": 0,
    "sessions_remaining_estimate": 10,
    "act_breakdown": {
      "act_1": {"planned": 4, "actual": 0, "variance": 0},
      "act_2": {"planned": 4, "actual": 0, "variance": 0},
      "act_3": {"planned": 4, "actual": 0, "variance": 0}
    }
  },
  "dm_notes": {
    "themes": ["theme1", "theme2"],
    "tone": "Dark fantasy with hope",
    "player_agency": "Notes on player choices",
    "memorable_moments": []
  }
}`, adv.Name, adv.Description, theme)

	// Call Anthropic API
	client := anthropic.NewClient(option.WithAPIKey(s.apiKey))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5, // Haiku for speed
		MaxTokens: 8192,
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: dmPersona,
			},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock(prompt),
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	// Extract JSON from response
	var responseText string
	for _, block := range response.Content {
		switch contentBlock := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText += contentBlock.Text
		}
	}

	if responseText == "" {
		return fmt.Errorf("empty response from API")
	}

	// Parse JSON (remove markdown code blocks if present)
	jsonText := strings.TrimSpace(responseText)
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(jsonText, "```")
	jsonText = strings.TrimSpace(jsonText)

	// Unmarshal into CampaignPlan
	var campaignPlan adventure.CampaignPlan
	if err := json.Unmarshal([]byte(jsonText), &campaignPlan); err != nil {
		// Log the response for debugging
		fmt.Printf("Failed to parse campaign plan JSON: %v\nResponse:\n%s\n", err, jsonText)
		return fmt.Errorf("failed to parse campaign plan: %w", err)
	}

	// Save to file
	if err := adv.SaveCampaignPlan(&campaignPlan); err != nil {
		return fmt.Errorf("failed to save campaign plan: %w", err)
	}

	fmt.Printf("✓ Generated campaign plan for '%s': %s\n", adv.Name, campaignPlan.Metadata.CampaignTitle)
	return nil
}

// copyGlobalCharactersToAdventure copies existing characters from data/characters/ to the new adventure.
func (s *Server) copyGlobalCharactersToAdventure(adv *adventure.Adventure) error {
	globalCharactersDir := filepath.Join("data", "characters")

	// Check if global characters directory exists
	if _, err := os.Stat(globalCharactersDir); os.IsNotExist(err) {
		// No global characters directory, skip
		return nil
	}

	// List all character JSON files in global directory
	entries, err := os.ReadDir(globalCharactersDir)
	if err != nil {
		return fmt.Errorf("failed to read global characters: %w", err)
	}

	var characterFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			characterFiles = append(characterFiles, entry.Name())
		}
	}

	if len(characterFiles) == 0 {
		// No characters to copy
		return nil
	}

	// Create characters directory for the adventure
	adventureCharactersDir := filepath.Join("data", "adventures", adv.Slug, "characters")
	if err := os.MkdirAll(adventureCharactersDir, 0755); err != nil {
		return fmt.Errorf("failed to create adventure characters directory: %w", err)
	}

	// Copy each character file
	var characterNames []string
	for _, charFile := range characterFiles {
		srcPath := filepath.Join(globalCharactersDir, charFile)
		dstPath := filepath.Join(adventureCharactersDir, charFile)

		// Read character file to extract name
		data, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Printf("Warning: failed to read character file %s: %v\n", charFile, err)
			continue
		}

		// Parse to get character name
		var charData map[string]interface{}
		if err := json.Unmarshal(data, &charData); err != nil {
			fmt.Printf("Warning: failed to parse character file %s: %v\n", charFile, err)
			continue
		}

		if name, ok := charData["name"].(string); ok && name != "" {
			characterNames = append(characterNames, name)
		}

		// Copy file
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			fmt.Printf("Warning: failed to copy character file %s: %v\n", charFile, err)
			continue
		}
	}

	if len(characterNames) == 0 {
		// No valid characters copied
		return nil
	}

	// Create party.json with the copied characters
	partyData := map[string]interface{}{
		"characters":     characterNames,
		"marching_order": characterNames,
		"formation":      "travel",
	}

	partyPath := filepath.Join("data", "adventures", adv.Slug, "party.json")
	partyJSON, err := json.MarshalIndent(partyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal party.json: %w", err)
	}

	if err := os.WriteFile(partyPath, partyJSON, 0644); err != nil {
		return fmt.Errorf("failed to write party.json: %w", err)
	}

	fmt.Printf("✓ Copied %d character(s) to adventure '%s': %v\n", len(characterNames), adv.Name, characterNames)
	return nil
}
