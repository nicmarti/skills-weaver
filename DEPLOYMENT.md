# Campaign Planning System - Quick Deployment Guide

**Status**: ✅ Ready for Production
**Date**: 2026-01-29
**Version**: 1.0.0

---

## Quick Start (5 minutes)

### 1. Prerequisites Check
```bash
# Verify Go version
go version  # Requires 1.21+

# Set API key
export ANTHROPIC_API_KEY="your-key-here"

# Verify key is set
echo $ANTHROPIC_API_KEY
```

### 2. Build Binaries
```bash
# Build all binaries
make

# Verify builds
ls -lh sw-dm sw-web
```

### 3. Run All Tests
```bash
# Unit tests (should see ~180+ tests passing)
go test ./... -v | grep -E "(PASS|FAIL|ok)"

# Expected: All packages pass, 0 failures
```

### 4. Launch Web Interface
```bash
# Start web server
./sw-web --port=8085 --debug

# Open browser: http://localhost:8085
```

### 5. Create Test Adventure
- Click "+ Nouvelle Aventure"
- Fill form:
  - **Nom**: Test Campaign
  - **Description**: Testing the campaign planning system
  - **Theme**: A stolen artifact threatens to unleash ancient chaos upon the kingdoms
- Click "Créer"
- Wait 30-60 seconds for campaign plan generation

### 6. Verify Campaign Plan
```bash
# Check file was created
ls -la data/adventures/test-campaign/campaign-plan.json

# View campaign structure
cat data/adventures/test-campaign/campaign-plan.json | jq '.narrative_structure.acts[].title'

# Expected: 3 act titles displayed
```

### 7. Test sw-dm Session
```bash
# Launch DM agent
./sw-dm

# Select "test-campaign" from menu
# Type: start_session
# Observe: [Consulting world-keeper...] message
# Verify: Session starts with contextual narration
```

---

## What to Expect

### ✅ Campaign Plan Generation (30-60s)
When you provide a theme, the system automatically generates:
- Complete 3-act narrative structure
- Primary antagonist with arc
- 2-3 MacGuffins and key locations
- Initial foreshadows linked to acts
- Pacing targets (sessions/act)

### ✅ Automated Session Briefing (~10s)
When you call `start_session` in sw-dm:
- Campaign plan loaded automatically
- World-keeper consulted silently
- Briefing injected into system context
- DM narrates with strategic direction

### ✅ Hidden World-Keeper Guidance
- Console shows: `[Consulting world-keeper...]`
- Player never sees the response
- DM integrates information naturally
- No spoilers leaked

---

## Validation Checklist

After deployment, verify:

- [ ] Campaign plan generation completes successfully
- [ ] campaign-plan.json contains valid JSON
- [ ] 3 acts present with meaningful titles
- [ ] session_start triggers world-keeper consultation
- [ ] World-keeper response is hidden from player
- [ ] DM narration is contextual (not generic)
- [ ] Legacy adventures (without campaign plan) still work
- [ ] Campaign tools (get_campaign_plan, etc.) function
- [ ] No errors in logs

---

## Troubleshooting

### Issue: Campaign plan not generated

**Symptoms**: No campaign-plan.json after adventure creation

**Checks**:
```bash
# Check web server logs
tail -50 sw-web.log

# Look for: "campaign plan generation failed"
```

**Solutions**:
1. Verify ANTHROPIC_API_KEY is set correctly
2. Check API rate limits (Anthropic console)
3. Try simpler, clearer theme description
4. Check network connectivity

---

### Issue: World-keeper response visible to player

**Symptoms**: Full world-keeper briefing displayed in console

**Checks**:
```bash
# Check if silent parameter is being used
grep "invoke_agent" data/adventures/*/sw-dm-session-*.log
```

**Solutions**:
1. Verify you're using latest build (rebuild if needed)
2. Check session_tools.go has silent=true for briefing
3. Review DM persona compliance (should not quote directly)

---

### Issue: Session start hangs

**Symptoms**: start_session doesn't complete

**Checks**:
```bash
# Check agent manager logs
grep "InvokeAgentSilent" data/adventures/*/sw-dm-session-*.log
```

**Solutions**:
1. Verify ANTHROPIC_API_KEY is valid
2. Check API timeout (80 seconds configured)
3. Try without campaign plan (legacy adventure)

---

## Performance Expectations

| Operation | Expected Time | Notes |
|-----------|--------------|-------|
| Campaign Plan Generation | 30-60s | Uses Haiku 4.5 |
| Session Start (with plan) | 10-15s | Includes world-keeper call |
| Session Start (without plan) | <1s | Legacy behavior |
| Campaign Tools Query | <100ms | Local file I/O |
| Campaign Tools Update | <500ms | Includes save |

---

## Monitoring

### Key Metrics to Track

**1. Campaign Plan Generation Success Rate**
```bash
# Count successful generations
grep "Generated campaign plan" data/adventures/*/sw-dm-session-*.log | wc -l

# Count failures
grep "campaign plan generation failed" data/adventures/*/sw-dm-session-*.log | wc -l
```

**2. World-Keeper Invocation Frequency**
```bash
# Count silent invocations (briefings)
grep "InvokeAgentSilent" data/adventures/*/sw-dm-session-*.log | wc -l

# Check average duration
grep "silent mode" data/adventures/*/sw-dm-session-*.log
```

**3. Campaign Tool Usage**
```bash
# Count tool invocations
grep -E "get_campaign_plan|update_campaign_progress" data/adventures/*/sw-dm-session-*.log | wc -l
```

---

## Rollback Procedure

If critical issues arise:

### 1. Disable Campaign Plan Generation
```go
// In internal/web/handlers.go, comment out:
// if theme != "" && s.apiKey != "" {
//     s.generateCampaignPlan(adv, theme)
// }
```

### 2. Disable Automated Briefing
```go
// In internal/dmtools/session_tools.go, comment out:
// if campaignPlan != nil && agentManager != nil {
//     // ... briefing logic
// }
```

### 3. Rebuild and Restart
```bash
make
./sw-web --port=8085
./sw-dm
```

**No Data Loss**: Existing adventures continue working normally.

---

## Success Criteria

System is working correctly if:

✅ All unit tests pass (go test ./...)
✅ Campaign plans generate with themes
✅ Session briefings are hidden from players
✅ DM narration is contextually appropriate
✅ Legacy adventures work without errors
✅ Campaign tools modify JSON correctly
✅ No data corruption after updates

---

## Next Steps

1. **Run E2E Tests**: Follow `E2E_TESTING_GUIDE.md`
   - Priority: E2E-3 (World-Keeper Confidentiality)
   - Time: 30-45 minutes total

2. **Create Real Campaign**: Test with actual game sessions
   - Provide detailed theme
   - Play 2-3 sessions
   - Verify briefings improve narration

3. **Monitor Production**: Track metrics above
   - Generation success rate
   - World-keeper confidentiality
   - Tool usage patterns

---

## Support

**Documentation**:
- `IMPLEMENTATION_SUMMARY.md` - Complete implementation details
- `E2E_TESTING_GUIDE.md` - Comprehensive testing procedures
- `CLAUDE.md` - User-facing documentation
- `core_agents/agents/dungeon-master.md` - DM persona instructions

**Logs**:
- `data/adventures/*/sw-dm-session-*.log` - DM agent logs
- `system-prompt.log` - System prompt (including briefings)
- `data/adventures/*/campaign-plan.json` - Campaign state

**Issues**: Report via GitHub repository

---

**Deployment Date**: 2026-01-29
**Version**: 1.0.0
**Status**: Production Ready ✅
