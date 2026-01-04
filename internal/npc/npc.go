package npc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dungeons/internal/names"
)

// Appearance holds physical appearance traits.
type Appearance struct {
	Build              string `json:"build"`
	Height             string `json:"height"`
	HairColor          string `json:"hair_color"`
	HairStyle          string `json:"hair_style"`
	EyeColor           string `json:"eye_color"`
	Skin               string `json:"skin"`
	FacialFeature      string `json:"facial_feature"`
	DistinctiveFeature string `json:"distinctive_feature"`
}

// Personality holds personality traits.
type Personality struct {
	TraitPrincipal  string `json:"trait_principal"`
	TraitSecondaire string `json:"trait_secondaire"`
	Defaut          string `json:"defaut"`
	Qualite         string `json:"qualite"`
}

// Motivation holds NPC goals, fears, and secrets.
type Motivation struct {
	Goal   string `json:"goal"`
	Fear   string `json:"fear"`
	Secret string `json:"secret"`
}

// Voice holds speaking characteristics.
type Voice struct {
	Tone   string `json:"tone"`
	Manner string `json:"manner"`
}

// NPC represents a non-player character.
type NPC struct {
	Name        string      `json:"name"`
	Race        string      `json:"race"`
	Gender      string      `json:"gender"`
	Occupation  string      `json:"occupation"`
	Appearance  Appearance  `json:"appearance"`
	Personality Personality `json:"personality"`
	Motivation  Motivation  `json:"motivation"`
	Voice       Voice       `json:"voice"`
	Quirk       string      `json:"quirk"`
	Attitude    string      `json:"attitude"`
}

// TraitsData holds all NPC trait data from JSON.
type TraitsData struct {
	Appearance struct {
		Build              []string `json:"build"`
		Height             []string `json:"height"`
		HairColor          []string `json:"hair_color"`
		HairStyle          []string `json:"hair_style"`
		EyeColor           []string `json:"eye_color"`
		Skin               []string `json:"skin"`
		FacialFeature      []string `json:"facial_feature"`
		DistinctiveFeature []string `json:"distinctive_feature"`
	} `json:"appearance"`
	Personality struct {
		TraitPrincipal  []string `json:"trait_principal"`
		TraitSecondaire []string `json:"trait_secondaire"`
		Defaut          []string `json:"defaut"`
		Qualite         []string `json:"qualite"`
	} `json:"personality"`
	Motivation struct {
		Goal   []string `json:"goal"`
		Fear   []string `json:"fear"`
		Secret []string `json:"secret"`
	} `json:"motivation"`
	Occupation struct {
		Commoner   []string `json:"commoner"`
		Skilled    []string `json:"skilled"`
		Authority  []string `json:"authority"`
		Underworld []string `json:"underworld"`
		Religious  []string `json:"religious"`
		Adventurer []string `json:"adventurer"`
	} `json:"occupation"`
	Relationship struct {
		AttitudePositive []string `json:"attitude_positive"`
		AttitudeNeutral  []string `json:"attitude_neutral"`
		AttitudeNegative []string `json:"attitude_negative"`
	} `json:"relationship"`
	Voice struct {
		Tone   []string `json:"tone"`
		Manner []string `json:"manner"`
	} `json:"voice"`
	Quirk []string `json:"quirk"`
}

// Generator generates random NPCs.
type Generator struct {
	traits   *TraitsData
	nameGen  *names.Generator
	rng      *rand.Rand
	dataDir  string
}

// NewGenerator creates a new NPC generator.
func NewGenerator(dataDir string) (*Generator, error) {
	// Load traits data
	path := filepath.Join(dataDir, "npc-traits.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading npc-traits.json: %w", err)
	}

	var traits TraitsData
	if err := json.Unmarshal(data, &traits); err != nil {
		return nil, fmt.Errorf("parsing npc-traits.json: %w", err)
	}

	// Create name generator
	nameGen, err := names.NewGenerator(dataDir)
	if err != nil {
		return nil, fmt.Errorf("creating name generator: %w", err)
	}

	return &Generator{
		traits:  &traits,
		nameGen: nameGen,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		dataDir: dataDir,
	}, nil
}

