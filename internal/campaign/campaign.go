// Package campaign builds and generates D&D 5e campaign plans.
//
// It centralizes the campaign-plan prompt assembly (shared until now by the
// web handlers and the wizard debug endpoint) and the generation call itself.
// The generation runs over llm.Client.Stream instead of a non-streaming
// Complete: the plan is a large structured JSON document (up to 8192
// completion tokens) that routinely takes more than the HTTP client's
// per-request cap to fully generate, and only streaming keeps the connection
// alive while the model works.
package campaign

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/agent"
	"dungeons/internal/llm"
	"dungeons/internal/npc"
	"dungeons/internal/tarot"
)

// Generation budget. The context timeout is the caller's responsibility
// (see DefaultTimeout); streaming keeps the HTTP connection alive between
// chunks so the shared client's per-request cap no longer bounds the total
// generation time.
const (
	// DefaultTimeout is the recommended wall-clock budget for one campaign
	// plan generation (prompt build + streaming + parse). 10 minutes per the
	// 24/09/2026 decision: Opus 5.5 at medium reasoning effort occasionally
	// needs more than 5 minutes for a complete plan.
	DefaultTimeout = 10 * time.Minute
	// MaxCompletionTokens caps the model output for one plan. Reasoning
	// tokens count toward this budget and vary per run: Sonnet 5 burned all
	// of 8K thinking (empty output) and later runs exceeded 16K mid-JSON, so
	// the budget must cover worst-case reasoning (measured 7-12K) plus the
	// ~19K-char plan (docs/issues-2026-09-24.md).
	MaxCompletionTokens = 32768

	// DefaultReasoningEffort constrains the thinking budget of reasoning
	// models. Opus 5.5 at its provider-default (high) effort exceeded the
	// 5-minute generation budget; "medium" keeps a campaign plan within it.
	DefaultReasoningEffort = "medium"
)

// Request is a fully assembled campaign-plan generation request. Build it
// with BuildRequest, set Model, then hand it to Generate.
type Request struct {
	// Model is the OpenRouter model ID used for generation.
	Model string
	// System is the dungeon-master persona system prompt.
	System string
	// Prompt is the full user prompt (theme, duration, JSON template...).
	Prompt string
	// WorldResources carries the optional world map image for multimodal
	// requests; nil when no map is available.
	WorldResources *agent.WorldResources
	// Duration is the resolved adventure duration preset.
	Duration adventure.AdventureDuration
	// ReasoningEffort constrains the thinking budget of reasoning models
	// (empty = DefaultReasoningEffort).
	ReasoningEffort string
}

