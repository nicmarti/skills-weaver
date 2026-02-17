package adventure

// AdventureType defines a classic D&D adventure archetype.
type AdventureType struct {
	ID          string `json:"id"`
	NameFR      string `json:"name_fr"`
	Description string `json:"description"`
	PromptGuide string `json:"prompt_guide"`
}

// AdventureDuration defines how long an adventure should last.
type AdventureDuration struct {
	ID          string `json:"id"`
	NameFR      string `json:"name_fr"`
	Acts        int    `json:"acts"`
	MinSessions int    `json:"min_sessions"`
	MaxSessions int    `json:"max_sessions"`
	MaxNPCs     int    `json:"max_npcs"`
	MaxLocations int   `json:"max_locations"`
	MaxForeshadows int `json:"max_foreshadows"`
}

// GetAdventureDurations returns the available duration options.
func GetAdventureDurations() []AdventureDuration {
	return []AdventureDuration{
		{
			ID:             "oneshot",
			NameFR:         "One-shot (1-2 sessions)",
			Acts:           1,
			MinSessions:    1,
			MaxSessions:    2,
			MaxNPCs:        3,
			MaxLocations:   2,
			MaxForeshadows: 1,
		},
		{
			ID:             "short",
			NameFR:         "Courte aventure (3-5 sessions)",
			Acts:           1,
			MinSessions:    3,
			MaxSessions:    5,
			MaxNPCs:        5,
			MaxLocations:   4,
			MaxForeshadows: 2,
		},
		{
			ID:             "campaign",
			NameFR:         "Campagne (8-12 sessions)",
			Acts:           3,
			MinSessions:    8,
			MaxSessions:    12,
			MaxNPCs:        8,
			MaxLocations:   6,
			MaxForeshadows: 3,
		},
	}
}

// GetDuration returns the duration config for the given ID, or the "short" default.
func GetDuration(id string) AdventureDuration {
	for _, d := range GetAdventureDurations() {
		if d.ID == id {
			return d
		}
	}
	// Default to short
	for _, d := range GetAdventureDurations() {
		if d.ID == "short" {
			return d
		}
	}
	return GetAdventureDurations()[1]
}

// antiCultConstraints is injected into ALL generation prompts regardless of type.
const antiCultConstraints = `CONTRAINTES CRITIQUES :
- NE PAS utiliser de cultes, societes secretes, ou entites cosmiques/divines comme antagonistes
- NE PAS impliquer de menaces de fin du monde ou d'apocalypse
- Les antagonistes doivent etre MUNDAINS : rivaux, nobles corrompus, bandits, monstres territoriaux,
  marchands jaloux, espions etrangers, betes sauvages, pirates, contrebandiers
- Les enjeux doivent etre LOCAUX et PERSONNELS, pas cosmiques
- Les secrets doivent etre HUMAINS : detournement de fonds, amour interdit, heritier cache,
  vieille rancune, contrebande, trahison politique -- PAS de rituels eldritch ou d'entites scellees
- Privilegier des antagonistes avec des motivations comprehensibles : cupidite, vengeance, ambition, peur, jalousie`

// GetAntiCultConstraints returns the anti-cult constraints text for prompt injection.
func GetAntiCultConstraints() string {
	return antiCultConstraints
}

