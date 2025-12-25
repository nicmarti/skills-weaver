package npcmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dungeons/internal/npc"
)

// ImportanceLevel represents how significant an NPC is to the story.
type ImportanceLevel string

const (
	ImportanceMentioned  ImportanceLevel = "mentioned"  // Generated but not interacted with
	ImportanceInteracted ImportanceLevel = "interacted" // Had dialogue or brief encounter
	ImportanceRecurring  ImportanceLevel = "recurring"  // Appeared multiple times
	ImportanceKey        ImportanceLevel = "key"        // Major story significance
)

// importanceValue returns a numeric value for importance level comparison.
func importanceValue(level ImportanceLevel) int {
	switch level {
	case ImportanceMentioned:
		return 1
	case ImportanceInteracted:
		return 2
	case ImportanceRecurring:
		return 3
	case ImportanceKey:
		return 4
	default:
		return 0
	}
}

// NPCRecord represents a generated NPC with metadata.
type NPCRecord struct {
	ID              string          `json:"id"`
	GeneratedAt     time.Time       `json:"generated_at"`
	SessionNumber   int             `json:"session_number"`
	NPC             *npc.NPC        `json:"npc"`
	Context         string          `json:"context"`          // Where/when encountered
	Importance      ImportanceLevel `json:"importance"`
	Notes           []string        `json:"notes"`            // DM notes added over time
	Appearances     int             `json:"appearances"`      // Number of times appeared
	PromotedToWorld bool            `json:"promoted_to_world"`
	WorldKeeperNotes string         `json:"world_keeper_notes"` // Validation/enrichment from world-keeper
}

// NPCDatabase holds all generated NPCs organized by session.
type NPCDatabase struct {
	Sessions map[string][]NPCRecord `json:"sessions"` // Key: "session_0", "session_1", etc.
	NextID   int                    `json:"next_id"`
}

// Manager handles NPC persistence for an adventure.
type Manager struct {
	adventurePath string
	dbPath        string
}

// NewManager creates a new NPC manager for an adventure.
func NewManager(adventurePath string) *Manager {
	return &Manager{
		adventurePath: adventurePath,
		dbPath:        filepath.Join(adventurePath, "npcs-generated.json"),
	}
}

// Load loads the NPC database from disk.
func (m *Manager) Load() (*NPCDatabase, error) {
	// Check if file exists
	if _, err := os.Stat(m.dbPath); os.IsNotExist(err) {
		// Create empty database
		return &NPCDatabase{
			Sessions: make(map[string][]NPCRecord),
			NextID:   1,
		}, nil
	}

	data, err := os.ReadFile(m.dbPath)
	if err != nil {
		return nil, fmt.Errorf("reading npcs-generated.json: %w", err)
	}

	var db NPCDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("parsing npcs-generated.json: %w", err)
	}

	// Ensure Sessions map is initialized
	if db.Sessions == nil {
		db.Sessions = make(map[string][]NPCRecord)
	}

	return &db, nil
}

// Save saves the NPC database to disk.
func (m *Manager) Save(db *NPCDatabase) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling npcs-generated.json: %w", err)
	}

	if err := os.WriteFile(m.dbPath, data, 0644); err != nil {
		return fmt.Errorf("writing npcs-generated.json: %w", err)
	}

	return nil
}

// AddNPC adds a new NPC to the database.
func (m *Manager) AddNPC(sessionNumber int, generatedNPC *npc.NPC, context string, worldKeeperNotes string) (*NPCRecord, error) {
	db, err := m.Load()
	if err != nil {
		return nil, err
	}

	// Create record
	record := NPCRecord{
		ID:               fmt.Sprintf("npc_%03d", db.NextID),
		GeneratedAt:      time.Now(),
		SessionNumber:    sessionNumber,
		NPC:              generatedNPC,
		Context:          context,
		Importance:       ImportanceMentioned,
		Notes:            []string{},
		Appearances:      1,
		PromotedToWorld:  false,
		WorldKeeperNotes: worldKeeperNotes,
	}

	// Add to session
	sessionKey := fmt.Sprintf("session_%d", sessionNumber)
	db.Sessions[sessionKey] = append(db.Sessions[sessionKey], record)
	db.NextID++

	// Save
	if err := m.Save(db); err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateImportance updates an NPC's importance and adds a note.
func (m *Manager) UpdateImportance(npcName string, importance ImportanceLevel, note string) error {
	db, err := m.Load()
	if err != nil {
		return err
	}

	// Find NPC by name
	found := false
	for sessionKey, records := range db.Sessions {
		for i, record := range records {
			if record.NPC.Name == npcName {
				// Update importance (only increase, never decrease)
				if importanceValue(importance) > importanceValue(record.Importance) {
					db.Sessions[sessionKey][i].Importance = importance
				}

				// Add note
				if note != "" {
					db.Sessions[sessionKey][i].Notes = append(db.Sessions[sessionKey][i].Notes, note)
				}

				// Increment appearances
				db.Sessions[sessionKey][i].Appearances++

				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("NPC not found: %s", npcName)
	}

	return m.Save(db)
}

// GetNPCHistory retrieves the full history of an NPC.
func (m *Manager) GetNPCHistory(npcName string) (*NPCRecord, error) {
	db, err := m.Load()
	if err != nil {
		return nil, err
	}

	// Find NPC by name
	for _, records := range db.Sessions {
		for _, record := range records {
			if record.NPC.Name == npcName {
				return &record, nil
			}
		}
	}

	return nil, fmt.Errorf("NPC not found: %s", npcName)
}

// ListNPCsForReview returns NPCs that should be reviewed for promotion.
func (m *Manager) ListNPCsForReview() ([]NPCRecord, error) {
	db, err := m.Load()
	if err != nil {
		return nil, err
	}

	var candidates []NPCRecord
	for _, records := range db.Sessions {
		for _, record := range records {
			// Include NPCs with importance >= interacted and not yet promoted
			if importanceValue(record.Importance) >= importanceValue(ImportanceInteracted) && !record.PromotedToWorld {
				candidates = append(candidates, record)
			}
		}
	}

	return candidates, nil
}

// MarkAsPromoted marks an NPC as promoted to world/npcs.json.
func (m *Manager) MarkAsPromoted(npcName string) error {
	db, err := m.Load()
	if err != nil {
		return err
	}

	// Find NPC by name
	found := false
	for sessionKey, records := range db.Sessions {
		for i, record := range records {
			if record.NPC.Name == npcName {
				db.Sessions[sessionKey][i].PromotedToWorld = true
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("NPC not found: %s", npcName)
	}

	return m.Save(db)
}

// GetCurrentSessionNumber retrieves the current session number from sessions.json.
func (m *Manager) GetCurrentSessionNumber() (int, error) {
	sessionsPath := filepath.Join(m.adventurePath, "sessions.json")

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		// If sessions.json doesn't exist, default to session 0
		return 0, nil
	}

	var sessions struct {
		Sessions []interface{} `json:"sessions"`
	}
	if err := json.Unmarshal(data, &sessions); err != nil {
		return 0, fmt.Errorf("parsing sessions.json: %w", err)
	}

	// Return the number of sessions
	return len(sessions.Sessions), nil
}
