package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CombatState is the canonical "is combat active" flag for an adventure.
//
// It is the single source of truth that gates combat/danger ambient music, so
// audio cues follow the engine state instead of the Dungeon Master's free-form
// prose. Without it, combat music is triggered the moment the DM *decides* there
// is a fight — potentially before the threat is revealed to the player — which
// leaks hidden state through the audio channel (an "audio oracle"). Tying the
// cue to this flag removes that spoiler vector.
type CombatState struct {
	Active    bool      `json:"active"`
	StartedAt time.Time `json:"started_at,omitempty"`
	SessionID int       `json:"session_id,omitempty"` // session in which combat started
}

func (a *Adventure) combatStatePath() string {
	return filepath.Join(a.basePath, "combat-state.json")
}

// LoadCombatState reads the canonical combat state. A missing file means no
// active combat (the default for any adventure that has never fought).
func (a *Adventure) LoadCombatState() (*CombatState, error) {
	data, err := os.ReadFile(a.combatStatePath())
	if os.IsNotExist(err) {
		return &CombatState{Active: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading combat-state.json: %w", err)
	}

	var cs CombatState
	if err := json.Unmarshal(data, &cs); err != nil {
		return nil, fmt.Errorf("parsing combat-state.json: %w", err)
	}
	return &cs, nil
}

// IsCombatActive reports whether combat is currently active. On any read error
// it returns false: the anti-spoil guardrail must fail safe (no combat cue)
// rather than risk emitting combat music outside of combat.
func (a *Adventure) IsCombatActive() bool {
	cs, err := a.LoadCombatState()
	if err != nil {
		return false
	}
	return cs.Active
}

// StartCombat marks combat as canonically active. It is idempotent: calling it
// again while combat is already active preserves the original StartedAt.
func (a *Adventure) StartCombat() (*CombatState, error) {
	cs, err := a.LoadCombatState()
	if err != nil {
		return nil, err
	}
	if cs.Active {
		return cs, nil
	}

	cs.Active = true
	cs.StartedAt = time.Now()
	if sess, err := a.GetCurrentSession(); err == nil && sess != nil {
		cs.SessionID = sess.ID
	}

	if err := a.saveJSON(a.combatStatePath(), cs); err != nil {
		return nil, err
	}
	return cs, nil
}

// EndCombat marks combat as canonically over. It is idempotent.
func (a *Adventure) EndCombat() (*CombatState, error) {
	cs := &CombatState{Active: false}
	if err := a.saveJSON(a.combatStatePath(), cs); err != nil {
		return nil, err
	}
	return cs, nil
}
