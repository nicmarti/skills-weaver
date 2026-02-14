package agent

import (
	"encoding/base64"
	"os"
	"strings"
	"sync"
)

// WorldResources holds the world map description and image for injection into agents.
type WorldResources struct {
	MapDescription    string // Detailed text from ai/world-map-prompt.md
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

// loadMapDescription extracts the detailed version from the world-map-prompt.md file.
// It extracts content between the "VERSION DETAILLEE" code block markers.
func loadMapDescription(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	text := string(content)

	// Find the "VERSION DETAILLEE" section and extract its code block
	marker := "### VERSION DETAILLEE"
	idx := strings.Index(text, marker)
	if idx == -1 {
		// Fallback: return the full file content
		return text, nil
	}

	// Find the opening ``` after the marker
	rest := text[idx:]
	startTick := strings.Index(rest, "```")
	if startTick == -1 {
		return text, nil
	}

	// Skip past the opening ``` line
	afterStart := rest[startTick+3:]
	newline := strings.Index(afterStart, "\n")
	if newline == -1 {
		return text, nil
	}
	blockContent := afterStart[newline+1:]

	// Find the closing ```
	endTick := strings.Index(blockContent, "```")
	if endTick == -1 {
		return text, nil
	}

	return strings.TrimSpace(blockContent[:endTick]), nil
}

// ResetWorldResources resets the cached world resources (for testing).
func ResetWorldResources() {
	worldResourcesOnce = sync.Once{}
	worldResourcesInstance = nil
}
