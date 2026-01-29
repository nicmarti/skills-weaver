# Campaign Planning System - End-to-End Testing Guide

**Purpose**: Validate the complete campaign planning system implementation
**Estimated Time**: 30-45 minutes for complete testing
**Prerequisites**: ANTHROPIC_API_KEY set, Go 1.21+, compiled binaries

---

## Test Suite Overview

| Test ID | Description | Expected Time | Priority |
|---------|-------------|---------------|----------|
| E2E-1 | Campaign Plan Generation | 5 min | HIGH |
| E2E-2 | Automated Session Briefing | 5 min | HIGH |
| E2E-3 | World-Keeper Confidentiality | 5 min | CRITICAL |
| E2E-4 | DM Narration Integration | 10 min | HIGH |
| E2E-5 | Campaign Tools Usage | 5 min | MEDIUM |
| E2E-6 | Backward Compatibility | 5 min | HIGH |
| E2E-7 | Pacing Metrics Update | 3 min | MEDIUM |

---

## E2E-1: Campaign Plan Generation

### Objective
Verify that providing a theme during adventure creation generates a complete campaign-plan.json file.

### Steps

1. **Start Web Server**
   ```bash
   go build -o sw-web ./cmd/web
   ./sw-web --port=8085 --debug
   ```

2. **Create Adventure with Theme**
   - Open http://localhost:8085
   - Click "+ Nouvelle Aventure"
   - Fill form:
     - **Nom**: Test Campaign Planning
     - **Description**: Test adventure for E2E validation
     - **Theme**: A mysterious artifact unleashes ancient magic. Heroes must stop it before kingdoms fall into chaos.
   - Click "Créer"

3. **Verify File Creation**
   ```bash
   ls -la data/adventures/test-campaign-planning/campaign-plan.json
   cat data/adventures/test-campaign-planning/campaign-plan.json | jq '.narrative_structure.objective'
   ```

### Expected Results

✅ **Success Criteria**:
- `campaign-plan.json` created within 60 seconds
- File contains valid JSON structure
- 3 acts present with status "pending"
- `objective` field populated with relevant text
- At least 1 foreshadow in `foreshadows.active`
- `metadata.campaign_title` is descriptive (not just adventure name)

❌ **Failure Indicators**:
- File not created after 60 seconds
- JSON parse error
- Empty or incomplete structure
- Generic/placeholder text in critical fields

### Troubleshooting

**Issue**: File not created
- **Check**: Server logs for "campaign plan generation failed"
- **Fix**: Verify ANTHROPIC_API_KEY is set correctly
- **Fix**: Check API rate limits

**Issue**: Invalid JSON
- **Check**: Server logs for JSON parsing errors
- **Fix**: The prompt may need adjustment (see `internal/web/handlers.go:generateCampaignPlan`)

---

## E2E-2: Automated Session Briefing

### Objective
Verify that `start_session` automatically loads campaign plan, consults world-keeper silently, and injects briefing into system context.

### Steps

1. **Build and Run DM Agent**
   ```bash
   go build -o sw-dm ./cmd/dm
   ./sw-dm
   ```

2. **Select Test Adventure**
   - Choose "test-campaign-planning" from list
   - Wait for initialization

3. **Start Session**
   - Type: `start_session`
   - Observe output

4. **Check Logs**
   ```bash
   tail -50 data/adventures/test-campaign-planning/sw-dm-session-1.log
   ```

### Expected Results

✅ **Success Criteria**:
- Console shows: `[Consulting world-keeper...]`
- Console shows: `✓ Session 1 démarrée`
- No world-keeper response visible to user
- Log file contains: `[SYSTEM] Injected campaign briefing`
- Log file shows world-keeper invocation with "(silent mode)"
- Agent responds with contextual opening narration (not generic)

❌ **Failure Indicators**:
- World-keeper response text displayed in console
- No `[Consulting world-keeper...]` notification
- Generic opening ("Bonjour, vous êtes dans une taverne...")
- Missing system briefing injection log

### Validation

**Check system prompt injection**:
```bash
# After session start, system-prompt.log should include campaign context
grep -A 30 "CAMPAIGN CONTEXT" system-prompt.log
```

Should contain:
- "Act 1: [Title]"
- "Campaign Objective: [Objective]"
- "World-Keeper Briefing:"

---

## E2E-3: World-Keeper Confidentiality

### Objective
**CRITICAL TEST**: Verify that world-keeper responses are NEVER displayed directly to players.

### Steps

1. **Continue from E2E-2** (session already started)

2. **Manually Invoke World-Keeper**
   - Type: `invoke_agent("world-keeper", "What should I focus on in this session?")`
   - Observe response

3. **Check Different Scenarios**

   **Scenario A - Silent Mode (Briefing)**:
   ```
   invoke_agent("world-keeper", "Briefing for upcoming encounter", silent=true)
   ```

   **Scenario B - Normal Mode (Query)**:
   ```
   invoke_agent("world-keeper", "What locations exist in Valdorine?")
   ```

### Expected Results

