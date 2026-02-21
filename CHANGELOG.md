# Changelog

All notable changes to SkillsWeaver will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Web Interface & Campaign Planning (February 2026)

- **Web Interface (`sw-web`)** - Complete browser-based UI for adventure creation and gameplay
  - Modern Dark Fantasy themed UI
  - Adventure creation with automatic campaign plan generation
  - Real-time streaming responses via Server-Sent Events (SSE)
  - Inline display of images, maps, and tactical scenes
  - Party status panel with live HP, AC, and level tracking
  - Live journal updates during gameplay
  - Runtime model selector (Haiku, Sonnet, Opus)
  - Session management with 30-minute timeout
  - Responsive design for desktop and tablet

- **Automatic Campaign Plan Generation**
  - 3-act structure (beginning, twists, final confrontation)
  - Main antagonist with motivations and narrative arc
  - 2-3 critical foreshadows linked to acts
  - MacGuffins and important locations
  - Pacing estimation (sessions, duration)
  - Automatic briefing at session start (hidden from players)

- **Character Progression Tools**
  - `update_character_stat` - Modify character stats during gameplay
  - `long_rest` - Handle long rest mechanics with HP/spell restoration

- **Combat Enhancements**
  - Automatic treasure generation after combat
  - Tactical map display in web interface
  - Monster HP rolling for combat instances

- **Agent System Improvements**
  - Agent-to-agent communication (dungeon-master → rules-keeper, world-keeper, character-creator)
  - Stateful nested agents with conversation history
  - Agent state persistence in `agent-states.json`
  - Token limits per agent (Main: 50K, Nested: 20K)
  - Security restrictions on nested agents (read-only consultants)
  - Performance metrics tracking (tokens, response time)

- **World Consistency**
  - NPC persistence system (`npcs-generated.json`)
  - NPC importance tracking (mentioned → interacted → recurring → key)
  - World-keeper validation for maps and locations
  - Geography and faction coherence checks

### Changed

- **Complete D&D 5e Migration**
  - Removed all Basic Fantasy RPG references
  - Updated to D&D 5e System Reference Document (SRD 5.1)
  - 9 official species (was 7)
  - 12 official classes (was 4)
  - 300+ monsters from D&D 5e bestiary
  - D&D 5e treasure tables and equipment

- **System Prompt Optimization**
  - Improved tool usage patterns
  - Better state management
  - Enhanced narrative consistency
  - Cleaner conversation history

- **Model Configuration**
  - Rules-keeper uses Haiku 4.5 (fast, cost-effective)
  - World-keeper uses Haiku 4.5 (geographical queries)
  - Character-creator uses Haiku 4.5 (build recommendations)
  - Main DM can switch between Haiku/Sonnet/Opus

### Fixed

- Fixed tactical maps not displaying in web interface
- Fixed log_event tool not being called consistently
- Fixed orphaned tool_results in agent conversation history
- Hardened message deserialization for state-modifying tools
- Expanded state-modifying tools list (combat, treasure, character updates)

---

## [0.2.0] - 2025-12-15

### Added

- **Autonomous Dungeon Master (`sw-dm`)**
  - Complete agent loop with Anthropic API
  - Full readline support with history
  - Streaming narrative responses
  - Context management with token limits
  - Tool registry system
  - Adventure auto-loading

- **Journal System**
  - Session-based journal files (`journal-session-N.json`)
  - Journal metadata tracking (`journal-meta.json`)
  - Automatic event logging
  - Image organization by session

- **Image Generation**
  - Integration with fal.ai API
  - Character portraits
  - Scene illustrations
  - Monster art
  - Tactical maps

- **NPC System**
  - Complete personality generation
  - Motivations and secrets
  - Species and occupation
  - Attitude system

- **Map Generation**
  - Tactical maps (combat scenes)
  - City maps (settlements)
  - Region maps (exploration)
  - Dungeon maps (interior exploration)

### Changed

- Migrated from BFRPG to D&D 5e rules
- Improved dice notation parsing
- Enhanced character sheet generation

---

## [0.1.0] - 2025-11-20

### Added

- **Core CLI Tools**
  - `sw-dice` - Dice rolling with RPG notation
  - `sw-character` - Character creation and management
  - `sw-adventure` - Adventure and session tracking
  - `sw-names` - Fantasy name generation
  - `sw-monster` - Monster stats and encounters
  - `sw-treasure` - Treasure generation
  - `sw-equipment` - Equipment catalog
  - `sw-spell` - Spell reference

- **Claude Code Skills**
  - dice-roller
  - character-generator
  - adventure-manager
  - name-generator
  - npc-generator
  - monster-manual
  - treasure-generator
  - equipment-browser
  - spell-reference

- **Claude Code Agents**
  - dungeon-master
  - character-creator
  - rules-keeper

- **Data Files**
  - monsters.json (Basic Fantasy bestiary)
  - treasure.json (treasure tables)
  - names.json (name dictionaries by species)
  - npc-traits.json (personality traits)

### Documentation

- Complete README with quick start
- CLAUDE.md with project instructions
- Skill documentation
- Agent personas

---

## Release Notes

### Version Numbering

- **Major (X.0.0)**: Breaking changes, major feature additions
- **Minor (0.X.0)**: New features, non-breaking changes
- **Patch (0.0.X)**: Bug fixes, minor improvements

### Upgrade Notes

#### Migrating to D&D 5e (0.2.0 → Current)

If you have adventures created with BFRPG rules:

1. **Characters**: Manual recreation recommended (species/class changes)
2. **Monsters**: Automatically migrated to D&D 5e stats
3. **Treasure**: Tables updated to D&D 5e (backward compatible)
4. **Journal**: No migration needed

#### Migrating to Session-Based Journals (0.1.0 → 0.2.0)

```bash
# Migrate old monolithic journal to session-based structure
./sw-adventure migrate-journal <adventure-name>
```

This converts `journal.json` to `journal-session-N.json` files.

---

## Planned Features

### Short Term
- [ ] Multi-player support (multiple users in one adventure)
- [ ] Voice narration (Text-to-Speech for DM responses)
- [ ] Mobile-responsive web UI improvements

### Long Term
- [ ] Additional RPG systems (Pathfinder, Call of Cthulhu)
- [ ] Community campaign marketplace
- [ ] Co-DM mode (player and AI collaborate as joint DMs)
- [ ] VTT integration (Roll20, Foundry)

---

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

SkillsWeaver is licensed under [CC BY-SA 4.0](LICENSE).
