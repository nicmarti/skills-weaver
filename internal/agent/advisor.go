// Package agent — Advisor tool integration for nested agents.
//
// The Anthropic Advisor tool (beta) lets a nested agent's executor model
// consult a stronger advisor model mid-generation for strategic guidance. It is
// a server-side tool: it resolves within a single Beta Messages call (the
// response already contains both the advisor_tool_result block and the final
// text), so no client-side tool loop is required.
//
// Integration is gated behind the SW_ADVISOR_ENABLED feature flag (off by
// default) and configured per-persona via the `advisor:` frontmatter field.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// defaultAdvisorMaxUses bounds advisor calls per request when a persona does not
// specify advisor_max_uses. Two covers an early planning call plus one mid-task
// course correction, matching the Phase 0 spike.
const defaultAdvisorMaxUses = 2

// advisorFeatureEnabled reports whether the Advisor tool integration is enabled
// for this process. Controlled by SW_ADVISOR_ENABLED ("1", "true", "yes" → on).
// Off by default so the feature can ship dark and be flipped without recompiling.
func advisorFeatureEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SW_ADVISOR_ENABLED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// resolveAdvisorConfig derives the advisor model and max-uses for a nested
// agent from its persona metadata, validating the executor/advisor pair.
// Returns ok=false (advisor disabled) when the feature flag is off, no advisor
// is configured, the advisor model is unrecognized, or the pair is invalid.
func resolveAdvisorConfig(metadata *PersonaMetadata, executor anthropic.Model) (model anthropic.Model, maxUses int, ok bool) {
	if !advisorFeatureEnabled() {
		return "", 0, false
	}
	if metadata == nil || strings.TrimSpace(metadata.Advisor) == "" {
		return "", 0, false
	}
	advisor, recognized := MapAdvisorModelToAnthropic(metadata.Advisor)
	if !recognized || !IsValidAdvisorPair(executor, advisor) {
		return "", 0, false
	}
	maxUses = metadata.AdvisorMaxUses
	if maxUses <= 0 {
		maxUses = defaultAdvisorMaxUses
	}
	return advisor, maxUses, true
}

// toBetaMessages converts standard MessageParam values into BetaMessageParam
// values via a JSON round-trip. The wire format for the block types used by
// nested agents (text, image, tool_use, tool_result) is identical between the
// standard and beta APIs, so this is lossless for our conversations.
func toBetaMessages(messages []anthropic.MessageParam) ([]anthropic.BetaMessageParam, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}
	var beta []anthropic.BetaMessageParam
	if err := json.Unmarshal(data, &beta); err != nil {
		return nil, fmt.Errorf("unmarshal beta messages: %w", err)
	}
	return beta, nil
}

// advisorCallResult holds the parsed outcome of a single advisor-enabled beta call.
type advisorCallResult struct {
	text          string    // executor's text this turn
	toolUses      []ToolUse // client-side tool calls requested by the executor
	advisorText   string    // advice returned by the advisor (empty if not consulted)
	advisorErr    string    // advisor error_code (empty if none)
	advisorCalled bool
	execInTokens  int64
	execOutTokens int64
	advInTokens   int64
	advOutTokens  int64
	advisorModel  string
}

// advisorToolParam builds the Advisor tool definition for a nested agent.
func advisorToolParam(nestedAgent *NestedAgentState) anthropic.BetaToolUnionParam {
	return anthropic.BetaToolUnionParam{OfAdvisorTool20260301: &anthropic.BetaAdvisorTool20260301Param{
		Model:   nestedAgent.advisorModel,
		MaxUses: anthropic.Int(int64(nestedAgent.advisorMaxUses)),
	}}
}