// BuildRequest assembles the system persona + the full user prompt for the
// campaign-plan generator, including the optional fortune-teller brief and
// world-coherence guidance. It performs no API call and writes nothing to
// disk; adv may be an unsaved in-memory adventure.
func BuildRequest(adv *adventure.Adventure, theme, duration, adventureType string, brief *tarot.CreativeBrief) (*Request, error) {
	// Load DM persona
	personaLoader := agent.NewPersonaLoader()
	_, dmPersona, err := personaLoader.LoadWithMetadata("dungeon-master")
	if err != nil {
		return nil, fmt.Errorf("failed to load DM persona: %w", err)
	}

	// If the form left the adventure type unset, let the reading decide.
	if adventureType == "" && brief != nil && brief.AdventureType != "" {
		adventureType = brief.AdventureType
	}

	// Get duration config
	dur := adventure.GetDuration(duration)

	// Build session list for JSON example
	actsJSON := buildActsJSONTemplate(dur)
	pacingJSON := buildPacingJSONTemplate(dur)

	// Build adventure type section
	typeSection := ""
	if advType := adventure.GetAdventureType(adventureType); advType != nil {
		typeSection = fmt.Sprintf("\n%s\n", advType.PromptGuide)
	}

	// Build NPC section based on duration. The role/race value lists are
	// generated from the npc taxonomy enums (single source of truth).
	npcSection := fmt.Sprintf(`
Génère %d à %d PNJ dans le tableau "npcs" de "plot_elements". Chaque PNJ doit avoir :
- "name" : nom complet (en français)
- "role" : l'une de ces valeurs EXACTES : %s
- "race" : l'une de ces valeurs EXACTES : %s
- "gender" : "m" ou "f"
- "occupation" : métier en français, varié (ex. marchand, garde, noble, artisan, aubergiste, prêtre, capitaine de navire, contrebandier, mercenaire, érudit, forgeron, herboriste, batelier, mendiant, espion, chasseur, fermier, scribe, ménestrel, mineur)
- "attitude" : valeur EXACTE "positive", "neutral" ou "negative"
- "motivation" : ce qui anime ce PNJ (en français)
- "secret" : une vérité cachée sur ce PNJ (mondaine, pas surnaturelle, en français)
- "narrative_context" : où et quand les joueurs le rencontrent pour la première fois (en français)
- "narrative_integration" : {"introduction_session": N, "plot_role": "description en français", "linked_to_act": N}

IMPORTANT — DIVERSITÉ : varie les races, genres et occupations des PNJ selon la région, le ton et le rôle de chacun, et ancre chaque métier dans le contexte de cette aventure. N'applique PAS de casting par défaut : évite en particulier les réflexes "informateur = contrebandier halfelin" et "rival = capitaine de la garde". Les exemples de structure ci-dessous ne sont QUE des gabarits de format : n'en recopie ni les races, ni les genres, ni les occupations.

L'antagoniste de plot_elements.antagonist DOIT aussi figurer dans le tableau npcs avec le role "antagoniste".`,
		dur.MaxNPCs-2, dur.MaxNPCs, npc.RoleEnumList(), npc.RaceEnumList())

	// Load world resources for geography context
	worldResources := agent.LoadWorldResources()

	// Build world geography section for the prompt
	worldGeographySection := ""
	if worldResources != nil && worldResources.MapDescription != "" {
		worldGeographySection = fmt.Sprintf(`
## Référence géographique du monde

Utilise cette description géographique des Quatre Royaumes pour situer l'aventure dans des lieux cohérents, avec des distances, routes commerciales et frontières exactes :

%s

`, worldResources.MapDescription)
	}

	// Append the fortune-teller's creative brief + world-coherence guidance.
	// It rides on the same %s placeholder as the geography section, so the
	// format string and argument list below stay untouched.
	if brief != nil {
		worldGeographySection += brief.PromptSection()
	}

	// Build prompt
	prompt := fmt.Sprintf(`Génère un plan de campagne D&D 5e pour cette aventure :

**Nom de l'aventure** : %s
**Description** : %s
**Thème** : %s
**Durée** : %d à %d sessions de 3 heures chacune
**Nombre d'actes** : %d
%s
Crée un plan de campagne comprenant :
1. Un titre de campagne et un objectif captivant
2. %d acte(s) avec titres, descriptions, événements clés et objectifs
3. Un antagoniste principal à la motivation MONDAINE et à l'arc narratif
4. %d à %d lieux clés avec niveaux de danger
5. 0 à %d présages (foreshadows) liés aux actes (seulement si la durée le permet) — chaque présage du tableau "active" suit EXACTEMENT la structure d'exemple ci-dessous (ids de la forme "fsh_001", "fsh_002"... ; "resolved" et "abandoned" restent vides à la création)
6. %d à %d PNJ aux rôles définis

%s

%s

%s
IMPÉRATIF : Réponds UNIQUEMENT par du JSON valide respectant EXACTEMENT cette structure (aucun markdown, aucune explication).
Conserve les CLÉS JSON en anglais (ex. "narrative_structure", "role", "race", "status"...) et les valeurs d'énumération techniques de "status"/"attitude"/"gender" et le "role" de l'antagoniste en anglais ("pending", "primary", "positive", "neutral", "negative", "m", "f"). Les valeurs de "role" et "race" des PNJ sont en FRANÇAIS (voir les listes EXACTES ci-dessus). Rédige en FRANÇAIS tout le reste (titres, descriptions, motivations, etc.) :
{
  "version": "1.0.0",
  "metadata": {
    "campaign_title": "Titre de la campagne",
    "theme": "Description du thème",
    "target_duration": {"sessions": %d, "hours_per_session": 3},
    "created_at": "2026-02-14T12:00:00Z",
    "generated_by": "dungeon-master",
    "last_updated": "2026-02-14T12:00:00Z"
  },
  "narrative_structure": {
    "objective": "Objectif principal de la campagne",
    "hook": "Accroche d'ouverture qui happe les joueurs",
    "acts": %s,
    "climax": {
      "description": "La confrontation décisive",
      "target_session": %d,
      "stakes": "Ce qui est en jeu en cas d'échec des héros"
    },
    "resolution": {
      "success_scenario": "Ce qui se passe si les héros réussissent",
      "failure_scenario": "Ce qui se passe si les héros échouent",
      "epilogue_notes": "Comment l'histoire se conclut"
    }
  },
  "plot_elements": {
    "antagonist": {
      "name": "Nom de l'antagoniste",
      "role": "primary",
      "motivation": "Pourquoi il agit ainsi (motivation MONDAINE)",
      "introduction_session": 1,
      "final_confrontation_session": %d,
      "arc": "Comment il évolue"
    },
    "secondary_antagonists": [],
    "supporting_characters": [],
    "macguffins": [],
    "key_locations": [],
    "npcs": [
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "donneur_de_quete",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "positive",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 1, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      },
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "informateur",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "neutral",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 2, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      },
      {
        "name": "<nom complet, inventé pour cette aventure>",
        "role": "rival",
        "race": "<race parmi la liste EXACTE ci-dessus>",
        "gender": "m ou f",
        "occupation": "<métier cohérent avec le rôle et la région>",
        "attitude": "negative",
        "motivation": "<ce qui anime ce PNJ>",
        "secret": "<sa vérité cachée, mondaine>",
        "narrative_context": "<où et quand les joueurs le rencontrent>",
        "narrative_integration": {"introduction_session": 2, "plot_role": "<rôle dans l'intrigue>", "linked_to_act": 1}
      }
    ]
  },
  "foreshadows": {
    "active": [
      {
        "id": "fsh_001",
        "description": "Un détail concret planté devant les joueurs, sans révélation prématurée",
        "introduced_session": 2,
        "introduced_act": 1,
        "expected_payoff_act": 2,
        "linked_element": "Élément de l'intrigue auquel ce présage se rattache (antagoniste, PNJ, lieu, objet...)"
      }
    ],
    "resolved": [],
    "abandoned": [],
    "next_id": 1
  },
  "progression": {
    "current_act": 1,
    "current_session": 0,
    "completed_plot_points": [],
    "active_threads": [],
    "pending_resolutions": []
  },
  "pacing": %s,
  "dm_notes": {
    "themes": ["thème1", "thème2"],
    "tone": "Dark fantasy teintée d'espoir",
    "player_agency": "Notes sur les choix des joueurs",
    "memorable_moments": []
  }
}`,
		adv.Name, adv.Description, theme,
		dur.MinSessions, dur.MaxSessions,
		dur.Acts,
		typeSection,
		dur.Acts,
		dur.MaxLocations-2, dur.MaxLocations,
		dur.MaxForeshadows,
		dur.MaxNPCs-2, dur.MaxNPCs,
		adventure.GetAntiCultConstraints(),
		npcSection,
		worldGeographySection,
		(dur.MinSessions+dur.MaxSessions)/2,
		actsJSON,
		dur.MaxSessions,
		dur.MaxSessions,
		pacingJSON,
	)

	return &Request{
		System:         dmPersona,
		Prompt:         prompt,
		WorldResources: worldResources,
		Duration:       dur,
	}, nil
}

