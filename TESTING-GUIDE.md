# Testing Guide: System Prompt Optimization

## Quick Start

```bash
# Build and restart sw-web
go build -o sw-web ./cmd/web
pkill -f sw-web
./sw-web

# Open in browser
open http://localhost:8085/play/les-naufrages-du-pierre-lune
```

## Test 1: Narrative Events (CRITICAL)

**Goal**: Verify `log_event` is called for narrative scenes without dice rolls.

**Baseline**:
```bash
# Check current journal entry count
cat data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json | jq '.entries | length'
# Expected: 4 entries
```

**Test Scenario**:
Interact with DM using pure narrative actions:

```
Player: "Marcus interroge Gareth sur la dame voilée. Qui est-elle ?"

Player: "Lyra propose : 'Nous devrions aller directement à Blackstone pour
détruire l'artefact. C'est notre seule chance.'"

Player: "Le groupe décide de partir immédiatement pour Blackstone."
```

**Verification**:
```bash
# Check new journal entry count
cat data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json | jq '.entries | length'
# Expected: 5-7 entries (+1-3 new)

# Check for narrative entries
cat data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json | jq '.entries[] | select(.type == "story" or .type == "npc")'
```

**Success Criteria**: At least +1 entry with `type: "story"` or `type: "npc"`

## Test 2: Combat Workflow

**Goal**: Verify post-combat workflow includes treasure generation.

**Test Scenario**:
```
Player: "Marcus attaque les cultistes avec son épée !"
[Resolve combat to victory]
```

**Verification**:
```bash
# Check tool call sequence
tail -50 data/adventures/les-naufrages-du-pierre-lune/sw-dm-session-5.log | grep "TOOL CALL"
```

**Expected Sequence**:
1. `roll_dice` (attack/damage)
2. `update_hp` (apply damage)
3. `log_event` (combat result)
4. `add_xp` (award XP)
5. `generate_treasure` (loot) ← MUST be present
6. `add_gold` (add to inventory)

**Success Criteria**: `generate_treasure` appears after combat victory

## Test 3: Crash Recovery (CRITICAL)

**Goal**: Verify narrative context survives session restarts.

**Test Scenario**:
1. Play narrative scene with critical decisions (Test 1)
2. Verify `log_event` was called
3. Restart sw-web:
   ```bash
   pkill -f sw-web
   ./sw-web
   ```
4. Reload browser page
5. Ask DM: "Résume la situation actuelle"

**Success Criteria**: DM mentions the decisions/dialogues from step 1

## Test 4: Regression Check

**Goal**: Verify existing tools still work correctly.

**Test Scenarios**:

```bash
# 1. Location tracking
Player: "Nous partons vers la Gorge du Passage"
# Verify: update_location called

# 2. Dice rolling
Player: "Marcus tente d'intimider le garde (jet d'Intimidation)"
# Verify: roll_dice called, result shown

# 3. XP attribution
Player: "Après la victoire contre les cultistes"
# Verify: add_xp called automatically

# 4. Treasure generation
Player: "Nous fouillons les corps"
# Verify: generate_treasure called

# 5. HP tracking
Player: "Caelian lance Soins sur Marcus"
# Verify: use_spell_slot + update_hp called
```

**Success Criteria**: All tools execute without errors

## Metrics to Track

### Before Fix (Baseline from Session 5)
- Journal entries: 4 total
- Story entries: 3
- Combat entries: 0
- Ratio story:combat: N/A
- log_event calls during narrative scenes: 0/5 (0%)

### After Fix (Expected)
- Journal entries: 8-10 total per session
- Story entries: 6-8
- Combat entries: 1-2
- Ratio story:combat: 3:1 or higher
- log_event calls during narrative scenes: 3-4/5 (70-80%)

## Quick Commands

```bash
# Count journal entries
jq '.entries | length' data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json

# List entry types
jq '.entries[] | {type, content: .content[:50]}' data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json

# Count by type
jq '[.entries[].type] | group_by(.) | map({type: .[0], count: length})' data/adventures/les-naufrages-du-pierre-lune/journal-session-5.json

# Check recent tool calls
tail -100 data/adventures/les-naufrages-du-pierre-lune/sw-dm-session-5.log | grep "TOOL CALL"

# Extract CLI commands (for debugging)
./scripts/extract-cli-commands.sh les-naufrages-du-pierre-lune
```

## Troubleshooting

### Issue: No new journal entries after narrative scene

**Diagnosis**:
```bash
# Check if log_event was called at all
tail -100 data/adventures/.../sw-dm-session-5.log | grep "log_event"
```

**Possible Causes**:
1. Reminder not loaded (check system-prompt.log for "RAPPEL CRITIQUE")
2. Narrative scene too simple (add more critical information)
3. Model chose not to log (check if reason was valid)

### Issue: Journal shows only auto-generated entries

**Diagnosis**:
```bash
# Check entry types
jq '.entries[] | .type' data/adventures/.../journal-session-5.json
```

**Possible Causes**:
1. Only combat/mechanical actions tested (add narrative scenes)
2. Model still confused about auto vs manual (enhance reminder text)

### Issue: sw-web won't start

**Diagnosis**:
```bash
# Check build errors
go build -o sw-web ./cmd/web

# Check port conflict
lsof -i :8085
```

**Solution**:
```bash
# Kill existing process
pkill -f sw-web

# Use different port
./sw-web --port=8086
```

## Rollback Procedure

If tests fail or regressions occur:

```bash
# 1. Restore backup
cp core_agents/agents/dungeon-master.md.backup core_agents/agents/dungeon-master.md

# 2. Revert agent.go
git checkout internal/agent/agent.go

# 3. Rebuild
go build -o sw-web ./cmd/web

# 4. Restart
pkill -f sw-web
./sw-web
```

## Success Indicators

✅ **Good Signs**:
- Journal entries increase by 1-2 per narrative scene
- Mix of `story`, `npc`, `discovery`, `quest` types
- DM remembers context after restart
- Combat workflow includes treasure generation
- All existing tools work correctly

❌ **Warning Signs**:
- Journal only has `xp`, `loot`, `combat` entries
- No new entries after significant narrative events
- DM forgets context after restart
- Existing tools broken or throwing errors

## Next Steps After Testing

1. **Document Results**: Update IMPLEMENTATION-SUMMARY.md with actual metrics
2. **Adjust Reminder**: If needed, fine-tune postJournalReminder text
3. **Monitor Sessions**: Track 5-10 sessions for consistent improvement
4. **Merge to Master**: If successful, merge fix/optimize-system-prompt-architecture
5. **Close Issue**: Update GitHub issue with results
