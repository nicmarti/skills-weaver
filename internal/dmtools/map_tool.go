package dmtools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dungeons/internal/ai"
	"dungeons/internal/image"
	"dungeons/internal/world"
)

// MapGeneratedNotifier is an interface for notifying about map generation.
// This avoids import cycle with internal/agent package.
type MapGeneratedNotifier interface {
	OnMapGenerated(location, mapPath string)
}

// GenerateMapTool generates 2D fantasy map prompts with world-keeper validation.
type GenerateMapTool struct {
	dataDir       string
	adventurePath string
	enricher      *ai.Enricher
	geography     *world.Geography
	factions      *world.Factions
	notifier      MapGeneratedNotifier
}

// NewGenerateMapTool creates a new map generation tool.
func NewGenerateMapTool(dataDir string, adventurePath string, notifier MapGeneratedNotifier) (*GenerateMapTool, error) {
	// Load world data
	geo, err := world.LoadGeography(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading geography: %w", err)
	}

	factions, err := world.LoadFactions(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading factions: %w", err)
	}

	// Create enricher
	enricher, err := ai.NewEnricher()
	if err != nil {
		return nil, fmt.Errorf("creating enricher: %w", err)
	}

	return &GenerateMapTool{
		dataDir:       dataDir,
		adventurePath: adventurePath,
		enricher:      enricher,
		geography:     geo,
		factions:      factions,
		notifier:      notifier,
	}, nil
}

// Name returns the tool name.
func (t *GenerateMapTool) Name() string {
	return "generate_map"
}

// Description returns the tool description.
func (t *GenerateMapTool) Description() string {
	return `Generate a detailed 2D fantasy map prompt to clarify narration for players. Validates locations against world-keeper data and applies kingdom-specific architectural styles.

WHEN TO USE:
- Players are confused about geography or layout
- Need to visualize a city, region, dungeon, or battle scene
- Want to show spatial relationships between locations
- Combat requires a tactical grid map

MAP TYPES:
- city: Aerial view of a city with districts, POIs, and infrastructure
- region: Bird's eye view of multiple settlements, routes, and terrain
- dungeon: Top-down floor plan with rooms, corridors, traps, and grid
- tactical: Combat grid with terrain, cover, obstacles, and elevation

The tool enriches prompts with Claude Haiku 3.5, caches results, and optionally generates images via fal.ai flux-2.`
}

// InputSchema returns the JSON schema for tool input.
func (t *GenerateMapTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"map_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of map to generate: 'city' (aerial view with districts), 'region' (bird's eye with multiple settlements), 'dungeon' (top-down floor plan with grid), 'tactical' (combat grid with terrain).",
				"enum":        []interface{}{"city", "region", "dungeon", "tactical"},
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the location, dungeon, or scene. For city/region: must exist in geography.json (e.g., 'Cordova'). For dungeon/tactical: any descriptive name (e.g., 'La Crypte des Ombres', 'Embuscade en forêt').",
			},
			"features": map[string]interface{}{
				"type":        "array",
				"description": "Optional list of additional POIs or features to include (e.g., ['Taverne du Voile Écarlate', 'Villa de Valorian']). For tactical maps, can specify terrain elements like ['Ruisseau', 'Pont de bois'].",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"scale": map[string]interface{}{
				"type":        "string",
				"description": "Map scale: 'small' (detailed, focused area), 'medium' (balanced), 'large' (wide view, multiple locations). Default: 'medium'.",
				"enum":        []interface{}{"small", "medium", "large"},
			},
			"style": map[string]interface{}{
				"type":        "string",
				"description": "Visual style: 'illustrated' (vibrant, colorful) or 'dark_fantasy' (atmospheric, dramatic). Default: 'illustrated'.",
				"enum":        []interface{}{"illustrated", "dark_fantasy"},
			},
			"level": map[string]interface{}{
				"type":        "integer",
				"description": "Dungeon level number (1, 2, 3, etc.). Only for map_type='dungeon'.",
			},
			"terrain": map[string]interface{}{
				"type":        "string",
				"description": "Terrain type for tactical maps (e.g., 'forêt', 'montagne', 'plaine', 'marais'). Only for map_type='tactical'.",
			},
			"scene": map[string]interface{}{
				"type":        "string",
				"description": "Scene description for context (e.g., 'Combat contre des bandits en forêt'). Only for map_type='tactical'.",
			},
			"generate_image": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, also generates the actual map image using fal.ai flux-2. Requires FAL_KEY environment variable. Default: false (prompt only).",
			},
		},
		"required": []interface{}{"map_type", "name"},
	}
}

