# Plan d'Implémentation - SkillsWeaver

## Vision du Projet

Créer un moteur de jeu de rôle interactif utilisant Claude Code comme orchestrateur, avec:
- **Skills** pour les mécaniques de jeu (dés, calculs, génération)
- **Sous-agents** spécialisés (Maître du Jeu, Créateur de personnages, Gardien des règles)
- **Scripts Go** pour la logique métier et les données de jeu

## Choix Techniques

| Aspect | Choix | Raison |
|--------|-------|--------|
| **Système de règles** | Basic Fantasy RPG | Open source, simple, gratuit |
| **Format de sortie** | JSON + Markdown | JSON pour données, Markdown pour affichage |
| **Persistance** | `data/characters/` | Sauvegarde automatique des personnages |
| **Langage** | Go | Performance, typage fort, CLI native |

---

## Progression

| Phase | Description | Statut |
|-------|-------------|--------|
| Phase 1 | Système de dés | Terminée |
| Phase 2 | Données BFRPG | Terminée |
| Phase 3 | Générateur de personnages | Terminée |
| Phase 3bis | Gestionnaire d'aventures | Terminée |
| Phase 4 | Sous-agents spécialisés | Terminée |
| Phase 4bis | Générateur de noms | Terminée |
| Phase 5 | Générateur de PNJ | Terminée |
| Phase 6 | Générateur d'images | Terminée |
| Phase 7 | Bestiaire BFRPG | Terminée |
| Phase 8 | Tables de trésors | Terminée |

---

## Phase 1: Système de Dés [TERMINÉE]

### Fichiers créés
- `internal/dice/dice.go` - Package avec support notation (2d6+3, 4d6kh3)
- `internal/dice/dice_test.go` - 10 tests unitaires
- `cmd/dice/main.go` - CLI interactive
- `.claude/skills/dice-roller/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Support de tous les dés : d4, d6, d8, d10, d12, d20, d100
- Notation standard : `2d6+3`, `4d6kh3` (keep highest)
- Avantage/Désavantage pour le d20
- Génération de caractéristiques (4d6kh3 × 6 ou 3d6 × 6)

### Usage
```bash
./sw-dice roll 2d6+3
./sw-dice roll 4d6kh3
./sw-dice roll d20 --advantage
./sw-dice stats
./sw-dice stats --classic
```

---

## Phase 2: Données BFRPG [TERMINÉE]

### Fichiers créés
- `data/races.json` - 4 races avec bonus/malus et restrictions
- `data/classes.json` - 4 classes avec tables XP, sauvegardes, sorts
- `data/equipment.json` - Armes, armures, équipement
- `internal/data/loader.go` - Package Go pour charger les données
- `internal/data/loader_test.go` - 10 tests unitaires

### Données disponibles

**Races** :
- Humain : toutes classes, niveau illimité
- Elfe : +1 DEX, -1 CON, Guerrier (6), Magicien (9), Voleur
- Nain : +1 CON, -1 CHA, Guerrier (7), Clerc (6), Voleur
- Halfelin : +1 DEX, -1 FOR, Guerrier (4), Voleur

**Classes** :
- Guerrier : d8 PV, toutes armes/armures
- Clerc : d6 PV, sorts divins niveau 2+, renvoi des morts-vivants
- Magicien : d4 PV, sorts arcaniques, pas d'armure
- Voleur : d4 PV, compétences spéciales, attaque sournoise

---

## Phase 3: Générateur de Personnages [TERMINÉE]

### Fichiers créés
- `internal/character/character.go` - Structure Character et méthodes
- `internal/character/character_test.go` - 14 tests unitaires
- `cmd/character/main.go` - CLI complète
- `.claude/skills/character-generator/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Génération de caractéristiques (4d6kh3 ou 3d6)
- Application des modificateurs raciaux
- Calcul des modificateurs BFRPG (-3 à +3)
- Points de vie (max au niveau 1 + CON)
- Or de départ (3d6×10 ou 2d6×10)
- Validation race/classe
- Sauvegarde JSON dans `data/characters/`
- Export Markdown et JSON

### Usage
```bash
./sw-character create "Aldric" --race=human --class=fighter
./sw-character create "Lyra" --race=elf --class=magic-user --method=classic
./sw-character list
./sw-character show "Aldric"
./sw-character delete "Aldric"
./sw-character export "Aldric" --format=json
```

---

