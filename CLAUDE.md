# SkillsWeaver - Moteur de Jeu de Rôle avec Claude Code

## Description

**SkillsWeaver** est un moteur de jeu de rôle interactif basé sur les règles de **D&D 5e** (5ème édition), orchestré par Claude Code. Il utilise des skills et des sous-agents pour gérer les différentes mécaniques du jeu.

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
│   │   ├── spell-reference/     # Grimoire des sorts
│   │   └── map-generator/       # Génération de prompts pour cartes 2D
│   └── agents/              # Sous-agents spécialisés
│       ├── character-creator.md
│       ├── rules-keeper.md
│       └── dungeon-master.md
├── cmd/
│   ├── dice/                # CLI sw-dice
│   ├── character/           # CLI sw-character
│   ├── character-sheet/     # CLI sw-character-sheet
│   ├── adventure/           # CLI sw-adventure
│   ├── names/               # CLI sw-names
│   ├── npc/                 # CLI sw-npc
│   ├── location-names/      # CLI sw-location-names
│   ├── image/               # CLI sw-image
│   ├── monster/             # CLI sw-monster
│   ├── treasure/            # CLI sw-treasure
│   ├── equipment/           # CLI sw-equipment
│   ├── spell/               # CLI sw-spell
│   └── map/                 # CLI sw-map
├── internal/
│   ├── dice/                # Package lancer de dés
│   ├── data/                # Chargement données JSON
│   ├── character/           # Package personnages
│   ├── charactersheet/      # Package génération fiches HTML
│   ├── adventure/           # Package aventures/campagnes
│   ├── names/               # Package génération de noms
│   ├── npc/                 # Package génération de PNJ
│   ├── locations/           # Package génération de noms de lieux
│   ├── image/               # Package génération d'images
│   ├── monster/             # Package bestiaire
│   ├── treasure/            # Package trésors
│   ├── equipment/           # Package catalogue équipement
│   ├── spell/               # Package grimoire des sorts
│   ├── map/                 # Package génération prompts cartes
│   └── world/               # Package données géographiques
├── data/
│   ├── names.json           # Dictionnaires de noms
│   ├── npc-traits.json      # Traits pour les PNJ
│   ├── location-names.json  # Dictionnaires de noms de lieux
│   ├── monsters.json        # Bestiaire D&D 5e
│   ├── treasure.json        # Tables de trésors D&D 5e
│   ├── characters/          # Personnages sauvegardés
│   ├── maps/                # Prompts et images de cartes
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

### Système de Persistance des PNJ

Les PNJ générés sont automatiquement sauvegardés et gérés via un système à deux niveaux :

#### 1. Fichier par Aventure : `npcs-generated.json`

**Localisation** : `data/adventures/<nom>/npcs-generated.json`

**Structure** :
```json
{
  "sessions": {
    "session_0": [
      {
        "id": "npc_001",
        "generated_at": "2025-12-24T19:39:02Z",
        "session_number": 0,
        "npc": { /* NPC complet */ },
        "context": "Taverne du Voile Écarlate, informateur",
        "importance": "mentioned",  // mentioned < interacted < recurring < key
        "notes": ["Note 1", "Note 2"],
        "appearances": 1,
        "promoted_to_world": false,
        "world_keeper_notes": "Validation world-keeper"
      }
    ],
    "session_1": [...]
  },
  "next_id": 2
}
```

**Niveaux d'importance** :
- `mentioned` : Généré mais pas d'interaction
- `interacted` : Dialogue ou rencontre brève
- `recurring` : Apparitions multiples
- `key` : Importance majeure pour l'intrigue

**Capture automatique** : Tous les PNJ générés via `generate_npc` sont automatiquement sauvegardés.

#### 2. Fichier Monde : `data/world/npcs.json`

**PNJ promus** : Seuls les PNJ récurrents et importants sont promus vers `npcs.json` après validation par le world-keeper.

