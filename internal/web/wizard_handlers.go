package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/image"
	"dungeons/internal/tarot"
)

// wizardCharacter is the trimmed view of a global hero shown in the selector.
type wizardCharacter struct {
	Name        string
	Class       string
	Species     string
	Level       int
	HasPortrait bool
	PortraitURL string
}

// listWizardCharacters returns the global heroes available for selection.
func listWizardCharacters() []wizardCharacter {
	dir := filepath.Join("data", "characters")
	chars, err := character.ListCharacters(dir)
	if err != nil {
		return nil
	}
	out := make([]wizardCharacter, 0, len(chars))
	for _, ch := range chars {
		wc := wizardCharacter{
			Name:    ch.Name,
			Class:   ch.Class,
			Species: ch.Species,
			Level:   ch.Level,
		}
		slug := character.SanitizeFilename(ch.Name)
		for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp"} {
			if _, err := os.Stat(filepath.Join(dir, slug+ext)); err == nil {
				wc.HasPortrait = true
				wc.PortraitURL = "/characters/images/" + slug + ext
				break
			}
		}
		out = append(out, wc)
	}
	return out
}

// handleAdventureWizard renders the guided "fortune teller" creation page.
func (s *Server) handleAdventureWizard(c *gin.Context) {
	deck, err := tarot.LoadDeck(tarot.DefaultDeckPath)
	if err != nil {
		s.renderError(c, http.StatusInternalServerError, fmt.Sprintf("Impossible de charger le jeu de cartes: %v", err))
		return
	}
	deckJSON, err := json.Marshal(deck)
	if err != nil {
		s.renderError(c, http.StatusInternalServerError, "Impossible de sérialiser le jeu de cartes")
		return
	}

	c.HTML(http.StatusOK, "adventure_wizard.html", gin.H{
		"Title":      "La Voyante",
		"DeckJSON":   string(deckJSON),
		"Characters": listWizardCharacters(),
	})
}

// suggestNamePayload is the JSON body for the name-suggestion endpoint.
type suggestNamePayload struct {
	CardIDs []string `json:"card_ids"`
	Theme   string   `json:"theme"`
}

// handleSuggestAdventureName proposes an evocative adventure title coherent with
// the card reading. It is best-effort: any failure (no API key, LLM error)
// returns an empty name so the client simply leaves the field blank/editable.
func (s *Server) handleSuggestAdventureName(c *gin.Context) {
	var p suggestNamePayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload invalide"})
		return
	}
	if s.apiKey == "" {
		c.JSON(http.StatusOK, gin.H{"name": ""})
		return
	}
	deck, err := tarot.LoadDeck(tarot.DefaultDeckPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"name": ""})
		return
	}
	brief, _ := deck.Aggregate(p.CardIDs) // partial brief is fine for a title

	name, err := s.generateAdventureName(brief, strings.TrimSpace(p.Theme))
	if err != nil {
		fmt.Printf("Warning: adventure name suggestion failed: %v\n", err)
		c.JSON(http.StatusOK, gin.H{"name": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": name})
}

// generateAdventureName asks a small, fast model for one evocative French title
// reflecting the reading's mood — without spoiling any twist.
func (s *Server) generateAdventureName(brief tarot.CreativeBrief, theme string) (string, error) {
	var b strings.Builder
	b.WriteString("Propose UN seul titre d'aventure de jeu de rôle médiéval-fantastique, en français.\n")
	b.WriteString("Contraintes: 2 à 6 mots, évocateur et mystérieux, SANS guillemets, SANS sous-titre, SANS deux-points, SANS ponctuation finale.\n")
	b.WriteString("Ne révèle aucun rebondissement. Inspire-toi de l'ambiance ci-dessous sans la citer littéralement:\n")
	if brief.Tone != "" {
		fmt.Fprintf(&b, "- Ton: %s\n", brief.Tone)
	}
	if brief.AntagonistNature != "" {
		fmt.Fprintf(&b, "- Menace: %s\n", brief.AntagonistNature)
	}
	if brief.Region != "" {
		fmt.Fprintf(&b, "- Région: %s\n", brief.Region)
	}
	if brief.DangerLevel != "" {
		fmt.Fprintf(&b, "- Danger: %s\n", brief.DangerLevel)
	}
	if brief.Stakes != "" {
		fmt.Fprintf(&b, "- Enjeu: %s\n", brief.Stakes)
	}
	if theme != "" {
		fmt.Fprintf(&b, "- Intuition du joueur: %s\n", theme)
	}
	b.WriteString("\nRéponds UNIQUEMENT par le titre, rien d'autre.")

	client := anthropic.NewClient(option.WithAPIKey(s.apiKey))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 64,
		System: []anthropic.TextBlockParam{
			{Type: "text", Text: "Tu crées des titres d'aventures D&D évocateurs. Tu réponds toujours par un seul titre, sans guillemets ni ponctuation superflue."},
		},
		Messages: []anthropic.MessageParam{
			{Role: "user", Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(b.String())}},
		},
	})
	if err != nil {
		return "", err
	}

	var txt string
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			txt += tb.Text
		}
	}
	return sanitizeAdventureName(txt), nil
}

