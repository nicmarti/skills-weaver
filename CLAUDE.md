# SkillsWeaver - Moteur de Jeu de Rôle avec Claude Code

## Description

**SkillsWeaver** est un moteur de jeu de rôle interactif basé sur les règles de **Basic Fantasy RPG** (BFRPG), orchestré par Claude Code. Il utilise des skills et des sous-agents pour gérer les différentes mécaniques du jeu.

Le préfixe `sw-` identifie toutes les commandes CLI du projet.

## But du Projet

Créer une expérience de jeu de rôle complète où Claude Code agit comme :
- **Lanceur de dés** via des scripts Go
- **Créateur de personnages** guidant le joueur
- **Maître du Jeu** pour les sessions de jeu

## Structure du Projet

```
skillsweaver/
├── .claude/
│   ├── skills/              # Skills Claude Code
│   │   ├── dice-roller/     # Lancer de dés
│   │   ├── character-generator/ # Création de personnages
│   │   ├── adventure-manager/   # Gestion des aventures
│   │   ├── name-generator/      # Génération de noms
│   │   ├── npc-generator/       # Génération de PNJ
│   │   ├── image-generator/     # Génération d'images
│   │   ├── journal-illustrator/ # Illustration de journaux
│   │   ├── monster-manual/      # Bestiaire
│   │   ├── treasure-generator/  # Génération de trésors
│   │   ├── equipment-browser/   # Catalogue d'équipement
│   │   └── spell-reference/     # Grimoire des sorts
│   └── agents/              # Sous-agents spécialisés
│       ├── character-creator.md
│       ├── rules-keeper.md
│       └── dungeon-master.md
├── cmd/
│   ├── dice/                # CLI sw-dice
│   ├── character/           # CLI sw-character
│   ├── adventure/           # CLI sw-adventure
│   ├── names/               # CLI sw-names
│   ├── npc/                 # CLI sw-npc
│   ├── location-names/      # CLI sw-location-names
│   ├── image/               # CLI sw-image
│   ├── monster/             # CLI sw-monster
│   ├── treasure/            # CLI sw-treasure
│   ├── equipment/           # CLI sw-equipment
│   └── spell/               # CLI sw-spell
├── internal/
│   ├── dice/                # Package lancer de dés
│   ├── data/                # Chargement données JSON
│   ├── character/           # Package personnages
│   ├── adventure/           # Package aventures/campagnes
│   ├── names/               # Package génération de noms
│   ├── npc/                 # Package génération de PNJ
│   ├── locations/           # Package génération de noms de lieux
│   ├── image/               # Package génération d'images
│   ├── monster/             # Package bestiaire
│   ├── treasure/            # Package trésors
│   ├── equipment/           # Package catalogue équipement
│   └── spell/               # Package grimoire des sorts
├── data/
│   ├── names.json           # Dictionnaires de noms
│   ├── npc-traits.json      # Traits pour les PNJ
│   ├── location-names.json  # Dictionnaires de noms de lieux
│   ├── monsters.json        # Bestiaire BFRPG
│   ├── treasure.json        # Tables de trésors BFRPG
│   ├── characters/          # Personnages sauvegardés
│   ├── adventures/          # Aventures sauvegardées
│   │   └── <nom-aventure>/
│   │       ├── adventure.json         # Métadonnées aventure
│   │       ├── sessions.json          # Historique sessions
│   │       ├── party.json             # Composition du groupe
│   │       ├── inventory.json         # Inventaire partagé
│   │       ├── journal-meta.json      # Métadonnées journal (NextID, Categories)
│   │       ├── journal-session-0.json # Journal hors session
│   │       ├── journal-session-1.json # Journal session 1
│   │       ├── journal-session-N.json # Journal session N
│   │       ├── images/
│   │       │   ├── session-0/         # Images hors session
│   │       │   ├── session-1/         # Images session 1
│   │       │   └── session-N/         # Images session N
│   │       └── characters/            # Personnages de l'aventure
│   └── images/              # Images générées (obsolète - maintenant par aventure)
├── ai/                      # Documentation et plans
└── CLAUDE.md                # Ce fichier
```

