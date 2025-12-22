package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dungeons/internal/adventure"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// EnrichmentResult holds generated descriptions.
type EnrichmentResult struct {
	Description   string `json:"description"`
	DescriptionFr string `json:"description_fr"`
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
		"expense":     "Transaction or town scene, bustling marketplace or shop interior",
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
