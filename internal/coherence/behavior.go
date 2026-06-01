package coherence

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"dungeons/internal/adventure"
)

// BehaviorProfile is a deterministic, log-derived view of how the DM agent
// actually played: which tools it called, what kinds of dice it rolled, and how
// often it consulted specialist agents. It is evidence for judging agent quality
// (e.g. the "lack of meaningful/social rolls" hypothesis), never a verdict.
type BehaviorProfile struct {
	SessionsAnalyzed   int               `json:"sessions_analyzed"`
	Sessions           []SessionBehavior `json:"sessions"`
	TotalToolCalls     map[string]int    `json:"total_tool_calls"`
	TotalRolls         RollBreakdown     `json:"total_rolls"`
	TotalAgentConsults map[string]int    `json:"total_agent_consults"`
	Notes              []string          `json:"notes"`
}

// SessionBehavior is the per-session behavior breakdown.
type SessionBehavior struct {
	SessionID     int            `json:"session_id"`
	ToolCalls     map[string]int `json:"tool_calls"`
	Rolls         RollBreakdown  `json:"rolls"`
	AgentConsults map[string]int `json:"agent_consults,omitempty"`
}

// RollBreakdown categorizes dice rolls by their narrative purpose.
type RollBreakdown struct {
	Total  int `json:"total"`
	Combat int `json:"combat"`
	Skill  int `json:"skill"`
	Social int `json:"social"`
	Save   int `json:"save"`
	Other  int `json:"other"`
}

var behaviorNotes = []string{
	"Profil dérivé des logs sw-dm ; la catégorisation des jets est heuristique (mots-clés FR sur la description).",
	"Une rencontre de combat = plusieurs jets : le nombre de jets ne reflète pas le nombre de rencontres.",
}

var toolCallRe = regexp.MustCompile(`TOOL CALL: ([a-zA-Z_]+)`)

// rollCategoryKeywords maps a category to the lowercased keywords that identify
// it in a roll's description. Order of evaluation is set by categorizeRoll.
var rollCategoryKeywords = map[string][]string{
	"combat": {"attaque", "attack", "dégât", "degat", "damage", "touche", "initiative", "coup critique"},
	"save":   {"sauvegarde", "jet de save", " save", "résistance au", "jet de résistance"},
	"social": {"persuasion", "intimidation", "tromperie", "deception", "représentation", "performance", "baratin", "diplomat", "persuad", "convaincre", "mentir"},
	"skill":  {"perception", "discrétion", "discretion", "furtiv", "investigation", "fouille", "athlétisme", "athletisme", "acrobat", "escamotage", "arcane", "histoire", "nature", "religion", "dressage", "médecine", "medecine", "survie", "escalade", "intuition", "perspicac", "crochetage"},
}

// categorizeRoll classifies a roll by its description. Combat and saves are checked
// first (unambiguous), then social, then other skills; unknown -> "other".
func categorizeRoll(description string) string {
	d := strings.ToLower(description)
	for _, cat := range []string{"combat", "save", "social", "skill"} {
		for _, kw := range rollCategoryKeywords[cat] {
			if strings.Contains(d, kw) {
				return cat
			}
		}
	}
	return "other"
}