### Structure du Journal par Session

Le journal est organisé en fichiers séparés par session pour optimiser la performance :

- **journal-meta.json** : Métadonnées globales (NextID, Categories, LastUpdate)
- **journal-session-N.json** : Entrées pour la session N
- **journal-session-0.json** : Entrées hors session

**Avantages** :
- Réduit l'utilisation de tokens (charge uniquement les sessions nécessaires)
- Scalable (pas de limite de taille de journal)
- Organisation claire par session de jeu
- Images organisées de manière cohérente

**Migration** : Utilisez `sw-adventure migrate-journal <aventure>` pour convertir un ancien journal.json monolithique vers la nouvelle structure.
```

## Architecture : Skills vs Agents

### Définitions

**Skills** = Outils automatisables avec CLI
- Invoqués via `/skill-name` ou automatiquement par Claude
- Exécutent des commandes `sw-*`
- Retournent des données structurées
- Autonomes : peuvent fonctionner seuls ou être utilisés par des agents

**Agents** = Personnalités/Rôles spécialisés
- Guident l'utilisateur avec contexte narratif
- Utilisent les skills comme outils
- Maintiennent un style et ton cohérent
- Orchestrent plusieurs skills pour accomplir des tâches complexes

### Hiérarchie

```
┌─────────────────────────────────────────────────────────┐
│                      UTILISATEUR                        │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                       AGENTS                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │
│  │ dungeon-    │ │ character-  │ │ rules-keeper    │   │
│  │ master      │ │ creator     │ │ (arbitre)       │   │
│  └──────┬──────┘ └──────┬──────┘ └────────┬────────┘   │
└─────────┼───────────────┼─────────────────┼────────────┘
          │               │                 │
┌─────────▼───────────────▼─────────────────▼────────────┐
│                       SKILLS                            │
│  ┌────────────┐ ┌────────────┐ ┌────────────────────┐  │
│  │dice-roller │ │character-  │ │adventure-manager   │  │
│  │            │ │generator   │ │                    │  │
│  └────────────┘ └────────────┘ └────────────────────┘  │
│  ┌────────────┐ ┌────────────┐ ┌────────────────────┐  │
│  │name-       │ │npc-        │ │image-generator     │  │
│  │generator   │ │generator   │ │                    │  │
│  └────────────┘ └────────────┘ └────────────────────┘  │
│  ┌────────────┐ ┌────────────┐ ┌────────────────────┐  │
│  │name-       │ │monster-    │ │treasure-generator  │  │
│  │location-   │ │manual      │ │                    │  │
│  │generator   │ │            │ │                    │  │
│  └────────────┘ └────────────┘ └────────────────────┘  │
│  ┌────────────┐ ┌────────────┐ ┌────────────────────┐  │
│  │equipment-  │ │spell-      │ │journal-illustrator │  │
│  │browser     │ │reference   │ │                    │  │
│  └────────────┘ └────────────┘ └────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                    CLI (sw-*)                           │
│  sw-dice, sw-character, sw-adventure, sw-names,        │
│  sw-npc, sw-location-names, sw-image, sw-monster,      │
│  sw-treasure, sw-equipment, sw-spell                   │
└─────────────────────────────────────────────────────────┘
```

### Workflow typique : Création de personnage

1. Utilisateur : "Je veux créer un personnage"
2. **Agent** `character-creator` guide la conversation (race, classe, nom)
3. **Skill** `dice-roller` lance les stats (4d6kh3)
4. **Skill** `name-generator` propose des noms
5. **Skill** `character-generator` sauvegarde le personnage

### Workflow typique : Session de jeu

1. Utilisateur : "Lançons une session"
2. **Agent** `dungeon-master` narre l'aventure
3. **Skill** `adventure-manager` gère l'état (session, journal)
4. **Skill** `dice-roller` résout les actions
5. **Skill** `monster-manual` fournit les stats des ennemis
6. **Skill** `treasure-generator` génère le butin
7. **Skill** `image-generator` illustre les moments clés

## Outils Disponibles

### CLI sw-dice

Lancer des dés avec notation standard RPG :

```bash
# Compiler
go build -o sw-dice ./cmd/dice