**Workflow de promotion** :
1. World-keeper review : `/world-review-npcs <adventure>`
2. Validation et enrichissement : `/world-promote-npc <adventure> <nom>`
3. Ajout à `data/world/npcs.json` avec contexte complet

#### Tools Disponibles dans sw-dm

**`generate_npc`** : Génère un PNJ et le sauvegarde automatiquement
```json
{
  "race": "human",
  "gender": "m",
  "occupation": "skilled",
  "attitude": "neutral",
  "context": "Taverne du Voile Écarlate, demande informations"
}
```

**`update_npc_importance`** : Met à jour l'importance d'un PNJ
```json
{
  "npc_name": "Grimbold Dreamcatcher",
  "importance": "interacted",
  "note": "A révélé information sur Vaskir"
}
```

**`get_npc_history`** : Consulte l'historique complet d'un PNJ
```json
{
  "npc_name": "Grimbold Dreamcatcher"
}
```

#### Avantages du Système

✅ **Aucune perte** : Tous les PNJ générés sont capturés automatiquement
✅ **Évolution naturelle** : L'importance augmente au fil des interactions
✅ **Validation centralisée** : World-keeper garantit la cohérence
✅ **Scalable** : Fonctionne avec 5 ou 50 PNJ par aventure
✅ **Séparation claire** : Adventure (brouillon) vs World (canon)

#### Exemple de Workflow Complet

```
┌─ PENDANT SESSION ─────────────────────────────┐
│ 1. DM: generate_npc → Grimbold               │
│ 2. ✓ Auto-saved dans npcs-generated.json    │
│    (section session_0, importance="mentioned")│
│                                               │
│ 3. Plus tard, PJ dialogue avec Grimbold      │
│ 4. DM: update_npc_importance("Grimbold",     │
│    importance="interacted", notes="Révélé    │
│    info sur Vaskir")                         │
└───────────────────────────────────────────────┘

┌─ POST-SESSION (World-Keeper) ─────────────────┐
│ 1. /world-keeper /world-review-npcs          │
│    "la-crypte-des-ombres"                    │
│ 2. Identifie PNJ avec importance >= interacted│
│ 3. /world-keeper /world-promote-npc          │
│    "la-crypte-des-ombres" "Grimbold"         │
│ 4. Validation, enrichissement, promotion      │
│ 5. ✓ Ajouté à data/world/npcs.json          │
└───────────────────────────────────────────────┘
```
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

**Gestion de session** (CRITIQUE pour le journal) :
- `start_session` : Démarrer une nouvelle session de jeu (OBLIGATOIRE au début)
- `end_session` : Terminer la session avec résumé (OBLIGATOIRE à la fin)
- `get_session_info` : Consulter l'état de la session active

**Mécanique de jeu** :
- `roll_dice` : Lancer des dés avec notation RPG
- `get_monster` : Consulter les stats d'un monstre
- `log_event` : Enregistrer un événement dans le journal
- `add_gold` : Modifier l'or du groupe
- `get_inventory` : Consulter l'inventaire partagé

**Génération de contenu** :
- `generate_treasure` : Générer un trésor D&D 5e
- `generate_npc` : Créer un PNJ complet (auto-sauvegardé)
- `generate_image` : Générer une illustration fantasy (requiert FAL_KEY)
- `generate_map` : Générer prompt carte 2D avec validation world-keeper

**NPC Management** :
- `update_npc_importance` : Mettre à jour l'importance d'un PNJ
- `get_npc_history` : Consulter l'historique d'un PNJ

**Consultation des Personnages** :
- `get_party_info` : Vue d'ensemble du groupe (PV, CA, niveau, stat principale)
- `get_character_info` : Fiche détaillée d'un personnage (caractéristiques, modificateurs, équipement, apparence)

**Consultation Équipement et Sorts** :
- `get_equipment` : Consulter armes, armures, équipement (dégâts, CA, coût, propriétés)
- `get_spell` : Consulter sorts par classe/niveau (portée, durée, effets, forme inversée)