✅ **Success Criteria - Silent Mode**:
- Console shows: `✓ world-keeper consulted (guidance injected in system context)`
- NO world-keeper response text displayed
- Next DM response incorporates guidance naturally

✅ **Success Criteria - Normal Mode**:
- Console shows brief notification
- World-keeper response IS displayed (this is intentional for queries)
- DM can use response for narration

❌ **Failure Indicators - Silent Mode**:
- Full world-keeper response displayed to player
- Player sees: "According to world-keeper..." in DM narration
- Direct quotes from world-keeper briefing

### Critical Validation

**Test DM Persona Compliance**:
After silent world-keeper consultation, ask DM to narrate the next scene.

❌ **UNACCEPTABLE**:
```
DM: Le world-keeper m'informe que Vaskir est arrivé à Shasseth il y a 2 jours.
```

✅ **ACCEPTABLE**:
```
DM: Les rumeurs dans les tavernes du port parlent d'un navire noir aperçu
près de Shasseth il y a deux jours. Les marins superstitieux évitent d'en
parler ouvertement.
```

---

## E2E-4: DM Narration Integration

### Objective
Verify that DM integrates campaign briefing naturally without direct quotes.

### Steps

1. **Continue Session** (from E2E-3)

2. **Player Action**:
   ```
   User: I ask the tavern keeper about recent ships arriving at port.
   ```

3. **Analyze DM Response**

### Expected Results

✅ **Success Criteria**:
- DM response shows awareness of campaign context
- Information delivered through NPC dialogue or environmental clues
- No meta-references to "world-keeper" or "campaign plan"
- Narrative advances campaign objectives naturally

**Example Good Response**:
```
DM: Le tavernier, un homme trapu aux mains calleuses, baisse la voix :
"Un navire étrange, deux jours passés. Noir comme la nuit, sans pavillon.
Les dockers refusent d'en parler. Mauvais présage, qu'ils disent."

Il essuie son comptoir nerveusement.

Que faites-vous ?
```

❌ **Failure Indicators**:
- "According to my briefing..."
- "The campaign plan indicates..."
- "World-keeper says..."
- Generic response unrelated to campaign context

---

## E2E-5: Campaign Tools Usage

### Objective
Verify that campaign plan tools work correctly during session.

### Steps

1. **Query Campaign State**
   ```
   get_campaign_plan(section="current_act")
   ```

2. **Complete Plot Point**
   ```
   update_campaign_progress(action="complete_plot_point", plot_point_id="first_encounter")
   ```

3. **Add Narrative Thread**
   ```
   add_narrative_thread(thread_name="mysterious_stranger_identity")
   ```

4. **Check Updated State**
   ```
   get_campaign_plan(section="progression")
   ```

5. **Verify Persistence**
   ```bash
   cat data/adventures/test-campaign-planning/campaign-plan.json | jq '.progression'
   ```

### Expected Results

✅ **Success Criteria**:
- `get_campaign_plan` returns formatted summary with emojis
- `update_campaign_progress` confirms plot point added
- `add_narrative_thread` shows thread in active list
- JSON file updated on disk
- Progression reflects changes immediately

❌ **Failure Indicators**:
- Tool errors or exceptions
- Changes not persisted to disk
- Corrupted JSON after updates

---

## E2E-6: Backward Compatibility

### Objective
Verify that adventures WITHOUT campaign-plan.json continue to work normally.

### Steps

1. **Create Legacy Adventure**
   ```bash
   # Via web UI - leave "Theme" field EMPTY
   ```
   - Name: Legacy Test
   - Description: Old-style adventure
   - Theme: **(leave blank)**

2. **Start Session in sw-dm**
   - Select "legacy-test" adventure
   - Call `start_session`

3. **Verify Legacy Behavior**
   ```bash
   ls data/adventures/legacy-test/
   # Should NOT have campaign-plan.json
   ```

### Expected Results

✅ **Success Criteria**:
- Session starts normally
- NO campaign-plan.json created
- No errors about missing campaign plan
- Legacy foreshadows.json still works (if present)
- No world-keeper silent consultation (no campaign plan to brief)
- DM responds normally without campaign context

❌ **Failure Indicators**:
- Errors about missing campaign-plan.json
- sw-dm crashes or hangs
- Features disabled for legacy adventures

---

## E2E-7: Pacing Metrics Update

### Objective
Verify that pacing metrics update correctly as campaign progresses.

### Steps

1. **Check Initial Pacing**
   ```bash
   cat data/adventures/test-campaign-planning/campaign-plan.json | jq '.pacing'
   ```

2. **Simulate Act Progression**
   ```
   # In sw-dm
   end_session(summary="Act 1 completed - heroes reached destination")
   start_session()
   # ... play several sessions ...
   update_campaign_progress(action="advance_act", act_number=2)
   ```

3. **Verify Pacing Recalculation**
   ```bash
   cat data/adventures/test-campaign-planning/campaign-plan.json | jq '.pacing.act_breakdown'
   ```

### Expected Results

✅ **Success Criteria**:
- `sessions_played` increments correctly
- `sessions_remaining_estimate` decreases
- `act_breakdown` shows actual vs planned sessions
- Variance calculated (positive or negative)