// sanitizeAdventureName keeps the first line and strips quotes/punctuation noise.
func sanitizeAdventureName(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	s = strings.Trim(s, " \t\"'«».")
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > 80 {
		s = strings.TrimSpace(string(r[:80]))
	}
	return s
}

// wizardCreatePayload is the JSON body POSTed by wizard.js on completion.
type wizardCreatePayload struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Theme         string   `json:"theme"`
	Duration      string   `json:"duration"`
	AdventureType string   `json:"adventure_type"`
	CardIDs       []string `json:"card_ids"`
	Characters    []string `json:"characters"`
	Debug         bool     `json:"debug"` // when true: return the reading, create nothing
}

// handleCreateAdventureWizard creates an adventure from the wizard's reading.
// It runs the campaign-plan generation synchronously (the client shows a
// spinner) and then kicks off the unique "destiny card" image asynchronously.
func (s *Server) handleCreateAdventureWizard(c *gin.Context) {
	var p wizardCreatePayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Payload invalide: %v", err)})
		return
	}

	p.Name = strings.TrimSpace(p.Name)
	if !p.Debug && p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le nom de l'aventure est requis"})
		return
	}

	deck, err := tarot.LoadDeck(tarot.DefaultDeckPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Jeu de cartes indisponible"})
		return
	}

	// One card per question, each a real card.
	if len(p.CardIDs) != len(deck.Questions) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Il faut choisir %d cartes (reçu %d)", len(deck.Questions), len(p.CardIDs)),
		})
		return
	}
	for _, id := range p.CardIDs {
		if _, ok := deck.CardByID(id); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Carte inconnue: %s", id)})
			return
		}
	}

	brief, err := deck.Aggregate(p.CardIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Attach world-coherence guidance (recent adventures + NPC names to avoid).
	coherence := s.buildWizardCoherence()
	brief.CoherenceText = coherence.PromptSection()

	// Debug mode: return the full reading (cards -> hidden brief -> prompt
	// fragment) WITHOUT creating an adventure or calling any API.
	if p.Debug {
		s.respondWizardDebug(c, deck, &p, brief, coherence)
		return
	}

	// Create the adventure.
	adv := adventure.New(p.Name, strings.TrimSpace(p.Description))
	if err := adv.Save(adventuresDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Création impossible: %v", err)})
		return
	}

	// Generate the campaign plan (blocking) when an API key is configured.
	if s.apiKey != "" {
		theme := strings.TrimSpace(p.Theme)
		if err := s.generateCampaignPlanWithBrief(adv, theme, p.Duration, p.AdventureType, &brief); err != nil {
			fmt.Printf("Warning: wizard campaign plan generation failed: %v\n", err)
		} else {
			if err := s.generateAdventureNPCs(adv); err != nil {
				fmt.Printf("Warning: wizard NPC generation failed: %v\n", err)
			}
			if plan, err := adv.LoadCampaignPlan(); err == nil {
				result := adventure.ValidateCampaignPlan(plan)
				if len(result.Errors) > 0 || len(result.Warnings) > 0 {
					fmt.Printf("Wizard campaign plan validation for '%s': score=%d, errors=%d, warnings=%d\n",
						adv.Name, result.Score, len(result.Errors), len(result.Warnings))
				}
				// Reveal the destiny card in the background.
				go s.generateDestinyCard(adv, plan.Metadata.CampaignTitle, brief.Tone)
			}
		}
	}

	// Copy only the selected heroes (nil/empty => all).
	if err := s.copyGlobalCharactersToAdventureSelected(adv, p.Characters); err != nil {
		fmt.Printf("Warning: wizard failed to copy characters: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"slug":     adv.Slug,
		"redirect": "/play/" + adv.Slug,
	})
}

