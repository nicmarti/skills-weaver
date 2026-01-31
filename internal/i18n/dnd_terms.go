// Package i18n provides translations for D&D terminology from English to French.
// This is used by the DM agent to use proper French terminology during sessions.
package i18n

// CombatAbilities maps common combat abilities to French translations.
var CombatAbilities = map[string]string{
	// Rogue abilities
	"Sneak Attack":    "Attaque sournoise",
	"Cunning Action":  "Ruse",
	"Evasion":         "Esquive totale",
	"Uncanny Dodge":   "Esquive instinctive",
	"Reliable Talent": "Talent fiable",
	"Blindsense":      "Perception aveugle",
	"Slippery Mind":   "Esprit fuyant",
	"Elusive":         "Insaisissable",
	"Stroke of Luck":  "Coup de chance",

	// Fighter abilities
	"Extra Attack":      "Attaque supplémentaire",
	"Action Surge":      "Sursaut d'action",
	"Second Wind":       "Second souffle",
	"Fighting Style":    "Style de combat",
	"Indomitable":       "Indomptable",
	"Martial Archetype": "Archétype martial",

	// Bard abilities
	"Bardic Inspiration": "Inspiration bardique",
	"Song of Rest":       "Chant reposant",
	"Countercharm":       "Contre-charme",
	"Magical Secrets":    "Secrets magiques",
	"Jack of All Trades": "Touche-à-tout",
	"Font of Inspiration": "Source d'inspiration",

	// Barbarian abilities
	"Rage":              "Rage",
	"Reckless Attack":   "Attaque téméraire",
	"Danger Sense":      "Sens du danger",
	"Feral Instinct":    "Instinct sauvage",
	"Brutal Critical":   "Critique brutal",
	"Relentless Rage":   "Rage implacable",
	"Persistent Rage":   "Rage persistante",
	"Primal Champion":   "Champion primitif",

	// Cleric abilities
	"Channel Divinity":  "Conduit divin",
	"Turn Undead":       "Renvoi des morts-vivants",
	"Divine Strike":     "Frappe divine",
	"Divine Intervention": "Intervention divine",
	"Destroy Undead":    "Destruction des morts-vivants",

	// Paladin abilities
	"Divine Smite":      "Châtiment divin",
	"Lay on Hands":      "Imposition des mains",
	"Divine Sense":      "Sens divin",
	"Aura of Protection": "Aura de protection",
	"Aura of Courage":   "Aura de courage",
	"Cleansing Touch":   "Toucher purificateur",

	// Ranger abilities
	"Favored Enemy":     "Ennemi juré",
	"Natural Explorer":  "Explorateur-né",
	"Primeval Awareness": "Conscience primitive",
	"Hunter's Mark":     "Marque du chasseur",
	"Hide in Plain Sight": "Se cacher en pleine lumière",
	"Vanish":            "Disparition",
	"Feral Senses":      "Sens sauvages",
	"Foe Slayer":        "Tueur d'ennemis",

	// Wizard abilities
	"Arcane Recovery":   "Récupération arcanique",
	"Spell Mastery":     "Maîtrise des sorts",
	"Signature Spells":  "Sorts de prédilection",

	// Sorcerer abilities
	"Sorcery Points":    "Points de sorcellerie",
	"Metamagic":         "Métamagie",
	"Quickened Spell":   "Sort rapide",
	"Twinned Spell":     "Sort jumelé",
	"Empowered Spell":   "Sort puissant",
	"Subtle Spell":      "Sort subtil",
	"Distant Spell":     "Sort distant",
	"Extended Spell":    "Sort prolongé",
	"Heightened Spell":  "Sort intense",
	"Careful Spell":     "Sort prudent",

	// Warlock abilities
	"Eldritch Invocations": "Manifestations occultes",
	"Pact Boon":            "Don du pacte",
	"Pact of the Blade":    "Pacte de la lame",
	"Pact of the Chain":    "Pacte de la chaîne",
	"Pact of the Tome":     "Pacte du grimoire",
	"Mystic Arcanum":       "Arcanum mystique",
	"Eldritch Master":      "Maître occulte",
	"Eldritch Blast":       "Décharge occulte",

	// Druid abilities
	"Wild Shape":         "Forme sauvage",
	"Druid Circle":       "Cercle druidique",
	"Beast Spells":       "Sorts de bête",
	"Archdruid":          "Archidruide",

	// Monk abilities
	"Ki":                  "Ki",
	"Martial Arts":        "Arts martiaux",
	"Flurry of Blows":     "Déluge de coups",
	"Patient Defense":     "Défense patiente",
	"Step of the Wind":    "Pas du vent",
	"Deflect Missiles":    "Parade de projectiles",
	"Stunning Strike":     "Frappe étourdissante",
	"Ki-Empowered Strikes": "Frappes de ki",
	"Stillness of Mind":   "Quiétude de l'esprit",
	"Purity of Body":      "Pureté du corps",
	"Tongue of the Sun and Moon": "Langue du soleil et de la lune",
	"Diamond Soul":        "Âme de diamant",
	"Timeless Body":       "Corps intemporel",
	"Empty Body":          "Corps vide",
	"Perfect Self":        "Perfection de l'être",

	// General combat actions
	"Opportunity Attack": "Attaque d'opportunité",
	"Bonus Action":       "Action bonus",
	"Reaction":           "Réaction",
	"Action":             "Action",
	"Movement":           "Déplacement",
	"Free Action":        "Action gratuite",

	// Combat maneuvers
	"Disarm":             "Désarmement",
	"Grapple":            "Empoignade",
	"Shove":              "Bousculade",
	"Dodge":              "Esquive",
	"Disengage":          "Se désengager",
	"Dash":               "Foncer",
	"Help":               "Aider",
	"Hide":               "Se cacher",
	"Ready":              "Préparer",
	"Search":             "Chercher",
	"Use an Object":      "Utiliser un objet",

	// Spellcasting
	"Concentration":      "Concentration",
	"Ritual":             "Rituel",
	"Cantrip":            "Tour de magie",
	"Spell Slot":         "Emplacement de sort",
	"Saving Throw":       "Jet de sauvegarde",
	"Attack Roll":        "Jet d'attaque",
	"Damage Roll":        "Jet de dégâts",
	"Ability Check":      "Test de caractéristique",
	"Skill Check":        "Test de compétence",
}

