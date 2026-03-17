package dmtools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/image"
)

// GenerateImageTool generates fantasy images from prompts.
type GenerateImageTool struct {
	adventure *adventure.Adventure
}

// NewGenerateImageTool creates a new image generation tool.
func NewGenerateImageTool(adv *adventure.Adventure) (*GenerateImageTool, error) {
	// Verify at least one image API key is configured
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("FAL_KEY") == "" {
		return nil, fmt.Errorf("no image API key configured (GEMINI_API_KEY or FAL_KEY required)")
	}

	return &GenerateImageTool{
		adventure: adv,
	}, nil
}

// getSessionImagesDir returns the images directory for the current session.
// Uses the active session number, or session-0 if no session is active.
func (t *GenerateImageTool) getSessionImagesDir() (string, error) {
	sessionNum := 0
	if session, err := t.adventure.GetCurrentSession(); err == nil && session != nil {
		sessionNum = session.ID
	}

	imagesDir := filepath.Join(t.adventure.BasePath(), "images", fmt.Sprintf("session-%d", sessionNum))
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("creating images directory: %w", err)
	}

	return imagesDir, nil
}

// Name returns the tool name.
func (t *GenerateImageTool) Name() string {
	return "generate_image"
}

// Description returns the tool description.
func (t *GenerateImageTool) Description() string {
	return `Generate a fantasy-style image from a text prompt. Use detailed descriptions with medieval fantasy aesthetics.

CHARACTER CONSISTENCY:
- Use "include_party: true" to automatically inject visual descriptions of ALL party members into the prompt.
- Use "characters: [\"Lyra\", \"Marcus\"]" to inject only specific characters by first name or full name.
  Both parameters can be combined. Named characters override include_party for partial scenes.
  Example: "Lyra entre seule dans la taverne" → characters: ["Lyra"]
  Example: "Le groupe affronte le dragon" → include_party: true

The image will be saved in the adventure's images folder. Returns the path to the generated image.`
}

// InputSchema returns the JSON schema for tool input.
func (t *GenerateImageTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "Detailed text description of the image to generate. Describe the setting, action, lighting, atmosphere. Character appearances will be injected automatically if include_party or characters is set.",
			},
			"style": map[string]interface{}{
				"type":        "string",
				"description": "Optional style hint (e.g., 'epic', 'dark_fantasy', 'watercolor'). Default is 'epic fantasy art'.",
			},
			"include_party": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, automatically injects visual descriptions of ALL party members (gender, appearance, equipment) into the prompt for consistent character rendering.",
			},
			"characters": map[string]interface{}{
				"type":        "array",
				"description": "List of character names (first name or full name) to include in the image. Use this when only specific party members are present in the scene (e.g., [\"Lyra\"] for a scene with only Lyra). Matched by partial name (case-insensitive).",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"combat_ready": map[string]interface{}{
				"type":        "boolean",
				"description": "Controls weapon posture in the image. true = weapons drawn and ready (combat scenes, ambush, confrontation). false (default) = weapons sheathed, relaxed posture (social scenes, exploration, rest, travel). Always set to false for tavern, council, travel or dialogue scenes.",
			},
		},
		"required": []interface{}{"prompt"},
	}
}

// Execute runs the tool.
func (t *GenerateImageTool) Execute(params map[string]interface{}) (interface{}, error) {
	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "prompt is required and must be a non-empty string",
		}, nil
	}

	// Resolve which characters to inject into the prompt
	resolvedChars, err := t.resolveCharacters(params)
	if err != nil {
		// Non-fatal: log warning but continue without character injection
		fmt.Printf("Warning: could not load character appearances: %v\n", err)
	}

	// Read combat_ready param (default false = armes rangées)
	combatReady, _ := params["combat_ready"].(bool)

	// Inject character visual descriptions into the prompt
	if len(resolvedChars) > 0 {
		charDesc := image.BuildCharacterVisualDescriptions(resolvedChars, combatReady)
		if charDesc != "" {
			prompt = injectCharacterDescriptions(prompt, charDesc)
		}
	}

	// Add style prefix
	if style, ok := params["style"].(string); ok && style != "" {
		prompt = fmt.Sprintf("%s style: %s", style, prompt)
	} else {
		prompt = fmt.Sprintf("Epic fantasy art style: %s", prompt)
	}

	// Get the images directory for the current session
	imagesDir, err := t.getSessionImagesDir()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to get images directory: %v", err),
		}, nil
	}

	// Create generator with session-specific directory (auto-selects Google or FAL)
	generator, err := image.NewGeneratorAuto(imagesDir)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create image generator: %v", err),
		}, nil
	}

	// Generate the image
	result, err := generator.Generate(prompt, image.WithModelInstance(image.ModelFlux2Pro))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate image: %v", err),
		}, nil
	}

	response := map[string]interface{}{
		"success":  true,
		"path":     result.LocalPath,
		"filename": filepath.Base(result.LocalPath),
		"url":      result.URL,
		"display":  fmt.Sprintf("Image generated successfully: %s", filepath.Base(result.LocalPath)),
	}
	if len(resolvedChars) > 0 {
		names := make([]string, len(resolvedChars))
		for i, c := range resolvedChars {
			names[i] = c.Name
		}
		response["characters_injected"] = names
	}
	return response, nil
}

// resolveCharacters returns the characters whose appearances should be injected.
// Priority: explicit "characters" list > "include_party: true" > none.
func (t *GenerateImageTool) resolveCharacters(params map[string]interface{}) ([]*character.Character, error) {
	// Load all party members once
	allChars, err := t.adventure.GetCharacters()
	if err != nil {
		return nil, fmt.Errorf("loading party: %w", err)
	}
	if len(allChars) == 0 {
		return nil, nil
	}

	// Explicit character list takes priority
	if namesRaw, ok := params["characters"].([]interface{}); ok && len(namesRaw) > 0 {
		var names []string
		for _, n := range namesRaw {
			if s, ok := n.(string); ok && s != "" {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			return matchCharactersByName(allChars, names), nil
		}
	}

	// include_party: true → all party members
	if includeParty, ok := params["include_party"].(bool); ok && includeParty {
		return allChars, nil
	}

	return nil, nil
}

// matchCharactersByName filters characters by partial name match (case-insensitive).
// "Lyra" matches "Lyra Dusavel", "lyra" also matches.
func matchCharactersByName(all []*character.Character, names []string) []*character.Character {
	var matched []*character.Character
	for _, c := range all {
		nameLower := strings.ToLower(c.Name)
		for _, query := range names {
			if strings.Contains(nameLower, strings.ToLower(query)) {
				matched = append(matched, c)
				break
			}
		}
	}
	return matched
}

// injectCharacterDescriptions inserts character appearance info after the first sentence
// of the prompt, before the scene description continues.
func injectCharacterDescriptions(prompt, charDesc string) string {
	// Try to inject after the first period followed by a space (end of opening sentence)
	if idx := strings.Index(prompt, ". "); idx != -1 {
		return prompt[:idx+2] + "Personnages présents : " + charDesc + ". " + prompt[idx+2:]
	}
	// Fallback: prepend
	return "Personnages présents : " + charDesc + ". " + prompt
}