// LLMRequest converts the built request into the neutral llm.Request with
// the campaign-plan JSON response format. Structured JSON output; providers
// without response-format support fail routing loudly (RequireParameters)
// instead of silently returning prose.
func (r *Request) LLMRequest() llm.Request {
	userMessage := llm.UserMessage(r.Prompt)
	if r.WorldResources != nil && r.WorldResources.MapImageBase64 != "" {
		userMessage = llm.UserMessageWithImage(r.Prompt, llm.ImagePart{
			MediaType: r.WorldResources.MapImageMediaType,
			Base64:    r.WorldResources.MapImageBase64,
		})
	}
	return llm.Request{
		Model:    r.Model,
		System:   r.System,
		Messages: []llm.Message{userMessage},
		// Structured JSON output; the tolerant fence-stripping parsing in
		// ParsePlanJSON stays as the fallback for endpoints without
		// response-format support.
		MaxCompletionTokens: MaxCompletionTokens,
		JSONResponse: &llm.JSONResponseFormat{
			Name:   "campaign_plan",
			Strict: false,
		},
		RequireParameters: true,
		// Keep reasoning models inside the 5-minute generation budget
		// (docs/issues-2026-09-24.md): provider-default effort made Opus
		// 5.5 think past every deadline.
		ReasoningEffort: r.reasoningEffort(),
	}
}

