package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/map"
	"dungeons/internal/world"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// EnrichmentResult holds generated descriptions.
type EnrichmentResult struct {
	Description   string `json:"description"`
	DescriptionFr string `json:"description_fr"`
}

// MapPromptRequest configures map prompt generation.
type MapPromptRequest struct {
	MapType      string   // "city", "region", "dungeon", "tactical"
	LocationName string   // Name of location (for city/region)
	Kingdom      string   // Kingdom ID (for validation)
	Scale        string   // "small", "medium", "large"
	Features     []string // Additional POIs to include
	Terrain      string   // Terrain type override
	Style        string   // "illustrated" or "dark_fantasy"
	DungeonLevel int      // For dungeon maps
	SceneDesc    string   // For tactical maps
}

// MapPromptResult holds the enriched map prompt.
type MapPromptResult struct {
	Prompt       string   `json:"prompt"`
	MapType      string   `json:"map_type"`
	LocationName string   `json:"location_name"`
	Kingdom      string   `json:"kingdom"`
	Features     []string `json:"features"`
	StyleHints   string   `json:"style_hints"`
	EnrichedAt   string   `json:"enriched_at"`
}

// Enricher generates descriptions using Claude API.
type Enricher struct {
	apiKey string
	model  string
}

// NewEnricher creates an enricher with Claude API.
func NewEnricher() (*Enricher, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	return &Enricher{
		apiKey: apiKey,
		model:  "claude-haiku-4-5-20251001", // Latest Haiku with improved capabilities
	}, nil
}

// EnrichEntry generates bilingual descriptions for a single journal entry.
func (e *Enricher) EnrichEntry(entry adventure.JournalEntry, ctx *adventure.EnrichmentContext) (*EnrichmentResult, error) {
	prompt := e.buildPrompt(entry, ctx)

	response, err := e.callClaude(prompt)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Strip markdown code fences if present
	jsonStr := stripMarkdownFences(response)

	var result EnrichmentResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parsing Claude response: %w\nResponse: %s", err, response)
	}

	// Validate results
	if result.Description == "" || result.DescriptionFr == "" {
		return nil, fmt.Errorf("incomplete descriptions returned (EN: %d words, FR: %d words)",
			len(strings.Fields(result.Description)),
			len(strings.Fields(result.DescriptionFr)))
	}

	return &result, nil
}

// EnrichMapPrompt generates an enriched map prompt with Claude API.
func (e *Enricher) EnrichMapPrompt(req MapPromptRequest, location *world.Location, kingdom *world.Kingdom) (*MapPromptResult, error) {
	// Build base prompt using internal/map builders
	basePrompt, err := e.buildBaseMapPrompt(req, location, kingdom)
	if err != nil {
		return nil, fmt.Errorf("building base prompt: %w", err)
	}

	// Build enrichment prompt with guidelines
	enrichPrompt := e.buildMapEnrichmentPrompt(req, basePrompt, location, kingdom)

	// Call Claude API
	response, err := e.callClaude(enrichPrompt)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Strip markdown fences if present
	jsonStr := stripMarkdownFences(response)

	// Parse response - expect just the prompt text or JSON with prompt field
	var result MapPromptResult

	// Try parsing as JSON first
	var jsonResponse struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &jsonResponse); err == nil && jsonResponse.Prompt != "" {
		result.Prompt = jsonResponse.Prompt
	} else {
		// If not JSON, use the response directly as the prompt
		result.Prompt = jsonStr
	}

	// Validate prompt
	if result.Prompt == "" {
		return nil, fmt.Errorf("empty prompt returned from Claude API")
	}

	wordCount := len(strings.Fields(result.Prompt))
	if wordCount < 80 {
		return nil, fmt.Errorf("prompt too short (%d words, minimum 80)", wordCount)
	}
	if wordCount > 250 {
		return nil, fmt.Errorf("prompt too long (%d words, maximum 250)", wordCount)
	}

	// Fill in metadata
	result.MapType = req.MapType
	result.LocationName = req.LocationName
	if kingdom != nil {
		result.Kingdom = kingdom.ID
	}
	result.Features = req.Features
	result.StyleHints = e.getStyleHints(req, kingdom)
	result.EnrichedAt = time.Now().Format(time.RFC3339)

	return &result, nil
}