**Génération de Rencontres** :
- `generate_encounter` : Générer rencontre équilibrée par table ou niveau de groupe
- `roll_monster_hp` : Créer instances de monstres avec PV aléatoires pour combat

**Gestion Inventaire** :
- `add_item` : Ajouter objet à l'inventaire partagé (avec log automatique)
- `remove_item` : Retirer objet de l'inventaire (consommation, vente)

**Génération de Noms** :
- `generate_name` : Noms de personnages par race/genre ou type PNJ
- `generate_location_name` : Noms de lieux par royaume et type

**IMPORTANT** : L'agent dungeon-master DOIT appeler `start_session` au début de chaque partie et `end_session` à la fin. Sans cela, tous les événements seront enregistrés dans `journal-session-0.json` au lieu d'être correctement organisés par session.

**Architecture** :
- `internal/agent/` : Orchestration de la boucle d'agent
  - `agent.go` : Boucle principale avec tool execution
  - `tools.go` : Système de registry des tools
  - `context.go` : Gestion contexte conversation/aventure
  - `streaming.go` : Traitement événements streaming
  - `register_tools.go` : Enregistrement de tous les tools
- `internal/dmtools/` : Wrappers des tools pour l'agent
  - `simple_tools.go` : Tools basiques (log_event, add_gold, etc.)
  - `session_tools.go` : Gestion de session (start/end/get_info)
  - `character_tools.go` : Consultation personnages (get_party_info, get_character_info)
  - `equipment_tools.go` : Consultation équipement (get_equipment)
  - `spell_tools.go` : Consultation sorts (get_spell)
  - `encounter_tools.go` : Génération rencontres (generate_encounter, roll_monster_hp)
  - `inventory_tools.go` : Gestion inventaire (add_item, remove_item)
  - `name_tools.go` : Génération noms (generate_name, generate_location_name)
  - `dice_tool.go`, `monster_tool.go`, `npc_management_tools.go`, etc.
- `cmd/dm/main.go` : Application REPL

**Prérequis** :
- Variable d'environnement `ANTHROPIC_API_KEY` configurée
- Une aventure existante dans `data/adventures/`

**Interface Utilisateur** :
- ✅ **Édition de ligne complète** : Utilise `readline` pour une expérience professionnelle
  - Touches fléchées (←, →) pour naviguer dans la ligne
  - Home/End, Ctrl+A/Ctrl+E pour début/fin de ligne
  - Backspace/Delete pour supprimer des caractères
  - Ctrl+W pour supprimer un mot
- ✅ **Historique des commandes** : Navigation avec ↑/↓
  - Historique persistant entre sessions (`/tmp/sw-dm-history.txt`)
  - Ctrl+R pour recherche dans l'historique
- ✅ **Gestion propre des signaux** :
  - Ctrl+C avec ligne vide = quitter
  - Ctrl+D = quitter proprement
  - Ctrl+L = effacer l'écran
- ✅ **Aucun caractère de contrôle visible** : Les séquences ANSI sont gérées en interne

**Note** : Voir `docs/readline-integration.md` pour plus de détails sur l'interface utilisateur.

**Logging automatique des commandes CLI** : Chaque tool appelé par sw-dm est automatiquement loggé avec sa commande CLI équivalente dans `data/adventures/<nom>/sw-dm-session-N.log` (un fichier par session pour éviter les fichiers trop gros). Cela permet de :
- Reproduire facilement les opérations (copier-coller la commande)
- Tester avec des paramètres différents
- Déboguer et améliorer les outils

Exemple de log :
```
[2025-12-25 19:30:45] TOOL CALL: generate_map (ID: toolu_01Abc...)
  Parameters:
  {
    "type": "city",
    "name": "Port-Sombre",
    "kingdom": "valdorine"
  }
  Equivalent CLI:
  ./sw-map generate city "Port-Sombre" --kingdom=valdorine
```