# Utiliser
./sw-dice roll d20              # Lance 1d20
./sw-dice roll 2d6+3            # Lance 2d6, ajoute 3
./sw-dice roll 4d6kh3           # Lance 4d6, garde les 3 plus hauts
./sw-dice roll d20 --advantage  # Avantage (2d20, garde le plus haut)
./sw-dice stats                 # Génère 6 caractéristiques (4d6kh3)
./sw-dice stats --classic       # Méthode classique (3d6)
```

### Skill dice-roller

La skill `dice-roller` permet à Claude de lancer des dés automatiquement pendant une session. Elle est découverte automatiquement quand on parle de jets de dés.

### CLI sw-dm (Dungeon Master Agent)

Application interactive de Maître du Jeu autonome avec boucle d'agent complète :

```bash
# Compiler
go build -o sw-dm ./cmd/dm

# Lancer l'application
./sw-dm

# L'application propose un menu pour sélectionner l'aventure
# Puis démarre une session REPL interactive avec streaming
```

**Fonctionnalités** :
- Boucle d'agent complète avec tool_use (Anthropic API)
- Streaming des réponses pour une expérience immersive
- Auto-chargement du contexte d'aventure (groupe, inventaire, journal)
- Accès direct aux packages Go (dice, monster, treasure, npc, etc.)
- Interface REPL avec historique de conversation

**Tools disponibles pour l'agent** :
- `roll_dice` : Lancer des dés avec notation RPG
- `get_monster` : Consulter les stats d'un monstre
- `log_event` : Enregistrer un événement dans le journal
- `add_gold` : Modifier l'or du groupe
- `get_inventory` : Consulter l'inventaire partagé
- `generate_treasure` : Générer un trésor BFRPG
- `generate_npc` : Créer un PNJ complet
- `generate_image` : Générer une illustration fantasy (requiert FAL_KEY)

**Architecture** :
- `internal/agent/` : Orchestration de la boucle d'agent
  - `agent.go` : Boucle principale avec tool execution
  - `tools.go` : Système de registry des tools
  - `context.go` : Gestion contexte conversation/aventure
  - `streaming.go` : Traitement événements streaming
- `internal/dmtools/` : Wrappers des tools pour l'agent
- `cmd/dm/main.go` : Application REPL

**Prérequis** :
- Variable d'environnement `ANTHROPIC_API_KEY` configurée
- Une aventure existante dans `data/adventures/`

### CLI sw-character

Créer et gérer des personnages BFRPG :

```bash
# Compiler
go build -o sw-character ./cmd/character

# Créer un personnage
./sw-character create "Aldric" --race=human --class=fighter
./sw-character create "Lyra" --race=elf --class=magic-user --method=classic

# Gérer
./sw-character list              # Liste tous les personnages
./sw-character show "Aldric"     # Affiche la fiche
./sw-character delete "Aldric"   # Supprime
./sw-character export "Aldric" --format=json
```

### Skill character-generator

La skill `character-generator` permet à Claude de créer des personnages en guidant le joueur étape par étape.

### CLI sw-adventure

Gérer des aventures et campagnes BFRPG :

```bash
# Compiler
go build -o sw-adventure ./cmd/adventure

# Créer une aventure
./sw-adventure create "La Mine Perdue" "Une aventure dans les montagnes"

# Gérer le groupe
./sw-adventure add-character "La Mine Perdue" "Aldric"
./sw-adventure party "La Mine Perdue"

