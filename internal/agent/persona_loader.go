// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PersonaLoader loads agent personas from disk with path resolution.
type PersonaLoader struct {
	basePaths []string
}

// PersonaMetadata represents the YAML frontmatter metadata from an agent file.
type PersonaMetadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"tools"`
	Model       string   `yaml:"model"`
}

// NewPersonaLoader creates a new PersonaLoader with default search paths.
// Searches in order: core_agents/agents, .claude/agents (fallback for backward compatibility)
func NewPersonaLoader() *PersonaLoader {
	return &PersonaLoader{
		basePaths: []string{
			"core_agents/agents",
			".claude/agents", // Fallback for backward compatibility
		},
	}
}

// NewPersonaLoaderWithPaths creates a PersonaLoader with custom search paths.
func NewPersonaLoaderWithPaths(paths []string) *PersonaLoader {
	return &PersonaLoader{
		basePaths: paths,
	}
}

// Load loads an agent persona by name, searching all configured base paths.
// Returns the full persona content (frontmatter + body).
func (pl *PersonaLoader) Load(agentName string) (string, error) {
	var searchedPaths []string

	for _, basePath := range pl.basePaths {
		path := filepath.Join(basePath, agentName+".md")
		searchedPaths = append(searchedPaths, path)

		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("persona not found: %s (searched: %v)",
		agentName, searchedPaths)
}

// LoadWithMetadata loads an agent persona and parses its YAML frontmatter.
// Returns the metadata, body content (without frontmatter), and any error.
func (pl *PersonaLoader) LoadWithMetadata(agentName string) (*PersonaMetadata, string, error) {
	content, err := pl.Load(agentName)
	if err != nil {
		return nil, "", err
	}

	metadata, body, err := pl.ParseFrontmatter(content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter for %s: %w", agentName, err)
	}

	return metadata, body, nil
}

// ParseFrontmatter extracts YAML frontmatter from markdown content.
// Expected format:
//   ---
//   name: agent-name
//   description: Agent description
//   tools: [Read, Write, Glob, Grep]
//   model: sonnet
//   ---
//
//   Markdown body content...
//
// Returns the parsed metadata, body content (without frontmatter), and any error.
func (pl *PersonaLoader) ParseFrontmatter(content string) (*PersonaMetadata, string, error) {
	// Check for frontmatter markers
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		// No frontmatter, return empty metadata and full content as body
		return &PersonaMetadata{}, content, nil
	}

	// Find the closing frontmatter marker
	lines := strings.Split(content, "\n")
	var endLine int
	foundEnd := false

	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "---" {
			endLine = i
			foundEnd = true
			break
		}
	}

	if !foundEnd {
		return nil, "", fmt.Errorf("frontmatter opening '---' found but no closing '---'")
	}

	// Extract frontmatter (lines 1 to endLine, excluding markers)
	frontmatterLines := lines[1:endLine]
	frontmatterYAML := strings.Join(frontmatterLines, "\n")

	// Extract body (lines after endLine+1)
	var bodyLines []string
	if endLine+1 < len(lines) {
		bodyLines = lines[endLine+1:]
	}
	body := strings.Join(bodyLines, "\n")
	body = strings.TrimSpace(body)

	// Parse YAML
	var metadata PersonaMetadata
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &metadata); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	return &metadata, body, nil
}

// GetBasePaths returns the configured base paths for persona loading.
func (pl *PersonaLoader) GetBasePaths() []string {
	return pl.basePaths
}
