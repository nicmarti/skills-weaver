# OpenRouter Migration Plan

Status (updated 2026-09-23): Phases 1–3 implemented; SDK pinned to `v0.8.19` and neutral model selection prepared; Phases 4–12 pending. Phase 0 live evidence remains incomplete (Advisor transcript and genuine provider mid-stream error). See "Decision gates before cutover" below.

Created: 2026-09-22

Scope: migrate the SkillsWeaver game-engine LLM runtime from direct Anthropic API access to OpenRouter. Keep FAL.ai, Google Imagen, Google Lyria, and other media transports unchanged.

The raw investigation transcript is archived in [`ai/MIGRATION_TO_OPENROUTER.md`](../ai/MIGRATION_TO_OPENROUTER.md). This document is the curated source of truth for implementation decisions, findings, tasks, and acceptance criteria.

## Confirmed Decisions

The migration will use the following architecture and product decisions:

- Replace `github.com/anthropics/anthropic-sdk-go` during this migration.
- Use the official OpenRouter Go SDK rather than retaining the Anthropic SDK as a compatibility client.
- Use OpenRouter Chat Completions because the official Go SDK does not currently expose OpenRouter's Anthropic-compatible `/messages` endpoint.
- Introduce provider-neutral messages and a provider-neutral client interface inside SkillsWeaver.
- Default **all OpenRouter game/utility roles to Sonnet 5 for now**; do not preserve Opus/Haiku as mandatory choices. Permit any explicitly selected provider-qualified OpenRouter model per DM or named subagent, including non-Claude providers, subject to required Chat/tool/image capabilities.
- Pin concrete **defaults** instead of mutable `~latest` aliases; an explicit user override may deliberately select a mutable alias.
- Use OpenRouter's native `openrouter:advisor` server tool.
- Configure the Advisor with `forward_transcript: true` for World Keeper.
- Keep `SW_ADVISOR_ENABLED` as the Advisor rollout flag.
- Remove unsupported `advisor_max_uses` and `advisor_caching` persona settings rather than pretending they retain their Anthropic semantics.
- Require `OPENROUTER_API_KEY`; do not silently fall back to `ANTHROPIC_API_KEY`.
- Preserve existing `agent-states.json` files through backward-compatible deserialization.
- Leave FAL.ai, Google Imagen, and Google Lyria generation transports unchanged.

## Target Architecture

Only a new `internal/llm` package should import OpenRouter SDK packages. Agent, tool, persistence, web, and utility packages should consume provider-neutral SkillsWeaver types.

```text
cmd/dm, cmd/web, cmd/adventure, cmd/character-sheet
                         |
                         v
                 internal/llm.Config
                         |
                         v
                 internal/llm.Client
                         |
                         v
          OpenRouter Go SDK / Chat Completions

internal/agent --------------------+
internal/web ----------------------+
internal/ai -----------------------+--> internal/llm.Client
internal/ambient ------------------+
internal/charactersheet -----------+
```

The intended client interface is conceptually:

```go
type Client interface {
	Complete(context.Context, Request) (Response, error)
	Stream(context.Context, Request, StreamObserver) (Response, error)
}
```

The adapter, not the agent loop, owns OpenRouter SDK unions, nullable values, wire conversion, streamed tool-call assembly, error normalization, retries, and usage extraction.

## Implementation checkpoint (2026-09-23)

This is the actionable current state. The investigation below is a **dated baseline from initial planning**, not a second checklist of current runtime behavior.