# Sessions de jeu
./sw-adventure start-session "La Mine Perdue"
./sw-adventure log "La Mine Perdue" combat "Combat contre 3 gobelins"
./sw-adventure add-gold "La Mine Perdue" 50 "Trésor gobelin"
./sw-adventure end-session "La Mine Perdue" "Premier niveau exploré"

# Consulter
./sw-adventure status "La Mine Perdue"    # Statut complet
./sw-adventure journal "La Mine Perdue"   # Journal de l'aventure
./sw-adventure sessions "La Mine Perdue"  # Historique des sessions
./sw-adventure inventory "La Mine Perdue" # Inventaire partagé

# Maintenance - Migration vers structure par session
./sw-adventure migrate-journal "La Mine Perdue"    # Migrer journal.json vers fichiers session
./sw-adventure validate-journal "La Mine Perdue"   # Valider intégrité des journaux
```

**Note** : Les aventures existantes avec `journal.json` monolithique sont automatiquement supportées. La migration vers la structure par session est optionnelle mais recommandée pour améliorer les performances.

### Skill adventure-manager

La skill `adventure-manager` permet à Claude de gérer les aventures, suivre les sessions et maintenir le journal automatique.

### CLI sw-names

Générer des noms de personnages fantasy :

```bash
# Compiler
go build -o sw-names ./cmd/names

# Générer des noms par race
./sw-names generate dwarf                    # Nom de nain
./sw-names generate elf --gender=f           # Nom d'elfe féminin
./sw-names generate human --count=5          # 5 noms humains
./sw-names generate halfling --first-only    # Prénom de halfelin

# Générer des noms de PNJ
./sw-names npc innkeeper                     # Nom de tavernier
./sw-names npc merchant                      # Nom de marchand
./sw-names npc villain                       # Nom de méchant

# Lister les options
./sw-names list                              # Toutes les options
```

### Skill name-generator

La skill `name-generator` permet à Claude de générer des noms pour les joueurs et les PNJ selon la race et le type.

### CLI sw-npc

Générer des PNJ complets :

```bash
# Compiler
go build -o sw-npc ./cmd/npc

# Générer un PNJ complet
./sw-npc generate                              # PNJ aléatoire
./sw-npc generate --race=dwarf --gender=m      # Nain masculin
./sw-npc generate --occupation=authority       # Figure d'autorité
./sw-npc generate --attitude=hostile           # PNJ hostile

# Génération rapide
./sw-npc quick --count=5                       # 5 PNJ en une ligne

# Formats de sortie
./sw-npc generate --format=md                  # Markdown (défaut)
./sw-npc generate --format=json                # JSON
./sw-npc generate --format=short               # Une ligne
```

### Skill npc-generator

La skill `npc-generator` permet à Claude de créer des PNJ complets avec apparence, personnalité, motivations et secrets.

### CLI sw-location-names

Générer des noms de lieux cohérents avec les 4 factions :

```bash
# Compiler
go build -o sw-location-names ./cmd/location-names

# Générer des noms par royaume
./sw-location-names city --kingdom=valdorine    # Cité maritime
./sw-location-names town --kingdom=karvath      # Bourg militaire
./sw-location-names village --kingdom=lumenciel # Village religieux
./sw-location-names region --kingdom=astrene    # Région mélancolique

# Lieux neutres
./sw-location-names ruin                        # Ruines anciennes
./sw-location-names generic                     # Lieu géographique
./sw-location-names special                     # Terres Brûlées, etc.

# Génération multiple
./sw-location-names city --kingdom=valdorine --count=5

# Lister les options
./sw-location-names list                        # Tout
./sw-location-names list kingdoms               # Royaumes
./sw-location-names list types                  # Types de lieux
```

### Skill name-location-generator

La skill `name-location-generator` permet à Claude de générer des noms de lieux (cités, villages, régions) cohérents avec les 4 factions. Utilise des styles distincts par royaume : valdorine maritime, karvath militaire, lumenciel religieux, astrène mélancolique.

### CLI sw-image

Générer des images heroic fantasy via fal.ai FLUX.1 :

```bash
# Compiler
go build -o sw-image ./cmd/image