// buildBaseMapPrompt constructs the base prompt using internal/map builders.
func (e *Enricher) buildBaseMapPrompt(req MapPromptRequest, location *world.Location, kingdom *world.Kingdom) (string, error) {
	ctx := &mapgen.MapContext{
		Location: location,
		Kingdom:  kingdom,
		Scale:    req.Scale,
	}

	opts := mapgen.PromptOptions{
		Features: req.Features,
		Terrain:  req.Terrain,
		Style:    req.Style,
	}

	switch req.MapType {
	case "city":
		if location == nil {
			return "", fmt.Errorf("location required for city map")
		}
		return mapgen.BuildCityMapPrompt(ctx, opts), nil

	case "region":
		if ctx.Region == nil && location != nil {
			// If we have a location but no region, we can still generate
			// Create a minimal region from location data
			ctx.Region = &world.Region{
				Name:        location.Name + " Region",
				Kingdom:     location.Kingdom,
				Description: location.Description,
				Cities:      []world.Location{*location},
			}
		}
		return mapgen.BuildRegionalMapPrompt(ctx, opts), nil

	case "dungeon":
		if req.LocationName == "" {
			return "", fmt.Errorf("location name required for dungeon map")
		}
		return mapgen.BuildDungeonMapPrompt(req.LocationName, req.DungeonLevel, opts), nil

	case "tactical":
		terrain := req.Terrain
		if terrain == "" {
			terrain = "terrain varié"
		}
		return mapgen.BuildTacticalMapPrompt(terrain, req.SceneDesc, opts), nil

	default:
		return "", fmt.Errorf("unknown map type: %s (valid: city, region, dungeon, tactical)", req.MapType)
	}
}

// buildMapEnrichmentPrompt constructs the full enrichment prompt.
func (e *Enricher) buildMapEnrichmentPrompt(req MapPromptRequest, basePrompt string, location *world.Location, kingdom *world.Kingdom) string {
	// Load guidelines from file
	guidelines, err := os.ReadFile("ai/map_prompt_guidelines.md")
	if err != nil {
		guidelines = []byte("Error loading guidelines. Use best practices for map prompts.")
	}

	// Build context information
	locationInfo := "No specific location"
	if location != nil {
		locationInfo = fmt.Sprintf("%s (%s, %s)", location.Name, location.Type, location.Kingdom)
		if location.Description != "" {
			locationInfo += "\nDescription: " + location.Description
		}
		if len(location.KeyLocations) > 0 {
			locationInfo += "\nPOIs: " + strings.Join(location.KeyLocations, ", ")
		}
	}

	kingdomInfo := "No kingdom context"
	if kingdom != nil {
		kingdomInfo = fmt.Sprintf("%s - Colors: %s", kingdom.Name, strings.Join(kingdom.Colors, ", "))
		kingdomInfo += "\nSymbol: " + kingdom.Symbol
		if len(kingdom.Values) > 0 {
			kingdomInfo += "\nValues: " + strings.Join(kingdom.Values, ", ")
		}
	}

	return fmt.Sprintf(`You are enriching a fantasy map prompt for image generation using fal.ai flux-2.

MAP TYPE: %s
LOCATION: %s
KINGDOM: %s
SCALE: %s
STYLE: %s

BASE PROMPT (to be enriched):
%s

GUIDELINES (full documentation):
%s

TASK:
Enrich the base prompt to create a detailed, vivid French description (100-200 words) suitable for generating a 2D fantasy map image. Follow the template structures and best practices from the guidelines.

REQUIREMENTS:
1. Output in French only
2. Length: 100-200 words (sweet spot: 150 words)
3. Include all elements from base prompt
4. Add specific visual details (colors, architecture, layout)
5. Maintain geographic precision (cardinal directions)
6. Use proper French typography (spaces before : and ;)
7. Make layout organic and natural (not monotone)
8. Integrate POIs naturally into the description
9. Match kingdom architectural style
10. Include scale/size information

OUTPUT FORMAT:
Return ONLY the enriched French prompt as plain text (no JSON, no markdown).
The prompt should be a single flowing paragraph with proper punctuation.`,
		req.MapType,
		locationInfo,
		kingdomInfo,
		req.Scale,
		req.Style,
		basePrompt,
		string(guidelines),
	)
}

// getStyleHints returns style hints based on request and kingdom.
func (e *Enricher) getStyleHints(req MapPromptRequest, kingdom *world.Kingdom) string {
	hints := []string{}

	// Map type hints
	switch req.MapType {
	case "city":
		hints = append(hints, "aerial view, organic layout, detailed districts")
	case "region":
		hints = append(hints, "bird's eye view, cartographic style, medieval fantasy map")
	case "dungeon":
		hints = append(hints, "top-down floor plan, grid overlay, D&D classic style")
	case "tactical":
		hints = append(hints, "combat grid, tactical elements, miniature-friendly")
	}

	// Kingdom style hints
	if kingdom != nil {
		switch strings.ToLower(kingdom.ID) {
		case "valdorine":
			hints = append(hints, "maritime Italian style, blue/gold colors")
		case "karvath":
			hints = append(hints, "militaristic Germanic style, red/black colors")
		case "lumenciel":
			hints = append(hints, "religious Latin style, white/gold colors")
		case "astrene":
			hints = append(hints, "melancholic Nordic style, gray/silver colors")
		}
	}

	// General style hints
	if req.Style == "dark_fantasy" {
		hints = append(hints, "dark atmosphere, dramatic lighting")
	} else {
		hints = append(hints, "illustrated style, vibrant colors")
	}

	return strings.Join(hints, ", ")
}

