// Package tarot implements the "fortune teller" (voyante) deck used by the
// guided adventure-creation wizard. The deck is a static, hand-authored set of
// tarot-style cards. Each card carries hidden parameters that silently shape the
// generated adventure (tone, antagonist, region, danger, pacing) — the player
// never sees these consequences, mirroring Ultima IV's gypsy character creation.
//
// The card->parameters mapping is fully deterministic Go code so it can be unit
// tested. The LLM is only used downstream to turn the assembled brief into a
// campaign plan.
package tarot

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DefaultDeckPath is the on-disk location of the deck definition, shared by the
// web server (mapping) and the sw-image CLI (art generation).
const DefaultDeckPath = "data/tarot-deck.json"

// HiddenParams are the adventure-shaping parameters a card silently contributes.
// Every field is optional; a card only sets the axis its owning question controls.
type HiddenParams struct {
	Tone             string `json:"tone,omitempty"`
	AdventureType    string `json:"adventure_type,omitempty"`
	AntagonistNature string `json:"antagonist_nature,omitempty"`
	Region           string `json:"region,omitempty"`
	DangerLevel      string `json:"danger_level,omitempty"`
	Pacing           string `json:"pacing,omitempty"`
	Stakes           string `json:"stakes,omitempty"`
}

// Card is a single tarot card.
type Card struct {
	ID       string       `json:"id"`       // stable id, also the art filename stem
	Name     string       `json:"name"`     // displayed name, e.g. "Le Soleil"
	Art      string       `json:"art"`      // art filename under web/static/cards/, e.g. "le_soleil.png"
	Question string       `json:"question"` // owning question id
	Flavor   string       `json:"flavor"`   // mystical blurb (shown to player + used as art hint)
	Params   HiddenParams `json:"params"`   // deterministic hidden mapping
}

// Question is one step of the wizard. The player draws a few cards from CardIDs
// and picks one; that choice sets the question's Axis.
type Question struct {
	ID      string   `json:"id"`
	Prompt  string   `json:"prompt"`
	Axis    string   `json:"axis"`
	CardIDs []string `json:"card_ids"`
}

// Deck is the full set of questions and cards.
type Deck struct {
	Questions []Question `json:"questions"`
	Cards     []Card     `json:"cards"`
}

// CreativeBrief is the aggregated result of a reading: the merged hidden
// parameters plus an optional pre-rendered coherence section (set by the web
// layer to avoid a circular dependency). It feeds the campaign-plan prompt.
type CreativeBrief struct {
	Tone             string
	AdventureType    string
	AntagonistNature string
	Region           string
	DangerLevel      string
	Pacing           string
	Stakes           string
	CoherenceText    string // pre-rendered "## World Coherence" section, optional
}

// LoadDeck reads and validates a deck definition from disk.
func LoadDeck(path string) (*Deck, error) {
	if path == "" {
		path = DefaultDeckPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tarot deck %s: %w", path, err)
	}
	var d Deck
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing tarot deck %s: %w", path, err)
	}
	if len(d.Questions) == 0 || len(d.Cards) == 0 {
		return nil, fmt.Errorf("tarot deck %s is empty", path)
	}
	return &d, nil
}

// CardByID returns the card with the given id.
func (d *Deck) CardByID(id string) (*Card, bool) {
	for i := range d.Cards {
		if d.Cards[i].ID == id {
			return &d.Cards[i], true
		}
	}
	return nil, false
}

// QuestionByID returns the question with the given id.
func (d *Deck) QuestionByID(id string) (*Question, bool) {
	for i := range d.Questions {
		if d.Questions[i].ID == id {
			return &d.Questions[i], true
		}
	}
	return nil, false
}

// CardsForQuestion returns every card eligible for a question.
func (d *Deck) CardsForQuestion(qID string) []Card {
	q, ok := d.QuestionByID(qID)
	if !ok {
		return nil
	}
	out := make([]Card, 0, len(q.CardIDs))
	for _, id := range q.CardIDs {
		if c, ok := d.CardByID(id); ok {
			out = append(out, *c)
		}
	}
	return out
}

// Aggregate merges the chosen cards' hidden parameters into one brief.
// Each question controls a distinct axis, so there is normally no conflict; if
// two cards set the same field, the later one in chosenCardIDs wins.
func (d *Deck) Aggregate(chosenCardIDs []string) (CreativeBrief, error) {
	var b CreativeBrief
	for _, id := range chosenCardIDs {
		card, ok := d.CardByID(id)
		if !ok {
			return CreativeBrief{}, fmt.Errorf("unknown card id %q", id)
		}
		p := card.Params
		if p.Tone != "" {
			b.Tone = p.Tone
		}
		if p.AdventureType != "" {
			b.AdventureType = p.AdventureType
		}
		if p.AntagonistNature != "" {
			b.AntagonistNature = p.AntagonistNature
		}
		if p.Region != "" {
			b.Region = p.Region
		}
		if p.DangerLevel != "" {
			b.DangerLevel = p.DangerLevel
		}
		if p.Pacing != "" {
			b.Pacing = p.Pacing
		}
		if p.Stakes != "" {
			b.Stakes = p.Stakes
		}
	}
	return b, nil
}

// French renderings of the short enum tokens, so the brief reads in one language.
var (
	toneFR   = map[string]string{"heroic": "héroïque", "mysterious": "mystérieux", "grim": "sombre", "epic": "épique"}
	dangerFR = map[string]string{"low": "faible", "moderate": "modéré", "high": "élevé", "deadly": "mortel"}
	stakesFR = map[string]string{"personal": "personnel", "local": "local", "regional": "régional", "kingdom": "royaume"}
	pacingFR = map[string]string{"slow_burn": "lent et progressif", "steady": "régulier", "breakneck": "effréné", "varied": "variable"}
)

func frOrRaw(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return key
}

// HasContent reports whether the brief carries any creative direction.
func (b CreativeBrief) HasContent() bool {
	return b.Tone != "" || b.AdventureType != "" || b.AntagonistNature != "" ||
		b.Region != "" || b.DangerLevel != "" || b.Pacing != "" || b.Stakes != "" ||
		strings.TrimSpace(b.CoherenceText) != ""
}

// PromptSection renders the brief as a prompt fragment for the campaign-plan
// generator. Only non-empty fields are included. The coherence text, if any, is
// appended verbatim.
func (b CreativeBrief) PromptSection() string {
	if !b.HasContent() {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n## Lecture de la voyante (contraintes créatives)\n")
	sb.WriteString("Respecte ces contraintes en façonnant la campagne :\n")
	add := func(label, val string) {
		if val != "" {
			fmt.Fprintf(&sb, "- %s : %s\n", label, val)
		}
	}
	add("Ton", frOrRaw(toneFR, b.Tone))
	add("Nature de l'antagoniste", b.AntagonistNature)
	add("Région principale (situe l'aventure ici, en cohérence avec la carte du monde)", b.Region)
	add("Niveau de danger", frOrRaw(dangerFR, b.DangerLevel))
	add("Portée des enjeux", frOrRaw(stakesFR, b.Stakes))
	add("Rythme", frOrRaw(pacingFR, b.Pacing))
	if txt := strings.TrimSpace(b.CoherenceText); txt != "" {
		sb.WriteString("\n")
		sb.WriteString(txt)
		sb.WriteString("\n")
	}
	return sb.String()
}
