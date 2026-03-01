package dmtools

import (
	"fmt"

	"dungeons/internal/ambient"
)

// SetAmbientMusicTool generates an optimized Lyria music prompt from a scene description.
type SetAmbientMusicTool struct {
	anthropicKey string
}

// NewSetAmbientMusicTool creates a new ambient music tool.
func NewSetAmbientMusicTool(anthropicKey string) *SetAmbientMusicTool {
	return &SetAmbientMusicTool{anthropicKey: anthropicKey}
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

// Execute runs the tool: generates Lyria parameters via Claude Haiku and returns them.
func (t *SetAmbientMusicTool) Execute(params map[string]interface{}) (interface{}, error) {
	sceneDesc, ok := params["scene_description"].(string)
	if !ok || sceneDesc == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "scene_description is required",
		}, nil
	}

	// Add mood hint if provided
	if mood, ok := params["mood"].(string); ok && mood != "" {
		sceneDesc = fmt.Sprintf("%s (mood: %s)", sceneDesc, mood)
	}

	// Generate Lyria parameters via Claude Haiku
	lyriaParams, err := ambient.GenerateLyriaPrompt(t.anthropicKey, sceneDesc)
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