## Phase 3bis: Gestionnaire d'Aventures [TERMINEE]

### Fichiers créés
- `internal/adventure/adventure.go` - Structure Adventure et méthodes
- `internal/adventure/party.go` - Groupe et inventaire partagé
- `internal/adventure/session.go` - Sessions de jeu
- `internal/adventure/journal.go` - Journal automatique
- `cmd/adventure/main.go` - CLI complète
- `.claude/skills/adventure-manager/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Création et gestion d'aventures/campagnes
- Ajout/retrait de personnages au groupe
- Inventaire partagé (or et objets)
- Sessions de jeu avec tracking du temps
- Journal automatique avec types d'événements
- Export Markdown du journal et des sessions

### Usage
```bash
./sw-adventure create "Nom" "Description"
./sw-adventure add-character "Aventure" "Personnage"
./sw-adventure start-session "Aventure"
./sw-adventure log "Aventure" combat "Description"
./sw-adventure end-session "Aventure" "Résumé"
./sw-adventure status "Aventure"
./sw-adventure journal "Aventure"
```

### Types de Journal
- `combat` ⚔️ - Rencontres et combats
- `loot` 💰 - Trésors trouvés
- `story` 📖 - Progression narrative
- `note` 📝 - Notes diverses
- `quest` 🎯 - Quêtes et objectifs
- `npc` 👤 - Interactions PNJ
- `location` 📍 - Nouveaux lieux
- `rest` 🏕️ - Repos
- `death` 💀 - Morts de personnages
- `levelup` ⬆️ - Montées de niveau

---

## Phase 4: Sous-agents [TERMINEE]

### Fichiers créés
- `.claude/agents/character-creator.md` - Guide de création de personnages
- `.claude/agents/rules-keeper.md` - Référence des règles BFRPG
- `.claude/agents/dungeon-master.md` - Maître du Jeu complet

### Agents disponibles

**character-creator**
- Guide la création de personnage étape par étape
- Explique les options et restrictions race/classe
- Utilise les skills dice-roller et character-generator
- Suggère des éléments de roleplay

**rules-keeper**
- Référence rapide des règles BFRPG
- Tables de combat, sauvegardes, compétences
- Arbitrage des situations ambiguës
- Réponses concises et précises

**dungeon-master**
- Narration immersive et concise
- Gestion des rencontres (combat, social, exploration)
- Incarnation des PNJ
- Tables de monstres et trésors
- Intégration avec adventure-manager pour le tracking

---

## Phase 4bis: Générateur de Noms [TERMINEE]

### Fichiers créés
- `data/names.json` - Dictionnaires de noms (~100 par race/genre)
- `internal/names/names.go` - Package génération de noms
- `cmd/names/main.go` - CLI complète
- `.claude/skills/name-generator/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Génération de noms par race (dwarf, elf, halfling, human)
- Support des genres (masculin, féminin, aléatoire)
- Génération multiple (--count=N)
- Prénoms seuls (--first-only)
- Noms de PNJ par type (innkeeper, merchant, guard, noble, wizard, villain)

### Usage
```bash
./sw-names generate dwarf --gender=m       # Nom de nain masculin
./sw-names generate elf --gender=f         # Nom d'elfe féminin
./sw-names generate human --count=5        # 5 noms humains
./sw-names npc innkeeper                   # Nom de tavernier
./sw-names npc villain                     # Nom de méchant
```

### Sources des noms
- Nains : collectés de fantasynamegenerators.com + classiques
- Elfes : style Tolkien/Sindarin
- Halfelins : style Hobbit (Tolkien)
- Humains : médiéval fantasy européen

---

## Phase 5: Générateur de PNJ [TERMINÉE]

