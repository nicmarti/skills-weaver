// Package adventure provides adventure/campaign management for BFRPG.
package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of an adventure.
type Status string

const (
	StatusActive   Status = "active"
	StatusPaused   Status = "paused"
	StatusComplete Status = "complete"
)

// Adventure represents a game campaign/adventure.
type Adventure struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	LastPlayed   time.Time `json:"last_played"`
	SessionCount int       `json:"session_count"`
	Status       Status    `json:"status"`

	// Runtime paths (not serialized)
	basePath string `json:"-"`
}

// GameState represents the current state of the adventure.
type GameState struct {
	CurrentLocation string            `json:"current_location"`
	Time            GameTime          `json:"time"`
	Quests          []Quest           `json:"quests"`
	Flags           map[string]bool   `json:"flags"`
	Variables       map[string]string `json:"variables"`
}

// GameTime represents in-game time.
type GameTime struct {
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// Quest represents a quest or objective.
type Quest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"` // active, completed, failed
	AddedAt     string `json:"added_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// New creates a new adventure.
func New(name, description string) *Adventure {
	now := time.Now()
	return &Adventure{
		ID:           uuid.New().String(),
		Name:         name,
		Slug:         slugify(name),
		Description:  description,
		CreatedAt:    now,
		LastPlayed:   now,
		SessionCount: 0,
		Status:       StatusActive,
	}
}

// SetBasePath sets the base path for the adventure files.
func (a *Adventure) SetBasePath(path string) {
	a.basePath = path
}

// BasePath returns the adventure's base path.
func (a *Adventure) BasePath() string {
	return a.basePath
}

// Save writes all adventure data to disk.
func (a *Adventure) Save(baseDir string) error {
	adventurePath := filepath.Join(baseDir, a.Slug)
	a.basePath = adventurePath

	// Create directory structure
	dirs := []string{
		adventurePath,
		filepath.Join(adventurePath, "characters"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	// Save adventure metadata
	if err := a.saveJSON(filepath.Join(adventurePath, "adventure.json"), a); err != nil {
		return fmt.Errorf("saving adventure.json: %w", err)
	}

	return nil
}

// Load reads an adventure from disk.
func Load(adventurePath string) (*Adventure, error) {
	data, err := os.ReadFile(filepath.Join(adventurePath, "adventure.json"))
	if err != nil {
		return nil, fmt.Errorf("reading adventure.json: %w", err)
	}

	var a Adventure
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parsing adventure.json: %w", err)
	}

	a.basePath = adventurePath
	return &a, nil
}

// LoadByName finds and loads an adventure by name or slug.
func LoadByName(baseDir, name string) (*Adventure, error) {
	slug := slugify(name)
	adventurePath := filepath.Join(baseDir, slug)

	if _, err := os.Stat(adventurePath); os.IsNotExist(err) {
		// Try to find by iterating
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			return nil, fmt.Errorf("reading adventures directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			a, err := Load(filepath.Join(baseDir, entry.Name()))
			if err != nil {
				continue
			}
			if strings.EqualFold(a.Name, name) || a.Slug == slug {
				return a, nil
			}
		}
		return nil, fmt.Errorf("adventure not found: %s", name)
	}

	return Load(adventurePath)
}

// ListAdventures returns all adventures in a directory.
func ListAdventures(baseDir string) ([]*Adventure, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Adventure{}, nil
		}
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var adventures []*Adventure
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		a, err := Load(filepath.Join(baseDir, entry.Name()))
		if err != nil {
			continue // Skip invalid adventures
		}
		adventures = append(adventures, a)
	}

	return adventures, nil
}

// Delete removes an adventure from disk.
func Delete(baseDir, name string) error {
	a, err := LoadByName(baseDir, name)
	if err != nil {
		return err
	}

	return os.RemoveAll(a.basePath)
}

// LoadState loads the game state.
func (a *Adventure) LoadState() (*GameState, error) {
	path := filepath.Join(a.basePath, "state.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return default state
		return &GameState{
			CurrentLocation: "Début de l'aventure",
			Time:            GameTime{Day: 1, Hour: 8, Minute: 0},
			Quests:          []Quest{},
			Flags:           make(map[string]bool),
			Variables:       make(map[string]string),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state.json: %w", err)
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state.json: %w", err)
	}

	return &state, nil
}

// SaveState saves the game state.
func (a *Adventure) SaveState(state *GameState) error {
	path := filepath.Join(a.basePath, "state.json")
	return a.saveJSON(path, state)
}

// UpdateLastPlayed updates the last played timestamp.
func (a *Adventure) UpdateLastPlayed() {
	a.LastPlayed = time.Now()
}

// ToMarkdown generates a summary of the adventure.
func (a *Adventure) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", a.Name))

	if a.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", a.Description))
	}

	sb.WriteString("## Informations\n\n")
	sb.WriteString(fmt.Sprintf("- **Statut** : %s\n", a.Status))
	sb.WriteString(fmt.Sprintf("- **Sessions jouées** : %d\n", a.SessionCount))
	sb.WriteString(fmt.Sprintf("- **Créée le** : %s\n", a.CreatedAt.Format("02/01/2006 15:04")))
	sb.WriteString(fmt.Sprintf("- **Dernière session** : %s\n", a.LastPlayed.Format("02/01/2006 15:04")))

	return sb.String()
}

// Helper methods

func (a *Adventure) saveJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, "é", "e")
	slug = strings.ReplaceAll(slug, "è", "e")
	slug = strings.ReplaceAll(slug, "ê", "e")
	slug = strings.ReplaceAll(slug, "à", "a")
	slug = strings.ReplaceAll(slug, "ù", "u")
	slug = strings.ReplaceAll(slug, "î", "i")
	slug = strings.ReplaceAll(slug, "ô", "o")
	return slug
}