Extraction des commandes :
```bash
# Toutes les commandes de toutes les aventures
./scripts/extract-cli-commands.sh

# Commandes d'une aventure spécifique
./scripts/extract-cli-commands.sh la-crypte-des-ombres

# Commandes d'un tool spécifique
./scripts/extract-cli-commands.sh la-crypte-des-ombres generate_map

# Grep manuel (cherche dans tous les fichiers de log)
grep "Equivalent CLI:" data/adventures/*/sw-dm*.log
```

**Note** : Les logs sont maintenant créés par session (`sw-dm-session-N.log`) pour éviter des fichiers trop gros. Le script d'extraction cherche automatiquement dans tous les fichiers. Voir `docs/log-rotation.md` pour plus de détails.

Voir `docs/cli-logging-example.md` pour plus d'exemples et de patterns d'utilisation.

### Skill character-generator

La skill `character-generator` permet à Claude de créer des personnages en guidant le joueur étape par étape.


### Skill adventure-manager

La skill `adventure-manager` permet à Claude de gérer les aventures, suivre les sessions et maintenir le journal automatique.


### Skill name-generator

La skill `name-generator` permet à Claude de générer des noms pour les joueurs et les PNJ selon la race et le type.


### Skill npc-generator

La skill `npc-generator` permet à Claude de créer des PNJ complets avec apparence, personnalité, motivations et secrets.


### Skill name-location-generator

La skill `name-location-generator` permet à Claude de générer des noms de lieux (cités, villages, régions) cohérents avec les 4 factions. Utilise des styles distincts par royaume : valdorine maritime, karvath militaire, lumenciel religieux, astrène mélancolique.


### Skill image-generator

La skill `image-generator` permet à Claude de générer des illustrations fantasy pour enrichir l'expérience de jeu : portraits, scènes, monstres, objets et lieux.

### Skill journal-illustrator

La skill `journal-illustrator` permet à Claude d'illustrer automatiquement les journaux d'aventures avec des prompts optimisés par type d'événement et une génération parallèle.

### Skill map-generator

La skill `map-generator` permet à Claude de générer des prompts enrichis pour cartes 2D fantasy avec validation world-keeper. Elle assure la cohérence des noms de lieux et des styles architecturaux des 4 royaumes.


### Skill monster-manual

La skill `monster-manual` permet à Claude de consulter les stats des monstres et générer des rencontres équilibrées pendant les sessions de jeu.


### Skill treasure-generator

La skill `treasure-generator` permet à Claude de générer des trésors appropriés après les combats, en respectant les types de trésors assignés aux monstres.


### Skill equipment-browser

La skill `equipment-browser` permet à Claude de consulter les armes, armures et équipement avec leurs statistiques (dégâts, CA, coût, propriétés).


### Skill spell-reference

La skill `spell-reference` permet à Claude de consulter les sorts par classe et niveau, avec leurs effets détaillés (portée, durée, descriptions).

## Sous-Agents Spécialisés

Les agents sont disponibles dans `.claude/agents/` :

### character-creator
Guide interactif pour créer des personnages étape par étape. Explique les races, classes, et aide à faire des choix cohérents.

### rules-keeper
Référence rapide des règles D&D 5e. Répond aux questions sur le combat, la magie, les jets de sauvegarde et arbitre les situations.

### dungeon-master
Maître du Jeu complet. Narration immersive, gestion des rencontres, incarnation des PNJ, et tracking automatique via les commandes sw-adventure.

## Système de Jeu D&D 5e

SkillsWeaver utilise les règles de **D&D 5e** (5ème édition) :

### Caractéristiques

- **9 espèces** : Humain, Drakéide, Elfe, Gnome, Goliath, Halfelin, Nain, Orc, Tieffelin
- **12 classes** : Barbare, Barde, Clerc, Druide, Ensorceleur, Guerrier, Magicien, Moine, Occultiste, Paladin, Rôdeur, Roublard
- **Niveaux** : 1 à 20 (pas de restrictions espèce/classe)
- **18 compétences** formelles

