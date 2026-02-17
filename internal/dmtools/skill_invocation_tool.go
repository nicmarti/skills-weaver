package dmtools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"dungeons/internal/skills"
)

// SkillRegistry interface defines the methods needed to work with skills.
type SkillRegistry interface {
	Get(skillName string) (*skills.Skill, bool)
	List() []string
	GetDescriptions() string
}

// NewInvokeSkillTool creates a tool to execute CLI-based skills.
// This allows the DM agent to use skills like dice-roller, name-generator, etc.
// adventureBasePath is the path to the current adventure (e.g., "data/adventures/my-adventure")
// and is used to provide context for commands that need it (like sw-character).
func NewInvokeSkillTool(registry SkillRegistry, adventureBasePath string) *SimpleTool {
	return &SimpleTool{
		name: "invoke_skill",
		description: fmt.Sprintf(`Execute a CLI-based skill to perform specialized tasks.

Available skills:
%s

To use a skill, provide the skill_name and the exact CLI command to execute.
The command should use the appropriate CLI binary (e.g., "./sw-dice roll 2d6+3").

Example parameters:
{
  "skill_name": "dice-roller",
  "command": "./sw-dice roll 2d6+3"
}`, registry.GetDescriptions()),
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"skill_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the skill to invoke",
					"enum":        registry.List(),
				},
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The exact CLI command to execute (e.g., './sw-dice roll 2d6+3')",
				},
			},
			"required": []string{"skill_name", "command"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			skillName := params["skill_name"].(string)
			command := params["command"].(string)

			// Validate skill exists
			skill, exists := registry.Get(skillName)
			if !exists {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("skill not found: %s", skillName),
					"display": fmt.Sprintf("Skill '%s' not found", skillName),
				}, nil
			}

			// Auto-inject --char-dir for sw-character commands when using adventure characters
			if adventureBasePath != "" && strings.HasPrefix(command, "./sw-character") {
				charDir := filepath.Join(adventureBasePath, "characters")
				// Only add --char-dir if not already specified
				if !strings.Contains(command, "--char-dir") {
					command = command + " --char-dir=" + charDir
				}
			}

			// Parse command into program and arguments
			// Handle quoted arguments properly
			cmdParts := parseCommand(command)
			if len(cmdParts) == 0 {
				return map[string]interface{}{
					"success": false,
					"error":   "empty command",
					"display": "Command is empty",
				}, nil
			}

			// Execute command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"output":  string(output),
					"display": fmt.Sprintf("Skill '%s' failed: %v", skillName, err),
				}, nil
			}

			return map[string]interface{}{
				"success":    true,
				"skill_name": skillName,
				"output":     string(output),
				"display":    fmt.Sprintf("✓ Executed %s", skill.Metadata.Name),
			}, nil
		},
	}
}

// parseCommand parses a command string into program and arguments.
// Handles quoted arguments (both single and double quotes).
func parseCommand(command string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range command {
		switch {
		case ch == '"' || ch == '\'':
			if !inQuote {
				// Start quote
				inQuote = true
				quoteChar = ch
			} else if ch == quoteChar {
				// End quote
				inQuote = false
				quoteChar = 0
			} else {
				// Different quote type inside quotes
				current.WriteRune(ch)
			}
		case ch == ' ' && !inQuote:
			// Space outside quotes - split here
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add final part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
