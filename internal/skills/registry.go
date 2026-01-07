package skills

import (
	"fmt"
	"sort"
)

// Registry manages available skills for the agent.
type Registry struct {
	skills map[string]*Skill
	parser *SkillParser
}

// NewRegistry creates a new skill registry and loads all available skills.
func NewRegistry() (*Registry, error) {
	parser := NewSkillParser()
	skills, err := parser.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	return &Registry{
		skills: skills,
		parser: parser,
	}, nil
}

// NewRegistryWithParser creates a registry with a custom parser.
func NewRegistryWithParser(parser *SkillParser) (*Registry, error) {
	skills, err := parser.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	return &Registry{
		skills: skills,
		parser: parser,
	}, nil
}

// Get returns a skill by name.
func (r *Registry) Get(skillName string) (*Skill, bool) {
	skill, exists := r.skills[skillName]
	return skill, exists
}

// List returns a sorted list of all available skill names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAll returns a map of all skills.
func (r *Registry) GetAll() map[string]*Skill {
	// Return a copy to prevent external modification
	result := make(map[string]*Skill, len(r.skills))
	for name, skill := range r.skills {
		result[name] = skill
	}
	return result
}

// Count returns the number of registered skills.
func (r *Registry) Count() int {
	return len(r.skills)
}

// Reload reloads all skills from disk.
func (r *Registry) Reload() error {
	skills, err := r.parser.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to reload skills: %w", err)
	}
	r.skills = skills
	return nil
}

// GetDescriptions returns a formatted string with all skill descriptions.
// Useful for providing context to the agent about available skills.
func (r *Registry) GetDescriptions() string {
	names := r.List()
	result := "Available Skills:\n\n"

	for _, name := range names {
		skill := r.skills[name]
		result += fmt.Sprintf("- **%s**: %s\n", name, skill.Metadata.Description)
	}

	return result
}

// GetCLIMapping returns a map of skill names to their CLI commands.
func (r *Registry) GetCLIMapping() map[string]string {
	mapping := make(map[string]string)
	for name := range r.skills {
		mapping[name] = GetCLIPrefix(name)
	}
	return mapping
}