// Generate creates a random NPC.
func (g *Generator) Generate(opts ...Option) (*NPC, error) {
	cfg := &config{
		name:           "",
		race:           "",
		gender:         "",
		occupationType: "",
		attitude:       "neutral",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Determine race (random if not specified)
	race := cfg.race
	if race == "" {
		races := []string{"human", "dwarf", "elf", "halfling"}
		weights := []int{60, 15, 15, 10} // Humans more common
		race = g.weightedChoice(races, weights)
	}

	// Determine gender (random if not specified)
	gender := cfg.gender
	if gender == "" {
		if g.rng.Intn(2) == 0 {
			gender = "male"
		} else {
			gender = "female"
		}
	} else {
		// Normalize gender value
		switch strings.ToLower(gender) {
		case "m", "male", "masculin", "homme":
			gender = "male"
		case "f", "female", "feminin", "femme":
			gender = "female"
		}
	}

	// Use specified name or generate random one
	name := cfg.name
	if name == "" {
		var err error
		name, err = g.nameGen.GenerateName(race, gender)
		if err != nil {
			return nil, fmt.Errorf("generating name: %w", err)
		}
	}

	// Generate occupation
	occupation := g.generateOccupation(cfg.occupationType)

	// Generate appearance
	appearance := g.generateAppearance(race)

	// Generate personality
	personality := g.generatePersonality()

	// Generate motivation
	motivation := g.generateMotivation()

	// Generate voice
	voice := g.generateVoice()

	// Generate quirk
	quirk := g.randomChoice(g.traits.Quirk)

	// Generate attitude
	attitude := g.generateAttitude(cfg.attitude)

	return &NPC{
		Name:        name,
		Race:        race,
		Gender:      gender,
		Occupation:  occupation,
		Appearance:  appearance,
		Personality: personality,
		Motivation:  motivation,
		Voice:       voice,
		Quirk:       quirk,
		Attitude:    attitude,
	}, nil
}

func (g *Generator) generateOccupation(occType string) string {
	occType = strings.ToLower(occType)

	// Check if occType is a specific occupation (exists in any category)
	allCategories := [][]string{
		g.traits.Occupation.Commoner,
		g.traits.Occupation.Skilled,
		g.traits.Occupation.Authority,
		g.traits.Occupation.Underworld,
		g.traits.Occupation.Religious,
		g.traits.Occupation.Adventurer,
	}

	for _, category := range allCategories {
		for _, occupation := range category {
			if occupation == occType {
				// Found exact match - use it directly
				return occupation
			}
		}
	}

	// Not a specific occupation, treat as category
	var pool []string
	switch occType {
	case "commoner", "roturier":
		pool = g.traits.Occupation.Commoner
	case "skilled", "artisan":
		pool = g.traits.Occupation.Skilled
	case "authority", "autorite":
		pool = g.traits.Occupation.Authority
	case "underworld", "criminel":
		pool = g.traits.Occupation.Underworld
	case "religious", "religieux":
		pool = g.traits.Occupation.Religious
	case "adventurer", "aventurier":
		pool = g.traits.Occupation.Adventurer
	default:
		// Random type with weights
		types := [][]string{
			g.traits.Occupation.Commoner,
			g.traits.Occupation.Skilled,
			g.traits.Occupation.Authority,
			g.traits.Occupation.Underworld,
			g.traits.Occupation.Religious,
			g.traits.Occupation.Adventurer,
		}
		weights := []int{40, 25, 10, 10, 10, 5}
		idx := g.weightedIndex(weights)
		pool = types[idx]
	}

	return g.randomChoice(pool)
}

func (g *Generator) generateAppearance(race string) Appearance {
	app := Appearance{
		Build:              g.randomChoice(g.traits.Appearance.Build),
		Height:             g.randomChoice(g.traits.Appearance.Height),
		HairColor:          g.randomChoice(g.traits.Appearance.HairColor),
		HairStyle:          g.randomChoice(g.traits.Appearance.HairStyle),
		EyeColor:           g.randomChoice(g.traits.Appearance.EyeColor),
		Skin:               g.randomChoice(g.traits.Appearance.Skin),
		FacialFeature:      g.randomChoice(g.traits.Appearance.FacialFeature),
		DistinctiveFeature: g.randomChoice(g.traits.Appearance.DistinctiveFeature),
	}

	// Adjust for race
	switch race {
	case "dwarf":
		// Dwarves are typically shorter and stockier
		app.Height = g.randomChoice([]string{"très petit", "petit", "petit"})
		app.Build = g.randomChoice([]string{"trapu", "musclé", "robuste", "râblé"})
	case "elf":
		// Elves are typically tall and slender
		app.Height = g.randomChoice([]string{"grand", "très grand", "de taille moyenne"})
		app.Build = g.randomChoice([]string{"mince", "svelte", "élancé"})
	case "halfling":
		// Halflings are very small
		app.Height = g.randomChoice([]string{"très petit", "très petit", "petit"})
		app.Build = g.randomChoice([]string{"mince", "bedonnant", "râblé"})
	}

	return app
}

func (g *Generator) generatePersonality() Personality {
	return Personality{
		TraitPrincipal:  g.randomChoice(g.traits.Personality.TraitPrincipal),
		TraitSecondaire: g.randomChoice(g.traits.Personality.TraitSecondaire),
		Defaut:          g.randomChoice(g.traits.Personality.Defaut),
		Qualite:         g.randomChoice(g.traits.Personality.Qualite),
	}
}

func (g *Generator) generateMotivation() Motivation {
	return Motivation{
		Goal:   g.randomChoice(g.traits.Motivation.Goal),
		Fear:   g.randomChoice(g.traits.Motivation.Fear),
		Secret: g.randomChoice(g.traits.Motivation.Secret),
	}
}

func (g *Generator) generateVoice() Voice {
	return Voice{
		Tone:   g.randomChoice(g.traits.Voice.Tone),
		Manner: g.randomChoice(g.traits.Voice.Manner),
	}
}

func (g *Generator) generateAttitude(attType string) string {
	attType = strings.ToLower(attType)

	switch attType {
	case "positive", "friendly", "amical":
		return g.randomChoice(g.traits.Relationship.AttitudePositive)
	case "negative", "hostile":
		return g.randomChoice(g.traits.Relationship.AttitudeNegative)
	case "neutral":
		return g.randomChoice(g.traits.Relationship.AttitudeNeutral)
	default:
		// Random attitude with weights (mostly neutral)
		types := [][]string{
			g.traits.Relationship.AttitudePositive,
			g.traits.Relationship.AttitudeNeutral,
			g.traits.Relationship.AttitudeNegative,
		}
		weights := []int{30, 50, 20}
		idx := g.weightedIndex(weights)
		return g.randomChoice(types[idx])
	}
}

// Helper functions
func (g *Generator) randomChoice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[g.rng.Intn(len(items))]
}

