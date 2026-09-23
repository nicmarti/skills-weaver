package dmtools

import (
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/ambient"
	"dungeons/internal/llm"
)

// SetAmbientMusicTool generates an optimized Lyria music prompt from a scene description.
type SetAmbientMusicTool struct {
	client    llm.Client
	fastModel string
	adv       *adventure.Adventure // canonical state, used to gate combat/danger cues
}

// NewSetAmbientMusicTool creates a new ambient music tool over the shared
// neutral client. adv supplies the canonical combat state used by the
// anti-spoil guardrail; it may be nil (the guardrail then simply lets every
// mood through).
func NewSetAmbientMusicTool(client llm.Client, fastModel string, adv *adventure.Adventure) *SetAmbientMusicTool {
	return &SetAmbientMusicTool{client: client, fastModel: fastModel, adv: adv}
}

// Name returns the tool name.
func (t *SetAmbientMusicTool) Name() string {
	return "set_ambient_music"
}

// Description returns the tool description.
func (t *SetAmbientMusicTool) Description() string {
	return "Generate and set ambient music for the current scene using Google Lyria RealTime. Call this to create atmospheric music that matches the current narrative moment. Returns the Lyria prompt and audio parameters that will be used."
}

// InputSchema returns the JSON schema for tool input.
func (t *SetAmbientMusicTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scene_description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the current scene or desired musical atmosphere (French or English). Examples: 'taverne bruyante et joyeuse', 'combat épique contre des orcs', 'forêt mystérieuse la nuit', 'donjon sombre et silencieux'",
			},
			"mood": map[string]interface{}{
				"type":        "string",
				"description": "Optional mood hint to guide music generation",
				"enum":        []string{"tavern", "combat", "exploration", "mystery", "nature", "danger"},
			},
		},
		"required": []interface{}{"scene_description"},
	}
}

// Execute runs the tool: generates Lyria parameters via Claude Sonnet and returns them.
func (t *SetAmbientMusicTool) Execute(params map[string]interface{}) (interface{}, error) {
	sceneDesc, ok := params["scene_description"].(string)
	if !ok || sceneDesc == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "scene_description is required",
		}, nil
	}

	mood, _ := params["mood"].(string)

	// Anti-spoil guardrail: combat/danger ambiance is gated on the canonical
	// engine combat state. If the DM asks for a combat/danger cue while no combat
	// is active, suppress it — the music must not reveal a hidden threat before
	// the fiction does. Combat music is meant to be triggered through start_combat
	// (at the initiative roll), not through free-form prose here.
	if (mood == "combat" || mood == "danger") && t.adv != nil && !t.adv.IsCombatActive() {
		return map[string]interface{}{
			"success":    true,
			"suppressed": true,
			"note":       "Cue combat/danger ignoré : aucun combat actif. Appelle start_combat au jet d'initiative pour déclencher la musique de combat (anti-spoil).",
			"display":    "🔇 (musique de combat ignorée hors combat)",
		}, nil
	}

	// Add mood hint if provided
	if mood != "" {
		sceneDesc = fmt.Sprintf("%s (mood: %s)", sceneDesc, mood)
	}

	// Generate Lyria parameters through the shared neutral client
	lyriaParams, err := ambient.GenerateLyriaPrompt(t.client, t.fastModel, sceneDesc)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate Lyria prompt: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success":      true,
		"lyria_prompt": lyriaParams.Prompt,
		"bpm":          lyriaParams.BPM,
		"temperature":  lyriaParams.Temperature,
		"display":      fmt.Sprintf("🎵 Ambient music: %s", lyriaParams.DisplayName),
		"scene_name":   lyriaParams.DisplayName,
	}, nil
}