// Execute runs the tool.
func (t *GenerateMapTool) Execute(params map[string]interface{}) (interface{}, error) {
	// Parse required parameters
	mapType, ok := params["map_type"].(string)
	if !ok || mapType == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "map_type is required (city, region, dungeon, tactical)",
		}, nil
	}

	name, ok := params["name"].(string)
	if !ok || name == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "name is required",
		}, nil
	}

	// Parse optional parameters
	var features []string
	if featuresRaw, ok := params["features"].([]interface{}); ok {
		for _, f := range featuresRaw {
			if fStr, ok := f.(string); ok {
				features = append(features, fStr)
			}
		}
	}

	scale := "medium"
	if scaleParam, ok := params["scale"].(string); ok && scaleParam != "" {
		scale = scaleParam
	}

	style := "illustrated"
	if styleParam, ok := params["style"].(string); ok && styleParam != "" {
		style = styleParam
	}

	level := 1
	if levelParam, ok := params["level"].(float64); ok {
		level = int(levelParam)
	}

	terrain := ""
	if terrainParam, ok := params["terrain"].(string); ok {
		terrain = terrainParam
	}

	scene := ""
	if sceneParam, ok := params["scene"].(string); ok {
		scene = sceneParam
	}

	generateImage := false
	if genImg, ok := params["generate_image"].(bool); ok {
		generateImage = genImg
	}

	// Validate and get location data (for city/region types)
	var location *world.Location
	var kingdom *world.Kingdom

	if mapType == "city" || mapType == "region" {
		exists, loc, _, _ := world.ValidateLocationExists(name, t.geography)
		if !exists {
			// Provide suggestions
			suggestions := world.GetSuggestions(name, t.geography, 5)
			suggestionStrs := []string{}
			for _, s := range suggestions {
				suggestionStrs = append(suggestionStrs, fmt.Sprintf("%s (%s)", s.Location.Name, s.Location.Kingdom))
			}

			return map[string]interface{}{
				"success":     false,
				"error":       fmt.Sprintf("Location '%s' not found in geography.json", name),
				"suggestions": suggestionStrs,
				"hint":        "For dungeons and tactical maps, location validation is not required.",
			}, nil
		}
		location = loc

		// Get kingdom data
		if location.Kingdom != "" {
			_, kd, err := world.ValidateKingdomExists(strings.ToLower(location.Kingdom), t.factions)
			if err == nil {
				kingdom = kd
			}
		}
	}

	// Build request
	req := ai.MapPromptRequest{
		MapType:      mapType,
		LocationName: name,
		Scale:        scale,
		Features:     features,
		Terrain:      terrain,
		Style:        style,
		DungeonLevel: level,
		SceneDesc:    scene,
	}

	if kingdom != nil {
		req.Kingdom = kingdom.ID
	}

	// Enrich prompt with AI
	result, err := t.enricher.EnrichMapPrompt(req, location, kingdom)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to enrich prompt: %v", err),
		}, nil
	}

	// Save to cache
	cacheFile := t.getCacheFile(name, mapType, scale)
	if err := t.saveCachedPrompt(cacheFile, result); err != nil {
		// Non-fatal: log warning but continue
		fmt.Printf("Warning: Failed to save cache: %v\n", err)
	}

	response := map[string]interface{}{
		"success":     true,
		"map_type":    result.MapType,
		"prompt":      result.Prompt,
		"style_hints": result.StyleHints,
		"cache_path":  cacheFile,
		"display":     formatMapDisplay(result),
	}

	// Generate image if requested
	if generateImage {
		imagePath, imageURL, err := t.generateImage(result.Prompt, name, mapType, scale)
		if err != nil {
			response["image_warning"] = fmt.Sprintf("Prompt generated but image generation failed: %v", err)
		} else {
			response["image_path"] = imagePath
			response["image_url"] = imageURL
			response["display"] = response["display"].(string) + fmt.Sprintf("\n\nImage générée: %s", filepath.Base(imagePath))

			// Notify that a map was generated (triggers minimap refresh in web UI)
			if t.notifier != nil && location != nil {
				t.notifier.OnMapGenerated(location.Name, imagePath)
			}
		}
	}

	if location != nil {
		response["location"] = location.Name
		response["kingdom"] = location.Kingdom
	}

	if len(features) > 0 {
		response["features"] = features
	}

	return response, nil
}

// getCacheFile returns the cache file path for a map prompt.
func (t *GenerateMapTool) getCacheFile(name, mapType, scale string) string {
	// Use hyphens for consistency in filenames
	safeName := strings.ReplaceAll(name, " ", "-")
	safeName = strings.ToLower(safeName)
	return filepath.Join(t.dataDir, "maps", fmt.Sprintf("%s_%s_%s_prompt.json", safeName, mapType, scale))
}

// saveCachedPrompt saves a prompt result to cache.
func (t *GenerateMapTool) saveCachedPrompt(cacheFile string, result *ai.MapPromptResult) error {
	// Ensure directory exists
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// generateImage generates an actual map image using fal.ai flux-2.
func (t *GenerateMapTool) generateImage(prompt, name, mapType, scale string) (string, string, error) {
	// Create image generator
	outputDir := filepath.Join(t.dataDir, "maps")
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return "", "", fmt.Errorf("creating generator: %w", err)
	}

	// Generate unique filename (use hyphens for consistency)
	safeName := strings.ReplaceAll(name, " ", "-")
	safeName = strings.ToLower(safeName)
	filename := fmt.Sprintf("%s_%s_%s", safeName, mapType, scale)

	// Generate with flux-2 model
	opts := []image.Option{
		image.WithModel("fal-ai/flux-2"),
		image.WithImageSize("landscape_16_9"),
		image.WithNumImages(1),
		image.WithFilenamePrefix(filename),
	}

	result, err := gen.Generate(prompt, opts...)
	if err != nil {
		return "", "", err
	}

	return result.LocalPath, result.URL, nil
}

// formatMapDisplay formats the map result for display to the user.
func formatMapDisplay(result *ai.MapPromptResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Carte %s générée ===\n\n", strings.ToUpper(result.MapType)))

	if result.LocationName != "" {
		sb.WriteString(fmt.Sprintf("Lieu: %s", result.LocationName))
		if result.Kingdom != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", result.Kingdom))
		}
		sb.WriteString("\n")
	}

	if len(result.Features) > 0 {
		sb.WriteString(fmt.Sprintf("POIs inclus: %s\n", strings.Join(result.Features, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\nStyle: %s\n", result.StyleHints))
	sb.WriteString(fmt.Sprintf("\nPROMPT:\n%s\n", result.Prompt))

	return sb.String()
}