// DamageTypes maps damage types to French translations.
var DamageTypes = map[string]string{
	"bludgeoning": "contondants",
	"slashing":    "tranchants",
	"piercing":    "perforants",
	"fire":        "feu",
	"cold":        "froid",
	"lightning":   "foudre",
	"thunder":     "tonnerre",
	"poison":      "poison",
	"psychic":     "psychiques",
	"necrotic":    "nécrotiques",
	"radiant":     "radieux",
	"force":       "force",
	"acid":        "acide",
}

// Conditions maps condition names to French translations.
var Conditions = map[string]string{
	"prone":         "à terre",
	"grappled":      "agrippé",
	"charmed":       "charmé",
	"frightened":    "effrayé",
	"blinded":       "aveuglé",
	"deafened":      "assourdi",
	"stunned":       "étourdi",
	"poisoned":      "empoisonné",
	"paralyzed":     "paralysé",
	"unconscious":   "inconscient",
	"restrained":    "entravé",
	"invisible":     "invisible",
	"incapacitated": "neutralisé",
	"exhausted":     "épuisé",
	"petrified":     "pétrifié",
}

// Abilities maps ability score names to French translations.
var Abilities = map[string]string{
	"Strength":     "Force",
	"Dexterity":    "Dextérité",
	"Constitution": "Constitution",
	"Intelligence": "Intelligence",
	"Wisdom":       "Sagesse",
	"Charisma":     "Charisme",
	"STR":          "FOR",
	"DEX":          "DEX",
	"CON":          "CON",
	"INT":          "INT",
	"WIS":          "SAG",
	"CHA":          "CHA",
}

// Skills maps skill names to French translations.
var Skills = map[string]string{
	"Acrobatics":      "Acrobaties",
	"Animal Handling": "Dressage",
	"Arcana":          "Arcanes",
	"Athletics":       "Athlétisme",
	"Deception":       "Tromperie",
	"History":         "Histoire",
	"Insight":         "Perspicacité",
	"Intimidation":    "Intimidation",
	"Investigation":   "Investigation",
	"Medicine":        "Médecine",
	"Nature":          "Nature",
	"Perception":      "Perception",
	"Performance":     "Représentation",
	"Persuasion":      "Persuasion",
	"Religion":        "Religion",
	"Sleight of Hand": "Escamotage",
	"Stealth":         "Discrétion",
	"Survival":        "Survie",
}

// Translate returns the French translation for a D&D term, or the original if not found.
func Translate(term string) string {
	if fr, ok := CombatAbilities[term]; ok {
		return fr
	}
	if fr, ok := DamageTypes[term]; ok {
		return fr
	}
	if fr, ok := Conditions[term]; ok {
		return fr
	}
	if fr, ok := Abilities[term]; ok {
		return fr
	}
	if fr, ok := Skills[term]; ok {
		return fr
	}
	return term
}

// TranslateDamageType returns the French translation for a damage type.
func TranslateDamageType(damageType string) string {
	if fr, ok := DamageTypes[damageType]; ok {
		return fr
	}
	return damageType
}

// TranslateCondition returns the French translation for a condition.
func TranslateCondition(condition string) string {
	if fr, ok := Conditions[condition]; ok {
		return fr
	}
	return condition
}

// TranslateAbility returns the French translation for an ability.
func TranslateAbility(ability string) string {
	if fr, ok := Abilities[ability]; ok {
		return fr
	}
	return ability
}

// TranslateSkill returns the French translation for a skill.
func TranslateSkill(skill string) string {
	if fr, ok := Skills[skill]; ok {
		return fr
	}
	return skill
}
