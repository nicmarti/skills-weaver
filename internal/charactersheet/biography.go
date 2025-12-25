package charactersheet

import (
	"context"
	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Biography represents character backstory
type Biography struct {
	CharacterName   string    `json:"character_name"`
	Origin          string    `json:"origin"`
	Background      string    `json:"background"`
	Motivation      string    `json:"motivation"`
	Personality     string    `json:"personality"`
	Bonds           []Bond    `json:"bonds"`
	Secrets         []string  `json:"secrets"`
	GeneratedAt     time.Time `json:"generated_at"`
	AdventureContext string   `json:"adventure_context"`
}

// Bond represents a relationship
type Bond struct {
	Type        string `json:"type"`        // "person", "place", "faction"
	Name        string `json:"name"`
	Description string `json:"description"`
	Sentiment   string `json:"sentiment"`   // "ally", "enemy", "neutral", "complicated"
}

// BiographyGenerator creates character backstories
type BiographyGenerator struct {
	apiKey string // Claude API key for AI-enhanced biographies
}

// AIBiographyResponse holds the structured response from Claude
type AIBiographyResponse struct {
	Origin       string   `json:"origin"`
	Background   string   `json:"background"`
	Motivation   string   `json:"motivation"`
	Personality  string   `json:"personality"`
	BondName     string   `json:"bond_name"`
	BondDesc     string   `json:"bond_description"`
	BondType     string   `json:"bond_type"`
	BondSentiment string  `json:"bond_sentiment"`
	Secrets      []string `json:"secrets"`
}

// NewBiographyGenerator creates a new biography generator
func NewBiographyGenerator() *BiographyGenerator {
	return &BiographyGenerator{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

// Generate creates a biography from character data
func (g *BiographyGenerator) Generate(c *character.Character, adventureName string) (*Biography, error) {
	// Try AI-enhanced generation if API key is available
	if g.apiKey != "" {
		bio, err := g.generateWithAI(c, adventureName)
		if err == nil {
			return bio, nil
		}
		// Fall back to templates if AI generation fails
		fmt.Fprintf(os.Stderr, "Warning: AI generation failed (%v), using templates\n", err)
	}

	// Fallback to template-based generation
	bio := &Biography{
		CharacterName:   c.Name,
		Origin:          g.generateOrigin(c),
		Background:      g.generateBackground(c),
		Motivation:      g.generateMotivation(c),
		Personality:     g.generatePersonality(c),
		Bonds:           g.generateBonds(c, adventureName),
		Secrets:         g.generateSecrets(c),
		GeneratedAt:     time.Now(),
		AdventureContext: adventureName,
	}

	return bio, nil
}

// Save writes biography to JSON cache
func (b *Biography) Save(dir string) error {
	filename := strings.ToLower(strings.ReplaceAll(b.CharacterName, " ", "-")) + "_bio.json"
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal biography: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write biography: %w", err)
	}

	return nil
}

// LoadBiography reads biography from cache
func LoadBiography(characterName string, dir string) (*Biography, error) {
	filename := strings.ToLower(strings.ReplaceAll(characterName, " ", "-")) + "_bio.json"
	path := filepath.Join(dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read biography: %w", err)
	}

	var bio Biography
	if err := json.Unmarshal(data, &bio); err != nil {
		return nil, fmt.Errorf("failed to unmarshal biography: %w", err)
	}

	return &bio, nil
}

// generateWithAI uses Claude API to generate rich, personalized biographies
func (g *BiographyGenerator) generateWithAI(c *character.Character, adventureName string) (*Biography, error) {
	// Build context for the prompt
	adventureContext := ""
	if adventureName != "" {
		adv, err := adventure.LoadByName("data/adventures", adventureName)
		if err == nil {
			journal, _ := adv.LoadJournal()
			if journal != nil && len(journal.Entries) > 0 {
				// Get last 5 entries as context
				recentEntries := []string{}
				start := len(journal.Entries) - 5
				if start < 0 {
					start = 0
				}
				for _, entry := range journal.Entries[start:] {
					recentEntries = append(recentEntries, entry.Content)
				}
				adventureContext = fmt.Sprintf("Recent adventure events: %s", strings.Join(recentEntries, " → "))
			}
		}
	}

	// Build the prompt
	prompt := g.buildBiographyPrompt(c, adventureContext)

	// Call Claude API
	client := anthropic.NewClient(option.WithAPIKey(g.apiKey))
	response, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model: anthropic.Model("claude-3-5-haiku-20241022"),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		MaxTokens:   1000,
		Temperature: anthropic.Float(0.8),
	})

	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude API")
	}

	// Parse JSON response
	jsonStr := stripMarkdownFences(response.Content[0].Text)
	var aiResp AIBiographyResponse
	if err := json.Unmarshal([]byte(jsonStr), &aiResp); err != nil {
		return nil, fmt.Errorf("parsing Claude response: %w\nResponse: %s", err, jsonStr)
	}

	// Convert to Biography struct
	bio := &Biography{
		CharacterName:    c.Name,
		Origin:           aiResp.Origin,
		Background:       aiResp.Background,
		Motivation:       aiResp.Motivation,
		Personality:      aiResp.Personality,
		GeneratedAt:      time.Now(),
		AdventureContext: adventureName,
	}

	// Add bond if provided
	if aiResp.BondName != "" {
		bio.Bonds = []Bond{{
			Type:        aiResp.BondType,
			Name:        aiResp.BondName,
			Description: aiResp.BondDesc,
			Sentiment:   aiResp.BondSentiment,
		}}
	}

	// Add secrets
	bio.Secrets = aiResp.Secrets

	return bio, nil
}