- `go.mod` declares Go `1.25.10` and now pins **both** OpenRouter Go SDK `v0.8.19` and Anthropic SDK `v1.46.0`. The [v0.8.19 release](https://github.com/OpenRouterTeam/go-sdk/releases/tag/v0.8.19) adds a switchyard-router plugin variant; `go build ./...`, offline adapter tests, and **targeted live Sonnet 5 text/tool/image tests** pass on this pin. Advisor context forwarding and genuine provider mid-stream failures remain unproven. Recheck whether future SDK releases expose a usable `/messages` operation before changing the Chat strategy; a Presets Messages operation is not itself a Messages client.
- `internal/llm` contains neutral types, model mapping, configuration, the Chat adapter, **and a temporary Anthropic tool-wire adapter** (`anthropic_compat.go`). OpenRouter SDK imports remain confined to this package. The actual streaming accumulator is in `internal/llm/openrouter.go`; there is no `openrouter_stream.go`.
- `ConversationContext` stores neutral messages and loads legacy agent state, but `GetMessages()` temporarily converts them back to Anthropic messages for old runtime paths. `ToolRegistry.ToolDefinitions()` exposes complete neutral schemas; the existing DM/nested call loops still call Anthropic. `cmd/dm` and `cmd/web` still require `ANTHROPIC_API_KEY`. Map/journal enrichment, ambient prompt generation, web campaign/title generation and character biographies still use Anthropic; FAL.ai, Imagen and Google Lyria transports are separate and unchanged. An exported `OPENROUTER_API_KEY` does **not** switch gameplay to OpenRouter yet.
- The **current legacy** main DM still honors `core_agents/agents/dungeon-master.md` (`model: opus` → direct Anthropic Opus 4.8), uses a **900,000-token local estimated-history limit**, and requests up to **32,000 output tokens**. The original Sonnet 4.6/16,384-token findings below predate that change. The user has now chosen **OpenRouter Sonnet 5 as the default for the DM and all nested/utility roles**; do not write OpenRouter IDs into the still-live Anthropic model mapping before Phases 4/5 cut over. Nested agents retain a 20,000-token local history limit.
- `internal/llm.Config` now resolves per-agent environment overrides for the three active nested personas and supports a validated `ModelOverrides` layer for future CLI flags. Sonnet 5 is the default for DM, nested, Fast, Campaign and Advisor roles (Advisor stays off pending its gate). **No `sw-dm` CLI flag or OpenRouter model selection affects gameplay yet**: `cmd/dm` still launches the Anthropic runtime until Phases 4/5/8 wire these settings into agents and entrypoints.
- The current web template still hardcodes the `sonnet`/`opus` aliases, displaying **Sonnet 4.6** and **Opus 4.8 (1M)** (`web/templates/game.html`). The investigation's claim that it labels Opus 4.6 is historical; Phase 9 still needs backend-owned labels and a Fast selector option.
- The only active nested personas are Rules Keeper, Character Creator, and World Keeper. `InvokeAgentSilent` serves World Keeper's session-start briefing. There is **no `scenario-critic` persona, `internal/narrativeai` package, four-agent narrative synthesis, or `coherence --ai` CLI path** in this checkout. `sw-adventure coherence` runs deterministic `internal/coherence.Analyze`; do not reintroduce an old optional feature as part of provider migration.
- Offline adapter/legacy-state/tool-schema tests and five opt-in live `internal/llm` checks already exist. The standalone `TestProbeRealOpenRouterAdvisorForwardedWorldContext` did **not** demonstrate context forwarding. A real provider-origin SSE error cannot currently be forced through a documented public API; local SSE fixtures cover the adapter error path.
- Public model endpoint metadata rechecked 2026-09-23: the pinned [Haiku 4.5](https://openrouter.ai/api/v1/models/anthropic/claude-haiku-4.5/endpoints), [Sonnet 5](https://openrouter.ai/api/v1/models/anthropic/claude-sonnet-5/endpoints) and [Opus 5](https://openrouter.ai/api/v1/models/anthropic/claude-opus-5/endpoints) IDs are available. Sonnet 5 endpoints support tools, images, and JSON response formats but do **not** list `temperature`; Haiku 4.5 endpoints vary on `response_format`. Check the routed endpoint for each required parameter, and use `provider.require_parameters: true` when the request depends on it.

## Initial runtime findings (historical planning baseline, 2026-09-22)

### Dependency and Configuration

- `go.mod` directly depends on `github.com/anthropics/anthropic-sdk-go v1.46.0`.
- The project currently declares Go `1.25.0`.
- The selected OpenRouter SDK release, `github.com/OpenRouterTeam/go-sdk v0.8.17`, requires Go `1.25.10` or newer.
- The local development toolchain is Go `1.26.5`.
- Runtime configuration is not centralized. Multiple packages call `os.Getenv` directly.
- `cmd/dm/main.go` and `cmd/web/main.go` require `ANTHROPIC_API_KEY` at startup.
- `cmd/adventure/main.go`, `cmd/advisor-ab/main.go`, `internal/ai/enricher.go`, `internal/agent/register_tools.go`, and `internal/charactersheet/biography.go` also read the Anthropic key independently.
- There are seven separate production `anthropic.NewClient` call sites.
- There is no configurable LLM base URL or provider abstraction.

### Direct Anthropic SDK Coupling

Production imports occur in:

- `internal/agent/agent.go`
- `internal/agent/agent_manager.go`
- `internal/agent/advisor.go`
- `internal/agent/context.go`
- `internal/agent/message_serialization.go`
- `internal/agent/model_mapping.go`
- `internal/agent/streaming.go`
- `internal/agent/tools.go`
- `internal/ai/enricher.go`
- `internal/ambient/prompt_generator.go`
- `internal/charactersheet/biography.go`
- `internal/web/handlers.go`
- `internal/web/wizard_handlers.go`

Test support is also Anthropic-specific through `internal/agent/mock_anthropic.go` and several agent tests.

### Main Agent Call Flow

The CLI and web interfaces share the same `internal/agent.Agent` implementation.

```text
cmd/dm or internal/web/session.go
  -> agent.New
  -> Agent.ProcessUserMessage
  -> buildSystemPrompt
  -> Anthropic Messages.NewStreaming
  -> parse text and tool_use blocks
  -> execute local tools
  -> append tool_result blocks
  -> repeat until no tool call remains
```

Important findings:

- The main client is a concrete `anthropic.Client`, so it cannot be mocked like the nested-agent client.
- The main request uses `context.Background()` and cannot be cancelled when a web client disconnects.
- The main tool loop has no iteration limit.
- The main agent does not record authoritative API token usage.
- The main conversation is in memory only and is not persisted across process restart or web-session expiration.
- The first system prompt is written to `system-prompt.log`.
- Browser streaming is a second SSE layer above provider streaming.
- `internal/web/web_output.go` can drop events when its channel buffer is full.
- Provider errors do not currently guarantee a normal completion event for web SSE consumers.

### Nested Agents

`internal/agent/agent_manager.go` implements:

- `InvokeAgent` with a bounded read-only tool loop.
- `InvokeAgentSilent` with a single non-streaming call.
- Persistent per-agent conversation contexts.
- Per-agent token and latency metrics.
- Optional Anthropic beta Advisor behavior.

Personas listed in the initial investigation (cross-check against the implementation checkpoint above):

| Agent (historical inventory) | Declared model at investigation | Capabilities at investigation |
|---|---|---|
| Character Creator | Sonnet | Filtered read-only tools |
| Rules Keeper | Sonnet | Filtered read-only tools |
| World Keeper | Sonnet + Opus Advisor | Filtered read-only tools and silent briefing |
| Scenario Critic | Sonnet | Historical proposal; not present in the current checkout |

Documentation stating that nested agents have no tools is stale. They currently receive policy-filtered read-only registries.

The automatic session briefing calls World Keeper through `InvokeAgentSilent`. The formerly described four-agent narrative judgment is absent from the current checkout (see the implementation checkpoint above).

### Model inventory at the initial investigation (superseded for the main DM)

| Path (historical inventory) | Model at investigation | Max output | Mode |
|---|---|---:|---|
| Main DM | Sonnet 4.6 | 16,384 | Streaming with tools |
| Nested agents | Usually Sonnet 4.6 | 4,096 | Non-streaming with filtered tools |
| Silent nested agents | Persona model | 4,096 | One non-streaming call |
| World Keeper Advisor | Opus 4.7 | 4,096 outer request | Anthropic beta server tool |
| Campaign plan | Sonnet 4.5 | 8,192 | Non-streaming, optional image |
| Adventure title | Haiku 4.5 | 64 | Non-streaming |
| Journal/map enrichment | Dated Haiku 4.5 | 500 | Non-streaming |
| Ambient prompt | Haiku 4.5 | 300 | Non-streaming |
| Character biography | Claude 3.5 Haiku | 1,000 | Non-streaming |

Inconsistencies recorded at the initial investigation (some have since been fixed):

- `core_agents/agents/dungeon-master.md` declares `model: opus`.
- `internal/agent/agent.go` hard-coded Sonnet 4.6 **at the time of the investigation**; it now honors the persona's Opus setting.
- The web selector labeled Opus as 4.6 **at the time**; it now labels it Opus 4.8 but still hardcodes the two aliases.
- The backend alias `opus` maps to Opus 4.8.
- README model claims do not consistently match runtime behavior.
- A `MaxTokens: 16384` comment incorrectly refers to Haiku despite the Sonnet runtime default.

### Tool Protocol

The repository `Tool` interface is already mostly provider-neutral:

```go
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(params map[string]interface{}) (interface{}, error)
}
```

At the time of the investigation, the coupling occurred when `ToolRegistry` converted tools into Anthropic standard and beta union types. Phase 3 replaced registry conversion with neutral definitions plus a temporary compatibility adapter inside `internal/llm`.

The current conversion reconstructs selected `properties` and `required` fields instead of forwarding the complete schema. This can discard constraints or nested schema information. The OpenRouter conversion should use the complete `InputSchema()` map as Chat function `parameters`.

### Streaming

`internal/agent/streaming.go` depends on:

- `ssestream.Stream[anthropic.MessageStreamEventUnion]`
- `anthropic.Message.Accumulate`
- `ContentBlockDeltaEvent`
- `TextDelta`
- Final Anthropic `ToolUseBlock` extraction

OpenRouter Chat streaming differs in several important ways:

- Text arrives through `choices[].delta.content`.
- Tool function arguments arrive as fragmented JSON strings.
- Parallel tool calls can interleave and must be assembled by index.
- Keepalive SSE comments can occur.
- An error can arrive inside an HTTP 200 stream.
- The final usage chunk repeats the terminal finish reason.
- Usage appears in the final chunk and includes OpenRouter-specific accounting.

### Advisor

Current World Keeper configuration is:

```yaml
model: sonnet
advisor: opus-4.7
advisor_max_uses: 2
advisor_caching: 5m
```

Current behavior uses Anthropic's beta `advisor-tool-2026-03-01` and records separate executor, Advisor, cache-write, and cache-read token counts from `usage.iterations`.

OpenRouter Chat supports the native server tool through this shape:

```json
{
  "type": "openrouter:advisor",
  "parameters": {
    "model": "anthropic/claude-opus-5",
    "forward_transcript": true,
    "max_completion_tokens": 4096
  }
}
```

OpenRouter differences:

- The consultation cap is controlled by OpenRouter rather than `advisor_max_uses`.
- Advisor-specific `advisor_caching` is not supported.
- Chat does not stream the advice separately.
- Chat aggregate usage does not provide the same executor-versus-Advisor iteration breakdown.
- `server_tool_use_details` provides requested and executed server-tool counts.
- `forward_transcript: true` is intended to avoid persisted Advisor-private replay for World Keeper **if actual access to the needed context is established**; the current differential probe did not establish this.
- Advisor and local client tools can be offered in the same request; the focused live contract already confirmed both were invoked in one request. This does not prove transcript forwarding.

### Usage and Cost Accounting

The current main DM has no authoritative provider usage accounting. Nested agents track input/output totals and Anthropic Advisor-specific totals.

OpenRouter Chat usage can provide:

- Prompt tokens
- Completion tokens
- Total tokens
- Cache-read tokens
- Cache-write tokens
- Reasoning tokens
- Aggregate request cost
- Cost details
- BYOK indicator
- Server-tool calls requested
- Server-tool calls executed
- Actual routed model
- Generation ID

OpenRouter Chat cannot preserve these existing metrics exactly:

- Separate Advisor input tokens
- Separate Advisor output tokens
- Separate Advisor cache creation tokens
- Separate Advisor cache read tokens
- Advisor-only cost

Existing Advisor fields in persisted state must remain readable as historical data. New OpenRouter aggregate metrics must use new fields rather than reinterpreting historical fields.

### Prompt Caching

Standard current calls do not enable provider prompt caching. Only the Anthropic Advisor configuration includes cache controls.

Repeated stable prefixes include:

- Main DM persona
- Full tool definitions
- Nested-agent personas
- World map description
- Adventure context
- Campaign directives

OpenRouter supports top-level Chat `cache_control` for Claude and `session_id` for provider stickiness. Initial implementation should use a stable per-adventure/per-agent session identifier and track cache reads/writes; **a session ID is a routing hint, not a cache breakpoint**. Add top-level `cache_control` to the neutral request and adapter only when its SDK serialization and actual cache behavior have been tested on selected Claude 5 endpoints. See [current prompt-caching guidance](https://openrouter.ai/docs/guides/best-practices/prompt-caching).

### Persistence

At the time of the investigation, only nested-agent states were persisted to `data/adventures/<slug>/agent-states.json` (this is still true for the main DM conversation).

The persisted DTO already contains mostly provider-neutral concepts:

- Role
- Text content
- Tool uses
- Tool results
- Token estimate
- Agent metrics

Compatibility findings:

- Existing tool results are persisted as user-role messages containing `tool_results`.
- Chat requires explicit `role: "tool"` messages with `tool_call_id`.
- The loader must therefore dual-read legacy and new forms.
- Historical model values must remain unchanged.
- Missing new OpenRouter fields must default safely.
- Current truncation can split tool calls from tool results.
- Image base64 data is intentionally omitted from persistence.
- World Keeper's map image is created before state restoration, then lost when the restored context replaces the initial context.
- Advisor-private blocks are not currently persisted.

### Retries, Cancellation, and Errors

Current findings:

- No application retry policy is configured.
- The runtime relies on SDK defaults.
- Main streaming has no timeout or cancellation.
- Map enrichment, ambient generation, and biography generation use `context.Background()`.
- Nested calls use one 120-second context for the whole invocation loop.
- Campaign planning uses 120 seconds.
- Title generation uses 30 seconds.
- Errors are mostly wrapped as opaque provider errors.

OpenRouter errors should be normalized into:

- Authentication failure
- Payment or credit exhaustion
- Permission or guardrail rejection
- Invalid request
- Unsupported parameter
- Context length exceeded
- Rate limit
- Provider overload
- Provider unavailable
- Timeout
- Cancellation
- Malformed streamed tool call
- Mid-stream provider error

Streaming requests must not be retried after any text or tool-call delta has been emitted.

### Test Coverage Findings

Existing useful coverage includes:

- Nested-agent invocation and reuse
- Recursion and invalid-agent handling
- Tool-access policy filtering
- Model mapping
- Advisor configuration and beta routing
- Some state save/load behavior
- UI Markdown streaming renderer
- CLI readline behavior

Important gaps:

- No provider-level test for the main streaming API path
- No main-agent streamed tool loop test
- No test for fragmented streamed tool arguments
- No interleaved parallel tool-call test
- No provider-to-browser SSE integration test
- No direct web-session model-call test
- No complete CLI production-agent test
- No full legacy message persistence round trip
- No explicit tool-call/result truncation-boundary test
- No test proving World Keeper map reinjection after state restoration
- No tests for ambient prompt generation
- No tests for biography AI generation and fallback
- No checked-in CI workflow running optional real-provider tests

## OpenRouter Compatibility Findings

### API Surfaces

OpenRouter exposes three relevant APIs:

| Surface | Compatibility | Decision |
|---|---|---|
| Anthropic Messages `/api/v1/messages` | Closest to current runtime | Not exposed by official Go SDK |
| Chat Completions `/api/v1/chat/completions` | Complete SDK support, normalized tools and streaming | Selected |
| Responses `/api/v1/responses` | Rich server-tool model, larger rewrite, stateless | Not selected |

OpenRouter's HTTP API supports Anthropic Messages, but the official Go SDK `v0.8.17` has no Messages service, Messages response model, or Messages SSE client. It contains some orphan generated request components but cannot perform the operation end to end.

The SDK does fully expose Chat and Responses streaming operations. The official SDK is beta and should be pinned to a reviewed version.

### Authentication

OpenRouter uses:

```text
Authorization: Bearer $OPENROUTER_API_KEY
Base URL: https://openrouter.ai/api/v1
```

Optional attribution headers are:

```text
HTTP-Referer: <application URL>
X-OpenRouter-Title: SkillsWeaver
```

### Message Conversion

| Internal semantic | OpenRouter Chat wire form |
|---|---|
| System prompt | System message |
| User text | User string content |
| User text plus image | User content parts with text and image URL |
| Base64 image | `data:<media-type>;base64,<payload>` URL |
| Assistant text | Assistant content |
| Assistant tool request | Assistant `tool_calls` |
| Tool input | JSON string in `function.arguments` |
| Tool result | Tool message with matching `tool_call_id` |

`ToolResultMessage.IsError` has no direct Chat field. SkillsWeaver should preserve it internally and on disk, while encoding failure information in the JSON tool-result content sent to the model.

### Streaming Requirements

The stream adapter must:

1. Check each chunk for an embedded error.
2. Forward non-empty text deltas immediately.
3. Assemble tool calls by streamed index.
4. Append argument fragments in arrival order.
5. Support interleaved parallel calls.
6. Reject incomplete calls at stream end.
7. Parse completed arguments once.
8. Capture final usage.
9. Record the response ID and actual model.
10. Treat `length`, `content_filter`, and `error` as incomplete responses.
11. Avoid retrying after observable output.
12. Close the stream on every exit path.

### Finish Reasons

Expected Chat finish reasons are:

- `stop`
- `tool_calls`
- `length`
- `content_filter`
- `error`

The agent loop must not infer completion solely from the absence of tool calls.

### Routing and Fallbacks

OpenRouter normally performs same-model provider fallback. Initial migration should not add cross-model fallbacks because that would combine provider migration, model upgrade, and behavioral fallback changes.

Requests using tools, images, structured output, or other required controls should use `provider.require_parameters: true` so unsupported parameters are not silently ignored.

Data collection, ZDR, provider ordering, and cross-model fallbacks should be configuration options rather than hard-coded without a product decision.

## Target Model Catalog

Model availability must be rechecked against OpenRouter's live catalog immediately before cutover. The original Claude-only role mapping from 2026-09-22 was superseded by the user's 2026-09-23 direction: **Sonnet 5 for every default role, with arbitrary provider-qualified OpenRouter model IDs allowed as explicit CLI/environment overrides**.

| Logical role | Current default OpenRouter model ID |
|---|---|
| Fast utility role | `anthropic/claude-sonnet-5` |
| Main DM default | `anthropic/claude-sonnet-5` |
| Nested-agent default | `anthropic/claude-sonnet-5` |
| Named nested-agent override (unset) | Inherit nested-agent default |
| Advisor (disabled by default) | `anthropic/claude-sonnet-5` until explicitly configured and Phase 6 verified |
| Campaign generation | `anthropic/claude-sonnet-5` |
| Title and other utility generation | `anthropic/claude-sonnet-5` |

Haiku 4.5 and Opus 5 remain **optional** named aliases/curated selector choices; neither is a forced runtime default. Explicit overrides may use any OpenRouter `provider/model` ID (e.g. other vendors) or a deliberately chosen `~` latest-resolution alias. Capability support, context/output budgets and price must be checked per selected model/endpoint; `provider.require_parameters: true` should prevent silently dropping necessary tools/images/structured-output controls. Do not silently fall back to a different model when an override is invalid or unsupported.

No automatic cross-model fallback or automatic mapping between model families is part of the first cutover.

## Configuration Plan

Required variable:

```text
OPENROUTER_API_KEY
```

Optional variables:

```text
OPENROUTER_MODEL_DM
OPENROUTER_MODEL_NESTED
OPENROUTER_MODEL_RULES_KEEPER
OPENROUTER_MODEL_CHARACTER_CREATOR
OPENROUTER_MODEL_WORLD_KEEPER
OPENROUTER_MODEL_FAST
OPENROUTER_MODEL_CAMPAIGN
OPENROUTER_MODEL_ADVISOR
OPENROUTER_HTTP_REFERER
OPENROUTER_APP_NAME
```

Existing feature flag retained:

```text
SW_ADVISOR_ENABLED
```

Configuration requirements:

- One config loader validates environment variables.
- For DM: explicit CLI `--model` > `OPENROUTER_MODEL_DM` > Sonnet 5. For a named subagent: explicit CLI `--agent-model name=model` > CLI `--nested-model` > matching `OPENROUTER_MODEL_<NAME>` > `OPENROUTER_MODEL_NESTED` > Sonnet 5. `internal/llm.Config.WithModelOverrides` implements this precedence for the neutral config; CLI parsing and active runtime wiring remain Phase 4/5/8 work.
- Reject unknown bare aliases and unknown subagent names; accept provider-qualified IDs from any OpenRouter provider. Log the resolved requested and actual routed model without logging credentials. An unavailable/incompatible model must produce an actionable error, not silently substitute Sonnet 5.
- One shared OpenRouter client is created per process.
- Packages receive dependencies through constructors.
- Utility packages do not read API keys independently.
- `ANTHROPIC_API_KEY` is not accepted as a fallback.
- Secrets are never logged or persisted.

Planned `sw-dm` invocation **after Phases 4/5/8 (not supported by the current Anthropic-based binary)**:

```bash
OPENROUTER_MODEL_DM=anthropic/claude-sonnet-5 \
OPENROUTER_MODEL_RULES_KEEPER=google/gemini-3-pro \
./sw-dm --model openai/gpt-6 --nested-model anthropic/claude-sonnet-5 \
  --agent-model rules-keeper=google/gemini-3-pro
```

CLI flags are deliberately not introduced while the REPL still sends requests to Anthropic; accepting and then ignoring them (or passing OpenRouter IDs to Anthropic) would mislead players. The `--agent-model` flag must be repeatable for all three named agents. Test CLI-over-env precedence, model capability failures, and effective-model reporting before enabling it for play.

## Implementation Tasks

### Phase 0: Contract Spike

- [x] Recheck the latest stable tagged OpenRouter Go SDK version.
- [x] Recheck the live model catalog for Haiku 4.5, Sonnet 5, and Opus 5.
- [x] Confirm each configured model supports the required Chat parameters.
- [x] Confirm Sonnet 5 supports client tools through OpenRouter Chat.
- [x] Confirm Sonnet 5 accepts image input through OpenRouter Chat.
- [x] Confirm streamed text works through the official Go SDK.
- [x] Confirm streamed tool-call arguments are complete and ordered.
- [x] Confirm parallel local tool calls work.
- [x] Confirm `openrouter:advisor` works with Sonnet 5 and Opus 5.
- [x] Confirm native Advisor and local client functions can coexist.
- [ ] Confirm `forward_transcript: true` passes World Keeper context.
- [x] Capture aggregate Advisor usage and cost behavior.
- [ ] Confirm provider errors and mid-stream errors are exposed by the SDK.
- [x] Record the observed SDK and API behavior in tests before production wiring.

Original Phase 0 exit criterion (not met): all selected models and required API behaviors have live evidence, or the gate is explicitly revised. The user requested Phases 2–3 despite the outstanding checks. Phase 4 fake-client implementation can be planned independently, but **do not interpret this as approval of the proposed real mid-stream-error gate change or as proof of Advisor context forwarding**. Resolve the "Decision gates before cutover" below before an OpenRouter-only release.

Static verifications (2026-09-22): the latest stable SDK tag is `v0.8.17`; the live catalog confirms `anthropic/claude-haiku-4.5`, `anthropic/claude-sonnet-5`, `anthropic/claude-opus-5` (and `anthropic/claude-sonnet-4.6`) exist with `tools`, image input, and `structured_outputs`/`response_format` support. Notable: Sonnet 5 does not list `temperature` in `supported_parameters` (Opus 5 and Haiku 4.5 do) — requests that require temperature must use `provider.require_parameters` or omit it. Paid behavior probes are opt-in and were run after the user attempted the Phase 0 live command.

Live checks (2026-09-22): `internal/llm/openrouter_live_test.go` is opt-in (`RUN_REAL_API_TESTS=1` and `OPENROUTER_API_KEY`). Run with `RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestRealOpenRouter' -v -count=1`. Sonnet 5 streamed text and two client function calls with complete JSON arguments; a tool-result follow-up succeeded; a generated PNG was accepted; Sonnet 5 invoked Opus 5 Advisor while a local function was available, and on a second check invoked the Advisor and requested the local function in the same turn. Live aggregate usage included cost and server-tool counts (three consultations on one run, one on another). A nonexistent model produced a normalized HTTP 400 invalid-request error. Live SSE chunk fragmentation was not independently observed; interleaved fragmented tool-call deltas and mid-stream errors are covered by local HTTP fixtures. `forward_transcript: true` is verified in the outgoing wire payload, but delivery of World Keeper's actual context to the Advisor is not yet independently established. No mid-stream provider error was induced in a live request. These remaining items keep Phase 0 open.

Further investigation (2026-09-22): run the separate diagnostic with `RUN_ADVISOR_FORWARD_PROBE=1 RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestProbeRealOpenRouterAdvisorForwardedWorldContext$' -v -count=1`. The synthetic World Keeper map ledger contains a unique sigil in a prior user message; the current turn and Advisor instructions contain no sigil. The executor is told to consult the Advisor and repeat its answer. Three `forward_transcript: true` live attempts returned `UNKNOWN` despite 2–3 reported server-tool executions; the `false` control also returned `UNKNOWN` in both runs where it was reached. An initial version placed the sigil in the system prompt; moving it to the prior user message (as with World Keeper's map) did not change the result. This is **not proof of successful forwarding** and does not isolate whether OpenRouter forwarded the conversation but the models ignored it, or the Advisor never received it. The diagnostic is intentionally excluded from the passing `TestRealOpenRouter` suite until the behavior is understood.

The documented production Chat API does not expose a deterministic switch to produce an authentic provider mid-stream failure. OpenRouter documents a `X-Simulate-Mid-Stream-Error` header only for a **local fake-provider service**, not the public API. A local SSE fixture already verifies the SDK/adapter rejects a 200-with-error chunk after partial output; it must not be reported as a real provider failure. Completion of this Phase 0 requirement needs an authentic OpenRouter incident captured with a generation ID, or a provider-operated test endpoint/error-injection mechanism. References: [Advisor transcript semantics](https://openrouter.ai/docs/guides/features/server-tools/advisor), [streaming error semantics](https://openrouter.ai/docs/api_reference/streaming), [provider error format](https://openrouter.ai/docs/api_reference/errors-and-debugging).

### Phase 1: Provider-Neutral LLM Package

- [x] Add `internal/llm/types.go`.
- [x] Add `internal/llm/client.go`.
- [x] Add `internal/llm/config.go`.
- [x] Add `internal/llm/models.go`.
- [x] Add `internal/llm/errors.go`.
- [x] Add `internal/llm/openrouter.go`.
- [x] Add `internal/llm/openrouter_messages.go`.
- [x] Implement SSE and streamed tool-call assembly in `internal/llm/openrouter.go` (no separate `openrouter_stream.go` exists).
- [x] Define neutral text, image, assistant tool-call, and tool-result messages.
- [x] Define neutral requests, responses, finish reasons, and usage.
- [x] Build complete Chat message conversion.
- [x] Build full tool JSON Schema conversion.
- [x] Add actual routed-model and generation-ID capture.
- [x] Add bounded retry configuration.
- [x] Add context cancellation.
- [x] Pin the OpenRouter SDK dependency.
- [ ] Remove the Anthropic SDK only after all callers compile against the new boundary (**Phase 12**, not a prerequisite for the completed Phase 1 boundary).
- [x] Raise the Go module requirement to at least 1.25.10 if the pinned SDK still requires it.

Exit criterion: the new package passes unit, wire-format, streaming, retry, and cancellation tests without importing it into the agent runtime yet.

### Phase 2: Conversation and Persistence

- [x] Rewrite `internal/agent/context.go` to store neutral messages.
- [x] Keep tool arguments as `json.RawMessage` until execution (Anthropic callers use a temporary wire conversion).
- [x] Represent tool results as internal tool-role messages.
- [x] Preserve `IsError` internally and on disk.
- [x] Rewrite `internal/agent/message_serialization.go` around neutral messages.
- [x] Preserve the existing serialized JSON fields.
- [x] Add optional `schema_version` and treat absence as legacy version 1.
- [x] Dual-read legacy user-role tool results and new tool-role messages.
- [x] Preserve historical model and Advisor metric values.
- [x] Add exchange-aware truncation.
- [x] Remove orphaned tool exchanges atomically.
- [x] Add lightweight image resource references where required.
- [x] Re-inject the World Keeper map after restored history is loaded.
- [x] Add a sanitized legacy `agent-states.json` fixture.

Exit criterion (verify with offline tests, no real-provider calls):

1. `ConversationContext` stores `llm.Message` values; only a temporary Anthropic wire adapter converts them for unmigrated Phase 4/5 callers. Tool-call arguments remain `json.RawMessage` in memory and survive disk round trips as JSON objects (not escaped JSON strings). Tool results use role `tool` internally and retain call ID and `IsError`.
2. A sanitized pre-migration `agent-states.json` fixture without `schema_version` loads with its original text, complete multi-tool exchange, historical model, and historical Advisor metrics. On save it emits schema version 2 and explicit tool-role result messages without breaking old field names. Reloading that output retains the same values. No adventure data files need migration.
3. Live-context truncation and disk-token-budget trimming never leave dangling tool calls or unmatched tool results, including multi-tool calls, an oversized latest exchange, and an in-progress exchange receiving results one at a time. Legacy incomplete/orphaned exchanges are removed as a unit on load.
4. No base64 world-map image is persisted; a stable image resource reference survives save/load and the actual map image is re-injected at the beginning of World Keeper history after state restoration, exactly once. Other agents do not receive the world map.
5. `go build ./...`, focused `internal/agent` persistence tests, and `make test` pass while the current Anthropic runtime remains functional until Phases 4/5 replace the temporary wire adapter.

User-directed sequencing (2026-09-22): Phase 2 was started at the user's request despite the two open Phase 0 live-evidence items. This does not imply that the Advisor transcript or genuine provider mid-stream error is verified. The temporary `ConversationContext.GetMessages()` conversion keeps the current Anthropic call paths running until Phase 4/5 cutover; the OpenRouter adapter skips references whose image bytes are unavailable instead of sending invalid image URLs.

### Phase 3: Tool Definitions

- [x] Replace `ToAnthropicTools` with provider-neutral definitions.
- [x] Consolidate standard and beta schema extraction (retain thin legacy wire wrappers until Phases 4/6).
- [x] Forward complete `InputSchema()` maps.
- [x] Do not enable strict schema mode initially.
- [x] Keep filtered nested-agent registries unchanged.
- [x] Verify every registered tool converts to a valid Chat function definition.
- [x] Verify tool results remain valid JSON on success and failure.

Exit criterion (offline, including a production-style registry constructed without making provider requests):

1. `ToolRegistry` exposes deterministically ordered `llm.ToolDefinition` values with name, description and the full original `InputSchema()` map. The OpenRouter Chat adapter emits them as non-strict function tools, preserving nested properties, array items, enums, constraints and both `[]string` / `[]interface{}` required lists.
2. All registered tools have JSON-serializable object schemas and valid function names. The existing filtered read-only nested registries contain only their allowed tools; optional map and image tools are included in the registry check using synthetic environment keys, without invoking providers.
3. Legacy standard and beta Anthropic call paths continue to compile through a single temporary wire-compatibility module. Both convert from the neutral definitions, preserving complete schema constraints; the original duplicate schema extraction is removed. The compat module is removed when Phases 4–6 retire those call paths.
4. Main and nested tool successes and failures produce valid JSON even when error strings contain quotes, backslashes or newlines or a tool returns an unsupported value; `IsError` remains correct. `internal/dmtools/` and the neutral registry in `internal/agent/tools.go` import no provider SDK; the legacy agent loops retain Anthropic imports until Phases 4/5.
5. `go build ./...`, focused tool/adapter tests and `make test` pass. No provider API call or TypeSafe service is needed for schema conversion.

TypeSafe.ai evaluation (2026-09-22): the [TypeSafe System One building guide](https://docs.typesafe.ai/concepts/how-to-build-with-system-one) reserves the model for semantic judgments and explicitly keeps exact rules and transformations in code. Phase 3 only copies and validates known JSON Schema and encodes known tool results, so adding a judgment call would introduce uncertainty into an exact conversion. No TypeSafe integration is needed here.

## Decision gates before cutover (reviewed 2026-09-23)

| Gate | Decision / evidence needed | When it blocks |
|---|---|---|
| Main DM default — **resolved 2026-09-23** | Use **Sonnet 5 by default for the DM and every role**. Retain optional user-selected OpenRouter models via CLI/environment; existing `model: opus` belongs to the unmigrated Anthropic runtime only. Do not pretend the current executable already applies the new defaults. | Apply during Phase 4 DM and Phase 5 nested-agent wiring; finish model selector in Phase 9. |
| World Keeper Advisor | The live `forward_transcript: true` probe returned `UNKNOWN` both with and without forwarding. It does not establish whether the Advisor saw the map/history. Keep `SW_ADVISOR_ENABLED` **off** on the OpenRouter path until a reproducible context-bearing contract test passes or an explicitly approved alternative supplies the necessary context. Sonnet 5 is the configurable initial Advisor default, not a mandatory Opus model; counted server-tool executions alone are insufficient. | Phase 6 enablement and final Advisor acceptance, not fake-client Phase 4 work. |
| Real provider mid-stream failure | No documented public fault-injection endpoint makes a genuine provider error reproducible. **Proposed gate change, awaiting approval:** use the HTTP-200/SSE-error-after-text adapter fixture plus a live authenticated stream, record real generation errors if observed, and require an actual provider failure only if OpenRouter supplies a test endpoint or an incident trace. Never label an injected fixture a real incident. | Production release acceptance; do not demand an uninducible error before writing Phase 4 tests. |
| Cutover sequencing | Phase 4/5 can be implemented and fake-tested while old web/utility paths still use Anthropic, but that is **not an OpenRouter-only production cutover**. Wire a shared `llm.Client` through DM/web session construction and map UI aliases at the same checkpoint as main-agent migration; finish the remaining utility callers and entrypoints in Phases 7–9 before removing the Anthropic key. Run live games only after the coherent cutover is built and tested. | Phase 4/5 design and Phase 8/12 release. |

### Phase 4: Main DM

- [ ] Inject `llm.Client` and model catalog into `agent.Agent`.
- [ ] Replace the concrete Anthropic client.
- [ ] Rename `callAnthropicAPI` to a provider-neutral operation.
- [ ] Send the system prompt as a Chat system message.
- [ ] Use the **already neutral** `ConversationContext.NeutralMessages()` directly; remove the main loop's temporary `GetMessages()` Anthropic conversion.
- [ ] Migrate the streamed local-tool loop.
- [ ] Add a minimal shared-client construction path into CLI and web session creation (final environment/utility wiring remains in Phase 8), and translate current web model aliases before they reach the neutral model catalog (finish UI catalog work in Phase 9).
- [ ] Wire `llm.Config.ModelDM` as the Sonnet 5 default, with CLI `--model` overriding `OPENROUTER_MODEL_DM`; validate/reject unknown IDs before gameplay. Reconcile the 900,000-token **estimated** history budget and 32,000 output tokens with the chosen model's actual context/output limits; reserve room for the changing system prompt, tools, images, reasoning and output. Do not assume a local character-count estimate equals provider tokens.
- [ ] Extend `llm.Request`/Chat conversion for `provider.require_parameters: true` on tool calls (and on later image/structured-output requests); test the outgoing wire payload. Sonnet 5 does not list `temperature` support, so omit it for the DM unless endpoint support is proven.
- [ ] Add a maximum tool-loop iteration count.
- [ ] Handle every finish reason explicitly.
- [ ] Aggregate per-API-call and per-turn usage/cost across tool-loop iterations, including cache/read/write/reasoning when present; record requested model, actual model, generation ID, duration and finish reason without mistaking absent cost for zero cost.
- [ ] Use a stable, bounded per-adventure/per-agent `session_id` for provider stickiness; **session_id alone does not activate Claude prompt caching**. Add an explicit `cache_control` option only after adapter wire coverage and measured cache-read/write behavior on the selected endpoint.
- [ ] Thread `context.Context` through `ProcessUserMessage`.
- [ ] Cancel generation on CLI shutdown and on web-turn cancellation; the web POST returns **before** the model finishes, so do not bind the turn lifetime solely to the POST request context. Define SSE disconnect/reconnect and session-expiration semantics explicitly.
- [ ] Preserve terminal rendering and fix browser SSE loss/termination paths: `WebOutput.sendEvent` currently drops text when its channel is full, and an agent error currently emits `error` without a terminal `complete`/`done` event. Verify a slow subscriber and both success/failure streams.
- [ ] On error/cancel/`length`/`content_filter`/`error`, do not commit partial assistant/tool-call history or retry after observable output. Preserve already executed tool side effects and avoid automatically replaying that turn. Keep automatic `start_session` and its hidden briefing before generation; journal events must not fall into session 0.
- [ ] Align Dungeon Master persona and selector with the confirmed Sonnet 5 default on the migrated runtime, while retaining explicit user overrides for any provider-qualified OpenRouter model.

Exit criterion: a fake `llm.Client` proves a streamed main turn, parallel/serial tool calls with matching IDs and JSON results, a normal finish, and aggregated usage; fault tests prove bounded iteration, cancellation, explicit incomplete finish handling, no replay after partial output, automatic session start, and complete/lossless web SSE. `go build ./...`, focused tests and `make test` pass. This is a **main-loop checkpoint**, not an OpenRouter-only runtime claim while nested/web/utility callers remain on Anthropic.

### Phase 5: Nested Agents

- [ ] Replace `anthropicKey` and Anthropic client factories with `llm.Client` injection.
- [ ] Remove standard-versus-beta Messages service abstractions.
- [ ] Replace `internal/agent/mock_anthropic.go` with a test-only neutral fake as nested-agent tests are migrated (do not keep a production Anthropic mock until Phase 10).
- [ ] Migrate `InvokeAgent` to neutral Chat messages and complete filtered `ToolDefinitions()`; preserve call IDs, JSON tool results and errors.
- [ ] Migrate `InvokeAgentSilent` to a single neutral Chat call without client tools (Advisor remains **off** until Phase 6).
- [ ] Resolve each nested model from the shared `llm.Config.ModelForAgent(name)` using CLI > environment > Sonnet 5 precedence; do not hardcode `MapPersonaModelToAnthropic` or reuse the DM selection for every agent.
- [ ] Decide and test whether an explicit persona `model:` value is honored **below** CLI/environment overrides or becomes metadata only; do not let old persona settings silently override the user's chosen agent model. Current nested personas all declare `sonnet`.
- [ ] Enforce required model capabilities for read-only tools and World Keeper's image. Define a tested text-only-map fallback or a clear unsupported-model error; do not silently route to another model.
- [ ] Preserve recursion limits.
- [ ] Preserve filtered read-only tools.
- [ ] Preserve per-agent iteration limits.
- [ ] Preserve the 120-second invocation timeout unless contract tests justify a change.
- [ ] Record requested and actual routed models plus aggregate OpenRouter usage/cost across nested tool loops in **new optional metrics fields**, keeping historical Anthropic Advisor metrics unchanged on load/save.
- [ ] Verify session-start World Keeper hidden briefing and restored image/legacy history on the neutral client. Currently `start_session` **silently skips** a failed World Keeper call while still starting the session; keep that best-effort session behavior, but add a safe diagnostic so a missing confidential briefing is observable without leaking it to players.
- [ ] Confirm `sw-adventure coherence` still uses deterministic `internal/coherence.Analyze` and `end_session` still closes the journal; do not add the absent `scenario-critic`/narrative synthesis pipeline.

Exit criterion: Rules Keeper, Character Creator and World Keeper work through a fake shared `llm.Client` with read-only tools, valid iteration limits, silent briefing, timeout/error handling and persisted legacy/new conversations. `SW_ADVISOR_ENABLED` remains off on this path. No `scenario-critic` test or Anthropic client factory remains in active nested runtime.

### Phase 6: Native OpenRouter Advisor

- [ ] Replace Anthropic beta Advisor types with `openrouter:advisor` configuration.
- [ ] Use `llm.Config.ModelAdvisor` (Sonnet 5 default, arbitrary explicit model override) instead of a fixed Opus ID; this optional server tool stays disabled until context transfer is proven.
- [ ] Set `forward_transcript: true`.
- [ ] Set a bounded Advisor output budget.
- [ ] Keep `SW_ADVISOR_ENABLED` as the rollout flag.
- [ ] Remove `advisor_max_uses` from persona metadata and World Keeper frontmatter.
- [ ] Remove `advisor_caching` from persona metadata and World Keeper frontmatter.
- [ ] Keep existing historical Advisor metrics readable.
- [ ] Add generic server-tool requested/executed metrics.
- [ ] Do not claim aggregate usage is Advisor-only usage.
- [ ] Keep Advisor failures non-fatal when OpenRouter returns a tool-level error.
- [ ] Update World Keeper's instructions for OpenRouter Advisor behavior.
- [ ] Update `cmd/advisor-ab` to use OpenRouter aggregate cost and server-tool metrics.
- [ ] Resolve the `forward_transcript` decision gate: test Advisor access to a prior World Keeper map/history fact **with a negative control**, distinguish an executed consultation from demonstrated context transfer, and keep the flag off if this remains unproven.

Exit criterion: offline fake-client tests cover flag-off/on, server-tool-level failure and historical metric preservation; focused opt-in live tests prove Advisor alone, Advisor plus local functions, and **actual context transfer** (or a separately approved context-supplying design). A passing `server_tool_use_details` count by itself does not satisfy this gate. Enable only after these checks.

### Phase 7: Direct Utility Calls

- [ ] Inject `llm.Client` into `internal/ai.Enricher`.
- [ ] Migrate journal enrichment to the configurable Fast role (Sonnet 5 default).
- [ ] Migrate map-prompt enrichment to the configurable Fast role (Sonnet 5 default).
- [ ] Inject `llm.Client` into ambient prompt generation.
- [ ] Migrate only Lyria parameter generation; keep Google Lyria transport unchanged.
- [ ] Inject an optional `llm.Client` into biography generation.
- [ ] Preserve template fallback when biography AI generation is unavailable.
- [ ] Migrate campaign-plan generation to the Campaign model.
- [ ] Preserve optional world-map image input.
- [ ] Migrate adventure-title suggestions to the configurable Fast role (Sonnet 5 default).
- [ ] Preserve the best-effort empty-title fallback.
- [ ] Add a provider-neutral optional `response_format`/JSON-schema control to `llm.Request` and test the SDK wire mapping before using structured output for JSON-producing calls. Require parameter support per **routed endpoint** (`provider.require_parameters: true`); Haiku 4.5 endpoints vary in `response_format` support. Preserve validated JSON parsing and existing fallbacks when structured output is unavailable.

Exit criterion: all journal/map enrichment, ambient-prompt, biography, campaign-plan and title paths use the injected shared neutral client or documented optional template fallback. Tests cover multimodal campaign requests and each utility's JSON/error behavior. Image generation (FAL.ai/Imagen) and Google Lyria transport are unchanged.

### Phase 8: Entrypoints and Dependency Injection

- [ ] Update `cmd/dm/main.go` to require `OPENROUTER_API_KEY`.
- [ ] Implement `sw-dm --model <provider/model>`, `--nested-model <provider/model>` and repeatable `--agent-model <rules-keeper|character-creator|world-keeper>=<provider/model>`; apply `llm.Config.WithModelOverrides` after the environment. Reject missing names/IDs and unknown agents before starting a session; show effective requested models without logging keys.
- [ ] Update `cmd/web/main.go` to require `OPENROUTER_API_KEY`.
- [ ] Update `cmd/adventure/main.go` **enrichment** path; preserve the existing deterministic `coherence` command (there is no `coherence --ai` path).
- [ ] Update `cmd/advisor-ab/main.go`.
- [ ] Update `cmd/character-sheet/main.go` optional AI path.
- [ ] Update `internal/web/server.go` configuration.
- [ ] Update `internal/web/session.go` session construction.
- [ ] Update `internal/agent/register_tools.go` dependency injection.
- [ ] Update `internal/dmtools/map_tool.go` construction.
- [ ] Update `internal/dmtools/ambient_tool.go` construction.
- [ ] Remove package-level API-key environment reads outside the config loader.
- [ ] Ensure one OpenRouter client is reused per process (including sessions and optional utility tools); do not create one per request or per nested agent.

Exit criterion: CLI, web and auxiliary entrypoints accept `OPENROUTER_API_KEY` via `llm.LoadConfig`, create a shared client and do not accept `ANTHROPIC_API_KEY` as fallback. No production caller creates an Anthropic client or reads its key. The only Anthropic SDK adapter left before final removal should be dead/test-only; remove it in Phase 12. No unintended call to unrelated provider transports occurs during construction.

### Phase 9: Model Selection

- [ ] Replace remaining Anthropic runtime constants with OpenRouter model IDs; the neutral catalog and Sonnet 5 role defaults are already prepared.
- [ ] Preserve `haiku`, `sonnet`, and `opus` persona aliases.
- [x] Accept provider-qualified OpenRouter IDs and validate environment/CLI-override values in neutral `llm.Config`; preserve aliases (`haiku`, `sonnet`, `opus`) as optional conveniences.
- [ ] Apply model override validation in the CLI and web selector: do not silently map mistyped values to Sonnet 5 (`llm.ResolveModel` still has a legacy unknown-bare-word fallback); decide whether to preserve that fallback for historical persona metadata only.
- [ ] Centralize selectable models and display names.
- [ ] Apply Sonnet 5 as the main DM and nested-agent **runtime** default, retaining explicit CLI/env overrides for any compatible provider/model ID.
- [ ] Keep Haiku 4.5 and Opus 5 as **optional** curated web-selector choices, not required default roles.
- [ ] Replace substring-based model detection in web handlers.
- [ ] Render the web selector from backend model metadata.
- [ ] Return stable IDs and display labels from model endpoints.
- [ ] Confirm longer labels fit the existing UI.
- [ ] Log the actual routed model, not only the requested alias.

Exit criterion: CLI, web, personas, logs, metrics and persisted request defaults agree on model identity. Tests cover all three aliases, pinned IDs, unknown inputs, frontmatter, backend selector endpoints and the actually routed model; the approved DM default decision is applied consistently.

### Phase 10: Tests and Mocks

- [x] Add non-streaming OpenRouter adapter tests using `httptest.Server`.
- [x] Add authentication and attribution-header tests.
- [x] Add system, user, assistant, image, tool-call, and tool-result conversion tests.
- [x] Add complete tool-schema preservation tests.
- [x] Add SSE keepalive, streamed text, fragmented/interleaved parallel tool-call tests.
- [x] Add final usage and duplicate finish-reason tests.
- [x] Add injected mid-stream HTTP-200/SSE error fixture tests (**not** a genuine provider incident).
- [x] Add basic stream cancellation and one transient-503 retry test.
- [ ] Prove stream closure/cancellation on **all** success and error paths, exhaustion of bounded retries, and no retry after emitted text/tool delta.
- [ ] Assert 401/402 are not retried with the **production** retry configuration (existing error-classification tests use retry disabled).
- [ ] Add main-agent tool-loop tests.
- [ ] Add nested-agent tool-loop tests.
- [ ] Add silent-agent tests.
- [ ] Add Advisor disabled and enabled tests.
- [ ] Add Advisor-plus-local-tools tests.
- [x] Add legacy state fixture and exchange-safe truncation tests.
- [x] Add World Keeper map restoration tests.
- [ ] Add campaign multimodal request tests.
- [ ] Add ambient prompt tests.
- [ ] Add biography AI and fallback tests.
- [ ] Add web-turn SSE slow-subscriber, error completion, and disconnect-cancellation tests; add CLI model selection/session start/end tests.
- [x] Gate `internal/llm` paid OpenRouter tests behind `OPENROUTER_API_KEY` and `RUN_REAL_API_TESTS=1`.
- [x] Test Sonnet 5 neutral defaults, named nested-agent env overrides, CLI-over-env config precedence and malformed model choices (no provider call).
- [ ] Test actual `sw-dm` flag parsing, requested/routed model per agent, non-Claude tool/image incompatibility and no silent cross-model fallback once Phases 4/5/8 wire the runtime.
- [ ] Retire/gate legacy Anthropic real tests in `internal/agent` before running any **repository-wide** paid test command; while both keys are present, run focused `RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestRealOpenRouter' -count=1` only.

Exit criterion: existing transport/persistence coverage remains passing, and the remaining main/nested/web/utility paths gain fake-client and integration tests **alongside the phases implementing them**, not as a late Phase 10 catch-up. Do not remove the Anthropic dependency before tests cover every replacement path.

### Phase 11: Documentation

- [ ] Update active runtime sections in `README.md`.
- [ ] Update `DEPLOYMENT.md`.
- [ ] Update remaining pre-migration runtime sections in `AGENTS.md`; `CLAUDE.md` is only a pointer and has no runtime sections to migrate.
- [ ] Update `docs/optional-features-summary.md`.
- [ ] Update `core_agents/skills/map-generator/SKILL.md`.
- [ ] Update `core_agents/skills/journal-illustrator/SKILL.md`.
- [ ] Update `internal/ui/README.md`.
- [ ] Add a new `CHANGELOG.md` entry.
- [ ] Keep historical changelog descriptions unchanged.
- [ ] Keep Claude Code references that describe the development tool rather than the game runtime.
- [ ] Document OpenRouter key limits and recommend a spending cap.
- [ ] Document model overrides and concrete production defaults.
- [ ] Document Advisor metric limitations.

Exit criterion: setup, deployment, troubleshooting, and runtime architecture documentation consistently describe OpenRouter.

### Phase 12: Final Removal and Validation

- [ ] Remove `github.com/anthropics/anthropic-sdk-go` from `go.mod` and `go.sum`.
- [ ] Remove all production imports of the Anthropic SDK.
- [ ] Remove all production reads of `ANTHROPIC_API_KEY`.
- [ ] Remove the temporary `internal/llm/anthropic_compat.go`, `ConversationContext.GetMessages()` Anthropic conversion, legacy `internal/agent/streaming.go`/Advisor paths, Anthropic model mapping and mock types once no caller needs them.
- [ ] Run `gofmt` on modified Go files.
- [ ] Run `go build ./...` after every Go edit batch, as required by repository policy.
- [ ] Run targeted `internal/llm` tests.
- [ ] Run targeted `internal/agent` tests.
- [ ] Run `go test ./...`.
- [ ] Run `make test`.
- [ ] Build `sw-dm`.
- [ ] Build `sw-web`.
- [ ] Build `sw-adventure`.
- [ ] Build `sw-character-sheet`.
- [ ] Run focused gated real OpenRouter contract tests (never `RUN_REAL_API_TESTS=1 go test ./...` while legacy Anthropic tests/keys are still present).
- [ ] Run one same-version reference adventure through the migrated engine.
- [ ] Verify FAL.ai, Google Imagen, and Google Lyria transports are unchanged.

Exit criterion: all automated checks and focused available live contracts pass; the reference-adventure evidence report satisfies the acceptance criteria below. Advisor enablement requires its **separate** Phase 6 context-transfer gate. A genuine provider mid-stream incident cannot be a deterministic test without an OpenRouter-controlled fixture/endpoint; obtain the explicit decision in "Decision gates before cutover" before marking the final release gate complete.

## Affected File Inventory

### New Core Package

| Proposed path | Responsibility |
|---|---|
| `internal/llm/types.go` | Neutral messages, requests, responses, and usage |
| `internal/llm/client.go` | Injectable client and stream observer interfaces |
| `internal/llm/config.go` | Environment and runtime configuration |
| `internal/llm/models.go` | Concrete models, aliases, and selectable catalog |
| `internal/llm/errors.go` | Provider-neutral errors |
| `internal/llm/openrouter.go` | Official SDK adapter, SSE consumption and tool-call assembly (already implemented) |
| `internal/llm/openrouter_messages.go` | Chat wire conversion |
| `internal/llm/anthropic_compat.go` | Temporary standard/beta Anthropic tool-wire conversion; remove in Phase 12 |

### Core Agent Runtime

| Existing path | Planned change |
|---|---|
| `internal/agent/agent.go` | Neutral client, Chat tool loop, context, metrics |
| `internal/agent/agent_manager.go` | Neutral nested and silent agent calls |
| `internal/agent/advisor.go` | OpenRouter native Advisor configuration |
| `internal/agent/context.go` | Neutral conversation storage |
| `internal/agent/streaming.go` | Retain output contract; move wire parsing to LLM adapter |
| `internal/agent/tools.go` | Neutral complete tool definitions |
| `internal/agent/message_serialization.go` | Neutral DTO conversion and legacy loading |
| `internal/agent/model_mapping.go` | Compatibility wrappers or replacement by model catalog |
| `internal/agent/agent_state.go` | New aggregate metrics and backward compatibility |
| `internal/agent/persona_loader.go` | Provider-neutral model and Advisor metadata |
| `internal/agent/register_tools.go` | Inject LLM dependencies |
| `internal/agent/mock_anthropic.go` | Remove and replace with test-only fake |

### Utilities and Web

| Existing path | Planned change |
|---|---|
| `internal/ai/enricher.go` | Inject neutral client and Fast model |
| `internal/ambient/prompt_generator.go` | Inject neutral client; leave Lyria transport unchanged |
| `internal/charactersheet/biography.go` | Optional neutral client and template fallback |
| `internal/web/server.go` | Shared LLM config and client |
| `internal/web/session.go` | Shared client per session manager |
| `internal/web/handlers.go` | Campaign generation and model endpoints |
| `internal/web/wizard_handlers.go` | Adventure-title generation |
| `internal/dmtools/map_tool.go` | Inject Enricher/client |
| `internal/dmtools/ambient_tool.go` | Inject prompt generator/client |
| `web/templates/game.html` | Catalog-driven model selector |
| `web/static/js/app.js` | Stable model IDs |

### Entrypoints and Build

| Existing path | Planned change |
|---|---|
| `cmd/dm/main.go` | OpenRouter configuration and cancellation |
| `cmd/web/main.go` | OpenRouter configuration |
| `cmd/adventure/main.go` | Anthropic-backed journal enrichment; preserve deterministic `coherence` unchanged |
| `cmd/advisor-ab/main.go` | OpenRouter Advisor and aggregate metrics |
| `cmd/character-sheet/main.go` | Optional OpenRouter biography client |
| `go.mod` | SDK replacement and Go version |
| `go.sum` | Dependency checksums |
| `Makefile` | New package dependency edges if required |

### Personas and Documentation

| Existing path | Planned change |
|---|---|
| `core_agents/agents/dungeon-master.md` | Align declared default model |
| `core_agents/agents/world-keeper.md` | OpenRouter Advisor semantics |
| `README.md` | OpenRouter setup and model documentation |
| `DEPLOYMENT.md` | OpenRouter deployment and troubleshooting |
| `AGENTS.md` (formerly CLAUDE.md) | Active runtime architecture only; CLAUDE.md is now a pointer |
| `docs/optional-features-summary.md` | New client and Advisor behavior |
| `core_agents/skills/map-generator/SKILL.md` | OpenRouter enrichment requirements |
| `core_agents/skills/journal-illustrator/SKILL.md` | OpenRouter enrichment requirements |
| `CHANGELOG.md` | New migration entry only |

## Persistence Compatibility Rules

The first OpenRouter release must follow these rules:

1. Existing `agent-states.json` files remain readable.
2. Existing serialized JSON field names remain supported.
3. A missing schema version means legacy version 1.
4. Historical model values are not rewritten.
5. Historical Advisor metrics are not reinterpreted.
6. New metrics use optional fields and safe zero defaults.
7. Legacy user-role tool results convert into Chat tool messages on the wire.
8. New tool-role messages can be saved and loaded.
9. Tool calls and all matching results are retained or dropped atomically during truncation.
10. Base64 image data remains outside persisted state.
11. World resources are re-injected after state restoration.
12. Adventure journals, campaigns, characters, inventories, sessions, and biographies require no data migration.

## Retry and Cancellation Policy

The adapter must use explicit bounded behavior rather than SDK defaults:

- Bound total retry time to seconds, not the SDK's potentially long default window.
- Retry transient connection failures and eligible 5xx/provider-overload responses only before observable streaming output.
- Do not retry authentication, permission, validation, ordinary payment, or unsupported-parameter failures.
- Treat rate-limit handling conservatively until SDK header access is verified.
- Let OpenRouter same-model provider fallback run before multiplying retries in application code.
- Honor context cancellation.
- Never replay a stream after text or tool deltas have reached the user.
- Distinguish cancellation from provider failure in terminal and web output.

## Acceptance Criteria

### Static Acceptance

- No production import of `github.com/anthropics/anthropic-sdk-go` remains.
- No production read of `ANTHROPIC_API_KEY` remains.
- OpenRouter SDK usage is isolated to `internal/llm`.
- Model IDs are centralized.
- No utility package creates its own provider client.
- FAL.ai and Google media transport code is unchanged except constructor wiring required to inject prompt-generation dependencies.

### Build and Test Acceptance

- `go build ./...` passes.
- `go test ./...` passes.
- `make test` passes.
- `sw-dm`, `sw-web`, `sw-adventure`, and `sw-character-sheet` build.
- Legacy state fixtures load and round-trip.
- Streaming fixtures cover text, tools, parallel calls, usage, errors, and cancellation.
- Paid contract tests pass when explicitly enabled.

### Runtime Acceptance

- The main DM streams narration through OpenRouter.
- The main DM can execute multiple local tools and continue generation.
- Rules Keeper can consult its read-only tools.
- Character Creator can consult its read-only tools.
- World Keeper produces a hidden session briefing.
- World Keeper can consult native OpenRouter Advisor when enabled.
- Advisor failure does not corrupt the nested-agent conversation.
- The existing deterministic `sw-adventure coherence` report works; no absent Scenario Critic or AI narrative synthesis is introduced.
- Campaign generation accepts the world-map image and saves valid JSON.
- Adventure-title suggestion remains best effort.
- Map enrichment works before unchanged image generation.
- Ambient prompt generation works before unchanged Google Lyria generation.
- Biography generation falls back to templates when OpenRouter is unavailable.
- Web model selection reports the requested and actual model correctly.
- Existing nested-agent state resumes without orphaned tool messages.

### Observability Acceptance

- Logs include requested model, actual model, generation ID, duration, and finish reason.
- Main and nested calls record input, output, cache, reasoning, and total tokens when available.
- OpenRouter aggregate cost is recorded when available.
- Server-tool request and execution counts are recorded.
- Logs never include API keys.

## Explicit Non-Goals

- Do not migrate FAL.ai image generation.
- Do not migrate Google Imagen.
- Do not migrate Google Lyria audio generation.
- Do not replace unrelated media SDKs.
- Do not require arbitrary provider IDs in the curated web model catalog; **CLI/environment overrides may use any compatible OpenRouter model**, including non-Claude models.
- Do not add cross-model fallback in the first cutover.
- Do not rewrite historical changelog entries.
- Do not rewrite Claude Code references that concern development tooling rather than game-runtime inference.
- Do not migrate adventure, character, journal, inventory, or campaign data formats beyond backward-compatible agent-state handling.

## Known Tradeoffs and Risks

- The official OpenRouter Go SDK is beta and generated APIs can change between releases.
- The SDK does not currently expose OpenRouter's Anthropic-compatible Messages endpoint.
- Chat migration changes message, tool-result, and streaming semantics.
- Model upgrades and provider migration happen together, increasing regression scope.
- OpenRouter Advisor cannot reproduce Anthropic's separate Advisor token and cache metrics.
- `advisor_max_uses` and `advisor_caching` cannot retain their old meaning.
- Provider routing can change latency or behavior even for a fixed model ID.
- Structured output and parameter support must be checked against eligible provider endpoints.
- Defaults remain pinned for reproducibility; a user who deliberately selects a mutable `~latest` alias accepts that the resolved model may change between sessions.
- A live Advisor contract test is mandatory before production enablement.
- The now-current **Anthropic** DM uses Opus 4.8 and ~900K estimated history; the user chose Sonnet 5 as the migrated default, with explicit per-role overrides. Estimation does not guarantee a request fits the selected provider's context window, especially for non-Claude models.
- The Chat adapter currently has no neutral `provider.require_parameters`, `response_format`, or `cache_control` options; Phases 4/7 must add and test only the controls their selected endpoints require.
- The production web SSE channel can drop deltas and lacks a terminal event after errors. A mock main-loop test alone is insufficient to claim the browser receives complete streamed output.

## Rollout Strategy

1. **Done:** Phases 1–3 established the neutral adapter, conversation/persistence compatibility and tool definitions; keep the old runtime working while replacements land.
2. **Before/with Phase 4:** apply the confirmed Sonnet 5 default with `--model`/`OPENROUTER_MODEL_DM`, validate required Chat controls and provide shared-client constructor/temporary entrypoint wiring plus backend model aliases. Fake-test main streaming, game-session auto-start and web SSE; do not claim an OpenRouter-only game yet.
3. **Phase 5:** migrate the three actual nested agents and hidden World Keeper briefing with Advisor **off**, and connect named/per-agent model overrides; keep deterministic coherence unchanged.
4. **Phase 7 then 8:** migrate every remaining direct Anthropic utility, then finish `OPENROUTER_API_KEY`-only process configuration and shared-client wiring across CLI/web/auxiliary binaries. The transition can remain mixed-provider during development but not at release.
5. **Phase 9:** finish web model catalog/labels and reconcile selected, requested and routed IDs with the approved default (the backend alias subset is needed already in step 2).
6. **Phase 6 + 10:** implement native Advisor behind the off-by-default flag and finish missing contract tests. Enable Advisor only after the World Keeper context-transfer gate passes or a different approved design demonstrably supplies context. Paid tests stay package-scoped while legacy Anthropic tests exist.
7. **Phases 11–12:** update active docs, remove old SDK/key/mocks and temporary adapters, run the full offline gate and focused OpenRouter live contracts, then play a same-version reference adventure with the stable migrated configuration. If Advisor is not enabled, report it as pending rather than treating the enabled-Advisor runtime criterion as passed.

Each phase that edits Go files must run `go build ./...` immediately, in accordance with repository policy.

## Official References

- [OpenRouter authentication](https://openrouter.ai/docs/api_reference/authentication)
- [OpenRouter Go SDK](https://openrouter.ai/docs/client-sdks/go)
- [OpenRouter Go SDK repository](https://github.com/OpenRouterTeam/go-sdk)
- [Chat Completions API](https://openrouter.ai/docs/api/api-reference/chat/create-a-chat-completion)
- [Anthropic Messages API](https://openrouter.ai/docs/api/api-reference/anthropic-messages/create-a-message)
- [Responses API](https://openrouter.ai/docs/api/api-reference/responses/create-a-response)
- [Client tool calling](https://openrouter.ai/docs/guides/features/tool-calling)
- [Advisor server tool](https://openrouter.ai/docs/guides/features/server-tools/advisor)
- [Prompt caching](https://openrouter.ai/docs/guides/best-practices/prompt-caching)
- [Provider routing](https://openrouter.ai/docs/guides/routing/provider-selection)
- [Model fallbacks](https://openrouter.ai/docs/guides/routing/model-fallbacks)
- [Live model catalog](https://openrouter.ai/api/v1/models)

## Progress Log

Implementation updates should be appended here with the date, completed phase, verification commands, evidence, and any approved deviation from this plan.

| Date | Phase | Result | Evidence or notes |
|---|---|---|---|
| 2026-09-22 | Planning and research | Complete | Architecture, model, SDK, Advisor, persistence, and test decisions recorded above |
| 2026-09-22 | Phase 0 (static) | Complete | Latest stable SDK tag: `v0.8.17`. Live catalog: `anthropic/claude-haiku-4.5`, `anthropic/claude-sonnet-5`, `anthropic/claude-opus-5` all present, supporting tools, image input, and structured outputs. Sonnet 5 does not list `temperature` in supported parameters — omit temperature or set `provider.require_parameters`. Live (paid) Phase 0 probes pending user go-ahead |
| 2026-09-22 | Phase 1 | Complete | `internal/llm` landed: neutral types/messages/usage/finish reasons, `Client` interface, env config with per-role defaults, model catalog with aliases, normalized errors, OpenRouter adapter (Chat wire conversion, Complete, SSE stream assembly with interleaved tool-call accumulation, bounded retry, cancellation). 22 offline tests pass. `go.mod`: pinned `github.com/OpenRouterTeam/go-sdk v0.8.17`, Go directive raised to `1.25.10`. SDK findings recorded: chat SSE wire format is standard `data: {chunk}` (no `{"data":...}` envelope); `WithHTTPReferer`/`WithXTitle` globals are not propagated by the generated Chat operation, so attribution headers are injected via a transport-level HTTP client in the adapter. Full `go build ./...` and `go test ./...` pass; the two failing `internal/dmtools` map tests are pre-existing and depend on `ANTHROPIC_API_KEY` (will be resolved in Phase 7) |
| 2026-09-22 | Phase 0 (live, partial) | Five opt-in live tests pass | `go build ./...`; `go test ./internal/llm -count=1`; `RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestRealOpenRouter' -v -count=1 -timeout=12m`; targeted rerun of `TestRealOpenRouterAdvisorWithLocalTools` after tightening its assertion; `make test` (all packages pass with the keys loaded in this shell). Streaming, two parallel tools and their result follow-up, image, Advisor plus local tool, usage/cost, and invalid-model classification verified; World Keeper transcript delivery and live mid-stream errors remain unproven. |
| 2026-09-22 | Phase 0 (live rerun) | Full gated suite passes | User reran `RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestRealOpenRouter' -v -count=1 -timeout=12m` after the Advisor test was tightened: all five live tests passed (19.957s), including Advisor plus local tool (`finish=tool_calls`, server tools executed/requested `1/1`). This confirms the complete current suite passes together; the two remaining Phase 0 evidence gaps above are unchanged. |
| 2026-09-22 | Phase 0 (additional investigation) | Still blocked | Paid synthetic World Keeper map-ledger probe returned `UNKNOWN` on three forwarding-enabled attempts and two disabled controls; a standalone opt-in diagnostic preserves the reproducer without silently passing the contract gate. No documented public method for forcing a genuine provider mid-stream failure; existing mid-stream fixture is simulated. `go build ./...` and `make test` pass. See detailed findings above. |
| 2026-09-22 | Phase 2 | Implemented | Neutral conversation history with a temporary Anthropic conversion for unmigrated callers; JSON-raw tool arguments and tool-role results; backward-compatible schema-v2 save/load, atomic multi-tool truncation/cleanup, and World Keeper map-resource restoration. Sanitized v1 fixture covers multi-tool history, historical metrics and model names; tests cover oversized/in-progress exchanges, incomplete/orphan cleanup, image omission/reinjection and missing legacy token limits. `go build ./...`, focused `internal/agent` and `internal/llm` tests, and `make test` passed (final verification below). User explicitly requested this phase while Phase 0's Advisor-forwarding and provider mid-stream evidence remain unresolved. |
| 2026-09-22 | Phase 3 | Implemented | `ToolRegistry.ToolDefinitions` produces sorted provider-neutral definitions and validates complete object schemas. Thin standard/beta Anthropic adapters in `internal/llm/anthropic_compat.go` share schema extraction and preserve extension keywords through `ExtraFields`; OpenRouter Chat receives the entire schema without `strict`. Offline tests check the full registered registry (including optional image, map and skill tools), filtered World Keeper tools, nested/array/enum/constraint fidelity on both provider formats, malformed schemas, and escaped JSON success/failure outputs in main and nested tool paths. `go build ./...`, targeted agent/LLM tests and `make test` pass. TypeSafe.ai was assessed and not used: schema transformation is deterministic. Phase 0 live evidence remains outstanding. |
| 2026-09-23 | Plan review before Phase 4 | Updated, with decisions pending at the time | Audited current Go runtime, personas, entrypoints, `internal/coherence`, old/new tests and docs; rechecked the three model endpoints and OpenRouter caching/routing guidance. Corrected the stale Sonnet/16K snapshot, missing streaming file, absent Scenario Critic/`coherence --ai` path, duplicate/premature Phase 10 tasks and CLAUDE.md pointer. Latest Go SDK release was `v0.8.19`; the dependency was still pinned to `v0.8.17` at review time. DM default and release handling of an uninducible genuine mid-stream error required decisions. This was a documentation audit; no provider calls or gameplay tests were run. |
| 2026-09-23 | SDK/model-selection preparation | Implemented for neutral config; CLI/runtime pending | User selected Sonnet 5 defaults for DM, all three nested agents, Fast, Campaign and optional Advisor roles, with arbitrary OpenRouter provider/model IDs as explicit overrides. Upgraded `github.com/OpenRouterTeam/go-sdk` from `v0.8.17` to `v0.8.19` (`go 1.25.10` unchanged); `go mod tidy`, `go build ./...`, `go test ./internal/llm -count=1`, and `make test` passed. Targeted gated live v0.8.19 checks passed for Sonnet 5 text streaming, parallel local tools with a follow-up, and image input (`RUN_REAL_API_TESTS=1 go test ./internal/llm -run '^TestRealOpenRouter(StreamText|StreamParallelTools|Image)$' -v -count=1 -timeout=5m`); Advisor was not retested. Added `OPENROUTER_MODEL_NESTED` and named agent environment overrides, strict model-choice validation, `ModelForAgent` and CLI-over-env `WithModelOverrides` with offline tests. Old Anthropic gameplay is intentionally unchanged: `sw-dm` CLI flags and active OpenRouter model selection are Phase 4/5/8 tasks. |