# Prérequis: variable d'environnement FAL_KEY
export FAL_KEY="votre_clé_fal_ai"

# Portrait de personnage existant
./sw-image character "Aldric" --style=epic

# Portrait de PNJ
./sw-image npc --race=dwarf --gender=m --occupation=skilled

# Scène d'aventure
./sw-image scene "Combat contre des gobelins" --type=battle

# Monstre
./sw-image monster dragon --style=dark_fantasy

# Objet magique
./sw-image item weapon "épée flamboyante"

# Lieu
./sw-image location dungeon "Les Mines Perdues"

# Prompt personnalisé
./sw-image custom "Un groupe d'aventuriers dans une taverne"

# Lister les options
./sw-image list
```

### Skill image-generator

La skill `image-generator` permet à Claude de générer des illustrations fantasy pour enrichir l'expérience de jeu : portraits, scènes, monstres, objets et lieux.

### Commande journal (sw-image)

Illustrer automatiquement le journal d'une aventure :

```bash
# Prévisualiser les prompts (sans générer d'images)
./sw-image journal "la-crypte-des-ombres" --dry-run

# Générer toutes les illustrations (parallèle)
./sw-image journal "la-crypte-des-ombres"

# Limiter le nombre d'images
./sw-image journal "la-crypte-des-ombres" --max=5

# Filtrer par type
./sw-image journal "la-crypte-des-ombres" --types=combat,discovery

# Ajuster le parallélisme (1-8)
./sw-image journal "la-crypte-des-ombres" --parallel=8
```

Types illustrables : `combat`, `exploration`, `discovery`, `loot`, `session`

Les images sont sauvegardées dans `data/adventures/<nom>/images/`

### Skill journal-illustrator

La skill `journal-illustrator` permet à Claude d'illustrer automatiquement les journaux d'aventures avec des prompts optimisés par type d'événement et une génération parallèle.

### CLI sw-monster

Consulter le bestiaire et générer des rencontres :

```bash
# Compiler
go build -o sw-monster ./cmd/monster

# Consulter un monstre
./sw-monster show goblin              # Fiche complète
./sw-monster show dragon_red_adult    # Dragon rouge adulte
./sw-monster search undead            # Recherche par type

# Lister les monstres
./sw-monster list                     # Tous les monstres
./sw-monster list --type=humanoid    # Par type
./sw-monster types                    # Types disponibles

# Générer une rencontre
./sw-monster encounter dungeon_level_1  # Niveau 1
./sw-monster encounter --level=3        # Par niveau de groupe
./sw-monster encounter forest           # En forêt

# Créer des ennemis avec PV
./sw-monster roll orc --count=4       # 4 orcs avec PV aléatoires
./sw-monster roll goblin --count=6    # 6 gobelins
```

### Skill monster-manual

La skill `monster-manual` permet à Claude de consulter les stats des monstres et générer des rencontres équilibrées pendant les sessions de jeu.

### CLI sw-treasure

Générer des trésors selon les tables BFRPG :

```bash
# Compiler
go build -o sw-treasure ./cmd/treasure

# Générer un trésor
./sw-treasure generate R              # Trésor type R (Gobelin)
./sw-treasure generate A              # Trésor type A (Dragon)
./sw-treasure generate B --count=3    # 3 trésors type B

# Lister les types de trésors
./sw-treasure types                   # Tous les types A-U

# Détails d'un type
./sw-treasure info A                  # Probabilités du type A