// buildBiographyPrompt constructs the prompt for Claude API
func (g *BiographyGenerator) buildBiographyPrompt(c *character.Character, adventureContext string) string {
	// Format character stats
	stats := fmt.Sprintf("FOR:%d DEX:%d CON:%d INT:%d SAG:%d CHA:%d",
		c.Abilities.Strength, c.Abilities.Dexterity, c.Abilities.Constitution,
		c.Abilities.Intelligence, c.Abilities.Wisdom, c.Abilities.Charisma)

	// Format appearance
	appearance := "Apparence inconnue"
	if c.Appearance != nil {
		appearance = fmt.Sprintf("Âge:%d ans, Corpulence:%s, Caractéristique distinctive:%s",
			c.Appearance.Age, c.Appearance.Build, c.Appearance.DistinctiveFeature)
	}

	// Translate race/class to French
	raceNames := map[string]string{
		"human": "humain", "elf": "elfe", "dwarf": "nain", "halfling": "halfelin",
	}
	classNames := map[string]string{
		"fighter": "guerrier", "cleric": "clerc", "magic-user": "magicien", "thief": "voleur",
	}

	raceFr := raceNames[c.Race]
	classFr := classNames[c.Class]

	return fmt.Sprintf(`Tu es un écrivain spécialisé dans les biographies de personnages de jeux de rôle fantasy. Crée une biographie IMMERSIVE et PERSONNALISÉE pour ce personnage Basic Fantasy RPG.

PERSONNAGE:
- Nom: %s
- Race: %s
- Classe: %s
- Niveau: %d
- Caractéristiques: %s
- %s
- XP: %d / PV: %d / CA: %d

%s

INSTRUCTIONS:
1. Écris en français avec un style narratif IMMERSIF et PERSONNALISÉ
2. Base-toi sur les caractéristiques pour créer une personnalité cohérente
3. Intègre les détails d'apparence de manière naturelle
4. Crée une origine unique et mémorable (pas générique)
5. Évite les clichés et les formulations stéréotypées
6. Utilise un ton littéraire et évocateur
7. Si un contexte d'aventure est fourni, intègre-le subtilement

STRUCTURE OBLIGATOIRE (format JSON):
{
  "origin": "Paragraphe sur l'origine (2-3 phrases). Où est-il né ? Quelle est son enfance ? Comment est devenu ce qu'il est ?",
  "background": "Paragraphe sur le passé récent (2-3 phrases). Qu'a-t-il fait avant de devenir aventurier ? Quelle formation ?",
  "motivation": "Une phrase courte sur sa motivation principale",
  "personality": "Paragraphe décrivant sa personnalité (2-3 phrases). Comment se comporte-t-il ? Quels sont ses traits dominants basés sur ses stats ?",
  "bond_name": "Nom d'une personne, faction ou lieu important",
  "bond_description": "Description du lien (1 phrase)",
  "bond_type": "person, faction ou place",
  "bond_sentiment": "ally, enemy, neutral ou complicated",
  "secrets": ["Secret 1", "Secret 2"]
}

EXEMPLES DE STYLE À ÉVITER:
❌ "Il a appris les techniques de combat par nécessité"
❌ "Personnage équilibré aux talents variés"
❌ "A servi dans la milice locale"

EXEMPLES DE BON STYLE:
✓ "Les cicatrices qui zèbrent ses avant-bras racontent mieux que les mots les années passées dans les arènes de Karvath"
✓ "Son regard trahit une intelligence vive, constamment en éveil, pesant chaque mot comme un marchand pèse l'or"
✓ "Ancien templier déchu, il porte encore l'armure de son ordre mais en a arraché les symboles sacrés dans un accès de rage"

OUTPUT: Retourne UNIQUEMENT le JSON, sans markdown, sans commentaire.`,
		c.Name, raceFr, classFr, c.Level, stats, appearance, c.XP, c.HitPoints, c.ArmorClass,
		adventureContext)
}

