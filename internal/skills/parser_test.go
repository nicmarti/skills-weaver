package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSkillParser_ValidSkillFile tests parsing a valid SKILL.md file.
func TestSkillParser_ValidSkillFile(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	// Create a valid skill file
	skillContent := `---
name: test-skill
description: A test skill for validation
allowed-tools:
  - tool1
  - tool2
---

# Test Skill

This is the skill content.
`
	createSkillFile(t, tmpDir, "test-skill", skillContent)

	parser := NewSkillParserWithPaths([]string{tmpDir})
	skill, err := parser.Load("test-skill")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if skill.Metadata.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got: %s", skill.Metadata.Name)
	}

	if skill.Metadata.Description != "A test skill for validation" {
		t.Errorf("Expected description, got: %s", skill.Metadata.Description)
	}

	if len(skill.Metadata.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got: %d", len(skill.Metadata.AllowedTools))
	}

	if skill.Content == "" {
		t.Error("Expected content to be extracted")
	}

	if !contains(skill.Content, "# Test Skill") {
		t.Error("Expected content to contain markdown header")
	}
}

// TestSkillParser_MissingFrontmatter tests error handling for missing frontmatter.
func TestSkillParser_MissingFrontmatter(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	skillContent := `# Test Skill

No frontmatter here.
`
	createSkillFile(t, tmpDir, "invalid-skill", skillContent)

	parser := NewSkillParserWithPaths([]string{tmpDir})
	_, err := parser.Load("invalid-skill")
	if err == nil {
		t.Fatal("Expected error for missing frontmatter, got nil")
	}

	if !contains(err.Error(), "missing YAML frontmatter") {
		t.Errorf("Expected error about missing frontmatter, got: %v", err)
	}
}

// TestSkillParser_MalformedFrontmatter tests error handling for malformed frontmatter.
func TestSkillParser_MalformedFrontmatter(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	// Missing closing ---
	skillContent := `---
name: test-skill
description: Test

# Content without closing frontmatter marker
`
	createSkillFile(t, tmpDir, "malformed-skill", skillContent)

	parser := NewSkillParserWithPaths([]string{tmpDir})
	_, err := parser.Load("malformed-skill")
	if err == nil {
		t.Fatal("Expected error for malformed frontmatter, got nil")
	}

	if !contains(err.Error(), "malformed YAML frontmatter") {
		t.Errorf("Expected error about malformed frontmatter, got: %v", err)
	}
}

// TestSkillParser_InvalidYAML tests error handling for invalid YAML.
func TestSkillParser_InvalidYAML(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	skillContent := `---
name: test-skill
description: [invalid yaml structure
---

Content
`
	createSkillFile(t, tmpDir, "invalid-yaml", skillContent)

	parser := NewSkillParserWithPaths([]string{tmpDir})
	_, err := parser.Load("invalid-yaml")
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}

	if !contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Expected error about YAML parsing, got: %v", err)
	}
}

