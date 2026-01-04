package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Importance levels for foreshadows.
type Importance string

const (
	ImportanceMinor    Importance = "minor"    // Background detail
	ImportanceModerate Importance = "moderate" // Notable hint
	ImportanceMajor    Importance = "major"    // Key plot point
	ImportanceCritical Importance = "critical" // Central to campaign
)

// ForeshadowStatus tracks the state of a foreshadow.
type ForeshadowStatus string

const (
	ForeshadowActive    ForeshadowStatus = "active"    // Not yet resolved
	ForeshadowResolved  ForeshadowStatus = "resolved"  // Payoff delivered
	ForeshadowAbandoned ForeshadowStatus = "abandoned" // No longer relevant
)

// ForeshadowCategory helps organize narrative threads.
type ForeshadowCategory string

const (
	CategoryVillain   ForeshadowCategory = "villain"
	CategoryArtifact  ForeshadowCategory = "artifact"
	CategoryProphecy  ForeshadowCategory = "prophecy"
	CategoryMystery   ForeshadowCategory = "mystery"
	CategoryFaction   ForeshadowCategory = "faction"
	CategoryLocation  ForeshadowCategory = "location"
	CategoryCharacter ForeshadowCategory = "character"
)

// Foreshadow represents a narrative seed planted for future payoff.
type Foreshadow struct {
	ID              string             `json:"id"`
	Description     string             `json:"description"`
	PlantedAt       time.Time          `json:"planted_at"`
	PlantedSession  int                `json:"planted_session"`
	Importance      Importance         `json:"importance"`
	Status          ForeshadowStatus   `json:"status"`
	Category        ForeshadowCategory `json:"category"`
	Tags            []string           `json:"tags,omitempty"`
	Context         string             `json:"context,omitempty"`
	PayoffSession   *int               `json:"payoff_session,omitempty"`
	ResolvedAt      *time.Time         `json:"resolved_at,omitempty"`
	ResolutionNotes string             `json:"resolution_notes,omitempty"`
	RelatedNPCs     []string           `json:"related_npcs,omitempty"`
	RelatedLocations []string          `json:"related_locations,omitempty"`
}

// ForeshadowHistory holds all foreshadows for an adventure.
type ForeshadowHistory struct {
	Foreshadows []Foreshadow `json:"foreshadows"`
	NextID      int          `json:"next_id"`
}

// LoadForeshadows loads the foreshadow history.
func (a *Adventure) LoadForeshadows() (*ForeshadowHistory, error) {
	path := filepath.Join(a.basePath, "foreshadows.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &ForeshadowHistory{
			Foreshadows: []Foreshadow{},
			NextID:      1,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading foreshadows.json: %w", err)
	}

	var history ForeshadowHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("parsing foreshadows.json: %w", err)
	}

	return &history, nil
}

// SaveForeshadows saves the foreshadow history.
func (a *Adventure) SaveForeshadows(history *ForeshadowHistory) error {
	path := filepath.Join(a.basePath, "foreshadows.json")
	return a.saveJSON(path, history)
}

// PlantForeshadow creates a new narrative seed.
func (a *Adventure) PlantForeshadow(description, context string, importance Importance, category ForeshadowCategory, tags []string, relatedNPCs, relatedLocations []string, payoffSession *int) (*Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	// Get current session
	currentSession := 0
	if session, err := a.GetCurrentSession(); err == nil && session != nil {
		currentSession = session.ID
	}

	foreshadow := Foreshadow{
		ID:               fmt.Sprintf("fsh_%03d", history.NextID),
		Description:      description,
		PlantedAt:        time.Now(),
		PlantedSession:   currentSession,
		Importance:       importance,
		Status:           ForeshadowActive,
		Category:         category,
		Tags:             tags,
		Context:          context,
		PayoffSession:    payoffSession,
		RelatedNPCs:      relatedNPCs,
		RelatedLocations: relatedLocations,
	}

	history.Foreshadows = append(history.Foreshadows, foreshadow)
	history.NextID++

	if err := a.SaveForeshadows(history); err != nil {
		return nil, err
	}

	// Log to journal
	a.LogEvent("story", fmt.Sprintf("Foreshadow planté: %s (%s, %s)", description, importance, category))

	return &foreshadow, nil
}

