package agent

import (
	"encoding/base64"
	"os"
	"strings"
	"sync"
)

// WorldResources holds the world map description and image for injection into agents.
type WorldResources struct {
	MapDescription    string // French geography reference from ai/world-map-prompt.md (layout + distances)
	MapImageBase64    string // Base64-encoded PNG image
	MapImageMediaType string // "image/png"
}

var (
	worldResourcesOnce     sync.Once
	worldResourcesInstance *WorldResources
)

// LoadWorldResources loads and caches world resources (map description + image).
// Returns nil if neither file exists (graceful degradation).
func LoadWorldResources() *WorldResources {
	worldResourcesOnce.Do(func() {
		var res WorldResources
		hasContent := false

		// Load map description from ai/world-map-prompt.md
		if desc, err := loadMapDescription("ai/world-map-prompt.md"); err == nil && desc != "" {
			res.MapDescription = desc
			hasContent = true
		}

		// Load and encode map image
		if imgData, err := os.ReadFile("data/world/skillsweaver_carte_principale.png"); err == nil {
			res.MapImageBase64 = base64.StdEncoding.EncodeToString(imgData)
			res.MapImageMediaType = "image/png"
			hasContent = true
		}

		if hasContent {
			worldResourcesInstance = &res
		}
	})
	return worldResourcesInstance
}

// loadMapDescription extracts the FRENCH geography reference from
// world-map-prompt.md — the "Schema de disposition geographique" section
// (ASCII layout + key points + travel distances) through the end of the file.
//
// This is intentionally NOT the English "VERSION DETAILLEE" block, which is an
// image-generation prompt full of drawing directives irrelevant to writing or
// reasoning about the world. The French section is more useful for the campaign
// generator and the world-keeper, and keeps every LLM context in French.
func loadMapDescription(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	text := string(content)

	// Extract from the French geography section to the end of the file.
	marker := "## Schema de disposition geographique"
	idx := strings.Index(text, marker)
	if idx == -1 {
		// Fallback: return the full file content
		return text, nil
	}

	return strings.TrimSpace(text[idx:]), nil
}

// ResetWorldResources resets the cached world resources (for testing).
func ResetWorldResources() {
	worldResourcesOnce = sync.Once{}
	worldResourcesInstance = nil
}