### Fichiers créés
- `data/npc-traits.json` - Dictionnaires de traits (apparence, personnalité, motivations)
- `internal/npc/npc.go` - Package génération procédurale de PNJ
- `cmd/npc/main.go` - CLI complète
- `.claude/skills/npc-generator/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Génération de PNJ complets avec description physique détaillée
- Traits de personnalité (principal, secondaire, qualité, défaut)
- Motivations secrètes (objectif, peur, secret)
- Voix et comportement (ton, manière de parler, tic)
- Attitude envers les PJ (positive, neutre, négative)
- Intégration avec le générateur de noms (Phase 4bis)
- Ajustement automatique de l'apparence selon la race
- Export Markdown, JSON, et description courte

### Usage
```bash
./sw-npc generate                              # PNJ aléatoire complet
./sw-npc generate --race=dwarf --gender=m      # Nain masculin
./sw-npc generate --occupation=authority       # Figure d'autorité
./sw-npc generate --attitude=negative          # PNJ hostile
./sw-npc quick --count=5                       # 5 PNJ en description courte
./sw-npc generate --format=json                # Sortie JSON
./sw-npc list                                  # Options disponibles
```

### Types d'Occupation
| Type | Description | Exemples |
|------|-------------|----------|
| `commoner` | Gens du peuple | fermier, boulanger, serveur |
| `skilled` | Artisans qualifiés | marchand, apothicaire, musicien |
| `authority` | Figures d'autorité | garde, noble, magistrat |
| `underworld` | Monde criminel | voleur, espion, assassin |
| `religious` | Religieux | prêtre, moine, inquisiteur |
| `adventurer` | Aventuriers | chasseur de primes, mercenaire |

---

## Phase 6: Générateur d'Images [TERMINÉE]

### Fichiers créés
- `internal/image/image.go` - Client API fal.ai pour FLUX.1 [schnell]
- `internal/image/prompts.go` - Templates de prompts fantasy optimisés
- `cmd/image/main.go` - CLI complète
- `.claude/skills/image-generator/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Génération d'images via fal.ai FLUX.1 [schnell]
- Portraits de personnages et PNJ
- Scènes d'aventure (taverne, donjon, forêt, bataille...)
- Illustrations de monstres (20 types prédéfinis)
- Objets magiques (armes, potions, artefacts...)
- Vues de lieux (ville, château, donjon...)
- Prompts personnalisés
- 5 styles artistiques (realistic, painted, illustrated, dark_fantasy, epic)
- Téléchargement automatique dans `data/images/`

### Usage
```bash
./sw-image character "Aldric" --style=epic
./sw-image npc --race=dwarf --occupation=skilled
./sw-image scene "Combat contre un dragon" --type=battle
./sw-image monster dragon --style=dark_fantasy
./sw-image item weapon "épée flamboyante"
./sw-image location dungeon "Les Mines Perdues"
./sw-image custom "Un elfe archer dans une forêt enchantée"
./sw-image list
```

### API
- **Fournisseur** : fal.ai
- **Modèle** : FLUX.1 [schnell]
- **Coût** : ~$0.003/image
- **Vitesse** : ~2-5 secondes
- **Variable requise** : `FAL_KEY`

---

## Phase 7: Bestiaire BFRPG [TERMINÉE]

### Fichiers créés
- `data/monsters.json` - 33 monstres avec stats complètes + tables de rencontres
- `internal/monster/monster.go` - Package de gestion du bestiaire
- `cmd/monster/main.go` - CLI complète
- `.claude/skills/monster-manual/SKILL.md` - Skill Claude Code

### Fonctionnalités
- Consultation des fiches de monstres (stats, attaques, capacités)
- Recherche par nom ou type
- Génération de rencontres aléatoires par table ou niveau
- Création d'instances avec PV aléatoires
- 6 tables de rencontres (donjon niv 1-4, forêt, crypte)

### Usage
```bash
./sw-monster show goblin              # Fiche complète
./sw-monster search dragon            # Recherche
./sw-monster list --type=undead       # Par type
./sw-monster encounter dungeon_level_1 # Rencontre aléatoire
./sw-monster encounter --level=3      # Par niveau de groupe
./sw-monster roll orc --count=4       # 4 orcs avec PV
./sw-monster types                    # Types disponibles
```

### Monstres inclus (33)
- **Animaux** : rat géant, chauve-souris, loup, loup sinistre, ours
- **Humanoïdes** : gobelin, hobgobelin, kobold, orc, bugbear, gnoll
- **Morts-vivants** : squelette, zombie, goule, wight, spectre, vampire, liche
- **Monstres** : hibours, minotaure, harpie, cocatrice, basilic, méduse, rouilleur
- **Géants** : ogre, troll
- **Dragons** : dragon rouge (jeune, adulte)
- **Vases** : gelée verte, cube gélatineux
- **Vermines** : araignée géante, mille-pattes géant

---

## Phase 8: Tables de Trésors [TERMINÉE]

### Fichiers créés
- `data/treasure.json` - Tables de trésors A-U avec objets magiques
- `internal/treasure/treasure.go` - Package génération de trésors
- `cmd/treasure/main.go` - CLI complète
- `.claude/skills/treasure-generator/SKILL.md` - Skill Claude Code

