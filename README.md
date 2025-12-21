# SkillsWeaver

![SkillsWeaver Logo](logo.png)

**SkillsWeaver** is an interactive tabletop RPG engine powered by [Claude Code](https://claude.ai/claude-code) created by Nicolas Martignole.
The engine is based on [Basic Fantasy RPG](https://www.basicfantasy.org/) rules. It combines AI orchestration with Go CLI tools to create a complete role-playing experience.


## See it live on YouTube

You can watch a sample game session on [my YouTube channel](https://youtu.be/K5CCB7MmegM) - English subtitles available

## What is this repository?

SkillsWeaver demonstrates how to build a complex, multi-tool AI application using Claude Code's skills and agents system. It includes:

- **Dice rolling** with standard RPG notation (2d6+3, 4d6kh3, advantage/disadvantage)
- **Character generation** following BFRPG rules (4 races, 4 classes)
- **Adventure management** with session tracking and automatic journaling
- **NPC generation** with personalities, motivations, and secrets
- **AI image generation** for characters, scenes, and monsters via fal.ai
- **Monster manual** with 33 classic fantasy creatures
- **Treasure generation** using official BFRPG tables
- **Journal illustration** - automatically generate images for adventure logs

## Prerequisites

### 1. Claude Code

Install [Claude Code](https://claude.ai/claude-code), Anthropic's official CLI for Claude. We recommend to use the native version.
A key is not required to play as the game can use the Haiku model. 

### 2. Go

Go 1.25 is required to build the CLI tools:

```bash
# macOS
brew install go

# Or download from https://go.dev/dl/
```

### 3. fal.ai API Key (for image generation) - OPTIONAL

Get your API key from [fal.ai](https://fal.ai) and set it:

```bash
export FAL_KEY="your_fal_ai_api_key"
```

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
- `character-generator` - Create BFRPG characters
- `adventure-manager` - Manage campaigns and sessions
- `image-generator` - Generate fantasy illustrations
- `journal-illustrator` - Illustrate adventure journals

### Agents (`.claude/agents/`)

Specialized sub-agents for complex tasks:
- `dungeon-master` - Run game sessions with narrative
- `character-creator` - Guide players through character creation
- `rules-keeper` - Answer rules questions

### CLI Tools (`cmd/`)

Go binaries that perform the actual work:
- `sw-dice` - Dice rolling engine
- `sw-character` - Character management
- `sw-adventure` - Adventure/campaign tracking
- `sw-image` - Image generation via fal.ai
- `sw-monster` - Monster stats and encounters
- `sw-treasure` - Treasure generation

## Example: Illustrating an Adventure Journal

After playing a session, automatically generate images for key moments:

```bash
# Preview what would be generated
./sw-image journal "my-adventure" --dry-run

# Generate images (parallel, fast)
./sw-image journal "my-adventure"

# Use a higher quality model
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
│   ├── skills/           # Claude Code skills
│   └── agents/           # Specialized sub-agents
├── cmd/                  # Go CLI source code
├── internal/             # Go packages
├── data/
│   ├── characters/       # Saved characters (JSON)
│   ├── adventures/       # Saved adventures (JSON)
│   └── images/           # Generated images
├── CLAUDE.md             # Project instructions for Claude
└── README.md             # This file
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
- **Basic Fantasy RPG** - © Chris Gonnerman and contributors (Open Game License)
- **Claude Code** - © Anthropic
- **fal.ai** - Image generation API

See the [LICENSE](LICENSE) file for full legal details.

## Author

This engine and the original idea is from **Nicolas MARTIGNOLE**, Principal Engineer at Back Market and Devoxx France's creator/organizer.

You can reach Nicolas by email: [nicolas.martignole@devoxx.fr](mailto:nicolas.martignole@devoxx.fr)