// reasoningEffort resolves the configured effort against the default.
func (r *Request) reasoningEffort() string {
	if r.ReasoningEffort != "" {
		return r.ReasoningEffort
	}
	return DefaultReasoningEffort
}

// Result reports one generation run: the parsed plan plus the raw model
// response and timing, so callers (web UI, benchmark CLI) can display usage
// and cost without re-deriving them.
type Result struct {
	// Plan is the parsed campaign plan.
	Plan *adventure.CampaignPlan
	// Response is the accumulated model response (usage, routed model...).
	Response llm.Response
	// Elapsed is the total wall-clock time of the generation call.
	Elapsed time.Duration
	// TimeToFirstDelta is the wall-clock time until the first streamed text
	// fragment arrived; zero when the stream produced no text.
	TimeToFirstDelta time.Duration
}

// ParseError carries the raw model output alongside a JSON parse failure so
// callers (benchmark CLI, logs) can inspect what the model actually returned.
type ParseError struct {
	// FinishReason is the stream's terminal finish reason.
	FinishReason llm.FinishReason
	// Raw is the complete raw model output.
	Raw string
	// Err is the underlying unmarshal error.
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse campaign plan (finish=%q, %d chars): %v",
		e.FinishReason, len(e.Raw), e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// Generate streams a campaign plan from the model and parses the JSON.
// obs may be nil; when set it receives every text delta (progress display,
// benchmark tooling). The ctx deadline (see DefaultTimeout) bounds the whole
// run.
func Generate(ctx context.Context, client llm.Client, req *Request, obs llm.StreamObserver) (*Result, error) {
	if client == nil {
		return nil, fmt.Errorf("llm client unavailable (OPENROUTER_API_KEY not set)")
	}

	start := time.Now()
	wrapped := &firstDeltaObserver{inner: obs, start: start}
	resp, err := client.Stream(ctx, req.LLMRequest(), wrapped)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	elapsed := time.Since(start)

	if resp.Message.Text == "" {
		return nil, fmt.Errorf("empty response from model (finish reason %q, usage: %d completion tokens — reasoning models can exhaust the output budget without emitting content)",
			resp.FinishReason, resp.Usage.CompletionTokens)
	}

	plan, err := ParsePlanJSON(resp.Message.Text)
	if err != nil {
		return nil, &ParseError{FinishReason: resp.FinishReason, Raw: resp.Message.Text, Err: err}
	}

	return &Result{
		Plan:             plan,
		Response:         resp,
		Elapsed:          elapsed,
		TimeToFirstDelta: wrapped.firstDelta(),
	}, nil
}

// ParsePlanJSON extracts a CampaignPlan from raw model output, tolerating
// markdown code fences around the JSON payload.
func ParsePlanJSON(text string) (*adventure.CampaignPlan, error) {
	jsonText := strings.TrimSpace(text)
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimPrefix(jsonText, "```")
	jsonText = strings.TrimSuffix(jsonText, "```")
	jsonText = strings.TrimSpace(jsonText)

	var plan adventure.CampaignPlan
	if err := json.Unmarshal([]byte(jsonText), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// firstDeltaObserver wraps an optional StreamObserver and records when the
// first text delta arrived relative to a start time.
type firstDeltaObserver struct {
	inner llm.StreamObserver
	start time.Time

	mu    sync.Mutex
	first time.Duration
	seen  bool
}

func (o *firstDeltaObserver) OnTextDelta(text string) {
	o.mu.Lock()
	if !o.seen {
		o.seen = true
		o.first = time.Since(o.start)
	}
	o.mu.Unlock()
	if o.inner != nil {
		o.inner.OnTextDelta(text)
	}
}

func (o *firstDeltaObserver) firstDelta() time.Duration {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.first
}

// buildActsJSONTemplate returns the acts JSON template adapted to the duration.
func buildActsJSONTemplate(dur adventure.AdventureDuration) string {
	if dur.Acts == 1 {
		sessions := "["
		for i := 1; i <= dur.MaxSessions; i++ {
			if i > 1 {
				sessions += ","
			}
			sessions += fmt.Sprintf("%d", i)
		}
		sessions += "]"
		return fmt.Sprintf(`[
      {
        "number": 1,
        "title": "Titre de l'acte 1",
        "description": "Ce qui se passe dans cet acte",
        "target_sessions": %s,
        "status": "pending",
        "key_events": ["Événement 1", "Événement 2"],
        "goals": ["Objectif 1", "Objectif 2"],
        "completion_criteria": {"milestone": "Ce qui marque la fin de l'acte"}
      }
    ]`, sessions)
	}

	// 3 acts for campaign
	sessPerAct := dur.MaxSessions / 3
	remainder := dur.MaxSessions % 3

	acts := "[\n"
	start := 1
	for a := 1; a <= 3; a++ {
		count := sessPerAct
		if a <= remainder {
			count++
		}
		sessions := "["
		for i := start; i < start+count; i++ {
			if i > start {
				sessions += ","
			}
			sessions += fmt.Sprintf("%d", i)
		}
		sessions += "]"
		if a > 1 {
			acts += ",\n"
		}
		acts += fmt.Sprintf(`      {
        "number": %d,
        "title": "Titre de l'acte %d",
        "description": "Ce qui se passe dans cet acte",
        "target_sessions": %s,
        "status": "pending",
        "key_events": ["Événement 1"],
        "goals": ["Objectif 1"],
        "completion_criteria": {"milestone": "Ce qui marque la fin de l'acte"}
      }`, a, a, sessions)
		start += count
	}
	acts += "\n    ]"
	return acts
}

// buildPacingJSONTemplate returns the pacing JSON template adapted to the duration.
func buildPacingJSONTemplate(dur adventure.AdventureDuration) string {
	avg := (dur.MinSessions + dur.MaxSessions) / 2

	if dur.Acts == 1 {
		return fmt.Sprintf(`{
    "sessions_played": 0,
    "sessions_remaining_estimate": %d,
    "act_breakdown": {
      "act_1": {"planned": %d, "actual": 0, "variance": 0}
    }
  }`, avg, avg)
	}

	sessPerAct := avg / 3
	return fmt.Sprintf(`{
    "sessions_played": 0,
    "sessions_remaining_estimate": %d,
    "act_breakdown": {
      "act_1": {"planned": %d, "actual": 0, "variance": 0},
      "act_2": {"planned": %d, "actual": 0, "variance": 0},
      "act_3": {"planned": %d, "actual": 0, "variance": 0}
    }
  }`, avg, sessPerAct, sessPerAct, avg-2*sessPerAct)
}