// stripMarkdownFences removes markdown code fences from JSON responses.
func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)

	// Remove ```json or ``` prefix
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}

	// Remove ``` suffix
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}

	return strings.TrimSpace(s)
}

// buildPrompt constructs the enrichment prompt using guidelines.
func (e *Enricher) buildPrompt(entry adventure.JournalEntry, ctx *adventure.EnrichmentContext) string {
	guidelines, err := os.ReadFile("ai/journal_description_guidelines.md")
	if err != nil {
		// Fallback to basic guidelines if file not found
		guidelines = []byte(fallbackGuidelines)
	}

	partyInfo := "Unknown party"
	if len(ctx.PartyMembers) > 0 {
		partyInfo = strings.Join(ctx.PartyMembers, ", ")
	}

	contextInfo := "No previous context"
	if len(ctx.RecentEntries) > 0 {
		contextInfo = strings.Join(ctx.RecentEntries, " → ")
	}

	sessionInfo := "Outside session"
	if ctx.SessionInfo != "" {
		sessionInfo = ctx.SessionInfo
	}

	return fmt.Sprintf(`You are enriching a Basic Fantasy RPG journal entry for AI image generation.

ADVENTURE: %s
PARTY: %s
SESSION: %s

RECENT CONTEXT:
%s

ENTRY TO ENRICH:
ID: %d
Type: %s
Content: "%s"

GUIDELINES (extract):
%s

TASK:
Generate TWO vivid, visual descriptions (English + French) following the template:
[Characters] + [Location] + [Action] + [Atmosphere] + [Visual Details]

RULES:
1. Third-person present tense
2. Visual, cinematic language (describe what you'd see)
3. 30-50 words per description (2-3 sentences)
4. Include character names when relevant
5. Describe lighting, mood, environment
6. Type-specific focus: %s
7. NO game mechanics, NO exact numbers
8. Show don't tell (e.g., "shadows dance" not "it's dark")

OUTPUT FORMAT (JSON only, no other text):
{
  "description": "English description here (30-50 words)",
  "description_fr": "Description française ici (30-50 mots)"
}`,
		ctx.AdventureName,
		partyInfo,
		sessionInfo,
		contextInfo,
		entry.ID,
		entry.Type,
		entry.Content,
		string(guidelines),
		getTypeFocus(entry.Type),
	)
}

// getTypeFocus returns type-specific guidance.
func getTypeFocus(entryType string) string {
	focus := map[string]string{
		"combat":      "Action and tension, dynamic battle scene with weapons and movement",
		"exploration": "Environment and discovery, exploration with lighting details",
		"discovery":   "Moment of revelation, surprising find with magical or revealing effects",
		"loot":        "Treasure itself, glittering gold and magical items with visual appeal",
		"note":        "Character interaction or observation, narrative moment with emotion",
		"session":     "Overall party mood and achievement, triumphant or somber atmosphere",
		"rest":        "Party recovering, campfire or inn scene with peaceful atmosphere",
		"npc":         "Character introduction, distinctive appearance and setting",
		"location":    "Place description, architectural details and great lighting",
		"quest":       "Mission objective, sense of purpose and destination",
	}

	if f, ok := focus[entryType]; ok {
		return f
	}
	return "General fantasy scene with atmospheric details and visual interest"
}

// callClaude sends the prompt to Claude API.
func (e *Enricher) callClaude(prompt string) (string, error) {
	client := anthropic.NewClient(
		option.WithAPIKey(e.apiKey),
	)

	response, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model: anthropic.Model(e.model),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		MaxTokens:   500,
		Temperature: anthropic.Float(0.7),
	})

	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return response.Content[0].Text, nil
}

// fallbackGuidelines provides basic guidelines if file is missing.
const fallbackGuidelines = `
## Template Structure
[Characters] + [Location] + [Action] + [Atmosphere] + [Visual Details]

## Length
- Target: 30-50 words (2-3 sentences)
- Minimum: 15 words
- Maximum: 80 words

## Best Practices
1. Be Specific: "torch-lit corridor" > "dark place"
2. Use Names: "Aldric swings his sword" > "the fighter attacks"
3. Show Don't Tell: "shadows dance on walls" > "it's dark"
4. Present Tense: "Aldric fights" > "Aldric fought"
5. Third Person: "The party enters" > "We enter"
6. Include Lighting: torch-lit, moonlit, flickering candlelight
7. Visual Language: glinting, crumbling, weathered, gleaming
`
