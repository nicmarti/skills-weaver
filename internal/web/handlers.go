package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	"dungeons/internal/coherence"
	"dungeons/internal/data"
	"dungeons/internal/narrativeai"
	"dungeons/internal/npc"
	"dungeons/internal/npcmanager"
	"dungeons/internal/tarot"
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
	duration := strings.TrimSpace(c.PostForm("duration"))
	adventureType := strings.TrimSpace(c.PostForm("adventure_type"))

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
		if err := s.generateCampaignPlan(adv, theme, duration, adventureType); err != nil {
			// Log warning but don't fail adventure creation
			fmt.Printf("Warning: campaign plan generation failed: %v\n", err)
		} else {
			// Generate NPCs from the campaign plan
			if err := s.generateAdventureNPCs(adv); err != nil {
				fmt.Printf("Warning: NPC generation failed: %v\n", err)
			}
			// Validate the campaign plan
			if plan, err := adv.LoadCampaignPlan(); err == nil {
				result := adventure.ValidateCampaignPlan(plan)
				if len(result.Errors) > 0 || len(result.Warnings) > 0 {
					fmt.Printf("Campaign plan validation for '%s': score=%d, errors=%d, warnings=%d\n",
						adv.Name, result.Score, len(result.Errors), len(result.Warnings))
					for _, e := range result.Errors {
						fmt.Printf("  ERROR: %s\n", e)
					}
					for _, w := range result.Warnings {
						fmt.Printf("  WARNING: %s\n", w)
					}
				}
			}
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

	// Determine current model for the selector
	currentModel := "sonnet" // default
	if session.Agent != nil {
		modelStr := agent.GetModelDisplayName(session.Agent.GetModel())
		switch {
		case strings.Contains(modelStr, "opus"):
			currentModel = "opus"
		case strings.Contains(modelStr, "sonnet"):
			currentModel = "sonnet"
		}
	}

	// Auto-start: a freshly generated adventure has never been played (no active
	// session, no recorded sessions, empty journal). The client kicks off the
	// DM's opening narration so the player isn't dropped on an empty page with
	// only a "type something to begin" hint.
	autoStart := !isSessionActive &&
		adv.SessionCount == 0 &&
		len(session.AdventureCtx.RecentJournal) == 0

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
		"CurrentModel":    currentModel,
		"AutoStart":       autoStart,
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

	// Get or recreate session (handles session expiration gracefully)
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
		return
	}

	// Check if already processing
	if session.IsProcessing() {
		c.JSON(http.StatusConflict, gin.H{"error": "Already processing a message"})
		return
	}

	// Start processing - this returns immediately with the output to read from
	_, err = session.ProcessMessage(message)
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

	// Get or recreate session
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
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

	// Get or recreate session
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
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

	// Get or recreate session (reload context)
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
		return
	}

	// Reload adventure context to get latest state
	if err := session.AdventureCtx.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Determine current model for selector
	currentModel := "sonnet"
	if session.Agent != nil {
		modelStr := agent.GetModelDisplayName(session.Agent.GetModel())
		if strings.Contains(modelStr, "opus") {
			currentModel = "opus"
		}
	}
	sonnetSelected := ""
	opusSelected := ""
	if currentModel == "sonnet" {
		sonnetSelected = " selected"
	} else {
		opusSelected = " selected"
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
</div>
<div class="info-item">
    <span class="info-label">Modele IA</span>
    <select id="model-selector" class="model-select">
        <option value="sonnet"%s>Sonnet 4.6</option>
        <option value="opus"%s>Opus 4.6</option>
    </select>
</div>`,
		session.AdventureCtx.State.CurrentLocation,
		session.AdventureCtx.Inventory.Gold,
		sonnetSelected,
		opusSelected)

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

// handleCoherence renders the read-only coherence analysis of an adventure.
// With ?format=json it returns the raw Report (the machine feed for Claude Code);
// otherwise it renders the coherence dashboard page.
func (s *Server) handleCoherence(c *gin.Context) {
	slug := c.Param("slug")

	adv, err := adventure.LoadByName(adventuresDir, slug)
	if err != nil {
		if c.Query("format") == "json" {
			c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		} else {
			s.renderError(c, http.StatusNotFound, "Aventure introuvable: "+slug)
		}
		return
	}

	report, err := coherence.Analyze(adv)
	if err != nil {
		if c.Query("format") == "json" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			s.renderError(c, http.StatusInternalServerError, "Analyse impossible: "+err.Error())
		}
		return
	}

	if c.Query("format") == "json" {
		c.JSON(http.StatusOK, report)
		return
	}

	// Load the cached AI narrative judgment (if any) and flag staleness.
	judgment, _ := coherence.LoadNarrativeJudgment(adv)
	played := 0
	if report.NarrativeBrief != nil {
		played = report.NarrativeBrief.PlayedSessions
	}
	judgmentStale := judgment != nil && judgment.Stale(played)
	var judgmentTime string
	if judgment != nil {
		judgmentTime = judgment.GeneratedAt.Format("02/01/2006 15:04")
	}

	c.HTML(http.StatusOK, "coherence.html", gin.H{
		"Slug":          slug,
		"AdventureName": report.AdventureName,
		"GeneratedAt":   report.GeneratedAt.Format("02/01/2006 15:04"),
		"Layers": []gin.H{
			layerView("Intégrité", report.Integrity),
			layerView("Dérive (squelette ↔ déroulé)", report.Drift),
			layerView("Qualité narrative (signaux)", report.Narrative),
		},
		"Brief":         report.NarrativeBrief,
		"Judgment":      judgment,
		"JudgmentTime":  judgmentTime,
		"JudgmentStale": judgmentStale,
		"AIAvailable":   s.apiKey != "",
	})
}

// handleCoherenceAnalyze runs the AI narrative judgment (3 lenses + synthesis)
// on demand, caches it, then redirects back to the coherence dashboard.
func (s *Server) handleCoherenceAnalyze(c *gin.Context) {
	slug := c.Param("slug")

	if s.apiKey == "" {
		s.renderError(c, http.StatusBadRequest, "Clé API Anthropic absente — le jugement narratif IA est indisponible.")
		return
	}

	// The game session owns a fully-wired AgentManager (the invoker).
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		s.renderError(c, http.StatusNotFound, "Session introuvable: "+err.Error())
		return
	}

	// Load the adventure fresh to guarantee basePath for the brief and cache IO.
	adv, err := adventure.LoadByName(adventuresDir, slug)
	if err != nil {
		s.renderError(c, http.StatusNotFound, "Aventure introuvable: "+slug)
		return
	}

	brief, err := coherence.BuildNarrativeBrief(adv)
	if err != nil {
		s.renderError(c, http.StatusInternalServerError, "Dossier narratif: "+err.Error())
		return
	}

	log.Printf("[coherence-ai][%s] analyse demandée — lancement du jugement narratif (3 agents + synthèse)", slug)
	start := time.Now()
	progress := func(msg string) {
		log.Printf("[coherence-ai][%s] %s", slug, msg)
	}
	judgment, err := narrativeai.Judge(brief, session.Agent.AgentManager(), progress)
	if err != nil {
		log.Printf("[coherence-ai][%s] ÉCHEC après %s: %v", slug, time.Since(start).Round(time.Second), err)
		s.renderError(c, http.StatusInternalServerError, "Analyse IA échouée: "+err.Error())
		return
	}
	log.Printf("[coherence-ai][%s] jugement terminé en %s (%d perspectives)", slug, time.Since(start).Round(time.Second), len(judgment.Lenses))

	if err := coherence.SaveNarrativeJudgment(adv, judgment); err != nil {
		s.renderError(c, http.StatusInternalServerError, "Sauvegarde du jugement: "+err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/play/"+slug+"/coherence")
}

// layerView adapts a coherence LayerReport for the dashboard template.
func layerView(title string, lr coherence.LayerReport) gin.H {
	errs, warns, infos := lr.Counts()
	scoreClass := "score-good"
	switch {
	case lr.Score < 70:
		scoreClass = "score-bad"
	case lr.Score < 90:
		scoreClass = "score-warn"
	}
	return gin.H{
		"Title":      title,
		"Score":      lr.Score,
		"ScoreClass": scoreClass,
		"Errors":     errs,
		"Warnings":   warns,
		"Infos":      infos,
		"Findings":   lr.Findings,
	}
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

	// Get or recreate session
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
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
	// Session is always serialized (no omitempty) so session-0 images render as
	// "S0" in the badge instead of "Sundefined".
	Session int `json:"session"`
}

// mapTypeTokens identify in-game generated maps by filename convention
// ({name}_{maptype}_{scale}_{model}.png). Used to categorize an adventure
// session image as a "map" (Cartes tab) versus a "session" scene.
var mapTypeTokens = []string{"_region_", "_city_", "_dungeon_", "_tactical_"}

// isMapFilename reports whether a session-dir image was produced by the map
// generator, based on its filename tokens.
func isMapFilename(name string) bool {
	lower := strings.ToLower(name)
	for _, tok := range mapTypeTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
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
							// In-game maps follow the {name}_{maptype}_{scale} naming
							// convention and belong in the Cartes tab; everything else
							// is a session scene/portrait.
							category := "session"
							if isMapFilename(file.Name()) {
								category = "map"
							}
							images = append(images, GalleryImage{
								URL:       fmt.Sprintf("/play/%s/images/%s", slug, filepath.Join(entry.Name(), file.Name())),
								Thumbnail: fmt.Sprintf("/play/%s/images/%s", slug, filepath.Join(entry.Name(), file.Name())),
								Title:     title,
								Category:  category,
								Session:   sessionNum,
							})
						}
					}
				}
			}
		}
	}

	// Note: the global data/maps/ pool is intentionally NOT listed here. It is a
	// shared scratch directory for the standalone sw-map CLI with no adventure
	// association, so listing it leaked maps from unrelated adventures into the
	// gallery. Maps generated in-game (via generate_map) are saved into the
	// adventure's images/session-N/ directory and are picked up above.

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

	// Class Features
	ClassFeatures []string

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

	// Get or recreate session
	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to restore session: " + err.Error()})
		return
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

	// Class features
	sheet.ClassFeatures = char.ClassFeatures

	// Magic section
	if classOk && classInfo.SpellcastingAbility != "" {
		sheet.IsSpellcaster = len(char.SpellSlots) > 0
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

	// Check for character portrait image (try multiple extensions)
	charSlug := character.SanitizeFilename(char.Name)
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp"} {
		portraitPath := filepath.Join("data", "characters", charSlug+ext)
		if _, err := os.Stat(portraitPath); err == nil {
			sheet.HasPortrait = true
			sheet.PortraitURL = "/characters/images/" + charSlug + ext
			break
		}
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


// handleArchiveAdventure archives an adventure and redirects to home.
func (s *Server) handleArchiveAdventure(c *gin.Context) {
	slug := c.Param("slug")

	if err := adventure.Archive(adventuresDir, slug); err != nil {
		s.renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to archive: %v", err))
		return
	}

	s.sessionManager.RemoveSession(slug)
	c.Redirect(http.StatusSeeOther, "/")
}

// handleDeleteAdventure syncs characters to global, then deletes the adventure.
func (s *Server) handleDeleteAdventure(c *gin.Context) {
	slug := c.Param("slug")

	// Sync characters before deleting to preserve progression
	adv, err := adventure.LoadByName(adventuresDir, slug)
	if err == nil {
		globalCharDir := filepath.Join("data", "characters")
		if synced, syncErr := adv.SyncCharactersToGlobal(globalCharDir); syncErr != nil {
			fmt.Printf("Warning: failed to sync characters before delete: %v\n", syncErr)
		} else if len(synced) > 0 {
			fmt.Printf("Synced %d character(s) before deleting '%s': %v\n", len(synced), slug, synced)
		}
	}

	if err := adventure.Delete(adventuresDir, slug); err != nil {
		s.renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete: %v", err))
		return
	}

	s.sessionManager.RemoveSession(slug)
	c.Redirect(http.StatusSeeOther, "/")
}

// handleGetModel returns the current model for an adventure session.
func (s *Server) handleGetModel(c *gin.Context) {
	slug := c.Param("slug")

	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	model := session.Agent.GetModel()
	displayName := agent.GetModelDisplayName(model)

	// Map to short name
	shortName := "sonnet"
	if strings.Contains(displayName, "opus") {
		shortName = "opus"
	}

	c.JSON(http.StatusOK, gin.H{
		"model":   shortName,
		"display": displayName,
	})
}

// handleSetModel changes the model used by the DM agent.
func (s *Server) handleSetModel(c *gin.Context) {
	slug := c.Param("slug")
	modelName := strings.TrimSpace(c.PostForm("model"))

	// Validate: only sonnet or opus
	if modelName != "sonnet" && modelName != "opus" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model. Use 'sonnet' or 'opus'."})
		return
	}

	session, err := s.sessionManager.GetOrCreateSession(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Map and apply
	previousModel := agent.GetModelDisplayName(session.Agent.GetModel())
	anthropicModel := agent.MapPersonaModelToAnthropic(modelName)
	session.Agent.SetModel(anthropicModel)

	displayName := agent.GetModelDisplayName(anthropicModel)
	fmt.Printf("[%s] Model changed: %s → %s\n", slug, previousModel, displayName)
	c.JSON(http.StatusOK, gin.H{
		"model":   modelName,
		"display": displayName,
	})
}

// generateCampaignPlan generates a campaign plan using the DM agent.
// The plan adapts to the chosen duration (oneshot/short/campaign) and adventure type.
// It is the backward-compatible entry point used by the quick-create form.
func (s *Server) generateCampaignPlan(adv *adventure.Adventure, theme, duration, adventureType string) error {
	return s.generateCampaignPlanWithBrief(adv, theme, duration, adventureType, nil)
}

// buildCampaignPlanRequest assembles the system persona + the full user prompt
// for the campaign-plan generator, including the optional fortune-teller brief
// and world-coherence guidance. It is shared by the real generator and by the
// wizard debug endpoint so the two can never drift. It performs no API call.
func (s *Server) buildCampaignPlanRequest(adv *adventure.Adventure, theme, duration, adventureType string, brief *tarot.CreativeBrief) (string, string, *agent.WorldResources, adventure.AdventureDuration, error) {
	// Load DM persona
	personaLoader := agent.NewPersonaLoader()
	_, dmPersona, err := personaLoader.LoadWithMetadata("dungeon-master")
	if err != nil {
		return "", "", nil, adventure.AdventureDuration{}, fmt.Errorf("failed to load DM persona: %w", err)
	}

	// If the form left the adventure type unset, let the reading decide.
	if adventureType == "" && brief != nil && brief.AdventureType != "" {
		adventureType = brief.AdventureType
	}

	// Get duration config
	dur := adventure.GetDuration(duration)

	// Build session list for JSON example
	actsJSON := s.buildActsJSONTemplate(dur)
	pacingJSON := s.buildPacingJSONTemplate(dur)

	// Build adventure type section
	typeSection := ""
	if advType := adventure.GetAdventureType(adventureType); advType != nil {
		typeSection = fmt.Sprintf("\n%s\n", advType.PromptGuide)
	}

	// Build NPC section based on duration. The role/race value lists are
	// generated from the npc taxonomy enums (single source of truth).
	npcSection := fmt.Sprintf(`
Génère %d à %d PNJ dans le tableau "npcs" de "plot_elements". Chaque PNJ doit avoir :
- "name" : nom complet (en français)
- "role" : l'une de ces valeurs EXACTES : %s
- "race" : l'une de ces valeurs EXACTES : %s
- "gender" : "m" ou "f"
- "occupation" : métier en français, varié (ex. marchand, garde, noble, artisan, aubergiste, prêtre, capitaine de navire, contrebandier, mercenaire, érudit, forgeron, herboriste, batelier, mendiant, espion, chasseur, fermier, scribe, ménestrel, mineur)
- "attitude" : valeur EXACTE "positive", "neutral" ou "negative"
- "motivation" : ce qui anime ce PNJ (en français)
- "secret" : une vérité cachée sur ce PNJ (mondaine, pas surnaturelle, en français)
- "narrative_context" : où et quand les joueurs le rencontrent pour la première fois (en français)
- "narrative_integration" : {"introduction_session": N, "plot_role": "description en français", "linked_to_act": N}

IMPORTANT — DIVERSITÉ : varie les races, genres et occupations des PNJ selon la région, le ton et le rôle de chacun, et ancre chaque métier dans le contexte de cette aventure. N'applique PAS de casting par défaut : évite en particulier les réflexes "informateur = contrebandier halfelin" et "rival = capitaine de la garde". Les exemples de structure ci-dessous ne sont QUE des gabarits de format : n'en recopie ni les races, ni les genres, ni les occupations.

L'antagoniste de plot_elements.antagonist DOIT aussi figurer dans le tableau npcs avec le role "antagoniste".`,
		dur.MaxNPCs-2, dur.MaxNPCs, npc.RoleEnumList(), npc.RaceEnumList())

	// Load world resources for geography context
	worldResources := agent.LoadWorldResources()

	// Build world geography section for the prompt
	worldGeographySection := ""
	if worldResources != nil && worldResources.MapDescription != "" {
		worldGeographySection = fmt.Sprintf(`
## Référence géographique du monde

Utilise cette description géographique des Quatre Royaumes pour situer l'aventure dans des lieux cohérents, avec des distances, routes commerciales et frontières exactes :

%s

`, worldResources.MapDescription)
	}

	// Append the fortune-teller's creative brief + world-coherence guidance.
	// It rides on the same %s placeholder as the geography section, so the
	// format string and argument list below stay untouched.
	if brief != nil {
		worldGeographySection += brief.PromptSection()
	}

	// Build prompt
	prompt := fmt.Sprintf(`Génère un plan de campagne D&D 5e pour cette aventure :

**Nom de l'aventure** : %s
**Description** : %s
**Thème** : %s
**Durée** : %d à %d sessions de 3 heures chacune
**Nombre d'actes** : %d
%s
Crée un plan de campagne comprenant :
1. Un titre de campagne et un objectif captivant
2. %d acte(s) avec titres, descriptions, événements clés et objectifs
3. Un antagoniste principal à la motivation MONDAINE et à l'arc narratif
4. %d à %d lieux clés avec niveaux de danger
5. 0 à %d présages (foreshadows) liés aux actes (seulement si la durée le permet)
6. %d à %d PNJ aux rôles définis

%s

%s

%s
IMPÉRATIF : Réponds UNIQUEMENT par du JSON valide respectant EXACTEMENT cette structure (aucun markdown, aucune explication).
Conserve les CLÉS JSON en anglais (ex. "narrative_structure", "role", "race", "status"...) et les valeurs d'énumération techniques de "status"/"attitude"/"gender" et le "role" de l'antagoniste en anglais ("pending", "primary", "positive", "neutral", "negative", "m", "f"). Les valeurs de "role" et "race" des PNJ sont en FRANÇAIS (voir les listes EXACTES ci-dessus). Rédige en FRANÇAIS tout le reste (titres, descriptions, motivations, etc.) :
{
  "version": "1.0.0",
  "metadata": {
    "campaign_title": "Titre de la campagne",
    "theme": "Description du thème",
    "target_duration": {"sessions": %d, "hours_per_session": 3},
    "created_at": "2026-02-14T12:00:00Z",
    "generated_by": "dungeon-master",
    "last_updated": "2026-02-14T12:00:00Z"
  },
  "narrative_structure": {
    "objective": "Objectif principal de la campagne",
    "hook": "Accroche d'ouverture qui happe les joueurs",
    "acts": %s,
    "climax": {
      "description": "La confrontation décisive",
      "target_session": %d,
      "stakes": "Ce qui est en jeu en cas d'échec des héros"
    },
    "resolution": {
      "success_scenario": "Ce qui se passe si les héros réussissent",
      "failure_scenario": "Ce qui se passe si les héros échouent",
      "epilogue_notes": "Comment l'histoire se conclut"
    }
  },
  "plot_elements": {
    "antagonist": {
      "name": "Nom de l'antagoniste",
      "role": "primary",
      "motivation": "Pourquoi il agit ainsi (motivation MONDAINE)",
      "introduction_session": 1,
      "final_confrontation_session": %d,
      "arc": "Comment il évolue"
    },
    "secondary_antagonists": [],
    "supporting_characters": [],
    "macguffins": [],
    "key_locations": [],
    "npcs": [
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "donneur_de_quete",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "positive",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 1, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      },
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "informateur",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "neutral",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 2, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      },
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "rival",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "negative",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 2, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      }
    ]
  },
  "foreshadows": {
    "active": [],
    "resolved": [],
    "abandoned": [],
    "next_id": 1
  },
  "progression": {
    "current_act": 1,
    "current_session": 0,
    "completed_plot_points": [],
    "active_threads": [],
    "pending_resolutions": []
  },
  "pacing": %s,
  "dm_notes": {
    "themes": ["thème1", "thème2"],
    "tone": "Dark fantasy teintée d'espoir",
    "player_agency": "Notes sur les choix des joueurs",
    "memorable_moments": []
  }
}`,
		adv.Name, adv.Description, theme,
		dur.MinSessions, dur.MaxSessions,
		dur.Acts,
		typeSection,
		dur.Acts,
		dur.MaxLocations-2, dur.MaxLocations,
		dur.MaxForeshadows,
		dur.MaxNPCs-2, dur.MaxNPCs,
		adventure.GetAntiCultConstraints(),
		npcSection,
		worldGeographySection,
		(dur.MinSessions+dur.MaxSessions)/2,
		actsJSON,
		dur.MaxSessions,
		dur.MaxSessions,
		pacingJSON,
	)

	return prompt, dmPersona, worldResources, dur, nil
}

