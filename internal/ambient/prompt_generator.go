package ambient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// LyriaSceneParams contains the generated parameters for a Lyria music scene.
type LyriaSceneParams struct {
	Prompt      string  // English prompt optimized for Lyria
	BPM         int     // 50-160
	Temperature float64 // 0.8-1.3
	DisplayName string  // Human-readable scene name
}

const lyriaSystemPrompt = `You are an expert in ambient RPG music generation. Given a scene description, generate parameters for Google Lyria RealTime music generation.

Return ONLY a JSON object with these exact fields:
{
  "prompt_en": "english music description optimized for Lyria (instruments, atmosphere, style, NO vocals)",
  "bpm": <integer 50-160>,
  "temperature": <float 0.8-1.3>,
  "scene_name": "short readable scene name in French"
}

Guidelines for the Lyria prompt:
- Write in English, 10-20 words
- Focus on: medieval instruments (lute, flute, drums, strings, horn), atmosphere, and RPG style
- Do NOT include: lyrics, vocals, specific artist names
- Examples:
  * Tavern: "lively medieval tavern folk music, lutes and flutes, cheerful festive atmosphere"
  * Combat: "epic battle orchestra, fast drums, heroic strings, intense combat fantasy"
  * Dungeon: "dark dungeon ambient, tense strings, mysterious atmosphere, low drones"
  * Forest: "peaceful enchanted forest, gentle flutes, nature ambient, soft adventure"
  * Mystery: "mysterious chamber music, harpsichord, tension, dark medieval RPG"

BPM guidelines:
- Calm/exploration: 55-80
- Tavern/market: 100-125
- Combat/danger: 130-155
- Mystery/dungeon: 60-75

Temperature guidelines:
- Predictable/calm: 0.8-0.95
- Normal: 1.0
- Creative/chaotic: 1.1-1.3`

// GenerateLyriaPrompt calls Claude Haiku to generate optimized Lyria parameters from a scene description.
func GenerateLyriaPrompt(apiKey, sceneDescription string) (*LyriaSceneParams, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	userPrompt := fmt.Sprintf("Generate Lyria music parameters for this D&D scene: %s", sceneDescription)

	response, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model: anthropic.ModelClaudeHaiku4_5,
		System: []anthropic.TextBlockParam{
			{Text: lyriaSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
		MaxTokens:   300,
		Temperature: anthropic.Float(0.3),
	})
	if err != nil {
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	rawText := response.Content[0].Text

	// Extract JSON from the response (handle markdown code blocks if present)
	jsonStr := extractJSON(rawText)

	var result struct {
		PromptEN    string  `json:"prompt_en"`
		BPM         int     `json:"bpm"`
		Temperature float64 `json:"temperature"`
		SceneName   string  `json:"scene_name"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Lyria params JSON: %w (raw: %s)", err, rawText)
	}

	// Validate and clamp values
	if result.BPM < 50 {
		result.BPM = 50
	}
	if result.BPM > 160 {
		result.BPM = 160
	}
	if result.Temperature < 0.8 {
		result.Temperature = 0.8
	}
	if result.Temperature > 1.3 {
		result.Temperature = 1.3
	}
	if result.PromptEN == "" {
		result.PromptEN = "medieval fantasy ambient music, atmospheric RPG"
	}
	if result.SceneName == "" {
		result.SceneName = sceneDescription
	}

	return &LyriaSceneParams{
		Prompt:      result.PromptEN,
		BPM:         result.BPM,
		Temperature: result.Temperature,
		DisplayName: result.SceneName,
	}, nil
}

// extractJSON extracts a JSON object from text that may contain markdown code blocks.
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	// Strip markdown code block if present
	if idx := strings.Index(text, "```json"); idx >= 0 {
		text = text[idx+7:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	} else if idx := strings.Index(text, "```"); idx >= 0 {
		text = text[idx+3:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	}

	// Find first { and last }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}

	return strings.TrimSpace(text)
}