// stripMarkdownFences removes markdown code fences from JSON responses
func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

// Template-based generation methods

func (g *BiographyGenerator) generateOrigin(c *character.Character) string {
	// Template based on race + build + age
	var origin string

	age := "d'âge mûr"
	if c.Appearance != nil && c.Appearance.Age > 0 {
		if c.Appearance.Age < 25 {
			age = "jeune"
		} else if c.Appearance.Age > 50 {
			age = "d'âge avancé"
		}
	}

	switch c.Race {
	case "dwarf":
		origin = fmt.Sprintf("%s est issu des halls montagneux des nains, un guerrier %s forgé par la pierre et le fer.", c.Name, age)
	case "elf":
		origin = fmt.Sprintf("%s est né dans les forêts anciennes, enfant des elfes, %s et marqué par la grâce de son peuple.", c.Name, age)
	case "halfling":
		origin = fmt.Sprintf("%s vient des shires paisibles des halfelins, un aventurier %s au cœur courageux malgré sa petite taille.", c.Name, age)
	case "human":
		origin = fmt.Sprintf("%s est né dans les terres variées des humains, un individu %s aux origines diverses.", c.Name, age)
	default:
		origin = fmt.Sprintf("%s est originaire de terres lointaines.", c.Name)
	}

	// Add build characteristic
	if c.Appearance != nil && c.Appearance.Build != "" {
		switch c.Appearance.Build {
		case "muscular":
			origin += " Sa carrure imposante témoigne d'années d'entraînement physique."
		case "slender":
			origin += " Sa silhouette élancée reflète agilité et grâce naturelle."
		case "stocky":
			origin += " Sa constitution robuste montre endurance et résistance."
		}
	}

	return origin
}

func (g *BiographyGenerator) generateBackground(c *character.Character) string {
	// Template based on class + abilities
	var background string

	switch c.Class {
	case "fighter":
		if c.Abilities.Strength > 14 {
			background = "Formé au combat dès l'enfance, il a perfectionné l'art de la guerre à travers d'innombrables batailles. Sa force exceptionnelle en fait un adversaire redoutable."
		} else {
			background = "Il a appris les techniques de combat par nécessité, développant tactique et discipline pour compenser une force modeste."
		}
	case "cleric":
		if c.Abilities.Wisdom > 14 {
			background = "Appelé par sa foi dès son plus jeune âge, il a étudié les textes sacrés et médité sur les mystères divins. Sa sagesse profonde guide ses actions."
		} else {
			background = "Il a trouvé sa foi plus tard dans la vie, apportant zèle et dévotion à servir sa divinité avec ferveur."
		}
	case "magic-user":
		if c.Abilities.Intelligence > 14 {
			background = "Des années d'étude intense des arcanes ont forgé son esprit brillant. Il a dévoré grimoires anciens et parchemins mystiques pour maîtriser la magie."
		} else {
			background = "Il a découvert la magie par hasard, apprenant les bases des arts arcaniques avec détermination malgré les difficultés."
		}
	case "thief":
		if c.Abilities.Dexterity > 14 {
			background = "Né dans les rues, il a développé une dextérité extraordinaire pour survivre. Ses doigts agiles et ses réflexes vifs sont légendaires."
		} else {
			background = "Il a choisi la vie de voleur par nécessité, apprenant les ficelles du métier avec ruse et intelligence plutôt que pure habileté."
		}
	default:
		background = "Son passé reste mystérieux, mais ses compétences parlent d'elles-mêmes."
	}

	return background
}

func (g *BiographyGenerator) generateMotivation(c *character.Character) string {
	// Template based on class + distinctive features
	motivations := map[string][]string{
		"fighter": {
			"Chercher gloire et honneur sur les champs de bataille",
			"Protéger les innocents et faire régner la justice",
			"Prouver sa valeur et gagner la reconnaissance",
			"Devenir le meilleur guerrier de sa génération",
		},
		"cleric": {
			"Servir fidèlement sa divinité et répandre sa foi",
			"Guérir les malades et protéger les faibles",
			"Combattre le mal et purifier les terres corrompues",
			"Accomplir une quête divine révélée en rêve",
		},
		"magic-user": {
			"Percer les secrets les plus profonds de la magie",
			"Découvrir des sorts perdus et des artefacts anciens",
			"Comprendre les mystères de l'univers",
			"Surpasser ses rivaux académiques",
		},
		"thief": {
			"Devenir riche et vivre dans le luxe",
			"Prendre aux riches pour donner aux pauvres",
			"Prouver son habileté en volant l'impossible",
			"Échapper à un passé sombre",
		},
	}

	if motList, ok := motivations[c.Class]; ok {
		// Use modulo to pick a consistent motivation based on character name
		idx := len(c.Name) % len(motList)
		return motList[idx]
	}

	return "Chercher aventure et fortune dans le vaste monde."
}