// ResolveForeshadow marks a foreshadow as resolved with notes.
func (a *Adventure) ResolveForeshadow(id, resolutionNotes string) (*Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	for i := range history.Foreshadows {
		if history.Foreshadows[i].ID == id {
			now := time.Now()
			history.Foreshadows[i].Status = ForeshadowResolved
			history.Foreshadows[i].ResolvedAt = &now
			history.Foreshadows[i].ResolutionNotes = resolutionNotes

			if err := a.SaveForeshadows(history); err != nil {
				return nil, err
			}

			// Log to journal
			a.LogEvent("story", fmt.Sprintf("Foreshadow résolu: %s - %s", history.Foreshadows[i].Description, resolutionNotes))

			return &history.Foreshadows[i], nil
		}
	}

	return nil, fmt.Errorf("foreshadow %s not found", id)
}

// AbandonForeshadow marks a foreshadow as abandoned (no longer relevant).
func (a *Adventure) AbandonForeshadow(id, reason string) (*Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	for i := range history.Foreshadows {
		if history.Foreshadows[i].ID == id {
			now := time.Now()
			history.Foreshadows[i].Status = ForeshadowAbandoned
			history.Foreshadows[i].ResolvedAt = &now
			history.Foreshadows[i].ResolutionNotes = fmt.Sprintf("Abandoned: %s", reason)

			if err := a.SaveForeshadows(history); err != nil {
				return nil, err
			}

			return &history.Foreshadows[i], nil
		}
	}

	return nil, fmt.Errorf("foreshadow %s not found", id)
}

// GetActiveForeshadows returns all active (unresolved) foreshadows.
func (a *Adventure) GetActiveForeshadows() ([]Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	var active []Foreshadow
	for _, f := range history.Foreshadows {
		if f.Status == ForeshadowActive {
			active = append(active, f)
		}
	}

	return active, nil
}

// GetStaleForeshadows returns foreshadows older than maxAge sessions.
func (a *Adventure) GetStaleForeshadows(maxAge int) ([]Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	// Get current session
	currentSession := 0
	if session, err := a.GetCurrentSession(); err == nil && session != nil {
		currentSession = session.ID
	}

	var stale []Foreshadow
	for _, f := range history.Foreshadows {
		if f.Status == ForeshadowActive {
			age := currentSession - f.PlantedSession
			if age >= maxAge {
				stale = append(stale, f)
			}
		}
	}

	return stale, nil
}

// GetForeshadowsByCategory returns foreshadows of a specific category.
func (a *Adventure) GetForeshadowsByCategory(category ForeshadowCategory) ([]Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	var result []Foreshadow
	for _, f := range history.Foreshadows {
		if f.Category == category {
			result = append(result, f)
		}
	}

	return result, nil
}

// GetForeshadowsByImportance returns foreshadows of a specific importance level.
func (a *Adventure) GetForeshadowsByImportance(importance Importance) ([]Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	var result []Foreshadow
	for _, f := range history.Foreshadows {
		if f.Importance == importance {
			result = append(result, f)
		}
	}

	return result, nil
}

// GetForeshadow returns a specific foreshadow by ID.
func (a *Adventure) GetForeshadow(id string) (*Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	for _, f := range history.Foreshadows {
		if f.ID == id {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("foreshadow %s not found", id)
}

// GetAllForeshadows returns all foreshadows regardless of status.
func (a *Adventure) GetAllForeshadows() ([]Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	return history.Foreshadows, nil
}

// UpdateForeshadow updates fields of an existing foreshadow.
func (a *Adventure) UpdateForeshadow(id string, updates map[string]interface{}) (*Foreshadow, error) {
	history, err := a.LoadForeshadows()
	if err != nil {
		return nil, err
	}

	for i := range history.Foreshadows {
		if history.Foreshadows[i].ID == id {
			// Apply updates
			if desc, ok := updates["description"].(string); ok {
				history.Foreshadows[i].Description = desc
			}
			if context, ok := updates["context"].(string); ok {
				history.Foreshadows[i].Context = context
			}
			if imp, ok := updates["importance"].(string); ok {
				history.Foreshadows[i].Importance = Importance(imp)
			}
			if cat, ok := updates["category"].(string); ok {
				history.Foreshadows[i].Category = ForeshadowCategory(cat)
			}
			if tags, ok := updates["tags"].([]string); ok {
				history.Foreshadows[i].Tags = tags
			}
			if payoff, ok := updates["payoff_session"].(int); ok {
				history.Foreshadows[i].PayoffSession = &payoff
			}

			if err := a.SaveForeshadows(history); err != nil {
				return nil, err
			}

			return &history.Foreshadows[i], nil
		}
	}

	return nil, fmt.Errorf("foreshadow %s not found", id)
}