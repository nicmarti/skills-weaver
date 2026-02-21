# SkillsWeaver - Moteur de Jeu de Rôle D&D 5e

## Description

**SkillsWeaver** est un moteur de jeu de rôle interactif basé sur les règles de **D&D 5e** (5ème édition). Il combine une interface web moderne (sw-web), un système d'agents autonomes (sw-dm), et un ensemble de skills CLI pour créer une expérience de jeu complète.

Le préfixe `sw-` identifie toutes les commandes CLI du projet.

## 🚨 Guidelines Claude Code

**Ces règles s'appliquent quand Claude Code modifie cette codebase. Elles sont distinctes du comportement interne des agents sw-dm.**

### Quand une Session de Jeu Live est Active

- **Ne PAS utiliser les outils MCP du navigateur** si l'utilisateur mentionne qu'il joue une session live (sw-dm ou sw-web en cours d'exécution).
- **Utiliser uniquement l'analyse de fichiers** : lire les logs (`sw-dm-session-N.log`), `agent-states.json`, et le code source. Ne pas ouvrir d'onglets navigateur.
- Si incertain qu'une session live est active, **demander avant d'utiliser tout outil navigateur**.

### Règles de Développement Go

- **Après toute modification de fichier `.go`**, exécuter `go build ./...` immédiatement. Si la compilation échoue, corriger avant toute autre action. **Ne jamais committer avec une compilation cassée**.
- **Si des fichiers `.templ` ont été modifiés**, exécuter `templ generate` d'abord, puis `go build ./...`.
- **Après ajout d'un nouveau tool sw-dm**, vérifier qu'il est accessible via le registry (checker `internal/agent/register_tools.go`) avant de déclarer terminé.

### Sécurité Git

- **Ne jamais exécuter `git add .` ou `git add -A`** : Stager uniquement les fichiers explicitement modifiés. Les hooks pre-commit gofmt et templ peuvent reformater des fichiers non liés, causant un staging accidentel depuis des sessions parallèles.
- **Ne jamais exécuter `git reset --hard` sans confirmation utilisateur**.
- Avant tout commit, exécuter `git status` pour vérifier que seuls les fichiers intentionnels sont stagés.

### Débogage du Comportement des Agents

- **Avant de suggérer un changement de prompt**, lire le fichier persona pertinent dans `core_agents/agents/` et tracer le chemin d'exécution dans `internal/agent/`. Les causes racines sont généralement architecturales (mauvaise registration d'outil, contexte non chargé, session non démarrée) — pas des problèmes de formulation.
- **Lors de modification des tools sw-dm**, suivre le processus en 5 étapes dans "Ajout de nouveaux tools pour sw-dm" ci-dessous. Ne pas sauter le CLI mapper ou l'étape de documentation dans dungeon-master.md.

---

## Architecture Globale

```
┌─────────────────────────────────────────────────────────┐
│                   UTILISATEUR                           │
└────────────┬────────────────────────────────────────────┘
             │
             ├──► sw-web (Interface Web Gin/HTMX/SSE)
             │    └──► internal/web/ + web/templates/
             │
             └──► sw-dm (REPL autonome)
                  └──► internal/agent/ (boucle d'agent complète)
                       ├──► dungeon-master (main agent, 50K tokens)
                       ├──► rules-keeper (nested, 20K tokens)
                       ├──► character-creator (nested, 20K tokens)
                       └──► world-keeper (nested, 20K tokens)
                            │
                            ▼
                  ┌─────────────────────────────────────────┐
                  │  SKILLS REGISTRY (12 skills)            │
                  │  dice-roller, character-generator,      │
                  │  adventure-manager, name-generator,     │
                  │  npc-generator, image-generator,        │
                  │  journal-illustrator, monster-manual,   │
                  │  treasure-generator, equipment-browser, │
                  │  spell-reference, map-generator         │
                  └─────────────────────────────────────────┘
                            │
                            ▼
                  ┌─────────────────────────────────────────┐
                  │  CLI BINARIES (sw-*)                    │
                  │  sw-dice, sw-character, sw-adventure,   │
                  │  sw-names, sw-npc, sw-location-names,   │
                  │  sw-image, sw-monster, sw-treasure,     │
                  │  sw-equipment, sw-spell, sw-map         │
                  └─────────────────────────────────────────┘
```

