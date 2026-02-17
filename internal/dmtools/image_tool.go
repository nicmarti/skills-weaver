package dmtools

import (
	"fmt"
	"os"
	"path/filepath"

	"dungeons/internal/adventure"
	"dungeons/internal/image"
)

// GenerateImageTool generates fantasy images from prompts.
type GenerateImageTool struct {
	adventure *adventure.Adventure
}

// NewGenerateImageTool creates a new image generation tool.
func NewGenerateImageTool(adv *adventure.Adventure) (*GenerateImageTool, error) {
	// Verify FAL_KEY is set
	if os.Getenv("FAL_KEY") == "" {
		return nil, fmt.Errorf("FAL_KEY environment variable not set")
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
	return "Generate a fantasy-style image from a text prompt. Use detailed descriptions with medieval fantasy aesthetics. The image will be saved in the adventure's images folder. Returns the path to the generated image."
}

// InputSchema returns the JSON schema for tool input.
func (t *GenerateImageTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "Detailed text description of the image to generate. Should include characters, setting, lighting, atmosphere, and art style (e.g., 'epic fantasy art', 'dark medieval style'). The more detailed, the better.",
			},
			"style": map[string]interface{}{
				"type":        "string",
				"description": "Optional style hint (e.g., 'epic', 'dark_fantasy', 'watercolor'). Default is 'epic fantasy art'.",
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

	// Add style prefix if specified
	if style, ok := params["style"].(string); ok && style != "" {
		prompt = fmt.Sprintf("%s style: %s", style, prompt)
	} else {
		// Default to epic fantasy art style
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

	// Create generator with session-specific directory
	generator, err := image.NewGenerator(imagesDir)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create image generator: %v", err),
		}, nil
	}

	// Generate the image using flux-2-pro (state-of-the-art quality)
	result, err := generator.Generate(prompt, image.WithModelInstance(image.ModelFlux2Pro))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate image: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success":  true,
		"path":     result.LocalPath,
		"filename": filepath.Base(result.LocalPath),
		"url":      result.URL,
		"display":  fmt.Sprintf("Image generated successfully: %s", filepath.Base(result.LocalPath)),
	}, nil
}