// GetAdventureTypes returns the catalog of 9 classic adventure types.
func GetAdventureTypes() []AdventureType {
	return []AdventureType{
		{
			ID:          "escort",
			NameFR:      "Mission d'escorte",
			Description: "Proteger quelqu'un pendant un voyage",
			PromptGuide: `TYPE D'AVENTURE : MISSION D'ESCORTE
Cette aventure est centree sur la protection d'une personne ou d'un convoi pendant un trajet dangereux.
FOCUS : Le voyage, les embuscades, la gestion des ressources, les rencontres en chemin, la relation avec le protege.
DEFIS TYPIQUES : Bandits de grand chemin, betes sauvages, terrain difficile, intemperies, traitres dans le convoi, rivaux qui veulent intercepter le protege.
ANTAGONISTES ADAPTES : Chef de bandits, seigneur de guerre local, chasseur de primes, rival commercial, creature territoriale.
ENJEUX : La vie du protege, une livraison cruciale, un traite diplomatique, un temoin a proteger.
NE PAS faire de l'objet du transport un artefact cosmique ou divin.`,
		},
		{
			ID:          "fetch_quest",
			NameFR:      "Quete d'objet",
			Description: "Localiser et recuperer un artefact ou objet",
			PromptGuide: `TYPE D'AVENTURE : QUETE D'OBJET
Cette aventure est centree sur la localisation et la recuperation d'un objet precis.
FOCUS : L'enquete pour trouver l'emplacement, le voyage, la recuperation de l'objet, le retour.
DEFIS TYPIQUES : Pieges, gardiens, enigmes, rivaux cherchant le meme objet, negociation avec le possesseur actuel.
ANTAGONISTES ADAPTES : Collectionneur rival, voleur professionnel, gardien ancien (golem, monstre), noble qui refuse de ceder l'objet.
ENJEUX : L'objet peut etre un heritage familial, une preuve juridique, un composant rare, un tresor historique.
L'objet ne doit PAS etre un artefact capable de detruire le monde ou d'invoquer des entites.`,
		},
		{
			ID:          "mystery",
			NameFR:      "Enquete / mystere",
			Description: "Resoudre un meurtre ou un crime",
			PromptGuide: `TYPE D'AVENTURE : ENQUETE / MYSTERE
Cette aventure est centree sur la resolution d'un crime, d'une disparition ou d'un evenement inexplique.
FOCUS : Interrogation de temoins, collecte d'indices, deductions, fausses pistes, confrontation du coupable.
DEFIS TYPIQUES : Temoins menteurs, indices caches, alibis fabriques, tentatives d'intimidation, course contre la montre.
ANTAGONISTES ADAPTES : Meurtrier avec mobile personnel, escroc, noble corrompu couvrant un crime, espion etranger.
ENJEUX : Justice pour la victime, sauver un innocent accuse a tort, decouvrir un reseau de contrebande.
Le mystere doit avoir une explication RATIONNELLE et HUMAINE, pas surnaturelle.`,
		},
		{
			ID:          "rescue",
			NameFR:      "Mission de sauvetage",
			Description: "Liberer un prisonnier ou une personne piegee",
			PromptGuide: `TYPE D'AVENTURE : MISSION DE SAUVETAGE
Cette aventure est centree sur la liberation d'une ou plusieurs personnes captives.
FOCUS : Localisation du prisonnier, infiltration ou assaut, extraction, fuite.
DEFIS TYPIQUES : Gardes, fortifications, pieges, negociation de rancon, le prisonnier est blesse ou ne veut pas partir.
ANTAGONISTES ADAPTES : Chef de brigands, seigneur tyrannique, esclavagiste, pirate, creature ayant capture des villageois.
ENJEUX : La vie du captif, une echance de rancon, des represailles si l'operation echoue.
Le captif doit etre une personne ordinaire (noble, marchand, enfant, artisan), pas un elu prophetique.`,
		},
		{
			ID:          "exploration",
			NameFR:      "Exploration / donjon",
			Description: "Explorer un lieu inconnu (ruines, cavernes)",
			PromptGuide: `TYPE D'AVENTURE : EXPLORATION / DONJON
Cette aventure est centree sur l'exploration d'un lieu inconnu ou oublie.
FOCUS : Cartographie, pieges mecaniques, monstres errants, puzzles, tresors, gestion des ressources (torches, rations).
DEFIS TYPIQUES : Couloirs effondres, salles piegees, monstres embusques, enigmes de portes, tresor garde.
ANTAGONISTES ADAPTES : Creature ayant elu domicile (dragon, troll, araignees geantes), groupe de pillards rival, gardien construit.
ENJEUX : Le tresor du donjon, une carte menant a d'autres lieux, un passage vers une region isolee.
Le donjon doit avoir une origine MUNDAINE (ancienne forteresse, mine abandonnee, tombe d'un noble), pas un temple a une entite cosmique.`,
		},
		{
			ID:          "defense",
			NameFR:      "Garde / defense",
			Description: "Proteger un lieu ou une personne contre des attaques",
			PromptGuide: `TYPE D'AVENTURE : GARDE / DEFENSE
Cette aventure est centree sur la defense d'un lieu, d'une communaute ou d'un evenement.
FOCUS : Preparation des defenses, reconnaissance de l'ennemi, gestion des ressources, batailles tactiques, moral des defenseurs.
DEFIS TYPIQUES : Vagues d'assaillants, sabotage interieur, siege, ravitaillement coupe, dissension parmi les defenseurs.
ANTAGONISTES ADAPTES : Bande de pillards, seigneur de guerre, horde de monstres (gobelins, orcs), pirates, armee rivale.
ENJEUX : La survie du village/fort, la protection d'un festival, empecher le vol de ressources precieuses.
L'attaque doit avoir une motivation CONCRETE (pillage, conquete, vengeance), pas un rituel ou une prophetie.`,
		},
		{
			ID:          "heist",
			NameFR:      "Braquage / infiltration",
			Description: "S'introduire dans un lieu securise",
			PromptGuide: `TYPE D'AVENTURE : BRAQUAGE / INFILTRATION
Cette aventure est centree sur la planification et l'execution d'une intrusion dans un lieu protege.
FOCUS : Reconnaissance, planification, deguisements, crochetage, desactivation de pieges, extraction discrete.
DEFIS TYPIQUES : Gardes patrouillant, serrures magiques ou mecaniques, alarmes, chiens de garde, timing serre.
ANTAGONISTES ADAPTES : Le proprietaire du lieu (noble, marchand, chef de guilde), capitaine de la garde, rival voleur.
ENJEUX : Voler un document compromettant, recuperer un bien vole, placer un objet, liberer un prisonnier discretement.
L'objectif du braquage doit etre CONCRET et HUMAIN, pas un artefact cosmique.`,
		},
		{
			ID:          "hunt",
			NameFR:      "Chasse au monstre",
			Description: "Traquer et eliminer une creature dangereuse",
			PromptGuide: `TYPE D'AVENTURE : CHASSE AU MONSTRE
Cette aventure est centree sur la traque et l'elimination d'une creature dangereuse.
FOCUS : Etude de la creature, pistage, preparation tactique, combat final, recompense.
DEFIS TYPIQUES : Terrain de chasse hostile, fausses pistes, la creature est plus intelligente que prevu, dommages collateraux.
ANTAGONISTES ADAPTES : La creature elle-meme (dragon, troll, loup-garou, basilic, hydre), un dresseur de monstres, un rival chasseur.
ENJEUX : Sauver un village terrorise, une prime genereuse, venger des victimes, recuperer un ingredient rare de la creature.
La creature doit etre un MONSTRE classique avec des motivations animales (faim, territoire), pas un serviteur d'entite cosmique.`,
		},
		{
			ID:          "diplomacy",
			NameFR:      "Diplomatie",
			Description: "Negocier entre factions rivales",
			PromptGuide: `TYPE D'AVENTURE : DIPLOMATIE
Cette aventure est centree sur la negociation et la resolution de conflits entre factions.
FOCUS : Rencontres diplomatiques, intrigues de cour, alliances, trahisons, compromis, equilibre des pouvoirs.
DEFIS TYPIQUES : Factions aux interets opposes, assassinats politiques, chantage, espionnage, manipulation.
ANTAGONISTES ADAPTES : Ambassadeur hostile, conseiller corrompu, espion, noble ambitieux, faux mediateur sabotant les negociations.
ENJEUX : Empecher une guerre, forger une alliance commerciale, resoudre un conflit frontalier, negocier un mariage politique.
Le conflit doit avoir des racines POLITIQUES et ECONOMIQUES, pas religieuses ou cosmiques.`,
		},
	}
}

// GetAdventureType returns the adventure type for the given ID, or nil if not found.
func GetAdventureType(id string) *AdventureType {
	for _, t := range GetAdventureTypes() {
		if t.ID == id {
			return &t
		}
	}
	return nil
}
