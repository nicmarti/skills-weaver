# SkillsWeaver

![SkillsWeaver Logo](logo.png)

**SkillsWeaver** is an interactive tabletop RPG engine powered by [Claude Code](https://claude.ai/claude-code) created by Nicolas Martignole.
The engine is based on [Dungeon&Dragon v5.2 French version](https://media.dndbeyond.com/compendium-images/srd/5.2/FR_SRD_CC_v5.2.1.pdf) rules. It combines AI orchestration with Go CLI tools to create a complete role-playing experience.


## See it live on YouTube

You can watch a sample game session on [my YouTube channel](https://youtu.be/K5CCB7MmegM) - English subtitles available

## How to Play

**To play a game session:**

```bash
# Build the tools
make build

# Set your Anthropic API key
export ANTHROPIC_API_KEY="your_key"

# Launch the Dungeon Master
./sw-dm
```

The `sw-dm` application provides an immersive, interactive RPG experience with:
- Streaming narrative responses
- Automatic dice rolling and rule application
- Adventure state management (party, inventory, journal)
- Optional AI image generation during gameplay

> **Note:** While Claude Code can also orchestrate gameplay using the agents and skills in this repository, `sw-dm` provides a more streamlined and immersive experience for actual game sessions.

## What is this repository?

SkillsWeaver demonstrates how to build a complex, multi-tool AI application using Claude Code's skills and agents system. It includes:

- **Autonomous Dungeon Master** (`sw-dm`) - Interactive REPL with full agent loop and tool use
- **Dice rolling** with standard RPG notation (2d6+3, 4d6kh3, advantage/disadvantage)
- **Character generation** following Dungeon&Dragon v5.2 rules
- **Adventure management** with session tracking and automatic journaling
- **NPC generation** with personalities, motivations, and secrets
- **AI image generation** for characters, scenes, and monsters via fal.ai
- **Monster manual** based on official D&D manual
- **Treasure generation** 
- **Journal illustration** - automatically generate images for adventure logs

## Prerequisites

### 1. Claude Code (optional)

You can install [Claude Code](https://claude.ai/claude-code), Anthropic's official CLI for Claude if you want to test skills or each Agents.
Else, use directly the go cli with a valid ANTHROPIC API key.

### 2. Go

Go 1.25 is required to build the CLI tools:

```bash
# macOS
brew install go

# Or download from https://go.dev/dl/
```

### 3. fal.ai API Key (for image generation) - OPTIONAL

**OPTIONAL for `sw-image` and `sw-dm` (image generation)**

Get your API key from [fal.ai](https://fal.ai) and set it:

```bash
export FAL_KEY="your_fal_ai_api_key"
```

The `sw-dm` Dungeon Master can generate images during gameplay if FAL_KEY is configured. The skill-based `sw-image` tool also uses it for character portraits and scene illustrations.

### 4. Anthropic API Key

**REQUIRED for `sw-dm` (Dungeon Master agent)**
**OPTIONAL for `sw-adventure enrich` (journal enrichment)**

The autonomous Dungeon Master (`sw-dm`) requires direct access to Claude API for the agent loop. The `sw-adventure enrich` command also uses it for bilingual journal descriptions.

Get your API key from [Anthropic Console](https://console.anthropic.com/) and set it:

```bash
export ANTHROPIC_API_KEY="your_anthropic_api_key"
```

**Usage:**
- `sw-dm`: Uses Claude Haiku 4.5 for fast, immersive game sessions (~$1/M input tokens, ~$5/M output tokens)
- `sw-adventure enrich`: Uses Claude Haiku 4.5 for cost-effective descriptions (~$0.0003 per entry)

## Quick Start

### 1. Build all CLI tools

```bash
make build
```

Or manually:

```bash
go build -o sw-dice ./cmd/dice
go build -o sw-character ./cmd/character
go build -o sw-adventure ./cmd/adventure
go build -o sw-names ./cmd/names
go build -o sw-npc ./cmd/npc
go build -o sw-image ./cmd/image
go build -o sw-monster ./cmd/monster
go build -o sw-treasure ./cmd/treasure
go build -o sw-equipment ./cmd/equipment  # Equipment browser
go build -o sw-spell ./cmd/spell          # Spell reference
go build -o sw-dm ./cmd/dm                # Autonomous Dungeon Master
```

### 2. Start Claude Code

```bash
claude
```

### 3. Try the skills

Once in Claude Code, the skills are automatically discovered. Try:

- *"Roll 4d6 keep highest 3"* → Uses dice-roller skill
- *"Create a dwarf fighter named Thorin"* → Uses character-generator skill
- *"Generate a portrait for Thorin"* → Uses image-generator skill
- *"Start a new adventure called The Lost Mine"* → Uses adventure-manager skill

## Autonomous Dungeon Master (sw-dm)

The `sw-dm` binary is a standalone Go application that acts as an autonomous Dungeon Master using the Anthropic API directly. Unlike the Claude Code skills that require manual orchestration, `sw-dm` runs a complete **agent loop** with tool use.

### Features

- **Full Agent Loop**: User → Claude → Tool Use → Execution → Claude → Response
- **Streaming Responses**: Real-time text streaming for immersive narrative
- **Adventure Auto-Loading**: Automatically loads party, inventory, journal, and game state
- **Direct Go Package Calls**: Tools call internal Go packages directly (no subprocess execution)
- **Interactive REPL**: Command-line interface with conversation history
- **Context Management**: Smart truncation to stay under token limits

### Available Tools

The Dungeon Master has access to these tools during gameplay:

| Tool | Purpose |
|------|---------|
| `roll_dice` | Roll dice with RPG notation (d20, 2d6+3, 4d6kh3) |
| `get_monster` | Look up monster stats from the bestiary |
| `log_event` | Record events in the adventure journal |
| `add_gold` | Modify the party's gold |
| `get_inventory` | Check the shared inventory |
| `generate_treasure` | Generate treasure using D&D 5e tables |
| `generate_npc` | Create NPCs with personality and motivations |
| `generate_image` | Generate fantasy-style images from prompts (requires FAL_KEY) |

### Usage

```bash
# Launch the Dungeon Master
./sw-dm

# Select your adventure from the menu
# The DM loads the full context and starts the REPL

> We enter the dark corridor cautiously
[DM narrates and rolls dice automatically]

> I attack the goblin with my longsword
[DM rolls attack, damage, updates journal]

> exit
```

### Example Session

```
============================================================
  La Crypte des Ombres
============================================================

📍 Lieu: Niveau Inférieur - Grande Chambre Funéraire
💰 Or: 1593 po

👥 Groupe:
   - Aldric (human fighter, niveau 1)
   - Lyra (elf magic-user, niveau 1)
   - Thorin (dwarf cleric, niveau 1)
   - Gareth (human fighter, niveau 1)

📖 Dernière action: Une créature scellée par des chaînes de
    mithril respire au centre de la chambre...
============================================================

> Aldric s'approche prudemment de la créature

[🎲 roll_dice: Perception check]
Le Maître du Jeu lance les dés...

Aldric avance avec précaution. Les chaînes de mithril tintent
légèrement à chaque respiration de la créature. Son marteau
Frappe-Juste émet une faible lueur bleue, réagissant à une
présence magique puissante...

[✓ log_event: Aldric approaches the chained creature]
```

### Architecture

```
sw-dm
├── internal/agent/
│   ├── agent.go          # Main agent loop
│   ├── tools.go          # Tool registry
│   ├── context.go        # Conversation & adventure context
│   └── streaming.go      # Event processing
├── internal/dmtools/     # Tool implementations
└── cmd/dm/main.go        # REPL application
```

### Why Two Approaches?

**Claude Code Skills** (`.claude/skills/`):
- Best for: Collaborative development, learning, exploring
- User controls flow and decisions
- Claude Code provides IDE integration

**Autonomous DM** (`sw-dm`):
- Best for: Actual gameplay sessions
- DM controls flow and makes decisions autonomously
- Faster, more immersive experience
- Direct API access with streaming

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                      Claude Code                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Skills      │  │ Agents      │  │ CLI Tools   │          │
│  │ (markdown)  │──│ (markdown)  │──│ (Go)        │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│        │                │                │                   │
│        ▼                ▼                ▼                   │
│  ┌─────────────────────────────────────────────────┐        │
│  │              Orchestration Layer                 │        │
│  │  - Skill discovery and invocation               │        │
│  │  - Agent delegation for complex tasks           │        │
│  │  - Tool execution (Bash, Read, Write...)        │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
   ┌──────────┐        ┌──────────┐        ┌──────────┐
   │ data/    │        │ fal.ai   │        │ Terminal │
   │ (JSON)   │        │ (images) │        │ (output) │
   └──────────┘        └──────────┘        └──────────┘
```

### Skills (`.claude/skills/`)

Markdown files that teach Claude how to use specific tools:
- `dice-roller` - Roll dice with RPG notation
- `character-generator` - Create D&D 5e characters
- `adventure-manager` - Manage campaigns and sessions
- `name-generator` - Generate fantasy names by species
- `npc-generator` - Create NPCs with personalities and secrets
- `image-generator` - Generate fantasy illustrations
- `journal-illustrator` - Illustrate adventure journals
- `monster-manual` - Monster stats and encounters
- `treasure-generator` - Generate treasure using D&D 5e tables
- `equipment-browser` - Browse weapons, armor, and gear
- `spell-reference` - Spell details by class and level

### Agents (`.claude/agents/`)

Specialized sub-agents for complex tasks:
- `dungeon-master` - Run game sessions with narrative
- `character-creator` - Guide players through character creation
- `rules-keeper` - Answer rules questions

### CLI Tools (`cmd/`)

Go binaries that perform the actual work:
- `sw-dm` - Autonomous Dungeon Master with full agent loop
- `sw-dice` - Dice rolling engine
- `sw-character` - Character management
- `sw-adventure` - Adventure/campaign tracking
- `sw-names` - Fantasy name generation
- `sw-npc` - NPC generation with personalities
- `sw-image` - Image generation via fal.ai
- `sw-monster` - Monster stats and encounters
- `sw-treasure` - Treasure generation
- `sw-equipment` - Equipment catalog (weapons, armor, gear)
- `sw-spell` - Spell reference (divine and arcane)

## Example: Enriching Journal Entries with AI

The journal enrichment feature automatically generates detailed, bilingual descriptions for your adventure log using Claude AI. These descriptions are optimized for image generation and provide rich context.

```bash
# Preview entries that need enrichment (dry-run mode)
./sw-adventure enrich "my-adventure" --dry-run

# Enrich all entries without descriptions
./sw-adventure enrich "my-adventure"

# Enrich only recent entries
./sw-adventure enrich "my-adventure" --recent=10

# Enrich entries from a specific session
./sw-adventure enrich "my-adventure" --session=2

# Re-generate existing descriptions (force mode)
./sw-adventure enrich "my-adventure" --force

# Preview with filters
./sw-adventure enrich "my-adventure" --from=50 --to=100 --dry-run
```

The AI generates descriptions following these guidelines:
- **Length:** 30-50 words per description
- **Format:** [Characters] + [Location] + [Action] + [Atmosphere] + [Visual Details]
- **Languages:** English (for image generation) + French (for readability)
- **Context-aware:** Uses party composition, recent events, and session info

Example output:
```
Entry 88 [note]:
📝 EN: Aldric, Lyra, Thorin, and Gareth stand before Mother Isabelle in the
       candlelit halls of the Convent, their burden lifted as the abbess
       accepts stewardship of the pilgrimage to Twilight Mountain.

📝 FR: Aldric, Lyra, Thorin et Gareth se tiennent devant Mère Isabelle dans
       les salles du Couvent illuminées aux chandelles, soulagés de leur
       fardeau tandis que l'abbesse accepte la garde du pèlerinage.
```

## Example: Illustrating an Adventure Journal

After enriching your journal, automatically generate images for key moments:

```bash
# Preview what would be generated
./sw-image journal "my-adventure" --dry-run

# Generate images (parallel, fast)
./sw-image journal "my-adventure"

# Use a higher quality model with a valid fal.ai key 
./sw-image journal "my-adventure" --model=banana

# Only combat scenes
./sw-image journal "my-adventure" --types=combat
```

Output:
```
data/adventures/my-adventure/images/
├── journal_002_exploration_schnell.png
├── journal_008_combat_schnell.png
├── journal_015_discovery_schnell.png
└── journal_029_session_schnell.png
```

## Available Models (fal.ai)

| Model | Speed | Cost/image | Use Case |
|-------|-------|------------|----------|
| `schnell` | ~3s | ~$0.003 | Fast iterations, drafts, testing |
| `banana` | ~5s | ~$0.039 | Better quality, final renders |

**Cost comparison for 20 images:**
- schnell: 20 × $0.003 = **$0.06**
- banana: 20 × $0.039 = **$0.78**

## Project Structure

```
skillsweaver/
├── .claude/
│   ├── skills/              # Claude Code skills
│   └── agents/              # Specialized sub-agents
├── cmd/                     # Go CLI source code
│   ├── dice/                # sw-dice
│   ├── character/           # sw-character
│   ├── adventure/           # sw-adventure
│   ├── names/               # sw-names
│   ├── npc/                 # sw-npc
│   ├── image/               # sw-image
│   ├── monster/             # sw-monster
│   ├── treasure/            # sw-treasure
│   ├── equipment/           # sw-equipment
│   ├── spell/               # sw-spell
│   └── dm/                  # sw-dm (Autonomous DM)
├── internal/                # Go packages
│   ├── agent/               # Agent loop orchestration
│   │   ├── agent.go         # Main agent loop
│   │   ├── tools.go         # Tool registry
│   │   ├── context.go       # Context management
│   │   └── streaming.go     # Event processing
│   ├── dmtools/             # Tool implementations for DM
│   ├── dice/                # Dice rolling logic
│   ├── character/           # Character management
│   ├── adventure/           # Adventure management
│   ├── monster/             # Bestiary
│   ├── treasure/            # Treasure generation
│   ├── npc/                 # NPC generation
│   ├── names/               # Name generation
│   ├── image/               # Image generation
│   ├── equipment/           # Equipment catalog
│   └── spell/               # Spell reference
├── data/
│   ├── characters/          # Saved characters (JSON)
│   ├── adventures/          # Saved adventures (JSON)
│   │   └── <adventure>/
│   │       ├── adventure.json
│   │       ├── party.json
│   │       ├── inventory.json
│   │       ├── journal-*.json
│   │       └── images/
│   ├── monsters.json        # Bestiary
│   ├── treasure.json        # Treasure tables
│   └── names.json           # Name dictionaries
├── CLAUDE.md                # Project instructions for Claude
└── README.md                # This file
```

## License

**SkillsWeaver** is licensed under the [Creative Commons Attribution-ShareAlike 4.0 International (CC BY-SA 4.0)](http://creativecommons.org/licenses/by-sa/4.0/) license.

This means you are free to:
- **Share** — copy and redistribute the material
- **Adapt** — remix, transform, and build upon the material

As long as you:
- **Give attribution** to Nicolas MARTIGNOLE (the original author)
- **Share alike** — distribute your contributions under the same license

SkillsWeaver builds upon:
- **Claude Code** - © Anthropic
- **fal.ai** - Image generation API

This work includes material taken from the Lazy GM's Resource Document by Michael E. Shea of SlyFlourish.com, available under a Creative Commons Attribution 4.0 International License.

This work includes material taken from the System Reference Document 5.1 ("SRD 5.1") by Wizards of the Coast LLC and available at https://dnd.wizards.com/resources/systems-reference-document. The SRD 5.1 is licensed under the Creative Commons Attribution 4.0 International License available at https://creativecommons.org/licenses/by/4.0/legalcode.

See the [LICENSE](LICENSE) file for full legal details.

## Author

This engine and the original idea is from **Nicolas MARTIGNOLE**, Principal Engineer at Back Market and Devoxx France's creator/organizer.

You can reach Nicolas by email: [nicolas.martignole@devoxx.fr](mailto:nicolas.martignole@devoxx.fr)