### Concepts Clés

**Skills** = Outils automatisables avec CLI
- Invoqués via `/skill-name` ou automatiquement par agents
- Exécutent des commandes `sw-*` (Go binaries)
- Retournent des données structurées JSON
- Autonomes : peuvent fonctionner seuls ou être utilisés par agents

**Agents** = Personnalités/Rôles spécialisés avec IA
- Guident l'utilisateur avec contexte narratif
- Utilisent les skills comme outils
- Maintiennent un style et ton cohérent
- Orchestrent plusieurs skills pour tâches complexes

**Agent-to-Agent Communication** :
- Le dungeon-master (main agent) peut invoquer des agents imbriqués via `invoke_agent`
- Les agents imbriqués sont des **consultants en lecture seule** (pas d'accès tools)
- Profondeur maximale de récursion = 1 (agents imbriqués ne peuvent pas invoquer d'autres agents)
- Conversations persistées dans `agent-states.json`

---

## 🌐 Interface Web (sw-web) - Interface Principale

Interface web moderne pour jouer à SkillsWeaver via navigateur :

```bash
# Compiler et lancer
go build -o sw-web ./cmd/web
./sw-web                    # Port 8085 par défaut
./sw-web --port=3000        # Port personnalisé
./sw-web --debug            # Mode debug Gin
```

### Fonctionnalités Principales

✅ **Interface Dark Fantasy Médiéval** avec thème immersif
✅ **Streaming temps réel** via SSE (Server-Sent Events)
✅ **Gestion d'aventures** : liste, création, sélection
✅ **Campaign Plan automatique** : génération narrative 3 actes si thème fourni
✅ **Copie auto des personnages** : personnages globaux vers nouvelle aventure
✅ **Session de jeu interactive** avec Dungeon Master agent
✅ **Affichage live** : groupe, inventaire, journal, images générées

### Routes Principales

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/` | Page d'accueil |
| GET | `/adventures` | Liste des aventures (HTMX) |
| POST | `/adventures` | Créer une aventure |
| GET | `/play/:slug` | Page de jeu |
| POST | `/play/:slug/message` | Envoyer un message au DM |
| GET | `/play/:slug/stream` | Endpoint SSE (streaming réponses) |
| GET | `/play/:slug/characters` | Liste des personnages |
| GET | `/play/:slug/info` | Info aventure (HTMX) |
| GET | `/play/:slug/images/*` | Images générées |

### Architecture Technique

```
cmd/web/main.go              # Entry point serveur
internal/web/
├── server.go                # Configuration Gin et routes
├── handlers.go              # Handlers HTTP
├── session.go               # SessionManager (sessions de jeu)
└── web_output.go            # OutputHandler pour SSE
web/
├── templates/               # Templates HTML (index, game, error)
└── static/                  # CSS (fantasy.css), JS (app.js)
```

### Prérequis

- Variable d'environnement `ANTHROPIC_API_KEY` configurée
- Personnages existants dans `data/characters/` (optionnel, créés auto si absents)
- Aventures existantes dans `data/adventures/` (ou créez-en via l'interface)

### Session Management

- Une session par aventure (mono-joueur actuellement)
- Sessions persistées en mémoire pendant 30 minutes d'inactivité
- Nettoyage automatique des sessions expirées
- Logs session-specific dans `data/adventures/<nom>/sw-dm-session-N.log`

---

## 🎲 Interface REPL (sw-dm) - Dungeon Master Autonome

Application interactive de Maître du Jeu autonome avec boucle d'agent complète. Alternative CLI à sw-web pour développement/tests.

```bash
# Compiler et lancer
go build -o sw-dm ./cmd/dm
./sw-dm

# L'application propose un menu pour sélectionner l'aventure
# Puis démarre une session REPL interactive avec streaming
```

### Fonctionnalités

- ✅ Boucle d'agent complète avec tool_use (Anthropic API)
- ✅ Streaming des réponses pour expérience immersive
- ✅ Auto-chargement du contexte d'aventure (groupe, inventaire, journal)
- ✅ Accès direct aux packages Go (dice, monster, treasure, npc, etc.)
- ✅ Interface readline avec historique de conversation persistant

### Tools Disponibles pour l'Agent DM

**Gestion de Session** (CRITIQUE) :
- `start_session`, `end_session`, `get_session_info`

**Mécanique de Jeu** :
- `roll_dice`, `get_monster`, `log_event`, `add_gold`, `get_inventory`

**Génération de Contenu** :
- `generate_treasure`, `generate_npc`, `generate_image`, `generate_map`

**NPC Management** :
- `update_npc_importance`, `get_npc_history`

**Consultation Personnages** :
- `get_party_info`, `get_character_info`, `create_character`

**Consultation Équipement/Sorts** :
- `get_equipment`, `get_spell`

**Génération Rencontres** :
- `generate_encounter`, `roll_monster_hp`

**Gestion Inventaire** :
- `add_item`, `remove_item`

**Génération Noms** :
- `generate_name`, `generate_location_name`

**Agent et Skill Invocation** :
- `invoke_agent` : Consulter agent spécialisé (rules-keeper, character-creator, world-keeper)
- `invoke_skill` : Exécuter directement une skill CLI (dice-roller, treasure-generator, etc.)

**IMPORTANT** : L'agent dungeon-master DOIT appeler `start_session` au début et `end_session` à la fin. Sinon, tous les événements vont dans `journal-session-0.json` au lieu d'être organisés par session.

### Logging Automatique CLI

Chaque tool appelé par sw-dm est automatiquement loggé avec sa commande CLI équivalente dans `data/adventures/<nom>/sw-dm-session-N.log`. Permet de :
- Reproduire facilement les opérations (copier-coller commande)
- Tester avec paramètres différents
- Déboguer et améliorer les outils

Extraction des commandes :
```bash
./scripts/extract-cli-commands.sh                           # Toutes aventures
./scripts/extract-cli-commands.sh la-crypte-des-ombres     # Aventure spécifique
./scripts/extract-cli-commands.sh la-crypte-des-ombres generate_map  # Tool spécifique
grep "Equivalent CLI:" data/adventures/*/sw-dm*.log        # Grep manuel
```

---

## 📊 Systèmes de Données Avancés

### 1. Système de Persistance des PNJ (Deux Niveaux)

#### Niveau 1 : Fichier par Aventure (`npcs-generated.json`)

**Localisation** : `data/adventures/<nom>/npcs-generated.json`

**Capture automatique** : Tous les PNJ générés via `generate_npc` sont auto-sauvegardés.

**Niveaux d'importance** :
- `mentioned` : Généré mais pas d'interaction
- `interacted` : Dialogue ou rencontre brève
- `recurring` : Apparitions multiples
- `key` : Importance majeure pour l'intrigue

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
        "importance": "mentioned",
        "notes": ["Note 1", "Note 2"],
        "appearances": 1,
        "promoted_to_world": false
      }
    ]
  }
}
```

#### Niveau 2 : Fichier Monde (`data/world/npcs.json`)

**PNJ promus** : Seuls les PNJ récurrents et importants sont promus vers `npcs.json` après validation par world-keeper.

**Workflow de promotion** :
1. World-keeper review : `/world-review-npcs <adventure>`
2. Validation et enrichissement : `/world-promote-npc <adventure> <nom>`
3. Ajout à `data/world/npcs.json` avec contexte complet

**Avantages** :
- ✅ Aucune perte (tous PNJ capturés automatiquement)
- ✅ Évolution naturelle (importance augmente avec interactions)
- ✅ Validation centralisée (world-keeper garantit cohérence)
- ✅ Scalable (5 ou 50 PNJ par aventure)
- ✅ Séparation claire (Adventure = brouillon, World = canon)

### 2. Structure du Journal par Session

Le journal est organisé en fichiers séparés par session pour optimiser la performance :

- `journal-meta.json` : Métadonnées globales (NextID, Categories, LastUpdate)
- `journal-session-N.json` : Entrées pour la session N
- `journal-session-0.json` : Entrées hors session

**Avantages** :
- Réduit l'utilisation de tokens (charge uniquement sessions nécessaires)
- Scalable (pas de limite de taille de journal)
- Organisation claire par session de jeu
- Images organisées de manière cohérente (session-0/, session-1/, etc.)

**Migration** : `sw-adventure migrate-journal <aventure>` pour convertir ancien journal.json monolithique.

### 3. 🎭 Système de Planification Narrative de Campagne

**Fichier** : `data/adventures/<nom>/campaign-plan.json`

**Génération automatique** : Si un thème est fourni lors de la création d'une aventure via sw-web, le DM génère automatiquement un plan structuré incluant :

- **Structure narrative 3 actes** avec objectifs, événements clés, critères de complétion
- **Antagoniste principal** avec arc narratif et sessions clés
- **MacGuffins et lieux importants** liés aux actes
- **Foreshadows critiques** avec liens aux actes et payoff planifiés
- **Progression et pacing** trackés automatiquement

#### Briefing Automatique au Démarrage de Session

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

**Ce briefing est caché du joueur** mais guide la narration pour :
- Avancer les threads narratifs actifs
- Résoudre les foreshadows critiques
- Respecter les objectifs de l'acte en cours
- Maintenir le pacing

#### Tools Campaign Plan

- `get_campaign_plan` : Retourne l'état complet du plan narratif
- `update_campaign_progress` : Marque des milestones comme complétés
- `add_narrative_thread` / `remove_narrative_thread` : Track intrigues secondaires

#### Règles pour le DM

✅ **CORRECT** - Intégrer le Briefing Naturellement :
- Transformer informations en dialogues PNJ, indices visuels, rumeurs
- **JAMAIS citer** : "Le world-keeper m'informe...", "Selon le briefing..."

❌ **INTERDIT** - Citer Directement :
- Pas de paraphrase mot-à-mot du briefing
- Pas de révélation directe des secrets

---

## 🛠️ Skills Disponibles (12 au total)

| Skill | CLI Binary | Description |
|-------|-----------|-------------|
| dice-roller | `sw-dice` | Lancer de dés avec notation RPG (1d20, 4d6kh3, etc.) |
| character-generator | `sw-character` | Création de personnages guidée étape par étape |
| adventure-manager | `sw-adventure` | Gestion d'aventures, sessions, journal automatique |
| name-generator | `sw-names` | Génération de noms par race/genre/type PNJ |
| npc-generator | `sw-npc` | Création PNJ complets (apparence, personnalité, secrets) |
| location-name-generator | `sw-location-names` | Noms de lieux cohérents avec les 4 royaumes |
| image-generator | `sw-image` | Illustrations fantasy (portraits, scènes, monstres, lieux) |
| journal-illustrator | `sw-adventure illustrate` | Illustration auto journaux avec prompts optimisés |
| map-generator | `sw-map` | Prompts enrichis pour cartes 2D fantasy |
| monster-manual | `sw-monster` | Stats monstres, génération rencontres équilibrées |
| treasure-generator | `sw-treasure` | Génération trésors D&D 5e par table de trésor |
| equipment-browser | `sw-equipment` | Catalogue armes, armures, équipement avec stats |
| spell-reference | `sw-spell` | Grimoire des sorts par classe/niveau avec effets |

---

## 🏗️ Architecture Technique Avancée

### Agent System - Fonctionnalités Avancées

#### 1. Historique de Conversation avec Optimisation Token

**Fichier** : `internal/agent/message_serialization.go`

- ✅ Sérialisation complète : texte, tool uses, tool results
- ✅ Optimisation : conserve seulement les 15K derniers tokens
- ✅ Persistance : sauvegardé dans `agent-states.json`
- ✅ Restauration : conversation continuée entre sessions

#### 2. Rotation et Compression Automatique des Logs

**Fichier** : `internal/agent/logger.go`

- ✅ Rotation automatique à 10MB (configurable)
- ✅ Compression gzip (~90% de réduction)
- ✅ Conservation de 5 rotations par défaut
- ✅ Nettoyage automatique des anciens fichiers

Exemple :
```
sw-dm-session-1.log        (10MB - rotation déclenchée)
  ↓
sw-dm-session-1.log        (0 bytes - nouveau fichier)
sw-dm-session-1.log.1.gz   (1MB compressé)
```

#### 3. Restrictions d'Outils par Agent

**Fichier** : `internal/agent/agent_manager.go`

Les agents imbriqués sont des **consultants en lecture seule** sans accès aux outils :

- ❌ **Rules-Keeper** : Ne peut PAS modifier l'état du jeu
- ❌ **Character-Creator** : Ne peut PAS invoquer de skills
- ❌ **World-Keeper** : Ne peut PAS modifier les données monde

Garanties de sécurité :
- ✅ Impossible d'invoquer d'autres agents (limite récursion = 1)
- ✅ Impossible d'invoquer des skills
- ✅ Impossible de modifier l'état du jeu
- ✅ Consultants purement informatifs

#### 4. Métriques de Performance des Agents

**Fichiers** : `internal/agent/agent_manager.go`, `internal/agent/agent_state.go`

Suivi complet des performances et coûts pour chaque agent :

**Métriques Trackées** :
- Total tokens used (input + output)
- Average tokens per call
- Average response time
- Model used
- Last call metrics

**Utilisation** :
```bash
# Voir les statistiques après une session
cat data/adventures/<nom>/agent-states.json | jq '.agents'
```

**Documentation complète** : Voir `docs/optional-features-summary.md` pour détails techniques et exemples.

### Agent State Persistence

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
      "token_estimate": 2340,
      "metrics": {
        "total_tokens_used": 12450,
        "average_tokens_per_call": 2490,
        "model_used": "claude-haiku-4-5"
      }
    }
  }
}
```

**Avantages** :
- Les agents se souviennent des consultations précédentes
- Continuité entre invocations dans même session
- Chargement automatique au démarrage de sw-dm
- Sauvegarde automatique après chaque message utilisateur
- Métriques persistées entre sessions

---

## 🎮 Système de Jeu D&D 5e

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

---

## 💻 Commandes de Développement

```bash
# Compiler tous les outils SkillsWeaver
make