**Example Pacing After 5 Sessions**:
```json
{
  "sessions_played": 5,
  "sessions_remaining_estimate": 5,
  "act_breakdown": {
    "act_1": {
      "planned": 4,
      "actual": 4,
      "variance": 0
    },
    "act_2": {
      "planned": 4,
      "actual": 1,
      "variance": -3
    }
  }
}
```

---

## Integration Test Matrix

| Feature | E2E-1 | E2E-2 | E2E-3 | E2E-4 | E2E-5 | E2E-6 | E2E-7 |
|---------|-------|-------|-------|-------|-------|-------|-------|
| Campaign Plan Generation | ✅ | | | | | | |
| Automated Session Briefing | | ✅ | | | | | |
| Silent World-Keeper | | ✅ | ✅ | | | | |
| DM Persona Compliance | | | ✅ | ✅ | | | |
| Campaign Tools | | | | | ✅ | | ✅ |
| Backward Compatibility | | | | | | ✅ | |
| Pacing Analytics | | | | | | | ✅ |

---

## Regression Testing

After completing all E2E tests, run unit tests to ensure no regressions:

```bash
# All tests
go test ./... -v

# Specific packages
go test ./internal/adventure -v
go test ./internal/agent -v
go test ./internal/dmtools -v
go test ./internal/web -v
```

**All tests must pass** ✅

---

## Test Report Template

```markdown
## Campaign Planning System - Test Report

**Date**: 2026-01-29
**Tester**: [Name]
**Environment**: Go [version], OS [platform]

### Test Results

| Test ID | Status | Notes |
|---------|--------|-------|
| E2E-1   | ✅/❌ | |
| E2E-2   | ✅/❌ | |
| E2E-3   | ✅/❌ | |
| E2E-4   | ✅/❌ | |
| E2E-5   | ✅/❌ | |
| E2E-6   | ✅/❌ | |
| E2E-7   | ✅/❌ | |

### Issues Found

1. [Issue description]
   - **Severity**: Critical/High/Medium/Low
   - **Steps to Reproduce**: [...]
   - **Expected**: [...]
   - **Actual**: [...]

### Recommendations

- [ ] Item 1
- [ ] Item 2

### Sign-Off

**Approved for Production**: ✅ / ❌
**Signature**: _______________
**Date**: _______________
```

---

## Automated Testing Script

For CI/CD integration:

```bash
#!/bin/bash
# test-campaign-planning.sh

set -e

echo "🧪 Testing Campaign Planning System..."

# Build binaries
echo "📦 Building binaries..."
go build -o sw-web ./cmd/web
go build -o sw-dm ./cmd/dm

# Unit tests
echo "✅ Running unit tests..."
go test ./internal/adventure ./internal/agent ./internal/dmtools -v

# Check for required files
echo "📁 Checking implementation files..."
test -f internal/adventure/campaign_plan.go || (echo "❌ Missing campaign_plan.go" && exit 1)
test -f internal/dmtools/campaign_plan_tools.go || (echo "❌ Missing campaign_plan_tools.go" && exit 1)

# Verify persona updates
echo "📝 Checking DM persona updates..."
grep -q "Confidentialité World-Keeper" core_agents/agents/dungeon-master.md || (echo "❌ Missing confidentiality section" && exit 1)

echo "✅ All automated checks passed!"
echo "⚠️  Manual E2E tests required (see E2E_TESTING_GUIDE.md)"
```

---

## Known Limitations

1. **Theme Quality**: Campaign plan quality depends on theme clarity
   - **Workaround**: Provide detailed, specific themes
   - **Example**: "Maritime conspiracy" → "A stolen sextant reveals an ancient entity sealed under a lost city"

2. **Generation Time**: Campaign plan generation takes 30-60 seconds
   - **Expected**: Using Haiku 4.5 for speed
   - **Timeout**: 60 seconds configured

3. **World-Keeper Depth**: Briefings limited to 4K tokens
   - **Expected**: Sufficient for session guidance
   - **Workaround**: Query specific topics for more detail

4. **Migration**: No automated migration from foreshadows.json
   - **Workaround**: Manual migration or create new adventures with themes

---

## Success Metrics

System considered **production-ready** if:

- ✅ 7/7 E2E tests pass
- ✅ All unit tests pass (0 failures)
- ✅ DM never quotes world-keeper directly in test sessions
- ✅ Campaign plan generation succeeds >95% of time
- ✅ Backward compatibility maintained (legacy adventures work)
- ✅ No data loss or corruption during testing
- ✅ Performance acceptable (<60s plan generation, <5s session start)

---

## Post-Deployment Monitoring

After deployment, monitor:

1. **Campaign Plan Generation Rate**
   - Log successful/failed generations
   - Track average generation time

2. **World-Keeper Confidentiality Compliance**
   - Review DM responses for direct quotes
   - User feedback on narrative quality

3. **Backward Compatibility**
   - Monitor errors from legacy adventures
   - Track migration requests

4. **Tool Usage**
   - Most/least used campaign tools
   - Tool error rates
