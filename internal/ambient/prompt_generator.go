package ambient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dungeons/internal/llm"
)

// LyriaSceneParams contains the generated parameters for a Lyria music scene.
type LyriaSceneParams struct {
	Prompt      string  // English prompt optimized for Lyria
	BPM         int     // 50-160
	Temperature float64 // 0.8-1.3
	DisplayName string  // Human-readable scene name
}

// lyriaTimeout bounds the parameter-generation call. The Google Lyria music
// transport itself is unchanged and has its own lifecycle.
const lyriaTimeout = 120 * time.Second

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

// GenerateLyriaPrompt generates optimized Lyria parameters from a scene
// description through the shared neutral client. Only the parameter
// generation is migrated; the Google Lyria music transport is unchanged.
func GenerateLyriaPrompt(client llm.Client, fastModel, sceneDescription string) (*LyriaSceneParams, error) {
	if client == nil {
		return nil, fmt.Errorf("LLM client not available for Lyria prompt generation")
	}
	if fastModel == "" {
		fastModel = llm.DefaultModelFast
	}

	userPrompt := fmt.Sprintf("Generate Lyria music parameters for this D&D scene: %s", sceneDescription)

	ctx, cancel := context.WithTimeout(context.Background(), lyriaTimeout)
	defer cancel()

	resp, err := client.Complete(ctx, llm.Request{
		Model:    fastModel,
		System:   lyriaSystemPrompt,
		Messages: []llm.Message{llm.UserMessage(userPrompt)},
		// Structured JSON output; keep the tolerant parsing below as a
		// fallback for endpoints without response-format support.
		MaxCompletionTokens: 300,
		JSONResponse: &llm.JSONResponseFormat{
			Name: "lyria_scene_params",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt_en":   map[string]interface{}{"type": "string"},
					"bpm":         map[string]interface{}{"type": "integer"},
					"temperature": map[string]interface{}{"type": "number"},
					"scene_name":  map[string]interface{}{"type": "string"},
				},
				"required": []string{"prompt_en", "bpm", "temperature", "scene_name"},
			},
		},
		RequireParameters: true,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	if resp.Message.Text == "" {
		return nil, fmt.Errorf("empty response from LLM")
	}

	rawText := resp.Message.Text

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