# Lancer les tests
make test

# Tester des packages spécifiques
go test ./internal/dice/... -v
go test ./internal/data/... -v
go test ./internal/character/... -v
```

---

## 📝 Conventions de Développement

### Ajout de Nouveaux Packages dans `internal/`

Lors de l'ajout d'un nouveau package dans `internal/` pour supporter une skill :

1. **Mettre à jour le Makefile** avec les nouvelles dépendances
   ```makefile
   $(BINARY_PREFIX)-adventure: cmd/adventure/main.go internal/adventure/*.go internal/<new>/*.go
   ```

2. **Créer des tests unitaires**
   - Tout nouveau package dans `internal/` doit avoir des tests
   - Créer `<package>_test.go` dans le même répertoire
   - Lancer `make test` pour vérifier que tous les tests passent

3. **Vérifier la compilation**
   ```bash
   make clean
   make
   touch internal/<package>/<file>.go
   make <binary-name>
   ```

### Ajout de Nouveaux Tools pour sw-dm

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

   func mapMonTool(params map[string]interface{}) string {
       return fmt.Sprintf("./sw-xxx ...")
   }
   ```

4. **Documenter le tool** :
   - `core_agents/agents/dungeon-master.md` : Ajouter dans la table "Tools API"

5. **Tester** :
   ```bash
   go build -o sw-dm ./cmd/dm
   go test ./...
   ```

### Packages dans `internal/`

| Package | Utilisé par | Tests | Makefile |
|---------|-------------|-------|----------|
| `agent` | `sw-dm`, `sw-web` | ✓ | ✓ |
| `adventure` | `sw-adventure` | ✓ | ✓ |
| `character` | `sw-character`, `sw-character-sheet` | ✓ | ✓ |
| `dice` | `sw-dice`, `sw-monster`, `sw-treasure` | ✓ | ✓ |
| `equipment` | `sw-equipment` | - | ✓ |
| `image` | `sw-image` | - | ✓ |
| `locations` | `sw-location-names` | ✓ | ✓ |
| `monster` | `sw-monster` | ✓ | ✓ |
| `names` | `sw-names`, `sw-npc` | ✓ | ✓ |
| `npc` | `sw-npc` | ✓ | ✓ |
| `spell` | `sw-spell` | - | ✓ |
| `treasure` | `sw-treasure` | ✓ | ✓ |
| `web` | `sw-web` | - | ✓ |

---

## 🔧 Conventions Git

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
git commit -m "docs: update rules-keeper with D&D 5e combat rules"
```

---

## 📚 Ressources

### Liens Externes

- [D&D Beyond](https://www.dndbeyond.com/) - Règles D&D 5e officielles
- [D&D 5e SRD](https://www.5esrd.com/) - System Reference Document (gratuit)
- [The Lazy GM's Resource Document](https://slyflourish.com/lazy_gm_resource_document.html) - Outils et tables pour améliorer le travail du MJ

---

## 📁 Structure du Projet (Vue d'Ensemble)

```
skillsweaver/
├── core_agents/             # Agent personas et skill definitions
│   ├── agents/              # dungeon-master, rules-keeper, character-creator, world-keeper
│   └── skills/              # 12 skills (SKILL.md files)
├── cmd/                     # CLI binaries (sw-*)
│   ├── web/                 # sw-web (Interface Web Gin)
│   ├── dm/                  # sw-dm (REPL autonome)
│   └── [12 autres CLIs]     # sw-dice, sw-character, sw-adventure, etc.
├── internal/                # Packages Go
│   ├── agent/               # Agent orchestration system
│   ├── dmtools/             # Tool wrappers pour sw-dm
│   ├── web/                 # Package web (Gin handlers, SSE)
│   └── [12 autres packages] # dice, character, adventure, monster, etc.
├── web/                     # Assets web
│   ├── templates/           # HTML templates (index, game, error)
│   └── static/              # CSS (fantasy.css), JS (app.js)
├── data/                    # Données et aventures
│   ├── adventures/          # Aventures sauvegardées
│   │   └── <nom>/
│   │       ├── adventure.json, sessions.json, party.json, inventory.json
│   │       ├── agent-states.json, campaign-plan.json
│   │       ├── npcs-generated.json
│   │       ├── journal-meta.json, journal-session-N.json
│   │       ├── sw-dm-session-N.log
│   │       └── images/session-N/
│   ├── characters/          # Personnages globaux
│   ├── world/               # Données monde (npcs.json, geography.json)
│   └── [JSON files]         # names, npc-traits, monsters, treasure, etc.
├── docs/                    # Documentation
│   ├── markdown-new/        # Règles D&D 5e complètes
│   └── [guides techniques]  # optional-features-summary, log-rotation, etc.
├── Makefile                 # Compilation tous les binaires
└── CLAUDE.md                # Ce fichier
```

---

**Version** : 2.0 (Février 2026)
**Dernière mise à jour** : Guidelines Claude Code intégrées, focus sw-web comme interface principale