# Lister les objets magiques
./sw-treasure items                   # Catégories disponibles
./sw-treasure items potions           # Toutes les potions
./sw-treasure items weapons           # Armes magiques
./sw-treasure items armor             # Armures magiques
```

### Skill treasure-generator

La skill `treasure-generator` permet à Claude de générer des trésors appropriés après les combats, en respectant les types de trésors assignés aux monstres.

### CLI sw-equipment

Consulter le catalogue d'équipement BFRPG :

```bash
# Compiler
go build -o sw-equipment ./cmd/equipment

# Lister les armes
./sw-equipment weapons                    # Toutes les armes
./sw-equipment weapons --type=melee      # Armes de mêlée
./sw-equipment weapons --type=ranged     # Armes à distance

# Lister les armures
./sw-equipment armor                      # Toutes les armures
./sw-equipment armor --type=heavy        # Armures lourdes

# Équipement d'aventure
./sw-equipment gear                       # Liste l'équipement
./sw-equipment ammo                       # Munitions

# Afficher un item
./sw-equipment show longsword            # Détails de l'épée longue
./sw-equipment search épée               # Recherche par nom FR/EN

# Équipement de départ
./sw-equipment starting fighter          # Équipement guerrier
./sw-equipment starting magic-user       # Équipement magicien
```

### Skill equipment-browser

La skill `equipment-browser` permet à Claude de consulter les armes, armures et équipement avec leurs statistiques (dégâts, CA, coût, propriétés).

### CLI sw-spell

Consulter le grimoire des sorts BFRPG :

```bash
# Compiler
go build -o sw-spell ./cmd/spell

# Lister les sorts
./sw-spell list                              # Tous les sorts
./sw-spell list --class=cleric              # Sorts de clerc
./sw-spell list --class=magic-user          # Sorts de magicien
./sw-spell list --class=cleric --level=1    # Clerc niveau 1

# Afficher un sort
./sw-spell show magic_missile               # Détails du projectile magique
./sw-spell show cure_light_wounds           # Soins légers

# Rechercher
./sw-spell search lumière                   # Recherche par nom FR/EN

# Sorts réversibles
./sw-spell reversible                       # Liste les sorts avec forme inversée
```

### Skill spell-reference

La skill `spell-reference` permet à Claude de consulter les sorts par classe et niveau, avec leurs effets détaillés (portée, durée, descriptions).

### CLI sw-validate

Valider les données de jeu :

```bash
# Compiler
go build -o sw-validate ./cmd/validate

# Valider toutes les données
./sw-validate                 # Affichage texte
./sw-validate --json          # Sortie JSON (CI/CD)
./sw-validate --data /path    # Répertoire personnalisé

# Aide
./sw-validate help
```

**Validations effectuées** :
- `races.json` : allowed_classes référencent des classes valides
- `equipment.json` : starting_equipment référence des items valides
- `monsters.json` : treasure_type valide (A-U ou 'none')
- `names.json` : toutes les races ont des entrées de noms
- `spells.json` : spell_lists référencent des sorts valides

## Sous-Agents Spécialisés

Les agents sont disponibles dans `.claude/agents/` :

### character-creator
Guide interactif pour créer des personnages étape par étape. Explique les races, classes, et aide à faire des choix cohérents.

### rules-keeper
Référence rapide des règles BFRPG. Répond aux questions sur le combat, la magie, les jets de sauvegarde et arbitre les situations.

### dungeon-master
Maître du Jeu complet. Narration immersive, gestion des rencontres, incarnation des PNJ, et tracking automatique via les commandes sw-adventure.

## Règles BFRPG

### Races Disponibles

| Race | Modificateurs | Classes Autorisées |
|------|--------------|-------------------|
| Humain | Aucun | Toutes |
| Elfe | +1 DEX, -1 CON | Guerrier (6), Magicien (9), Voleur |
| Nain | +1 CON, -1 CHA | Guerrier (7), Clerc (6), Voleur |
| Halfelin | +1 DEX, -1 FOR | Guerrier (4), Voleur |

### Classes Disponibles

| Classe | Dé de Vie | Armes | Armures |
|--------|-----------|-------|---------|
| Guerrier | d8 | Toutes | Toutes |
| Clerc | d6 | Contondantes | Toutes |
| Magicien | d4 | Dague, bâton | Aucune |
| Voleur | d4 | Toutes | Cuir |


## Règles d'Utilisation des CLI

### Accès Direct (Claude Code)
Les CLI `sw-*` peuvent être utilisés directement pour :
- Jets de dés ponctuels
- Consultation de données (show, list, status)
- Commandes de debug

### Via Agents/Skills
Utilisez les sous-agents specialisés pour :
- Sessions de jeu complètes (dungeon-master)
- Création guidée de personnages (character-creator)
- Arbitrage de règles (rules-keeper)

## Commandes de Développement

```bash
# Compiler tous les outils SkillsWeaver
make