// BuildBehaviorProfile parses an adventure's sw-dm session logs into a profile.
// Missing or unreadable logs yield an empty profile (no error): not every
// adventure has logs.
func BuildBehaviorProfile(adv *adventure.Adventure) (*BehaviorProfile, error) {
	pattern := filepath.Join(adv.BasePath(), "sw-dm-session-*.log*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("finding session logs: %w", err)
	}

	bySession := make(map[int]*SessionBehavior)
	for _, f := range files {
		sid, ok := sessionIDFromLogName(f)
		if !ok {
			continue
		}
		sb, err := parseSessionLogFile(f, sid)
		if err != nil {
			continue // tolerate unreadable/partial logs
		}
		if existing, ok := bySession[sid]; ok {
			mergeSessionBehavior(existing, sb) // rotated logs for the same session
		} else {
			bySession[sid] = sb
		}
	}

	bp := &BehaviorProfile{
		TotalToolCalls:     map[string]int{},
		TotalAgentConsults: map[string]int{},
		Notes:              behaviorNotes,
	}
	sids := make([]int, 0, len(bySession))
	for sid := range bySession {
		sids = append(sids, sid)
	}
	sort.Ints(sids)
	for _, sid := range sids {
		sb := bySession[sid]
		bp.Sessions = append(bp.Sessions, *sb)
		for tool, n := range sb.ToolCalls {
			bp.TotalToolCalls[tool] += n
		}
		for agent, n := range sb.AgentConsults {
			bp.TotalAgentConsults[agent] += n
		}
		bp.TotalRolls.add(sb.Rolls)
	}
	bp.SessionsAnalyzed = len(bp.Sessions)
	return bp, nil
}

func (r *RollBreakdown) add(o RollBreakdown) {
	r.Total += o.Total
	r.Combat += o.Combat
	r.Skill += o.Skill
	r.Social += o.Social
	r.Save += o.Save
	r.Other += o.Other
}

var logNameRe = regexp.MustCompile(`sw-dm-session-(\d+)\.log`)

func sessionIDFromLogName(path string) (int, bool) {
	m := logNameRe.FindStringSubmatch(filepath.Base(path))
	if m == nil {
		return 0, false
	}
	var sid int
	if _, err := fmt.Sscanf(m[1], "%d", &sid); err != nil {
		return 0, false
	}
	return sid, true
}

func parseSessionLogFile(path string, sid int) (*SessionBehavior, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	return parseSessionLog(r, sid), nil
}

// parseSessionLog scans a session log, counting tool calls and categorizing
// roll_dice / invoke_agent calls from their Parameters JSON block.
func parseSessionLog(r io.Reader, sid int) *SessionBehavior {
	sb := &SessionBehavior{
		SessionID:     sid,
		ToolCalls:     map[string]int{},
		AgentConsults: map[string]int{},
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	pendingTool := ""
	collecting := false
	depth := 0
	var jsonBuf strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if m := toolCallRe.FindStringSubmatch(line); m != nil {
			pendingTool = m[1]
			sb.ToolCalls[pendingTool]++
			collecting = false
			continue
		}

		if pendingTool != "" && !collecting && strings.Contains(line, "Parameters:") {
			collecting = true
			depth = 0
			jsonBuf.Reset()
			continue
		}

		if collecting {
			jsonBuf.WriteString(line)
			jsonBuf.WriteByte('\n')
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if depth <= 0 && strings.Contains(jsonBuf.String(), "{") {
				var params map[string]any
				if json.Unmarshal([]byte(jsonBuf.String()), &params) == nil {
					applyToolParams(sb, pendingTool, params)
				}
				collecting = false
				pendingTool = ""
			}
		}
	}
	return sb
}

func applyToolParams(sb *SessionBehavior, tool string, params map[string]any) {
	switch tool {
	case "roll_dice":
		sb.Rolls.Total++
		desc, _ := params["description"].(string)
		switch categorizeRoll(desc) {
		case "combat":
			sb.Rolls.Combat++
		case "skill":
			sb.Rolls.Skill++
		case "social":
			sb.Rolls.Social++
		case "save":
			sb.Rolls.Save++
		default:
			sb.Rolls.Other++
		}
	case "invoke_agent":
		if name, ok := params["agent_name"].(string); ok && name != "" {
			sb.AgentConsults[name]++
		}
	}
}

func mergeSessionBehavior(dst, src *SessionBehavior) {
	for k, v := range src.ToolCalls {
		dst.ToolCalls[k] += v
	}
	for k, v := range src.AgentConsults {
		dst.AgentConsults[k] += v
	}
	dst.Rolls.add(src.Rolls)
}