// doBetaCall performs a single beta Messages call with the given tools, parsing
// the response into text, client-side tool uses, advisor advice, and a per-
// iteration token breakdown. The Advisor tool (if present in tools) resolves
// server-side within this one call; client-side tool uses are returned for the
// caller to execute and loop.
func (am *AgentManager) doBetaCall(ctx context.Context, nestedAgent *NestedAgentState, systemPrompt string, tools []anthropic.BetaToolUnionParam) (advisorCallResult, error) {
	var res advisorCallResult

	betaMessages, err := toBetaMessages(nestedAgent.conversationCtx.GetMessages())
	if err != nil {
		return res, err
	}

	msg, err := nestedAgent.client.GetMessages().NewBeta(ctx, anthropic.BetaMessageNewParams{
		Model:     nestedAgent.model,
		MaxTokens: 4096,
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaAdvisorTool2026_03_01},
		System:    []anthropic.BetaTextBlockParam{{Text: systemPrompt}},
		Tools:     tools,
		Messages:  betaMessages,
	})
	if err != nil {
		return res, err
	}

	for _, block := range msg.Content {
		switch b := block.AsAny().(type) {
		case anthropic.BetaTextBlock:
			res.text += b.Text
		case anthropic.BetaToolUseBlock:
			input := map[string]interface{}{}
			if raw, mErr := json.Marshal(b.Input); mErr == nil {
				_ = json.Unmarshal(raw, &input)
			}
			res.toolUses = append(res.toolUses, ToolUse{ID: b.ID, Name: b.Name, Input: input})
		case anthropic.BetaAdvisorToolResultBlock:
			res.advisorCalled = true
			// The content union flattens variant fields; discriminate on Type.
			switch b.Content.Type {
			case "advisor_result":
				res.advisorText = b.Content.Text
			case "advisor_redacted_result":
				res.advisorText = "[encrypted advice]"
			default: // advisor_tool_result_error
				res.advisorErr = string(b.Content.ErrorCode)
			}
		}
	}

	// Per-iteration token breakdown: executor vs advisor (billed separately).
	for _, it := range msg.Usage.Iterations {
		switch v := it.AsAny().(type) {
		case anthropic.BetaMessageIterationUsage:
			res.execInTokens += v.InputTokens
			res.execOutTokens += v.OutputTokens
		case anthropic.BetaAdvisorMessageIterationUsage:
			res.advInTokens += v.InputTokens
			res.advOutTokens += v.OutputTokens
			res.advisorModel = string(v.Model)
		}
	}
	// Fallback when iterations are absent (e.g. advisor not consulted): use the
	// top-level executor usage so metrics stay populated.
	if res.execInTokens == 0 && res.execOutTokens == 0 {
		res.execInTokens = msg.Usage.InputTokens
		res.execOutTokens = msg.Usage.OutputTokens
	}

	return res, nil
}

// callWithAdvisor performs a single beta Messages call with ONLY the Advisor
// tool enabled (no client-side tools). Used by the silent briefing path, where
// the advisor resolves server-side within this one call.
func (am *AgentManager) callWithAdvisor(ctx context.Context, nestedAgent *NestedAgentState, systemPrompt string) (advisorCallResult, error) {
	return am.doBetaCall(ctx, nestedAgent, systemPrompt, []anthropic.BetaToolUnionParam{advisorToolParam(nestedAgent)})
}

// recordAdvisorMetrics folds an advisor call's token usage into the agent's
// metrics, tracking advisor tokens separately from executor tokens.
func recordAdvisorMetrics(m *AgentMetrics, res advisorCallResult, duration time.Duration) {
	execTokens := res.execInTokens + res.execOutTokens
	m.TotalTokensUsed += execTokens
	m.TotalInputTokens += res.execInTokens
	m.TotalOutputTokens += res.execOutTokens
	m.TotalResponseTime += duration
	m.LastCallTokens = execTokens
	m.LastCallDuration = duration

	if res.advisorCalled {
		m.AdvisorCalls++
		m.AdvisorInputTokens += res.advInTokens
		m.AdvisorOutputTokens += res.advOutTokens
		if res.advisorModel != "" {
			m.AdvisorModelUsed = res.advisorModel
		}
	}
}