func (g *BiographyGenerator) generatePersonality(c *character.Character) string {
	// Template based on ability scores
	traits := []string{}

	if c.Abilities.Strength > 14 {
		traits = append(traits, "imposant et intimidant")
	} else if c.Abilities.Strength < 8 {
		traits = append(traits, "frêle mais déterminé")
	}

	if c.Abilities.Intelligence > 14 {
		traits = append(traits, "brillant et analytique")
	} else if c.Abilities.Intelligence < 8 {
		traits = append(traits, "simple mais pragmatique")
	}

	if c.Abilities.Wisdom > 14 {
		traits = append(traits, "sage et perspicace")
	} else if c.Abilities.Wisdom < 8 {
		traits = append(traits, "impulsif et téméraire")
	}

	if c.Abilities.Charisma > 14 {
		traits = append(traits, "charismatique et persuasif")
	} else if c.Abilities.Charisma < 8 {
		traits = append(traits, "bourru et direct")
	}

	if len(traits) == 0 {
		return "Personnage équilibré aux talents variés, sans trait dominant particulier."
	}

	if len(traits) == 1 {
		return fmt.Sprintf("Personnage %s.", traits[0])
	}

	// Join traits with comma, except last one with "et"
	personality := "Personnage " + strings.Join(traits[:len(traits)-1], ", ") + " et " + traits[len(traits)-1] + "."
	return personality
}

func (g *BiographyGenerator) generateBonds(c *character.Character, adventureName string) []Bond {
	bonds := []Bond{}

	// Add class-specific bond
	switch c.Class {
	case "fighter":
		bonds = append(bonds, Bond{
			Type:        "faction",
			Name:        "Garde du royaume",
			Description: "A servi dans la milice locale, conserve des contacts",
			Sentiment:   "ally",
		})
	case "cleric":
		bonds = append(bonds, Bond{
			Type:        "faction",
			Name:        "Temple local",
			Description: "Lieu de formation religieuse et de méditation",
			Sentiment:   "ally",
		})
	case "magic-user":
		bonds = append(bonds, Bond{
			Type:        "person",
			Name:        "Mentor arcanique",
			Description: "Ancien maître qui a enseigné les bases de la magie",
			Sentiment:   "ally",
		})
	case "thief":
		bonds = append(bonds, Bond{
			Type:        "faction",
			Name:        "Guilde des ombres",
			Description: "Réseau secret de voleurs et d'informateurs",
			Sentiment:   "complicated",
		})
	}

	// TODO: Parse adventure journal for additional bonds in future enhancement

	return bonds
}

func (g *BiographyGenerator) generateSecrets(c *character.Character) []string {
	secrets := []string{}

	// Add distinctive feature as secret if present
	if c.Appearance != nil && c.Appearance.DistinctiveFeature != "" {
		switch {
		case strings.Contains(strings.ToLower(c.Appearance.DistinctiveFeature), "scar"):
			secrets = append(secrets, "Porte une cicatrice dont il ne révèle jamais l'origine")
		case strings.Contains(strings.ToLower(c.Appearance.DistinctiveFeature), "tattoo"):
			secrets = append(secrets, "Ses tatouages cachent un symbole interdit")
		case strings.Contains(strings.ToLower(c.Appearance.DistinctiveFeature), "burn"):
			secrets = append(secrets, "Ses brûlures témoignent d'une rencontre traumatisante")
		}
	}

	// Add class-specific secret
	classSecrets := map[string]string{
		"fighter": "Cache une peur secrète d'être considéré comme lâche",
		"cleric":  "Doute parfois de la foi qu'il professe publiquement",
		"magic-user": "Possède un grimoire volé qu'il ne devrait pas avoir",
		"thief":   "Doit encore une dette importante à quelqu'un de dangereux",
	}

	if secret, ok := classSecrets[c.Class]; ok {
		secrets = append(secrets, secret)
	}

	return secrets
}

// LoadCharacterForBio is a helper to load character for bio command
func LoadCharacterForBio(path string) (*character.Character, error) {
	return character.Load(path)
}