func (g *Generator) weightedChoice(items []string, weights []int) string {
	idx := g.weightedIndex(weights)
	return items[idx]
}

func (g *Generator) weightedIndex(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}

	r := g.rng.Intn(total)
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1
}

// ToMarkdown returns a formatted markdown description of the NPC.
func (n *NPC) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s\n\n", n.Name))

	// Basic info
	genderFr := "Homme"
	if n.Gender == "female" {
		genderFr = "Femme"
	}
	raceFr := map[string]string{
		"human":    "Humain",
		"dwarf":    "Nain",
		"elf":      "Elfe",
		"halfling": "Halfelin",
	}[n.Race]

	sb.WriteString(fmt.Sprintf("**%s %s** - %s\n\n", raceFr, genderFr, n.Occupation))

	// Appearance
	sb.WriteString("### Apparence\n")
	sb.WriteString(fmt.Sprintf("%s %s, de stature %s. ",
		strings.Title(n.Appearance.Height), n.Appearance.Build, n.Appearance.Build))
	sb.WriteString(fmt.Sprintf("Cheveux %s %s, yeux %s, peau %s. ",
		n.Appearance.HairColor, n.Appearance.HairStyle, n.Appearance.EyeColor, n.Appearance.Skin))
	sb.WriteString(fmt.Sprintf("A %s. ", n.Appearance.FacialFeature))
	sb.WriteString(fmt.Sprintf("Se distingue par %s.\n\n", n.Appearance.DistinctiveFeature))

	// Personality
	sb.WriteString("### Personnalité\n")
	sb.WriteString(fmt.Sprintf("- **Trait principal** : %s\n", n.Personality.TraitPrincipal))
	sb.WriteString(fmt.Sprintf("- **Trait secondaire** : %s\n", n.Personality.TraitSecondaire))
	sb.WriteString(fmt.Sprintf("- **Qualité** : %s\n", n.Personality.Qualite))
	sb.WriteString(fmt.Sprintf("- **Défaut** : %s\n\n", n.Personality.Defaut))

	// Voice & Quirk
	sb.WriteString("### Comportement\n")
	sb.WriteString(fmt.Sprintf("- **Voix** : %s, %s\n", n.Voice.Tone, n.Voice.Manner))
	sb.WriteString(fmt.Sprintf("- **Tic** : %s\n", n.Quirk))
	sb.WriteString(fmt.Sprintf("- **Attitude** : %s\n\n", n.Attitude))

	// Motivation (for GM only)
	sb.WriteString("### Secrets (MJ seulement)\n")
	sb.WriteString(fmt.Sprintf("- **Objectif** : %s\n", n.Motivation.Goal))
	sb.WriteString(fmt.Sprintf("- **Peur** : %s\n", n.Motivation.Fear))
	sb.WriteString(fmt.Sprintf("- **Secret** : %s\n", n.Motivation.Secret))

	return sb.String()
}