### Fonctionnalités
- 21 types de trésors (A-U) selon les règles BFRPG
- Génération de pièces (cp, sp, ep, gp, pp)
- Génération de gemmes (6 tiers de valeur)
- Génération de bijoux (5 tiers de valeur)
- Objets magiques : potions (10), parchemins (6), anneaux (5), armes (11), armures (7), baguettes (5), objets divers (10)
- Probabilités configurables par type de trésor
- Export Markdown et JSON

### Usage
```bash
./sw-treasure generate R              # Trésor type R (Gobelin)
./sw-treasure generate A              # Trésor type A (Dragon)
./sw-treasure generate B --count=3    # 3 trésors type B
./sw-treasure types                   # Liste des types A-U
./sw-treasure info A                  # Probabilités du type A
./sw-treasure items potions           # Liste des potions
./sw-treasure items weapons           # Liste des armes magiques
```

### Types de Trésors
| Type | Description | Exemple |
|------|-------------|---------|
| A-H | Trésors de repaire | Dragon, Ogre, Orc |
| I-O | Trésors individuels | Garde, Mage |
| P-U | Trésors mineurs | Gobelin, Paysan |

---

## Architecture Finale

```
dungeons/
├── .claude/
│   ├── skills/
│   │   ├── dice-roller/         # Lancer de dés
│   │   ├── character-generator/ # Création de personnages
│   │   ├── adventure-manager/   # Gestion des aventures
│   │   ├── name-generator/      # Génération de noms
│   │   ├── npc-generator/       # Génération de PNJ
│   │   ├── image-generator/     # Génération d'images
│   │   ├── monster-manual/      # Bestiaire
│   │   └── treasure-generator/  # Génération de trésors
│   └── agents/
│       ├── dungeon-master.md    # Maître du Jeu
│       ├── rules-keeper.md      # Gardien des règles
│       └── character-creator.md # Guide de création
├── ai/
│   └── PLAN.md                  # Ce fichier
├── cmd/
│   ├── dice/main.go             # CLI dés
│   ├── character/main.go        # CLI personnages
│   ├── adventure/main.go        # CLI aventures
│   ├── names/main.go            # CLI noms
│   ├── npc/main.go              # CLI PNJ
│   ├── image/main.go            # CLI images
│   ├── monster/main.go          # CLI bestiaire
│   └── treasure/main.go         # CLI trésors
├── internal/
│   ├── dice/                    # Package dés
│   ├── data/                    # Chargement JSON
│   ├── character/               # Package personnages
│   ├── adventure/               # Package aventures
│   ├── names/                   # Package noms
│   ├── npc/                     # Package PNJ
│   ├── image/                   # Package images (fal.ai)
│   ├── monster/                 # Package bestiaire
│   └── treasure/                # Package trésors
├── data/
│   ├── races.json               # Données races BFRPG
│   ├── classes.json             # Données classes BFRPG
│   ├── equipment.json           # Équipement
│   ├── names.json               # Dictionnaires de noms
│   ├── npc-traits.json          # Traits de PNJ
│   ├── monsters.json            # Bestiaire BFRPG
│   ├── treasure.json            # Tables de trésors BFRPG
│   ├── characters/              # Personnages sauvegardés
│   ├── adventures/              # Aventures sauvegardées
│   └── images/                  # Images générées
├── CLAUDE.md                    # Instructions Claude Code
└── go.mod
```

---

## Améliorations Futures (non planifiées)

| # | Amélioration | Description |
|---|--------------|-------------|
| 1 | **Système de combat** | Résolution automatique des combats |
| 2 | **Carte de donjon** | Génération procédurale de donjons |
| 3 | **Progression de personnages** | Gestion XP et montée de niveau |
| 4 | **Filtrage des traits par genre** | Éviter "moustache" pour les femmes dans le générateur de PNJ |

---

## Anciennes sections (historique)

### cmd/
│   ├── dice/main.go
│   ├── character/main.go
│   └── adventure/main.go
├── internal/
│   ├── dice/
│   ├── data/
│   ├── character/
│   ├── adventure/
│   └── npc/
├── data/
│   ├── races.json
│   ├── classes.json
│   ├── equipment.json
│   ├── characters/
│   └── adventures/
├── CLAUDE.md
└── go.mod
```
