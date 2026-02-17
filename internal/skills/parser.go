// Package skills provides skill parsing and management for sw-dm.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMetadata represents the frontmatter of a skill definition.
type SkillMetadata struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	AllowedTools []string `yaml:"allowed-tools"`
}

// Skill represents a complete skill with metadata and content.
type Skill struct {
	Metadata SkillMetadata
	Content  string // The markdown content without frontmatter
	Path     string // Full path to the SKILL.md file
}

// SkillParser handles loading and parsing skill definitions.
type SkillParser struct {
	basePaths []string
}

// NewSkillParser creates a new skill parser with default search paths.
// It searches core_agents/skills/ first, then .claude/skills/ as fallback.
func NewSkillParser() *SkillParser {
	return &SkillParser{
		basePaths: []string{
			"core_agents/skills",
			".claude/skills",
		},
	}
}

// NewSkillParserWithPaths creates a skill parser with custom search paths.
func NewSkillParserWithPaths(basePaths []string) *SkillParser {
	return &SkillParser{
		basePaths: basePaths,
	}
}

// Load loads a skill by name, searching in configured base paths.
func (sp *SkillParser) Load(skillName string) (*Skill, error) {
	var lastErr error

	// Try each base path
	for _, basePath := range sp.basePaths {
		skillPath := filepath.Join(basePath, skillName, "SKILL.md")

		// Check if file exists
		if _, err := os.Stat(skillPath); err == nil {
			// File exists, parse it
			skill, err := sp.parseSkillFile(skillPath)
			if err != nil {
				// Save error and try next path
				lastErr = err
				continue
			}
			skill.Path = skillPath
			return skill, nil
		}
	}

	// If we found files but failed to parse them, return the parsing error
	if lastErr != nil {
		return nil, lastErr
	}

	// No files found at all
	return nil, fmt.Errorf("skill not found: %s (searched in: %v)", skillName, sp.basePaths)
}

// LoadAll loads all skills from the configured base paths.
func (sp *SkillParser) LoadAll() (map[string]*Skill, error) {
	skills := make(map[string]*Skill)

	for _, basePath := range sp.basePaths {
		// Read directory
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue // Path doesn't exist, try next
		}

		// Process each subdirectory
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillName := entry.Name()
			// Skip if already loaded (first path takes precedence)
			if _, exists := skills[skillName]; exists {
				continue
			}

			// Try to load skill
			skill, err := sp.Load(skillName)
			if err != nil {
				continue // Skip invalid skills
			}

			skills[skillName] = skill
		}
	}

	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found in: %v", sp.basePaths)
	}

	return skills, nil
}

// parseSkillFile parses a SKILL.md file, extracting frontmatter and content.
func (sp *SkillParser) parseSkillFile(filePath string) (*Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	content := string(data)

	// Check for YAML frontmatter (starts with ---)
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("skill file missing YAML frontmatter: %s", filePath)
	}

	// Find end of frontmatter
	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed YAML frontmatter in: %s", filePath)
	}

	frontmatter := parts[0]
	bodyContent := strings.TrimSpace(parts[1])

	// Parse YAML frontmatter
	var metadata SkillMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Validate required fields
	if metadata.Name == "" {
		return nil, fmt.Errorf("skill missing required field 'name' in: %s", filePath)
	}
	if metadata.Description == "" {
		return nil, fmt.Errorf("skill missing required field 'description' in: %s", filePath)
	}

	return &Skill{
		Metadata: metadata,
		Content:  bodyContent,
	}, nil
}

// GetCLIPrefix returns the CLI binary name prefix for a skill.
// Most skills use the pattern: sw-<skill-name>
// Special cases:
//   - dice-roller -> sw-dice
//   - character-generator -> sw-character
//   - adventure-manager -> sw-adventure
//   - name-generator -> sw-names
//   - npc-generator -> sw-npc
//   - image-generator -> sw-image
//   - journal-illustrator -> sw-image (uses journal subcommand)
//   - monster-manual -> sw-monster
//   - treasure-generator -> sw-treasure
//   - equipment-browser -> sw-equipment
//   - spell-reference -> sw-spell
//   - map-generator -> sw-map
//   - name-location-generator -> sw-location-names
func GetCLIPrefix(skillName string) string {
	// Handle special cases
	specialCases := map[string]string{
		"dice-roller":             "sw-dice",
		"character-generator":     "sw-character",
		"adventure-manager":       "sw-adventure",
		"name-generator":          "sw-names",
		"npc-generator":           "sw-npc",
		"image-generator":         "sw-image",
		"journal-illustrator":     "sw-image",
		"monster-manual":          "sw-monster",
		"treasure-generator":      "sw-treasure",
		"equipment-browser":       "sw-equipment",
		"spell-reference":         "sw-spell",
		"map-generator":           "sw-map",
		"name-location-generator": "sw-location-names",
	}

	if cli, ok := specialCases[skillName]; ok {
		return cli
	}

	// Default: sw-<skill-name>
	return "sw-" + skillName
}
