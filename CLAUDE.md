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
├── core_agents/             # ⭐ NEW: Core agent/skill definitions
│   ├── agents/              # Agent personas (markdown with YAML frontmatter)
│   │   ├── dungeon-master.md      # Main DM agent
│   │   ├── character-creator.md   # Character creation guide
│   │   ├── rules-keeper.md        # D&D 5e rules arbiter
│   │   └── world-keeper.md        # World consistency guardian
│   └── skills/              # Skill definitions (SKILL.md files)
│       ├── dice-roller/     # Lancer de dés
│       ├── character-generator/ # Création de personnages
│       ├── adventure-manager/   # Gestion des aventures
│       ├── name-generator/      # Génération de noms
│       ├── npc-generator/       # Génération de PNJ
│       ├── image-generator/     # Génération d'images
│       ├── journal-illustrator/ # Illustration de journaux
│       ├── monster-manual/      # Bestiaire
│       ├── treasure-generator/  # Génération de trésors
│       ├── equipment-browser/   # Catalogue d'équipement
│       ├── spell-reference/     # Grimoire des sorts
│       └── map-generator/       # Génération de prompts pour cartes 2D
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
│   ├── map/                 # CLI sw-map
│   ├── dm/                  # CLI sw-dm (Dungeon Master REPL)
│   └── web/                 # CLI sw-web (Interface Web)
├── internal/
│   ├── agent/               # ⭐ NEW: Agent orchestration system
│   │   ├── agent.go         # Main agent loop with tool execution
│   │   ├── agent_manager.go # Nested agent invocation management
│   │   ├── agent_state.go   # Agent conversation persistence
│   │   ├── persona_loader.go # Dynamic persona loading
│   │   ├── context.go       # Conversation context with token limits
│   │   ├── tools.go         # Tool registry and execution
│   │   └── streaming.go     # Streaming response handling
│   ├── dmtools/             # ⭐ NEW: Tool wrappers for sw-dm
│   │   ├── agent_invocation_tool.go  # invoke_agent tool
│   │   ├── skill_invocation_tool.go  # invoke_skill tool
│   │   ├── simple_tools.go           # Basic game tools
│   │   └── session_tools.go          # Session management
│   ├── skills/              # ⭐ NEW: Skill management
│   │   ├── parser.go        # SKILL.md parser (YAML + markdown)
│   │   └── registry.go      # Skill discovery and registration
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
│   ├── world/               # Package données géographiques
│   └── web/                 # ⭐ NEW: Interface web Gin
│       ├── server.go        # Configuration Gin et routes
│       ├── handlers.go      # Handlers HTTP
│       ├── session.go       # Gestion sessions de jeu
│       └── web_output.go    # OutputHandler pour SSE
├── web/                     # ⭐ NEW: Assets web
│   ├── templates/           # Templates HTML (index, game, error)
│   └── static/              # CSS et JavaScript
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
│   │       ├── agent-states.json      # ⭐ NEW: Nested agent conversation history
│   │       ├── journal-meta.json      # Métadonnées journal (NextID, Categories)
│   │       ├── journal-session-0.json # Journal hors session
│   │       ├── journal-session-1.json # Journal session 1
│   │       ├── journal-session-N.json # Journal session N
│   │       ├── sw-dm-session-N.log    # ⭐ NEW: Session-specific DM logs
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

### ⭐ NEW: Architecture Standalone sw-dm

**sw-dm est désormais autonome** - Il n'a plus besoin de Claude Code pour fonctionner !

#### Agent-to-Agent Communication

Le système implémente désormais une **communication agent-à-agent** permettant au dungeon-master d'invoquer des agents spécialisés :

**Architecture à 2 niveaux** :
- **Main Agent (dungeon-master)** : Orchestrateur principal avec accès complet aux tools
- **Nested Agents** : Consultants spécialisés (rules-keeper, character-creator, world-keeper)