### Mécaniques Principales

- **Modificateurs** : `(Score - 10) ÷ 2`
- **Bonus de maîtrise** : +2 à +6 selon niveau
- **Initiative** : d20 + DEX (pas d6)
- **Avantage/Désavantage** : 2d20 (garde meilleur/pire)
- **Challenge Rating (CR)** : Difficulté des monstres (0, 1/8, 1/4, 1/2, 1-30)

### Documentation

Les règles complètes D&D 5e sont disponibles dans `docs/markdown-new/` :
- `regles_de_bases_SRD_CCv5.2.1.md` (règles fondamentales)
- `personnages.md` (création de personnage)
- `monstres.md` (bestiaire)
- `equipements.md` (équipement)

Les agents `rules-keeper` et `dungeon-master` consultent ces fichiers via Read/Grep/Glob.

Ces agents ne sont pas destinés à être utilisé de Claude Code directement, mais via sw-dm.

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

### Ajout de nouveaux tools pour sw-dm

**IMPORTANT** : Quand une nouvelle fonctionnalité est ajoutée au projet (skill, CLI), elle doit également être exposée comme tool dans sw-dm pour que l'agent DM puisse l'utiliser pendant les sessions de jeu.

1. **Créer le tool** dans `internal/dmtools/<category>_tools.go`
   ```go
   func NewMonToolTool(dep *package.Type) *SimpleTool {
       return &SimpleTool{
           name:        "mon_tool",
           description: "Description pour Claude...",
           schema: map[string]interface{}{
               "type": "object",
               "properties": map[string]interface{}{...},
           },
           execute: func(params map[string]interface{}) (interface{}, error) {
               // Appeler le package internal/...
               return map[string]interface{}{"success": true, ...}, nil
           },
       }
   }
   ```

2. **Enregistrer le tool** dans `internal/agent/register_tools.go`
   ```go
   // Créer l'instance du package si nécessaire
   myPackage, err := package.New(dataDir)
   if err != nil {
       return fmt.Errorf("failed to create package: %w", err)
   }
   registry.Register(dmtools.NewMonToolTool(myPackage))
   ```

3. **Ajouter le mapping CLI** dans `internal/agent/cli_mapper.go`
   ```go
   case "mon_tool":
       return mapMonTool(params)
   // ...
   func mapMonTool(params map[string]interface{}) string {
       return fmt.Sprintf("./sw-xxx ...")
   }
   ```

4. **Documenter le tool** :
   - `.claude/agents/dungeon-master.md` : Ajouter dans la table "Tools API"

5. **Tester** :
   ```bash
   go build -o sw-dm ./cmd/dm
   go test ./...
   ```

### Packages actuellement dans `internal/`

| Package | Utilisé par | Tests | Makefile |
|---------|-------------|-------|----------|
| `adventure` | `sw-adventure` | ✓ | ✓ |
| `ai` | `sw-adventure`, `sw-character-sheet` | - | ✓ |
| `character` | `sw-character`, `sw-character-sheet` | ✓ | ✓ |
| `charactersheet` | `sw-character-sheet` | - | ✓ |
| `combat` | (orphelin) | ✓ | - |
| `data` | `sw-character`, `sw-character-sheet` | ✓ | ✓ |
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

### Liens Externes

- [D&D Beyond](https://www.dndbeyond.com/) - Règles D&D 5e officielles
- [D&D 5e SRD](https://www.5esrd.com/) - System Reference Document (gratuit)
- [The Lazy GM's resource Document](https://slyflourish.com/lazy_gm_resource_document.html#treasuregenerator) - Site contenant de nombreuses idées, outils, tables pour améliorer le travail du MJ (Maitre du jeu). A utiliser pour améliorer le système actuel.

