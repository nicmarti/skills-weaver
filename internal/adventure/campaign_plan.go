package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CampaignPlan represents the complete narrative structure of a campaign.
type CampaignPlan struct {
	Version            string             `json:"version"`
	Metadata           CampaignMetadata   `json:"metadata"`
	NarrativeStructure NarrativeStructure `json:"narrative_structure"`
	PlotElements       PlotElements       `json:"plot_elements"`
	Foreshadows        ForeshadowsContainer `json:"foreshadows"`
	Progression        Progression        `json:"progression"`
	Pacing             Pacing             `json:"pacing"`
	DMNotes            DMNotes            `json:"dm_notes"`
}

// CampaignMetadata holds basic information about the campaign.
type CampaignMetadata struct {
	CampaignTitle  string         `json:"campaign_title"`
	Theme          string         `json:"theme"`
	TargetDuration TargetDuration `json:"target_duration,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	GeneratedBy    string         `json:"generated_by"`
	LastUpdated    time.Time      `json:"last_updated"`
}

// TargetDuration specifies the planned campaign length.
type TargetDuration struct {
	Sessions        int `json:"sessions"`
	HoursPerSession int `json:"hours_per_session"`
}

// NarrativeStructure defines the story arc.
type NarrativeStructure struct {
	Objective  string      `json:"objective"`
	Hook       string      `json:"hook"`
	Acts       []Act       `json:"acts"`
	Climax     Climax      `json:"climax"`
	Resolution Resolution  `json:"resolution"`
}

// Act represents a major story arc (typically 3 acts per campaign).
type Act struct {
	Number             int                `json:"number"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	TargetSessions     []int              `json:"target_sessions"`
	Status             string             `json:"status"` // pending|in_progress|completed
	KeyEvents          []string           `json:"key_events"`
	Goals              []string           `json:"goals"`
	CompletionCriteria CompletionCriteria `json:"completion_criteria"`
}