// respondWizardDebug returns the deterministic reading produced by the chosen
// cards — the per-card hidden params, the aggregated brief, the world-coherence
// context and the exact prompt fragment that WOULD be sent to the campaign-plan
// generator. Nothing is persisted and no API is called.
func (s *Server) respondWizardDebug(c *gin.Context, deck *tarot.Deck, p *wizardCreatePayload, brief tarot.CreativeBrief, coherence WizardCoherence) {
	type chosenCard struct {
		Question string             `json:"question"`
		Prompt   string             `json:"prompt"`
		ID       string             `json:"id"`
		Name     string             `json:"name"`
		Flavor   string             `json:"flavor"`
		Params   tarot.HiddenParams `json:"params"`
	}

	chosen := make([]chosenCard, 0, len(p.CardIDs))
	for _, id := range p.CardIDs {
		card, ok := deck.CardByID(id)
		if !ok {
			continue
		}
		prompt := ""
		if q, ok := deck.QuestionByID(card.Question); ok {
			prompt = q.Prompt
		}
		chosen = append(chosen, chosenCard{
			Question: card.Question,
			Prompt:   prompt,
			ID:       card.ID,
			Name:     card.Name,
			Flavor:   card.Flavor,
			Params:   card.Params,
		})
	}

	recent := make([]string, 0, len(coherence.RecentAdventures))
	for _, a := range coherence.RecentAdventures {
		recent = append(recent, a.Name)
	}

	// Build the FULL prompt that would be sent to the campaign-plan generator,
	// so the reader can see exactly how the brief shapes the multi-act request.
	// Uses a throwaway (unsaved) adventure — nothing is persisted.
	campaignPrompt := ""
	previewName := p.Name
	if previewName == "" {
		previewName = "Aventure (aperçu)"
	}
	tmpAdv := adventure.New(previewName, strings.TrimSpace(p.Description))
	if pr, _, _, _, err := s.buildCampaignPlanRequest(tmpAdv, strings.TrimSpace(p.Theme), p.Duration, p.AdventureType, &brief); err == nil {
		campaignPrompt = pr
	}

	c.JSON(http.StatusOK, gin.H{
		"debug":               true,
		"name":                p.Name,
		"duration":            p.Duration,
		"selected_characters": p.Characters,
		"chosen_cards":        chosen,
		"brief": gin.H{
			"tone":              brief.Tone,
			"adventure_type":    brief.AdventureType,
			"antagonist_nature": brief.AntagonistNature,
			"region":            brief.Region,
			"danger_level":      brief.DangerLevel,
			"stakes":            brief.Stakes,
			"pacing":            brief.Pacing,
		},
		"coherence": gin.H{
			"recent_adventures":  recent,
			"npc_names_to_avoid": coherence.NPCNamesToAvoid,
		},
		"prompt_section":  brief.PromptSection(),
		"campaign_prompt": campaignPrompt,
	})
}

// destinyCardPath is the stable on-disk location polled by the client.
func destinyCardPath(slug string) string {
	return filepath.Join(adventuresDir, slug, "images", "session-0", "destiny.png")
}

// generateDestinyCard renders a single unique tarot-style card illustrating the
// freshly created adventure. Best-effort: failures never affect creation.
func (s *Server) generateDestinyCard(adv *adventure.Adventure, title, tone string) {
	outDir := filepath.Join(adventuresDir, adv.Slug, "images", "session-0")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Warning: destiny card dir: %v\n", err)
		return
	}
	gen, err := image.NewGeneratorAuto(outDir)
	if err != nil {
		// No image provider configured — silently skip.
		return
	}
	subject := title
	if subject == "" {
		subject = adv.Name
	}
	prompt := fmt.Sprintf(
		"A single ornate tarot 'card of destiny' illustrating the adventure '%s'. Mood/tone: %s. "+
			"Art-nouveau gold filigree border, mystical symbolism, vertical card composition, painted fantasy style, %s",
		subject, tone, image.BasePromptSuffix)

	result, err := gen.Generate(prompt,
		image.WithModelInstance(image.ModelSchnell),
		image.WithImageSize("portrait_4_3"),
		image.WithFilenamePrefix("destiny_raw"),
		image.WithOutputFormat("png"),
	)
	if err != nil {
		fmt.Printf("Warning: destiny card generation failed: %v\n", err)
		return
	}
	// Normalize to the stable filename the client polls for.
	dest := destinyCardPath(adv.Slug)
	if result.LocalPath != dest {
		if err := os.Rename(result.LocalPath, dest); err != nil {
			fmt.Printf("Warning: destiny card rename failed: %v\n", err)
		}
	}
}

// handleDestinyCard reports whether the destiny card is ready (200 + url) or not
// yet (404). The client polls this after the wizard POST returns.
func (s *Server) handleDestinyCard(c *gin.Context) {
	slug := c.Param("slug")
	// Validate the slug resolves to a real adventure before touching the
	// filesystem (avoids probing paths like ".." outside the adventures dir).
	if _, err := adventure.LoadByName(adventuresDir, slug); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ready": false})
		return
	}
	if _, err := os.Stat(destinyCardPath(slug)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ready": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ready": true,
		"url":   fmt.Sprintf("/play/%s/images/session-0/destiny.png", slug),
	})
}