**Caractéristiques** :
- ✅ **Conversations stateful** : Les agents gardent l'historique de leurs consultations pendant la session
- ✅ **Token limits** : Main agent 50K, nested agents 20K
- ✅ **Récursion prévenue** : Profondeur maximale = 1 (agents imbriqués ne peuvent pas invoquer d'autres agents)
- ✅ **Persistance** : L'historique de conversation est sauvegardé dans `agent-states.json`
- ✅ **Logging complet** : Toutes les invocations sont enregistrées dans `sw-dm-session-N.log`

#### Nouveaux Tools Disponibles

**1. invoke_agent** : Consulte un agent spécialisé

```json
{
  "agent_name": "rules-keeper|character-creator|world-keeper",
  "question": "Question pour l'agent",
  "context": "Contexte additionnel (optionnel)"
}
```

Exemples d'utilisation :
```json
// Consulter rules-keeper pour arbitrer une règle
{"agent_name": "rules-keeper", "question": "Comment fonctionne le désavantage sur les jets d'attaque en D&D 5e ?"}

// Demander conseil à character-creator
{"agent_name": "character-creator", "question": "Quelles sont les meilleures cantrips pour un magicien niveau 1 ?"}

// Vérifier la cohérence avec world-keeper
{"agent_name": "world-keeper", "question": "Quels PNJ sont actuellement à Cordova ?", "context": "Session 3, après la bataille"}
```

**2. invoke_skill** : Exécute directement une skill CLI

```json
{
  "skill_name": "dice-roller|treasure-generator|...",
  "command": "./sw-<skill> <args>"
}
```

Exemples :
```json
{"skill_name": "dice-roller", "command": "./sw-dice roll 4d6kh3"}
{"skill_name": "treasure-generator", "command": "./sw-treasure generate H"}
{"skill_name": "name-generator", "command": "./sw-names generate elf --gender=f"}
```

#### Agent State Persistence

Le système sauvegarde automatiquement l'état des agents imbriqués :

**Fichier** : `data/adventures/<nom>/agent-states.json`

**Structure** :
```json
{
  "session_id": 3,
  "last_updated": "2026-01-07T14:30:00Z",
  "agents": {
    "rules-keeper": {
      "invocation_count": 5,
      "last_invoked": "2026-01-07T14:25:00Z",
      "conversation_history": [...],
      "token_estimate": 2340
    },
    "world-keeper": {
      "invocation_count": 3,
      "last_invoked": "2026-01-07T14:20:00Z",
      "conversation_history": [...],
      "token_estimate": 1850
    }
  }
}
```

**Avantages** :
- Les agents se souviennent des consultations précédentes
- Continuité entre les invocations dans une même session
- Chargement automatique au démarrage de sw-dm
- Sauvegarde automatique après chaque message utilisateur

### Hiérarchie (Architecture v2.0 avec Agent-to-Agent)

```
┌──────────────────────────────────────────────────────────┐
│                      UTILISATEUR                         │
└────────────────────────┬─────────────────────────────────┘
                         │
                         │ ./sw-dm
                         ▼
┌────────────────────────────────────────────────────────────┐
│                    MAIN AGENT (sw-dm)                      │
│                   dungeon-master.md                        │
│  ┌──────────────────────────────────────────────────┐     │
│  │ • 50K token limit                                │     │
│  │ • Full tool access (dice, monsters, treasure...)│     │
│  │ • Can invoke nested agents ─────────────────┐   │     │
│  │ • Can invoke skills directly                │   │     │
│  └──────────────────────────────────────────────┘   │     │
└────┬───────────────────────────────────────────┬────┼─────┘
     │                                           │    │
     │ invoke_agent                              │    │ invoke_skill
     ▼                                           │    ▼
┌──────────────────────────────────────┐         │  ┌──────────┐
│       NESTED AGENTS                  │         │  │  SKILLS  │
│  (Read-only consultants)             │         │  └─────┬────┘
│  ┌─────────────────────────────────┐ │         │        │
│  │ rules-keeper (20K tokens)       │◄┼─────────┘        │
│  │ • D&D 5e rules expert           │ │                  │
│  │ • Maintains conversation history│ │                  │
│  └─────────────────────────────────┘ │                  │
│  ┌─────────────────────────────────┐ │                  │
│  │ character-creator (20K tokens)  │ │                  │
│  │ • Character build guidance      │ │                  │
│  │ • Race/class recommendations    │ │                  │
│  └─────────────────────────────────┘ │                  │
│  ┌─────────────────────────────────┐ │                  │
│  │ world-keeper (20K tokens)       │ │                  │
│  │ • World consistency validation  │ │                  │
│  │ • Geography/faction coherence   │ │                  │
│  └─────────────────────────────────┘ │                  │
└──────────────────────────────────────┘                  │
                                                          ▼
┌────────────────────────────────────────────────────────────┐
│                    SKILL REGISTRY                          │
│  dice-roller, character-generator, adventure-manager,      │
│  name-generator, npc-generator, image-generator,           │
│  monster-manual, treasure-generator, equipment-browser,    │
│  spell-reference, map-generator, journal-illustrator       │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                    CLI BINARIES (sw-*)                     │
│  sw-dice, sw-character, sw-adventure, sw-names,           │
│  sw-npc, sw-location-names, sw-image, sw-monster,         │
│  sw-treasure, sw-equipment, sw-spell, sw-map              │
└────────────────────────────────────────────────────────────┘

**Flux Agent-to-Agent** :
1. User → sw-dm : "Le magicien lance Boule de Feu"
2. sw-dm → invoke_agent(rules-keeper, "Comment résoudre Boule de Feu ?")
3. rules-keeper → Response : "8d6 dégâts, JDS DEX DD 15..."
4. sw-dm → invoke_skill(dice-roller, "./sw-dice roll 8d6")
5. sw-dm → User : "La boule explose ! 35 dégâts de feu..."

**Persistance** :
- Conversation history saved in agent-states.json
- Agents remember previous consultations within session
- Automatic load on startup, save after each user message
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
- `create_character` : Créer un personnage complet et l'ajouter au groupe (sauvegarde aventure + global + party.json)

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

**⭐ NEW: Agent et Skill Invocation** :
- `invoke_agent` : Consulter un agent spécialisé (rules-keeper, character-creator, world-keeper)
- `invoke_skill` : Exécuter directement une skill CLI (dice-roller, treasure-generator, etc.)

**IMPORTANT** : L'agent dungeon-master DOIT appeler `start_session` au début de chaque partie et `end_session` à la fin. Sans cela, tous les événements seront enregistrés dans `journal-session-0.json` au lieu d'être correctement organisés par session.

**Architecture** :
- `internal/agent/` : ⭐ Orchestration de la boucle d'agent avec agent-to-agent
  - `agent.go` : Boucle principale avec tool execution et state persistence
  - `agent_manager.go` : ⭐ NEW - Gestion des agents imbriqués (rules-keeper, etc.)
  - `agent_state.go` : ⭐ NEW - Persistance conversations agents dans agent-states.json
  - `persona_loader.go` : ⭐ NEW - Chargement dynamique personas depuis core_agents/
  - `tools.go` : Système de registry des tools
  - `context.go` : Gestion contexte conversation/aventure avec token limits
  - `streaming.go` : Traitement événements streaming
  - `register_tools.go` : Enregistrement de tous les tools
- `internal/dmtools/` : Wrappers des tools pour l'agent
  - `agent_invocation_tool.go` : ⭐ NEW - Tool invoke_agent pour consulter agents
  - `skill_invocation_tool.go` : ⭐ NEW - Tool invoke_skill pour exécuter skills
  - `simple_tools.go` : Tools basiques (log_event, add_gold, etc.)
  - `session_tools.go` : Gestion de session (start/end/get_info)
  - `character_tools.go` : Consultation personnages (get_party_info, get_character_info)
  - `create_character_tool.go` : Création de personnage (create_character)
  - `equipment_tools.go` : Consultation équipement (get_equipment)
  - `spell_tools.go` : Consultation sorts (get_spell)
  - `encounter_tools.go` : Génération rencontres (generate_encounter, roll_monster_hp)
  - `inventory_tools.go` : Gestion inventaire (add_item, remove_item)
  - `name_tools.go` : Génération noms (generate_name, generate_location_name)
  - `dice_tool.go`, `monster_tool.go`, `npc_management_tools.go`, etc.
- `internal/skills/` : ⭐ NEW - Système de gestion des skills
  - `parser.go` : Parser SKILL.md (YAML frontmatter + markdown)
  - `registry.go` : Découverte et enregistrement des skills
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

### CLI sw-web (Interface Web)

Interface web basée sur Gin pour jouer à SkillsWeaver via navigateur :

```bash
# Compiler
go build -o sw-web ./cmd/web

# Lancer le serveur (port 8085 par défaut)
./sw-web

# Options
./sw-web --port=3000        # Port personnalisé
./sw-web --debug            # Mode debug avec logs Gin
```

**Fonctionnalités** :
- Interface web avec thème Dark Fantasy Médiéval
- Streaming des réponses en temps réel via SSE (Server-Sent Events)
- Liste et création d'aventures
- Génération automatique de campaign plan (si thème fourni)
- Copie automatique des personnages globaux vers nouvelle aventure
- Session de jeu interactive avec le Dungeon Master
- Affichage du groupe, inventaire et journal
- Images générées affichées inline

**Gestion Automatique des Personnages** :
Lors de la création d'une nouvelle aventure, le système copie automatiquement tous les personnages présents dans `data/characters/` vers le répertoire de l'aventure et crée le fichier `party.json`. Si aucun personnage global n'existe, vous devrez créer des personnages via `sw-character` avant de démarrer la session.

**Architecture** :
- `cmd/web/main.go` : Entry point du serveur
- `internal/web/` : Package web
  - `server.go` : Configuration Gin et routes
  - `handlers.go` : Handlers HTTP
  - `session.go` : Gestion des sessions de jeu (SessionManager)
  - `web_output.go` : OutputHandler pour SSE (WebOutput)
- `web/templates/` : Templates HTML
  - `index.html` : Page d'accueil avec liste des aventures
  - `game.html` : Interface de jeu
  - `error.html` : Page d'erreur
- `web/static/` : Assets statiques
  - `css/fantasy.css` : Thème Dark Fantasy
  - `js/app.js` : Client JavaScript pour SSE

**Routes** :

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/` | Page d'accueil |
| GET | `/adventures` | Liste des aventures (HTMX) |
| POST | `/adventures` | Créer une aventure |
| GET | `/play/:slug` | Page de jeu |
| POST | `/play/:slug/message` | Envoyer un message au DM |
| GET | `/play/:slug/stream` | Endpoint SSE |
| GET | `/play/:slug/characters` | Liste des personnages |
| GET | `/play/:slug/info` | Info aventure (HTMX) |
| GET | `/play/:slug/images/*` | Images générées |

**Prérequis** :
- Variable d'environnement `ANTHROPIC_API_KEY` configurée
- Des aventures existantes dans `data/adventures/` (ou créez-en via l'interface)

**Session Management** :
- Une session par aventure (mono-joueur)
- Sessions persistées en mémoire pendant 30 minutes d'inactivité
- Nettoyage automatique des sessions expirées

---

## 🚀 Agent System - Fonctionnalités Avancées

Le système d'agents de SkillsWeaver inclut 4 fonctionnalités avancées pour une expérience professionnelle :

### 1. ✅ Historique de Conversation Complet avec Optimisation Token

**Fichier** : `internal/agent/message_serialization.go`

Le système sauvegarde maintenant l'historique complet des conversations des agents imbriqués :

**Fonctionnalités** :
- ✅ Sérialisation complète : texte, tool uses, tool results
- ✅ Optimisation : conserve seulement les 15K derniers tokens
- ✅ Persistance : sauvegardé dans `agent-states.json`
- ✅ Restauration : conversation continuée entre sessions

**Détails Techniques** :
```go
// Serialization automatique avec limite de tokens
conversationHistory, _ := SerializeConversationContextWithOptimization(
    state.conversationCtx,
    15000, // Garde les 15K derniers tokens
)
```

**Avantages** :
- Les agents se souviennent des discussions précédentes
- Continuité contextuelle entre sessions
- Optimisation de la taille des fichiers d'état
- Balance entre contexte et performance

---

### 2. ✅ Rotation et Compression Automatique des Logs

**Fichier** : `internal/agent/logger.go`

Les logs sont automatiquement gérés pour éviter les fichiers trop volumineux :

**Fonctionnalités** :
- ✅ Rotation automatique à 10MB (configurable)
- ✅ Compression gzip (~90% de réduction)
- ✅ Conservation de 5 rotations par défaut
- ✅ Nettoyage automatique des anciens fichiers

**Configuration** :
```go
logger.SetMaxSize(20)        // Rotation à 20MB
logger.SetMaxRotations(10)   // Garde 10 fichiers compressés
```

**Exemple de Rotation** :
```
sw-dm-session-1.log        (10MB - rotation déclenchée)
  ↓
sw-dm-session-1.log        (0 bytes - nouveau fichier)
sw-dm-session-1.log.1.gz   (1MB compressé)
  ↓ (après seconde rotation)
sw-dm-session-1.log        (0 bytes)
sw-dm-session-1.log.1.gz   (1MB)
sw-dm-session-1.log.2.gz   (1MB)
```

**Avantages** :
- Gestion automatique de l'espace disque
- Logs compressés pour archivage
- Performance améliorée (fichiers plus petits)
- Maintenance zéro

---

### 3. ✅ Restrictions d'Outils par Agent

**Fichier** : `internal/agent/agent_manager.go`

Les agents imbriqués sont des **consultants en lecture seule** sans accès aux outils :

**Restrictions Enforced** :
- ❌ **Rules-Keeper** : Ne peut PAS modifier l'état du jeu
- ❌ **Character-Creator** : Ne peut PAS invoquer de skills
- ❌ **World-Keeper** : Ne peut PAS modifier les données monde

**Implémentation** :
```go
// Appel API SANS paramètre Tools
response, err := nestedAgent.client.Messages.New(ctx, anthropic.MessageNewParams{
    Model:     anthropic.ModelClaudeHaiku4_5,
    MaxTokens: 4096,
    System:    []anthropic.TextBlockParam{...},
    Messages:  nestedAgent.conversationCtx.GetMessages(),
    // Tools intentionnellement omis - agents imbriqués sans outils
})
```

**Garanties de Sécurité** :
- ✅ Impossible d'invoquer d'autres agents (limite de récursion = 1)
- ✅ Impossible d'invoquer des skills
- ✅ Impossible de modifier l'état du jeu
- ✅ Consultants purement informatifs

**Avantages** :
- Sécurité : Aucune modification involontaire
- Prévisibilité : Agents imbriqués = consultants purs
- Architecture claire : Seul le DM principal contrôle l'état

---

### 4. ✅ Métriques de Performance des Agents

**Fichiers** : `internal/agent/agent_manager.go`, `internal/agent/agent_state.go`

Suivi complet des performances et coûts pour chaque agent :

**Métriques Trackées** :
```go
type AgentMetrics struct {
    TotalTokensUsed      int64         // Tokens cumulés
    TotalInputTokens     int64         // Tokens d'entrée
    TotalOutputTokens    int64         // Tokens de sortie
    TotalResponseTime    time.Duration // Temps cumulé
    AverageTokensPerCall int64         // Moyenne par appel
    AverageResponseTime  time.Duration // Temps moyen
    ModelUsed            string        // Modèle utilisé
    LastCallTokens       int64         // Dernier appel
    LastCallDuration     time.Duration // Durée dernier appel
}
```

**API d'Accès** :
```go
// Statistiques de tous les agents
stats := agentManager.GetStatistics()

// Métriques d'un agent spécifique
metrics, exists := agentManager.GetAgentMetrics("rules-keeper")
```

**Exemple de Sortie** :
```json
{
  "rules-keeper": {
    "invocation_count": 5,
    "total_tokens_used": 12450,
    "total_input_tokens": 8200,
    "total_output_tokens": 4250,
    "average_tokens_per_call": 2490,
    "average_response_time_ms": 3064,
    "model_used": "claude-haiku-4-5",
    "last_call_tokens": 2680
  }
}
```

**Avantages** :
- 💰 Suivi des coûts : Tokens utilisés par agent
- 📊 Optimisation : Identifie les agents lents
- 📈 Analytics : Données pour améliorer le système
- 💾 Persisté : Métriques sauvegardées entre sessions

**Utilisation** :
```bash
# Voir les statistiques après une session
cat data/adventures/<nom>/agent-states.json | jq '.agents'
```

---

### Documentation Complète

Voir `docs/optional-features-summary.md` pour :
- Guide détaillé de chaque fonctionnalité
- Exemples d'utilisation
- Détails techniques d'implémentation
- Résultats des tests

---

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


---

## 🎭 Système de Planification Narrative de Campagne

### Vue d'Ensemble

SkillsWeaver dispose d'un système avancé de planification narrative en 3 actes qui guide les sessions de jeu. Ce système automatise les briefings pré-session et maintient la cohérence de l'intrigue sur plusieurs sessions.

### Fichier campaign-plan.json

**Localisation** : `data/adventures/<nom>/campaign-plan.json`

**Génération automatique** : Si un thème est fourni lors de la création d'une aventure via l'interface web, le DM génère automatiquement un plan structuré incluant :

- **Structure narrative 3 actes** avec objectifs, événements clés, et critères de complétion
- **Antagoniste principal** avec arc narratif et sessions clés
- **MacGuffins et lieux importants** liés aux actes
- **Foreshadows critiques** avec liens aux actes et payoff planifiés
- **Progression et pacing** trackés automatiquement

### Fonctionnement Automatique

#### 1. Création d'Aventure avec Thème

Dans l'interface web :
```
Nom : Le Sextant Magique de Cordova
Description : Conspiration maritime dans le royaume de Valdorine
Thème : Un sextant magique révèle l'emplacement d'une entité ancienne 
        scellée sous Shasseth. Plusieurs factions cherchent à l'atteindre.
```

Le DM génère automatiquement :
- 3 actes structurés (début, rebondissements, confrontation finale)
- Antagonistes avec motivations et arcs
- 2-3 foreshadows critiques liés aux actes
- Pacing cible (ex: 10 sessions, 3h chacune)

#### 2. Briefing Automatique au Démarrage de Session

Quand vous appelez `start_session` dans sw-dm :

```
✓ Session 12 démarrée

=== CAMPAIGN CONTEXT (CONFIDENTIAL - DO NOT QUOTE DIRECTLY) ===

Act 3: Confrontation à Shasseth
Les PJ arrivent à la cité perdue. Vaskir prépare le rituel final.

Campaign Objective: Empêcher le réveil de l'entité divine ancienne

Active Threads:
  • vaskir_ritual_countdown
  • cinquieme_acteur_identity

Critical Foreshadows (2):
  • [fsh_002] Entité scellée (planted 5 sessions ago, critical)
  • [fsh_004] Trahison d'allié (planted 3 sessions ago, major)

World-Keeper Briefing:
[Guidance stratégique pour la session...]

=== INSTRUCTIONS ===
• Use this context to guide your narration naturally
• DO NOT quote world-keeper directly to players
• Integrate information organically into the story
===
```

**Ce briefing est caché du joueur** mais guide votre narration pour :
- Avancer les threads narratifs actifs
- Résoudre les foreshadows critiques
- Respecter les objectifs de l'acte en cours
- Maintenir le pacing

#### 3. Consultation Silencieuse World-Keeper

Le système consulte automatiquement le world-keeper en mode silencieux :
- **Notification visible** : `[Consulting world-keeper...]`
- **Réponse cachée** : Injectée dans le contexte système uniquement
- **Utilisation** : Guide votre narration sans révéler les secrets

### Tools Disponibles pour Campaign Plan

#### get_campaign_plan

```json
{"section": "current_act"}
{"section": "foreshadows"}
{"section": "progression"}
{"section": "all"}
```

Retourne l'état complet du plan narratif.

#### update_campaign_progress

```json
{"action": "complete_plot_point", "plot_point_id": "valorian_alliance"}
{"action": "advance_act", "act_number": 2}
```

Marque des milestones comme complétés.

#### add_narrative_thread / remove_narrative_thread

```json
{"thread_name": "mysterious_stranger_identity"}
{"thread_name": "alliance_betrayal"}
```

Track les intrigues secondaires actives.

### Migration depuis Foreshadows.json

Les anciennes aventures utilisent `foreshadows.json`. Le nouveau système utilise `campaign-plan.json` qui intègre les foreshadows avec des liens vers les actes.

**Backward Compatibility** : Les aventures sans campaign-plan continuent de fonctionner normalement avec foreshadows.json legacy.

**Migration manuelle** (optionnelle) :
1. Créer `campaign-plan.json` avec structure par défaut
2. Importer foreshadows existants avec liens actes estimés
3. Enrichir manuellement : objectif, actes, antagonistes

### Règles Importantes pour le DM

#### ✅ CORRECT - Intégrer le Briefing Naturellement

**Briefing** : "Vaskir est à Shasseth depuis 2 jours, préparant le rituel dans les ruines du temple."

**Narration** :
```
Les rumeurs dans les tavernes du port parlent d'un navire noir aperçu
près de Shasseth il y a deux jours. Les marins superstitieux murmurent
que personne n'en est revenu vivant.

Que faites-vous ?
```

#### ❌ INTERDIT - Citer Directement

**JAMAIS faire** :
- "Le world-keeper m'informe que Vaskir est à Shasseth."
- "Selon le briefing, l'entité se réveille bientôt."
- Paraphraser mot-à-mot le briefing

#### Transformation de l'Information

Le briefing te donne la **direction stratégique**. Les joueurs découvrent par :
- **Dialogues PNJ** : "Un marin tremble : 'J'ai vu ce navire... noir comme la nuit...'"
- **Indices visuels** : "Des runes anciennes gravées pâlissent lentement."
- **Rumeurs** : "Les prêtres parlent à voix basse de tremblements souterrains."

### Avantages du Système

1. **Cohérence Narrative** : Objectif clair et structure 3 actes dès le début
2. **Foreshadows Organisés** : Liés aux actes, pas orphelins
3. **Briefings Automatiques** : Direction narrative au début de chaque session
4. **Confidentialité** : Secrets restent secrets (world-keeper en mode silencieux)
5. **Pacing Trackéé** : Comparaison sessions planifiées vs réelles par acte

### Fichiers Concernés

- `data/adventures/<nom>/campaign-plan.json` - Plan narratif complet
- `data/adventures/<nom>/foreshadows.json` - Legacy (deprecated)
- `data/adventures/<nom>/agent-states.json` - Historique consultations agents