// CompletionCriteria defines when an act is complete.
type CompletionCriteria struct {
	Milestone   string     `json:"milestone"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Climax represents the campaign's climactic moment.
type Climax struct {
	Description   string `json:"description"`
	TargetSession int    `json:"target_session"`
	Stakes        string `json:"stakes"`
}

// Resolution defines potential endings.
type Resolution struct {
	SuccessScenario string `json:"success_scenario"`
	FailureScenario string `json:"failure_scenario"`
	EpilogueNotes   string `json:"epilogue_notes,omitempty"`
}

// PlotElements contains NPCs, locations, and MacGuffins.
type PlotElements struct {
	Antagonist           Character   `json:"antagonist"`
	SecondaryAntagonists []Character `json:"secondary_antagonists,omitempty"`
	SupportingCharacters []Character `json:"supporting_characters,omitempty"`
	MacGuffins           []MacGuffin `json:"macguffins,omitempty"`
	KeyLocations         []Location  `json:"key_locations,omitempty"`
}

// Character represents an NPC with a story arc.
type Character struct {
	Name                   string `json:"name"`
	Role                   string `json:"role"` // primary|mastermind|ally|rival
	Motivation             string `json:"motivation"`
	IntroductionSession    int    `json:"introduction_session,omitempty"`
	FinalConfrontationSession int `json:"final_confrontation_session,omitempty"`
	Arc                    string `json:"arc,omitempty"`
	Mystery                bool   `json:"mystery,omitempty"`
	RevealSession          int    `json:"reveal_session,omitempty"`
	KeySessions            []int  `json:"key_sessions,omitempty"`
}

// MacGuffin represents an important object or artifact.
type MacGuffin struct {
	Name              string `json:"name"`
	Type              string `json:"type"` // artifact|treasure|knowledge
	Significance      string `json:"significance"`
	IntroducedSession int    `json:"introduced_session"`
	Resolution        string `json:"resolution"`
}

// Location represents a key campaign location.
type Location struct {
	Name        string `json:"name"`
	Kingdom     string `json:"kingdom,omitempty"`
	Type        string `json:"type"` // city|ruins|dungeon|wilderness
	Role        string `json:"role"` // Act 1 hub|Act 3 destination
	Sessions    []int  `json:"sessions,omitempty"`
	DangerLevel string `json:"danger_level,omitempty"` // safe|moderate|dangerous|extreme
}

// ForeshadowsContainer organizes foreshadows by status with act linkage.
type ForeshadowsContainer struct {
	Active    []ForeshadowLinked `json:"active"`
	Resolved  []ForeshadowLinked `json:"resolved"`
	Abandoned []ForeshadowLinked `json:"abandoned"`
	NextID    int                `json:"next_id"`
}

// ForeshadowLinked extends Foreshadow with campaign plan linkage.
type ForeshadowLinked struct {
	ID                  string             `json:"id"`
	Description         string             `json:"description"`
	PlantedAt           time.Time          `json:"planted_at"`
	PlantedSession      int                `json:"planted_session"`
	Importance          Importance         `json:"importance"`
	Category            ForeshadowCategory `json:"category"`
	Tags                []string           `json:"tags,omitempty"`
	Context             string             `json:"context,omitempty"`
	RelatedNPCs         []string           `json:"related_npcs,omitempty"`
	RelatedLocations    []string           `json:"related_locations,omitempty"`
	LinkedToAct         int                `json:"linked_to_act"`
	LinkedToPlotPoint   string             `json:"linked_to_plot_point,omitempty"`
	TargetPayoffSession int                `json:"target_payoff_session,omitempty"`
	PayoffType          string             `json:"payoff_type,omitempty"` // revelation|resolution|twist
	ResolvedAt          *time.Time         `json:"resolved_at,omitempty"`
	ResolutionNotes     string             `json:"resolution_notes,omitempty"`
}

// Progression tracks the campaign's current state.
type Progression struct {
	CurrentAct           int      `json:"current_act"`
	CurrentSession       int      `json:"current_session"`
	CompletedPlotPoints  []string `json:"completed_plot_points"`
	ActiveThreads        []string `json:"active_threads"`
	PendingResolutions   []string `json:"pending_resolutions"` // Foreshadow IDs
}

// Pacing tracks planned vs actual session counts.
type Pacing struct {
	SessionsPlayed           int                 `json:"sessions_played"`
	SessionsRemainingEstimate int                `json:"sessions_remaining_estimate"`
	ActBreakdown             map[string]ActPacing `json:"act_breakdown"`
}

// ActPacing compares planned vs actual sessions for an act.
type ActPacing struct {
	Planned  int `json:"planned"`
	Actual   int `json:"actual"`
	Variance int `json:"variance"`
}

// DMNotes holds private DM observations.
type DMNotes struct {
	Themes           []string `json:"themes,omitempty"`
	Tone             string   `json:"tone,omitempty"`
	PlayerAgency     string   `json:"player_agency,omitempty"`
	MemorableMoments []string `json:"memorable_moments,omitempty"`
}

// LoadCampaignPlan loads the campaign plan from disk.
func (a *Adventure) LoadCampaignPlan() (*CampaignPlan, error) {
	path := filepath.Join(a.basePath, "campaign-plan.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil // No campaign plan exists (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("reading campaign-plan.json: %w", err)
	}

	var plan CampaignPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parsing campaign-plan.json: %w", err)
	}

	return &plan, nil
}

// SaveCampaignPlan saves the campaign plan to disk.
func (a *Adventure) SaveCampaignPlan(plan *CampaignPlan) error {
	plan.Metadata.LastUpdated = time.Now()
	path := filepath.Join(a.basePath, "campaign-plan.json")
	return a.saveJSON(path, plan)
}

// GetCurrentAct returns the currently active act.
func (cp *CampaignPlan) GetCurrentAct() *Act {
	for i := range cp.NarrativeStructure.Acts {
		if cp.NarrativeStructure.Acts[i].Number == cp.Progression.CurrentAct {
			return &cp.NarrativeStructure.Acts[i]
		}
	}
	return nil
}

// GetCriticalForeshadows returns foreshadows that are >= 3 sessions old and critical/major importance.
func (cp *CampaignPlan) GetCriticalForeshadows() []ForeshadowLinked {
	currentSession := cp.Progression.CurrentSession
	var critical []ForeshadowLinked

	for _, f := range cp.Foreshadows.Active {
		age := currentSession - f.PlantedSession
		if age >= 3 && (f.Importance == ImportanceCritical || f.Importance == ImportanceMajor) {
			critical = append(critical, f)
		}
	}

	return critical
}

// AdvanceAct marks current act as completed and moves to next act.
func (cp *CampaignPlan) AdvanceAct(actNumber int) error {
	if actNumber < 1 || actNumber > len(cp.NarrativeStructure.Acts) {
		return fmt.Errorf("invalid act number: %d", actNumber)
	}

	// Mark previous act as completed
	if actNumber > 1 {
		for i := range cp.NarrativeStructure.Acts {
			if cp.NarrativeStructure.Acts[i].Number == actNumber-1 {
				now := time.Now()
				cp.NarrativeStructure.Acts[i].Status = "completed"
				cp.NarrativeStructure.Acts[i].CompletionCriteria.CompletedAt = &now
			}
		}
	}

	// Set new current act
	cp.Progression.CurrentAct = actNumber
	for i := range cp.NarrativeStructure.Acts {
		if cp.NarrativeStructure.Acts[i].Number == actNumber {
			cp.NarrativeStructure.Acts[i].Status = "in_progress"
		}
	}

	return nil
}

// CompletePlotPoint marks a plot point as completed.
func (cp *CampaignPlan) CompletePlotPoint(plotPointID string) error {
	// Check if already completed
	for _, id := range cp.Progression.CompletedPlotPoints {
		if id == plotPointID {
			return fmt.Errorf("plot point %s already completed", plotPointID)
		}
	}

	cp.Progression.CompletedPlotPoints = append(cp.Progression.CompletedPlotPoints, plotPointID)
	return nil
}

// AddActiveThread adds a new narrative thread to track.
func (cp *CampaignPlan) AddActiveThread(name string) error {
	// Check if already exists
	for _, thread := range cp.Progression.ActiveThreads {
		if thread == name {
			return fmt.Errorf("thread %s already active", name)
		}
	}

	cp.Progression.ActiveThreads = append(cp.Progression.ActiveThreads, name)
	return nil
}

// RemoveActiveThread removes a resolved narrative thread.
func (cp *CampaignPlan) RemoveActiveThread(name string) error {
	for i, thread := range cp.Progression.ActiveThreads {
		if thread == name {
			cp.Progression.ActiveThreads = append(cp.Progression.ActiveThreads[:i], cp.Progression.ActiveThreads[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("thread %s not found", name)
}

// PlantForeshadowLinked creates a new foreshadow with campaign plan linkage.
func (cp *CampaignPlan) PlantForeshadowLinked(foreshadow *ForeshadowLinked) error {
	// Generate ID if not set
	if foreshadow.ID == "" {
		foreshadow.ID = fmt.Sprintf("fsh_%03d", cp.Foreshadows.NextID)
		cp.Foreshadows.NextID++
	}

	// Set planted time if not set
	if foreshadow.PlantedAt.IsZero() {
		foreshadow.PlantedAt = time.Now()
	}

	// Set planted session if not set
	if foreshadow.PlantedSession == 0 {
		foreshadow.PlantedSession = cp.Progression.CurrentSession
	}

	cp.Foreshadows.Active = append(cp.Foreshadows.Active, *foreshadow)
	cp.Progression.PendingResolutions = append(cp.Progression.PendingResolutions, foreshadow.ID)

	return nil
}

// ResolveForeshadowLinked marks a foreshadow as resolved.
func (cp *CampaignPlan) ResolveForeshadowLinked(id string, notes string) error {
	for i, f := range cp.Foreshadows.Active {
		if f.ID == id {
			now := time.Now()
			f.ResolvedAt = &now
			f.ResolutionNotes = notes

			// Move from active to resolved
			cp.Foreshadows.Resolved = append(cp.Foreshadows.Resolved, f)
			cp.Foreshadows.Active = append(cp.Foreshadows.Active[:i], cp.Foreshadows.Active[i+1:]...)

			// Remove from pending resolutions
			for j, pendingID := range cp.Progression.PendingResolutions {
				if pendingID == id {
					cp.Progression.PendingResolutions = append(cp.Progression.PendingResolutions[:j], cp.Progression.PendingResolutions[j+1:]...)
					break
				}
			}

			return nil
		}
	}

	return fmt.Errorf("foreshadow %s not found", id)
}

// AbandonForeshadowLinked marks a foreshadow as abandoned.
func (cp *CampaignPlan) AbandonForeshadowLinked(id string, reason string) error {
	for i, f := range cp.Foreshadows.Active {
		if f.ID == id {
			now := time.Now()
			f.ResolvedAt = &now
			f.ResolutionNotes = fmt.Sprintf("Abandoned: %s", reason)

			// Move from active to abandoned
			cp.Foreshadows.Abandoned = append(cp.Foreshadows.Abandoned, f)
			cp.Foreshadows.Active = append(cp.Foreshadows.Active[:i], cp.Foreshadows.Active[i+1:]...)

			// Remove from pending resolutions
			for j, pendingID := range cp.Progression.PendingResolutions {
				if pendingID == id {
					cp.Progression.PendingResolutions = append(cp.Progression.PendingResolutions[:j], cp.Progression.PendingResolutions[j+1:]...)
					break
				}
			}

			return nil
		}
	}

	return fmt.Errorf("foreshadow %s not found", id)
}

// GetForeshadow returns a specific foreshadow by ID (searches all lists).
func (cp *CampaignPlan) GetForeshadow(id string) *ForeshadowLinked {
	// Search active
	for _, f := range cp.Foreshadows.Active {
		if f.ID == id {
			return &f
		}
	}

	// Search resolved
	for _, f := range cp.Foreshadows.Resolved {
		if f.ID == id {
			return &f
		}
	}

	// Search abandoned
	for _, f := range cp.Foreshadows.Abandoned {
		if f.ID == id {
			return &f
		}
	}

	return nil
}

// UpdatePacing recalculates pacing metrics based on current state.
func (cp *CampaignPlan) UpdatePacing() {
	cp.Pacing.SessionsPlayed = cp.Progression.CurrentSession

	// Calculate sessions remaining estimate
	totalPlanned := 0
	if cp.Metadata.TargetDuration.Sessions > 0 {
		totalPlanned = cp.Metadata.TargetDuration.Sessions
	} else {
		// Estimate from acts
		for _, act := range cp.NarrativeStructure.Acts {
			totalPlanned += len(act.TargetSessions)
		}
	}
	cp.Pacing.SessionsRemainingEstimate = totalPlanned - cp.Pacing.SessionsPlayed

	// Update act breakdown
	if cp.Pacing.ActBreakdown == nil {
		cp.Pacing.ActBreakdown = make(map[string]ActPacing)
	}

	for _, act := range cp.NarrativeStructure.Acts {
		planned := len(act.TargetSessions)
		actual := 0

		// Count actual sessions for this act
		if act.Status == "completed" {
			for _, sessionNum := range act.TargetSessions {
				if sessionNum <= cp.Progression.CurrentSession {
					actual++
				}
			}
		} else if act.Status == "in_progress" {
			for _, sessionNum := range act.TargetSessions {
				if sessionNum <= cp.Progression.CurrentSession {
					actual++
				}
			}
		}

		cp.Pacing.ActBreakdown[fmt.Sprintf("act_%d", act.Number)] = ActPacing{
			Planned:  planned,
			Actual:   actual,
			Variance: actual - planned,
		}
	}
}

// AddMemorableMoment adds a memorable moment to DM notes.
func (cp *CampaignPlan) AddMemorableMoment(moment string) {
	cp.DMNotes.MemorableMoments = append(cp.DMNotes.MemorableMoments, moment)
}