// TestSkillParser_MissingRequiredFields tests validation of required fields.
func TestSkillParser_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedErr string
	}{
		{
			name: "missing name",
			content: `---
description: Test skill
---

Content`,
			expectedErr: "missing required field 'name'",
		},
		{
			name: "missing description",
			content: `---
name: test-skill
---

Content`,
			expectedErr: "missing required field 'description'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := setupTempSkillDir(t)
			defer cleanup()

			createSkillFile(t, tmpDir, "invalid-skill", tt.content)

			parser := NewSkillParserWithPaths([]string{tmpDir})
			_, err := parser.Load("invalid-skill")
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestSkillParser_SkillNotFound tests error handling for non-existent skills.
func TestSkillParser_SkillNotFound(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	parser := NewSkillParserWithPaths([]string{tmpDir})
	_, err := parser.Load("non-existent-skill")
	if err == nil {
		t.Fatal("Expected error for non-existent skill, got nil")
	}

	if !contains(err.Error(), "skill not found") {
		t.Errorf("Expected 'skill not found' error, got: %v", err)
	}
}

// TestSkillParser_MultiplePaths tests fallback to secondary paths.
func TestSkillParser_MultiplePaths(t *testing.T) {
	tmpDir1, cleanup1 := setupTempSkillDir(t)
	defer cleanup1()
	tmpDir2, cleanup2 := setupTempSkillDir(t)
	defer cleanup2()

	// Create skill only in second path
	skillContent := `---
name: test-skill
description: Test skill
---

Content from path 2`
	createSkillFile(t, tmpDir2, "test-skill", skillContent)

	// Parse with both paths (first path is empty)
	parser := NewSkillParserWithPaths([]string{tmpDir1, tmpDir2})
	skill, err := parser.Load("test-skill")
	if err != nil {
		t.Fatalf("Expected skill to be found in second path, got error: %v", err)
	}

	if !contains(skill.Content, "Content from path 2") {
		t.Error("Expected content from second path")
	}
}

// TestSkillParser_PathPrecedence tests that first path takes precedence.
func TestSkillParser_PathPrecedence(t *testing.T) {
	tmpDir1, cleanup1 := setupTempSkillDir(t)
	defer cleanup1()
	tmpDir2, cleanup2 := setupTempSkillDir(t)
	defer cleanup2()

	// Create skill in both paths
	skill1 := `---
name: test-skill
description: From path 1
---

Content from path 1`
	skill2 := `---
name: test-skill
description: From path 2
---

Content from path 2`

	createSkillFile(t, tmpDir1, "test-skill", skill1)
	createSkillFile(t, tmpDir2, "test-skill", skill2)

	// Parse with both paths
	parser := NewSkillParserWithPaths([]string{tmpDir1, tmpDir2})
	skill, err := parser.Load("test-skill")
	if err != nil {
		t.Fatalf("Expected skill to be found, got error: %v", err)
	}

	// Should load from first path
	if skill.Metadata.Description != "From path 1" {
		t.Errorf("Expected description from path 1, got: %s", skill.Metadata.Description)
	}

	if !contains(skill.Content, "Content from path 1") {
		t.Error("Expected content from first path")
	}
}

// TestSkillParser_LoadAll tests loading all skills from directories.
func TestSkillParser_LoadAll(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	// Create multiple skills
	skill1 := `---
name: skill1
description: First skill
---

Content 1`
	skill2 := `---
name: skill2
description: Second skill
---

Content 2`

	createSkillFile(t, tmpDir, "skill1", skill1)
	createSkillFile(t, tmpDir, "skill2", skill2)

	parser := NewSkillParserWithPaths([]string{tmpDir})
	skills, err := parser.LoadAll()
	if err != nil {
		t.Fatalf("Expected to load all skills, got error: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got: %d", len(skills))
	}

	if _, exists := skills["skill1"]; !exists {
		t.Error("Expected skill1 to be loaded")
	}

	if _, exists := skills["skill2"]; !exists {
		t.Error("Expected skill2 to be loaded")
	}
}

// TestSkillParser_LoadAllEmpty tests LoadAll with no valid skills.
func TestSkillParser_LoadAllEmpty(t *testing.T) {
	tmpDir, cleanup := setupTempSkillDir(t)
	defer cleanup()

	parser := NewSkillParserWithPaths([]string{tmpDir})
	_, err := parser.LoadAll()
	if err == nil {
		t.Fatal("Expected error when no skills found, got nil")
	}

	if !contains(err.Error(), "no skills found") {
		t.Errorf("Expected 'no skills found' error, got: %v", err)
	}
}

// TestGetCLIPrefix tests CLI prefix mapping.
func TestGetCLIPrefix(t *testing.T) {
	tests := []struct {
		skillName string
		expected  string
	}{
		{"dice-roller", "sw-dice"},
		{"character-generator", "sw-character"},
		{"adventure-manager", "sw-adventure"},
		{"name-generator", "sw-names"},
		{"npc-generator", "sw-npc"},
		{"image-generator", "sw-image"},
		{"journal-illustrator", "sw-image"},
		{"monster-manual", "sw-monster"},
		{"treasure-generator", "sw-treasure"},
		{"equipment-browser", "sw-equipment"},
		{"spell-reference", "sw-spell"},
		{"map-generator", "sw-map"},
		{"name-location-generator", "sw-location-names"},
		{"custom-skill", "sw-custom-skill"}, // Default case
		{"another-skill", "sw-another-skill"}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.skillName, func(t *testing.T) {
			result := GetCLIPrefix(tt.skillName)
			if result != tt.expected {
				t.Errorf("GetCLIPrefix(%s) = %s, want %s", tt.skillName, result, tt.expected)
			}
		})
	}
}

// Helper functions

func setupTempSkillDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "skills-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createSkillFile(t *testing.T, baseDir, skillName, content string) {
	t.Helper()

	skillDir := filepath.Join(baseDir, skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write skill file: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
