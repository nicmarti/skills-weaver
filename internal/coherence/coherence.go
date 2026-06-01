// Package coherence provides read-only analysis of an adventure's data.
//
// It is the shared analysis engine the council called for: deterministic
// checks produce structured Findings, and thin front-ends (a web route, a
// sw-validate CLI, a JSON feed for Claude Code) render them. The engine never
// mutates adventure data — repairing belongs to sw-adventure repair.
//
// Analysis is organized in layers, each scored independently (there is
// deliberately NO single blended score — a green data-integrity layer must
// never mask a broken narrative layer):
//
//   - integrity: is each file well-formed and are cross-file references valid?
//     (deterministic, implemented here)
//   - drift:     does the played flow match the planned skeleton? (deterministic,
//     planned)
//   - narrative: does the well-formed story actually hold together? (a judgement
//     delegated to an LLM; this package only assembles the evidence, planned)
//
// coherence imports adventure (one-way); adventure must never import coherence.
package coherence

import (
	"time"

	"dungeons/internal/adventure"
)

// Severity classifies how serious a Finding is.
type Severity string

const (
	// SeverityError marks data that is structurally broken or self-contradictory.
	SeverityError Severity = "error"
	// SeverityWarning marks data that is suspect and likely a bug, but not fatal.
	SeverityWarning Severity = "warning"
	// SeverityInfo marks an observation that may be intentional (e.g. deleted entries).
	SeverityInfo Severity = "info"
)

// Layer names. Only integrity is implemented today.
const (
	LayerIntegrity = "integrity"
	LayerDrift     = "drift"
	LayerNarrative = "narrative"
)

// Finding is a single issue detected by a layer.
type Finding struct {
	Layer    string         `json:"layer"`
	Severity Severity       `json:"severity"`
	Rule     string         `json:"rule"`              // stable id, e.g. "journal.duplicate_id"
	Message  string         `json:"message"`           // human-readable, French
	Context  map[string]any `json:"context,omitempty"` // structured detail for machine consumers
}

// LayerReport holds all findings for one layer plus that layer's own score.
type LayerReport struct {
	Layer    string    `json:"layer"`
	Findings []Finding `json:"findings"`
	Score    int       `json:"score"` // 0-100, computed from this layer's findings only
}

// Counts returns the number of findings per severity.
func (lr LayerReport) Counts() (errors, warnings, infos int) {
	for _, f := range lr.Findings {
		switch f.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		case SeverityInfo:
			infos++
		}
	}
	return
}

// HasErrors reports whether the layer contains any error-severity finding.
func (lr LayerReport) HasErrors() bool {
	e, _, _ := lr.Counts()
	return e > 0
}

// Report is the full coherence analysis of an adventure.
type Report struct {
	AdventureName string      `json:"adventure_name"`
	AdventureSlug string      `json:"adventure_slug"`
	GeneratedAt   time.Time   `json:"generated_at"`
	Integrity     LayerReport `json:"integrity"`
	Drift         LayerReport `json:"drift"`
	Narrative     LayerReport `json:"narrative"`
	// NarrativeBrief is the deterministic evidence dossier for an LLM to judge
	// narrative quality (repetition, stagnation, what did not work). The actual
	// AI judgment is wired on top of this in a later iteration.
	NarrativeBrief *NarrativeBrief `json:"narrative_brief,omitempty"`
}

// HasErrors reports whether any layer contains an error-severity finding.
// Front-ends (e.g. a CI gate) can use this for exit codes. The narrative layer
// is deliberately excluded: narrative observations are never errors and must
// not gate a pipeline.
func (r *Report) HasErrors() bool {
	return r.Integrity.HasErrors() || r.Drift.HasErrors()
}

// Analyze runs every implemented layer over an adventure and returns a Report.
func Analyze(adv *adventure.Adventure) (*Report, error) {
	integrity, err := AnalyzeIntegrity(adv)
	if err != nil {
		return nil, err
	}
	drift, err := AnalyzeDrift(adv)
	if err != nil {
		return nil, err
	}
	narrative, err := AnalyzeNarrative(adv)
	if err != nil {
		return nil, err
	}
	brief, err := BuildNarrativeBrief(adv)
	if err != nil {
		return nil, err
	}
	return &Report{
		AdventureName:  adv.Name,
		AdventureSlug:  adv.Slug,
		GeneratedAt:    time.Now(),
		Integrity:      integrity,
		Drift:          drift,
		Narrative:      narrative,
		NarrativeBrief: brief,
	}, nil
}

// scoreFor computes a layer score from its findings. Each severity subtracts a
// fixed weight from 100; the result is floored at 0. The weighting is
// intentionally simple and transparent so the number is explainable.
func scoreFor(findings []Finding) int {
	score := 100
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			score -= 15
		case SeverityWarning:
			score -= 5
		case SeverityInfo:
			score -= 1
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

// newLayerReport assembles a LayerReport, guaranteeing a non-nil Findings slice
// and a computed score.
func newLayerReport(layer string, findings []Finding) LayerReport {
	if findings == nil {
		findings = []Finding{}
	}
	return LayerReport{
		Layer:    layer,
		Findings: findings,
		Score:    scoreFor(findings),
	}
}