// ToShortDescription returns a brief one-line description.
func (n *NPC) ToShortDescription() string {
	genderFr := "homme"
	if n.Gender == "female" {
		genderFr = "femme"
	}
	raceFr := map[string]string{
		"human":    "humain",
		"dwarf":    "nain",
		"elf":      "elfe",
		"halfling": "halfelin",
	}[n.Race]

	return fmt.Sprintf("%s - %s %s, %s (%s, %s)",
		n.Name, raceFr, genderFr, n.Occupation,
		n.Personality.TraitPrincipal, n.Attitude)
}

// ToJSON returns the NPC as JSON string.
func (n *NPC) ToJSON() (string, error) {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Options

type config struct {
	name           string
	race           string
	gender         string
	occupationType string
	attitude       string
}

// Option is a functional option for NPC generation.
type Option func(*config)

// WithName sets the NPC's name (skips random name generation).
func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

// WithRace sets the NPC's race.
func WithRace(race string) Option {
	return func(c *config) {
		c.race = strings.ToLower(race)
	}
}

// WithGender sets the NPC's gender.
func WithGender(gender string) Option {
	return func(c *config) {
		c.gender = strings.ToLower(gender)
	}
}

// WithOccupationType sets the occupation type.
func WithOccupationType(occType string) Option {
	return func(c *config) {
		c.occupationType = occType
	}
}

// WithAttitude sets the NPC's attitude.
func WithAttitude(attitude string) Option {
	return func(c *config) {
		c.attitude = attitude
	}
}

// GetAvailableOccupationTypes returns the list of occupation types.
func (g *Generator) GetAvailableOccupationTypes() []string {
	return []string{"commoner", "skilled", "authority", "underworld", "religious", "adventurer"}
}

// GetAvailableRaces returns the list of available races.
func (g *Generator) GetAvailableRaces() []string {
	return []string{"human", "dwarf", "elf", "halfling"}
}

// GetAvailableAttitudes returns the list of attitude types.
func (g *Generator) GetAvailableAttitudes() []string {
	return []string{"positive", "neutral", "negative"}
}