// generateCampaignPlanWithBrief builds the request, calls the model, then parses
// and saves the JSON campaign plan. When brief is non-nil (the fortune-teller
// wizard path) its creative constraints and world-coherence guidance are part of
// the prompt. The output JSON schema is unchanged either way.
func (s *Server) generateCampaignPlanWithBrief(adv *adventure.Adventure, theme, duration, adventureType string, brief *tarot.CreativeBrief) error {
	if s.apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	prompt, dmPersona, worldResources, dur, err := s.buildCampaignPlanRequest(adv, theme, duration, adventureType, brief)
	if err != nil {
		return err
	}

	// Call Anthropic API with Sonnet for better narrative quality
	client := anthropic.NewClient(option.WithAPIKey(s.apiKey))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Build user message content blocks (text + optional map image)
	userContent := []anthropic.ContentBlockParamUnion{
		anthropic.NewTextBlock(prompt),
	}
	if worldResources != nil && worldResources.MapImageBase64 != "" {
		userContent = append(userContent,
			anthropic.NewImageBlockBase64(worldResources.MapImageMediaType, worldResources.MapImageBase64),
		)
	}

	response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 8192,
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: dmPersona,
			},
		},
		Messages: []anthropic.MessageParam{
			{
				Role:    "user",
				Content: userContent,
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

	fmt.Printf("✓ Generated campaign plan for '%s': %s (%s, %d act(s))\n",
		adv.Name, campaignPlan.Metadata.CampaignTitle, duration, dur.Acts)
	return nil
}

// buildActsJSONTemplate returns the acts JSON template adapted to the duration.
func (s *Server) buildActsJSONTemplate(dur adventure.AdventureDuration) string {
	if dur.Acts == 1 {
		sessions := "["
		for i := 1; i <= dur.MaxSessions; i++ {
			if i > 1 {
				sessions += ","
			}
			sessions += fmt.Sprintf("%d", i)
		}
		sessions += "]"
		return fmt.Sprintf(`[
      {
        "number": 1,
        "title": "Titre de l'acte 1",
        "description": "Ce qui se passe dans cet acte",
        "target_sessions": %s,
        "status": "pending",
        "key_events": ["Événement 1", "Événement 2"],
        "goals": ["Objectif 1", "Objectif 2"],
        "completion_criteria": {"milestone": "Ce qui marque la fin de l'acte"}
      }
    ]`, sessions)
	}

	// 3 acts for campaign
	sessPerAct := dur.MaxSessions / 3
	remainder := dur.MaxSessions % 3

	acts := "[\n"
	start := 1
	for a := 1; a <= 3; a++ {
		count := sessPerAct
		if a <= remainder {
			count++
		}
		sessions := "["
		for i := start; i < start+count; i++ {
			if i > start {
				sessions += ","
			}
			sessions += fmt.Sprintf("%d", i)
		}
		sessions += "]"
		if a > 1 {
			acts += ",\n"
		}
		acts += fmt.Sprintf(`      {
        "number": %d,
        "title": "Titre de l'acte %d",
        "description": "Ce qui se passe dans cet acte",
        "target_sessions": %s,
        "status": "pending",
        "key_events": ["Événement 1"],
        "goals": ["Objectif 1"],
        "completion_criteria": {"milestone": "Ce qui marque la fin de l'acte"}
      }`, a, a, sessions)
		start += count
	}
	acts += "\n    ]"
	return acts
}

// buildPacingJSONTemplate returns the pacing JSON template adapted to the duration.
func (s *Server) buildPacingJSONTemplate(dur adventure.AdventureDuration) string {
	avg := (dur.MinSessions + dur.MaxSessions) / 2

	if dur.Acts == 1 {
		return fmt.Sprintf(`{
    "sessions_played": 0,
    "sessions_remaining_estimate": %d,
    "act_breakdown": {
      "act_1": {"planned": %d, "actual": 0, "variance": 0}
    }
  }`, avg, avg)
	}

	sessPerAct := avg / 3
	return fmt.Sprintf(`{
    "sessions_played": 0,
    "sessions_remaining_estimate": %d,
    "act_breakdown": {
      "act_1": {"planned": %d, "actual": 0, "variance": 0},
      "act_2": {"planned": %d, "actual": 0, "variance": 0},
      "act_3": {"planned": %d, "actual": 0, "variance": 0}
    }
  }`, avg, sessPerAct, sessPerAct, avg-2*sessPerAct)
}

// generateAdventureNPCs generates real NPC records from the campaign plan's NPC definitions.
// applyCampaignPlanNPCOverrides reconciles a randomly generated NPC with the
// narrative facts defined in the campaign plan. The random generator picks an
// occupation/motivation/secret unrelated to the planned role; without this the
// saved record contradicts the plan (e.g. an "Inquisiteur" antagonist generated
// as a "caravan escort", or a female ambassador given a beard). Only non-empty
// plan fields override the generated ones, so a sparse plan keeps random flavor.
func applyCampaignPlanNPCOverrides(n *npc.NPC, def adventure.NPCDefinition) {
	if n == nil {
		return
	}
	if def.Occupation != "" {
		n.Occupation = def.Occupation
	}
	if def.Motivation != "" {
		n.Motivation.Goal = def.Motivation
	}
	if def.Secret != "" {
		n.Motivation.Secret = def.Secret
	}
}

func (s *Server) generateAdventureNPCs(adv *adventure.Adventure) error {
	// Load campaign plan
	plan, err := adv.LoadCampaignPlan()
	if err != nil {
		return fmt.Errorf("failed to load campaign plan: %w", err)
	}

	if len(plan.PlotElements.NPCs) == 0 {
		return nil // No NPCs defined in plan
	}

	// Create NPC generator
	gen, err := npc.NewGenerator("data")
	if err != nil {
		return fmt.Errorf("failed to create NPC generator: %w", err)
	}

	// Create NPC manager for this adventure
	adventurePath := filepath.Join(adventuresDir, adv.Slug)
	mgr := npcmanager.NewManager(adventurePath)

	for _, npcDef := range plan.PlotElements.NPCs {
		// Build generation options
		opts := []npc.Option{
			npc.WithName(npcDef.Name),
		}
		if npcDef.Race != "" {
			opts = append(opts, npc.WithRace(npcDef.Race))
		}
		if npcDef.Gender != "" {
			opts = append(opts, npc.WithGender(npcDef.Gender))
		}
		if npcDef.Attitude != "" {
			opts = append(opts, npc.WithAttitude(npcDef.Attitude))
		}

		// Generate the NPC with random traits
		generatedNPC, err := gen.Generate(opts...)
		if err != nil {
			fmt.Printf("Warning: failed to generate NPC '%s': %v\n", npcDef.Name, err)
			continue
		}

		// Reconcile the random profile with the narrative facts from the plan
		// (occupation/motivation/secret) so the saved record matches the role
		// instead of the generator's unrelated random pick.
		applyCampaignPlanNPCOverrides(generatedNPC, npcDef)

		// Determine importance from role
		importance := npcImportanceFromRole(npcDef.Role)

		// Build context string
		npcContext := npcDef.NarrativeContext
		if npcContext == "" {
			npcContext = fmt.Sprintf("Campaign plan: %s", npcDef.Role)
		}

		// Save NPC (session 0 = pre-session / creation time)
		record, err := mgr.AddNPC(0, generatedNPC, npcContext, "")
		if err != nil {
			fmt.Printf("Warning: failed to save NPC '%s': %v\n", npcDef.Name, err)
			continue
		}

		// Set importance (AddNPC defaults to "mentioned", we may need higher)
		if importance != npcmanager.ImportanceMentioned {
			if err := mgr.UpdateImportance(npcDef.Name, importance, "Set from campaign plan role: "+npcDef.Role); err != nil {
				fmt.Printf("Warning: failed to update importance for '%s': %v\n", npcDef.Name, err)
			}
		}

		fmt.Printf("  ✓ NPC generated: %s (%s, %s, importance=%s)\n",
			record.NPC.Name, npcDef.Role, npcDef.Race, importance)
	}

	fmt.Printf("✓ Generated %d NPC(s) for adventure '%s'\n", len(plan.PlotElements.NPCs), adv.Name)
	return nil
}

// npcImportanceFromRole maps a campaign plan NPC role to an importance level.
// The role is normalized (FR canonical, legacy EN tolerated) before matching.
func npcImportanceFromRole(role string) npcmanager.ImportanceLevel {
	switch npc.NormalizeRole(role) {
	case npc.RoleAntagoniste, npc.RoleDonneurDeQuete:
		return npcmanager.ImportanceKey
	case npc.RoleAllie, npc.RoleRival:
		return npcmanager.ImportanceRecurring
	default:
		return npcmanager.ImportanceMentioned
	}
}

// copyGlobalCharactersToAdventure copies ALL existing global characters into the
// new adventure. Backward-compatible wrapper used by the quick-create form.
func (s *Server) copyGlobalCharactersToAdventure(adv *adventure.Adventure) error {
	return s.copyGlobalCharactersToAdventureSelected(adv, nil)
}

// copyGlobalCharactersToAdventureSelected copies the chosen global characters
// into the new adventure and writes party.json from that subset. A nil/empty
// selected slice means "copy all" (matched case-insensitively on the JSON name).
func (s *Server) copyGlobalCharactersToAdventureSelected(adv *adventure.Adventure, selected []string) error {
	globalCharactersDir := filepath.Join("data", "characters")

	selectedSet := make(map[string]bool, len(selected))
	for _, name := range selected {
		if n := strings.TrimSpace(name); n != "" {
			selectedSet[strings.ToLower(n)] = true
		}
	}

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

		name, _ := charData["name"].(string)
		// Skip characters the player did not pick (when a selection was made).
		if len(selectedSet) > 0 && !selectedSet[strings.ToLower(strings.TrimSpace(name))] {
			continue
		}
		if name != "" {
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

// handleAmbientStream streams continuous WAV audio from Lyria to the browser.
// It writes a streaming WAV header followed by raw PCM16 chunks.
func (s *Server) handleAmbientStream(c *gin.Context) {
	slug := c.Param("slug")

	if s.geminiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GEMINI_API_KEY not set"})
		return
	}

	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	mgr := session.GetOrCreateLyriaManager(s.geminiKey)
	if mgr == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lyria not available"})
		return
	}

	// Subscribe to audio stream
	_, audioCh, cancel := mgr.Subscribe()
	defer cancel()

	// Set streaming WAV headers
	c.Header("Content-Type", "audio/wav")
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Content-Type-Options", "nosniff")

	w := c.Writer

	// Write streaming WAV header (RIFF with size=0xFFFFFFFF for live streaming)
	writeStreamingWAVHeader(w)
	w.(http.Flusher).Flush()

	// Stream PCM chunks until client disconnects
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case pcm, ok := <-audioCh:
			if !ok {
				return
			}
			if _, err := w.Write(pcm); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

// handleAmbientSet applies a new scene to the Lyria manager.
func (s *Server) handleAmbientSet(c *gin.Context) {
	slug := c.Param("slug")

	if s.geminiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GEMINI_API_KEY not set"})
		return
	}

	var req struct {
		LyriaPrompt string  `json:"lyria_prompt"`
		BPM         int     `json:"bpm"`
		Temperature float64 `json:"temperature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	mgr := session.GetOrCreateLyriaManager(s.geminiKey)
	if mgr == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lyria not available"})
		return
	}

	bpm := req.BPM
	if bpm == 0 {
		bpm = 100
	}
	temp := req.Temperature
	if temp == 0 {
		temp = 1.0
	}

	if err := mgr.SetScene(req.LyriaPrompt, bpm, temp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleAmbientStop stops Lyria music generation for the session. The audio
// stream ends (subscribers disconnect) and the WebSocket to Lyria is closed.
// A later /ambient/set reconnects a fresh manager, so playback can be resumed.
func (s *Server) handleAmbientStop(c *gin.Context) {
	slug := c.Param("slug")

	session, exists := s.sessionManager.GetSession(slug)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	stopped := session.StopLyriaManager()
	fmt.Printf("[ambient] stop requested for %s (stopped=%v)\n", slug, stopped)
	c.JSON(http.StatusOK, gin.H{"success": true, "stopped": stopped})
}

// writeStreamingWAVHeader writes a WAV header suitable for live/infinite streaming.
// Uses RIFF chunk size 0xFFFFFFFF and data chunk size 0xFFFFFFFF to indicate streaming.
// Format: PCM16, 44100 Hz, stereo (2 channels), 16-bit samples.
func writeStreamingWAVHeader(w io.Writer) {
	const (
		sampleRate   = 44100
		channels     = 2
		bitsPerSample = 16
		byteRate     = sampleRate * channels * bitsPerSample / 8
		blockAlign   = channels * bitsPerSample / 8
		infiniteSize = uint32(0xFFFFFFFF)
	)

	// RIFF chunk
	w.Write([]byte("RIFF"))
	writeUint32LE(w, infiniteSize) // file size - 8 (infinite)
	w.Write([]byte("WAVE"))

	// fmt chunk
	w.Write([]byte("fmt "))
	writeUint32LE(w, 16)            // chunk size
	writeUint16LE(w, 1)             // PCM format
	writeUint16LE(w, channels)
	writeUint32LE(w, sampleRate)
	writeUint32LE(w, byteRate)
	writeUint16LE(w, blockAlign)
	writeUint16LE(w, bitsPerSample)

	// data chunk header
	w.Write([]byte("data"))
	writeUint32LE(w, infiniteSize) // data size (infinite)
}

func writeUint32LE(w io.Writer, v uint32) {
	b := [4]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	w.Write(b[:])
}

func writeUint16LE(w io.Writer, v uint16) {
	b := [2]byte{byte(v), byte(v >> 8)}
	w.Write(b[:])
}
