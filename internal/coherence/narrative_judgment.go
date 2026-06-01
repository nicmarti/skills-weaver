package coherence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dungeons/internal/adventure"
)

// narrativeJudgmentFile is the cache filename for the AI narrative judgment.
const narrativeJudgmentFile = "coherence-narrative.json"

// NarrativeJudgment is the AI judgment produced on top of the NarrativeBrief by
// the three analytical lenses (world / rules / scenario) plus a light synthesis.
// It is assembled by the narrativeai orchestrator (which calls the LLM agents)
// and cached here; this package only defines its shape and persistence so it
// stays free of any agent/LLM dependency.
type NarrativeJudgment struct {
	GeneratedAt    time.Time     `json:"generated_at"`
	PlayedSessions int           `json:"played_sessions"` // how far the campaign had progressed when judged
	Lenses         []LensVerdict `json:"lenses"`
	Synthesis      string        `json:"synthesis"`
}

// LensVerdict is one analytical lens's prose assessment.
type LensVerdict struct {
	Lens       string `json:"lens"`       // human label, e.g. "Monde", "Règles", "Scénario"
	Agent      string `json:"agent"`      // backing persona, e.g. "world-keeper"
	Assessment string `json:"assessment"` // the agent's prose analysis
}

// Stale reports whether the judgment predates the given number of played
// sessions, i.e. new sessions were played since it was generated.
func (j *NarrativeJudgment) Stale(currentPlayedSessions int) bool {
	return j == nil || j.PlayedSessions < currentPlayedSessions
}

// LoadNarrativeJudgment reads the cached judgment, returning (nil, nil) if none
// exists yet.
func LoadNarrativeJudgment(adv *adventure.Adventure) (*NarrativeJudgment, error) {
	path := filepath.Join(adv.BasePath(), narrativeJudgmentFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", narrativeJudgmentFile, err)
	}
	var j NarrativeJudgment
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", narrativeJudgmentFile, err)
	}
	return &j, nil
}

// SaveNarrativeJudgment writes the judgment to the adventure's cache file.
func SaveNarrativeJudgment(adv *adventure.Adventure, j *NarrativeJudgment) error {
	path := filepath.Join(adv.BasePath(), narrativeJudgmentFile)
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling narrative judgment: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", narrativeJudgmentFile, err)
	}
	return nil
}