# Lancer les tests
make test

# Tester le système de dés
go test ./internal/dice/... -v

# Tester le chargement des données
go test ./internal/data/... -v

# Tester le générateur de personnages
go test ./internal/character/... -v
```

## Conventions de Développement

### Ajout de nouveaux packages dans `internal/`

Lors de l'ajout d'un nouveau package dans `internal/` pour supporter une skill :

1. **Mettre à jour le Makefile** avec les nouvelles dépendances
   - Ajouter le package aux dépendances du binaire concerné
   - Exemple : Si vous créez `internal/combat/` utilisé par `cmd/adventure`, modifier :
     ```makefile
     $(BINARY_PREFIX)-adventure: cmd/adventure/main.go internal/adventure/*.go internal/combat/*.go
     ```

2. **Créer des tests unitaires**
   - Tout nouveau package dans `internal/` doit avoir des tests
   - Créer `<package>_test.go` dans le même répertoire
   - Lancer `make test` pour vérifier que tous les tests passent

3. **Vérifier la compilation**
   ```bash
   # Nettoyer et recompiler pour vérifier les dépendances
   make clean
   make

   # Vérifier que les modifications du package déclenchent la recompilation
   touch internal/<package>/<file>.go
   make <binary-name>
   ```

### Packages actuellement dans `internal/`

| Package | Utilisé par | Tests | Makefile |
|---------|-------------|-------|----------|
| `adventure` | `sw-adventure` | ✓ | ✓ |
| `ai` | `sw-adventure` | - | ✓ |
| `character` | `sw-character` | ✓ | ✓ |
| `combat` | (orphelin) | ✓ | - |
| `data` | `sw-character` | ✓ | ✓ |
| `dice` | `sw-dice`, `sw-monster`, `sw-treasure` | ✓ | ✓ |
| `equipment` | `sw-equipment` | - | ✓ |
| `image` | `sw-image` | - | ✓ |
| `locations` | `sw-location-names` | ✓ | ✓ |
| `monster` | `sw-monster` | ✓ | ✓ |
| `names` | `sw-names`, `sw-npc` | ✓ | ✓ |
| `npc` | `sw-npc` | ✓ | ✓ |
| `spell` | `sw-spell` | - | ✓ |
| `treasure` | `sw-treasure` | ✓ | ✓ |

## Conventions Git

### Commits
- **Langue** : Anglais uniquement
- **Format** : `<type>: <description>`
- **Types** : `feat`, `fix`, `refactor`, `test`, `docs`, `chore`
- **Ne pas mentionner** : Claude Code, Claude, AI, ou LLM dans les messages de commit

### Exemples
```bash
git commit -m "feat: add combat system with initiative tracking"
git commit -m "fix: validate race/class combinations in character creation"
git commit -m "test: add unit tests for dice roller"
git commit -m "docs: update rules-keeper with BFRPG combat rules"
```

## Ressources

- [Basic Fantasy RPG](https://www.basicfantasy.org/) - Règles complètes (gratuit)
- [SRD BFRPG](https://www.basicfantasy.org/srd/) - System Reference Document
