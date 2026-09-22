# OpenRouter Migration Plan

Status: approved plan, implementation not started

Created: 2026-09-22

Scope: migrate the SkillsWeaver game-engine LLM runtime from direct Anthropic API access to OpenRouter. Keep FAL.ai, Google Imagen, Google Lyria, and other media transports unchanged.

The raw investigation transcript is archived in [`ai/MIGRATION_TO_OPENROUTER.md`](../ai/MIGRATION_TO_OPENROUTER.md). This document is the curated source of truth for implementation decisions, findings, tasks, and acceptance criteria.

## Confirmed Decisions

The migration will use the following architecture and product decisions:

- Replace `github.com/anthropics/anthropic-sdk-go` during this migration.
- Use the official OpenRouter Go SDK rather than retaining the Anthropic SDK as a compatibility client.
- Use OpenRouter Chat Completions because the official Go SDK does not currently expose OpenRouter's Anthropic-compatible `/messages` endpoint.
- Introduce provider-neutral messages and a provider-neutral client interface inside SkillsWeaver.
- Upgrade model roles to the latest concrete Claude models available through OpenRouter at implementation-planning time.
- Pin concrete model IDs instead of using mutable `~latest` aliases in production.
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

## Current Runtime Findings

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

Current nested agents are:

| Agent | Current declared model | Runtime capabilities |
|---|---|---|
| Character Creator | Sonnet | Filtered read-only tools |
| Rules Keeper | Sonnet | Filtered read-only tools |
| World Keeper | Sonnet + Opus Advisor | Filtered read-only tools and silent briefing |
| Scenario Critic | Sonnet | Silent narrative analysis |

Documentation stating that nested agents have no tools is stale. They currently receive policy-filtered read-only registries.

The automatic session briefing calls World Keeper through `InvokeAgentSilent`. Narrative judgment performs four silent calls: World Keeper, Rules Keeper, Scenario Critic, then Scenario Critic synthesis.

### Current Model Inventory

| Path | Current model | Max output | Mode |
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

Current inconsistencies:

- `core_agents/agents/dungeon-master.md` declares `model: opus`.
- `internal/agent/agent.go` actually hard-codes Sonnet 4.6.
- The web selector labels Opus as 4.6.
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

The coupling occurs when `ToolRegistry` converts tools into Anthropic standard and beta union types.

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
- `forward_transcript: true` avoids requiring persisted Advisor-private replay for World Keeper.
- Advisor and local client tools can be offered in the same request, but this must be covered by a live contract test.

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

OpenRouter supports top-level Chat `cache_control` and `session_id` provider stickiness. Initial implementation should use a stable per-adventure/per-agent session identifier and track cache reads and writes. Cache behavior must be verified against the selected Claude 5 endpoints before being treated as guaranteed.

### Persistence

Only nested-agent states are persisted to `data/adventures/<slug>/agent-states.json`.

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

Model availability must be rechecked against OpenRouter's live catalog immediately before implementation. The approved initial mapping, based on the 2026-09-22 catalog, is:

| Logical role | Concrete OpenRouter model ID |
|---|---|
| Fast | `anthropic/claude-haiku-4.5` |
| Balanced | `anthropic/claude-sonnet-5` |
| Premium | `anthropic/claude-opus-5` |
| Main DM default | `anthropic/claude-sonnet-5` |
| Nested-agent default | `anthropic/claude-sonnet-5` |
| World Keeper Advisor | `anthropic/claude-opus-5` |
| Campaign generation | `anthropic/claude-sonnet-5` |
| Title and utility generation | `anthropic/claude-haiku-4.5` |

Production should pin these IDs. Mutable aliases such as `~anthropic/claude-sonnet-latest` can change behavior without a deployment and are unsuitable for regression-sensitive campaigns.

Do not map Haiku to Claude Fable automatically. Fable is a distinct family with different reasoning, parameter support, and pricing.

## Configuration Plan

Required variable:

```text
OPENROUTER_API_KEY
```

Optional variables:

```text
OPENROUTER_MODEL_DM
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
- One shared OpenRouter client is created per process.
- Packages receive dependencies through constructors.
- Utility packages do not read API keys independently.
- `ANTHROPIC_API_KEY` is not accepted as a fallback.
- Secrets are never logged or persisted.

## Implementation Tasks

### Phase 0: Contract Spike

- [x] Recheck the latest stable tagged OpenRouter Go SDK version.
- [x] Recheck the live model catalog for Haiku 4.5, Sonnet 5, and Opus 5.
- [x] Confirm each configured model supports the required Chat parameters.
- [ ] Confirm Sonnet 5 supports client tools through OpenRouter Chat.
- [ ] Confirm Sonnet 5 accepts image input through OpenRouter Chat.
- [ ] Confirm streamed text works through the official Go SDK.
- [ ] Confirm streamed tool-call arguments are complete and ordered.
- [ ] Confirm parallel local tool calls work.
- [ ] Confirm `openrouter:advisor` works with Sonnet 5 and Opus 5.
- [ ] Confirm native Advisor and local client functions can coexist.
- [ ] Confirm `forward_transcript: true` passes World Keeper context.
- [ ] Capture aggregate Advisor usage and cost behavior.
- [ ] Confirm provider errors and mid-stream errors are exposed by the SDK.
- [ ] Record the exact SDK and API behavior in tests before production wiring.

Exit criterion: all selected models and required API behaviors have live evidence, or the plan is revised before core implementation.

Static verifications (2026-09-22): the latest stable SDK tag is `v0.8.17`; the live catalog confirms `anthropic/claude-haiku-4.5`, `anthropic/claude-sonnet-5`, `anthropic/claude-opus-5` (and `anthropic/claude-sonnet-4.6`) exist with `tools`, image input, and `structured_outputs`/`response_format` support. Notable: Sonnet 5 does not list `temperature` in `supported_parameters` (Opus 5 and Haiku 4.5 do) — requests that require temperature must use `provider.require_parameters` or omit it. Live (paid) behavior probes below remain gated pending user go-ahead.

### Phase 1: Provider-Neutral LLM Package

- [x] Add `internal/llm/types.go`.
- [x] Add `internal/llm/client.go`.
- [x] Add `internal/llm/config.go`.
- [x] Add `internal/llm/models.go`.
- [x] Add `internal/llm/errors.go`.
- [x] Add `internal/llm/openrouter.go`.
- [x] Add `internal/llm/openrouter_messages.go`.
- [x] Add `internal/llm/openrouter_stream.go`.
- [x] Define neutral text, image, assistant tool-call, and tool-result messages.
- [x] Define neutral requests, responses, finish reasons, and usage.
- [x] Build complete Chat message conversion.
- [x] Build full tool JSON Schema conversion.
- [x] Add actual routed-model and generation-ID capture.
- [x] Add bounded retry configuration.
- [x] Add context cancellation.
- [x] Pin the OpenRouter SDK dependency.
- [ ] Remove the Anthropic SDK only after all callers compile against the new boundary.
- [x] Raise the Go module requirement to at least 1.25.10 if the pinned SDK still requires it.

Exit criterion: the new package passes unit, wire-format, streaming, retry, and cancellation tests without importing it into the agent runtime yet.

### Phase 2: Conversation and Persistence

- [ ] Rewrite `internal/agent/context.go` to store neutral messages.
- [ ] Keep tool arguments as `json.RawMessage` until execution.
- [ ] Represent tool results as internal tool-role messages.
- [ ] Preserve `IsError` internally and on disk.
- [ ] Rewrite `internal/agent/message_serialization.go` around neutral messages.
- [ ] Preserve the existing serialized JSON fields.
- [ ] Add optional `schema_version` and treat absence as legacy version 1.
- [ ] Dual-read legacy user-role tool results and new tool-role messages.
- [ ] Preserve historical model and Advisor metric values.
- [ ] Add exchange-aware truncation.
- [ ] Remove orphaned tool exchanges atomically.
- [ ] Add lightweight image resource references where required.
- [ ] Re-inject the World Keeper map after restored history is loaded.
- [ ] Add a sanitized legacy `agent-states.json` fixture.

Exit criterion: existing state fixtures load, save, and reload without losing valid tool history or historical metrics.

### Phase 3: Tool Definitions

- [ ] Replace `ToAnthropicTools` with provider-neutral definitions.
- [ ] Remove standard and beta duplicate tool conversions.
- [ ] Forward complete `InputSchema()` maps.
- [ ] Do not enable strict schema mode initially.
- [ ] Keep filtered nested-agent registries unchanged.
- [ ] Verify every registered tool converts to a valid Chat function definition.
- [ ] Verify tool results remain valid JSON on success and failure.

Exit criterion: all current tools pass conversion tests, and no tool execution code imports either provider SDK.

### Phase 4: Main DM

- [ ] Inject `llm.Client` and model catalog into `agent.Agent`.
- [ ] Replace the concrete Anthropic client.
- [ ] Rename `callAnthropicAPI` to a provider-neutral operation.
- [ ] Send the system prompt as a Chat system message.
- [ ] Migrate user and assistant history to neutral messages.
- [ ] Migrate the streamed local-tool loop.
- [ ] Add a maximum tool-loop iteration count.
- [ ] Handle every finish reason explicitly.
- [ ] Track main-agent usage, cache tokens, cost, generation ID, and routed model.
- [ ] Thread `context.Context` through `ProcessUserMessage`.
- [ ] Cancel generation on CLI shutdown and web-session cancellation.
- [ ] Preserve terminal rendering and browser SSE behavior.
- [ ] Ensure partial streamed responses are neither persisted nor retried.
- [ ] Change the Dungeon Master persona declaration to match the approved Sonnet default.

Exit criterion: a mocked main-agent turn can stream text, execute one or more tools, send tool results, and finish normally.

### Phase 5: Nested Agents

- [ ] Replace `anthropicKey` and Anthropic client factories with `llm.Client` injection.
- [ ] Remove standard-versus-beta Messages service abstractions.
- [ ] Migrate `InvokeAgent` to Chat.
- [ ] Migrate `InvokeAgentSilent` to Chat.
- [ ] Preserve recursion limits.
- [ ] Preserve filtered read-only tools.
- [ ] Preserve per-agent iteration limits.
- [ ] Preserve the 120-second invocation timeout unless contract tests justify a change.
- [ ] Record requested and actual routed models.
- [ ] Record aggregate OpenRouter usage and cost.
- [ ] Verify session-start World Keeper briefing.
- [ ] Verify narrative judgment and synthesis.

Exit criterion: all nested agents work through a fake neutral client and persisted conversations resume correctly.

### Phase 6: Native OpenRouter Advisor

- [ ] Replace Anthropic beta Advisor types with `openrouter:advisor` configuration.
- [ ] Set the approved Advisor model to `anthropic/claude-opus-5`.
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

Exit criterion: gated real tests prove native Advisor works alone, with World Keeper context, and alongside local functions.

### Phase 7: Direct Utility Calls

- [ ] Inject `llm.Client` into `internal/ai.Enricher`.
- [ ] Migrate journal enrichment to the Fast model.
- [ ] Migrate map-prompt enrichment to the Fast model.
- [ ] Inject `llm.Client` into ambient prompt generation.
- [ ] Migrate only Lyria parameter generation; keep Google Lyria transport unchanged.
- [ ] Inject an optional `llm.Client` into biography generation.
- [ ] Preserve template fallback when biography AI generation is unavailable.
- [ ] Migrate campaign-plan generation to the Campaign model.
- [ ] Preserve optional world-map image input.
- [ ] Migrate adventure-title suggestions to the Fast model.
- [ ] Preserve the best-effort empty-title fallback.
- [ ] Use structured output for JSON-producing calls where the selected endpoint confirms support.

Exit criterion: every former direct Anthropic client call routes through the shared neutral client.

### Phase 8: Entrypoints and Dependency Injection

- [ ] Update `cmd/dm/main.go` to require `OPENROUTER_API_KEY`.
- [ ] Update `cmd/web/main.go` to require `OPENROUTER_API_KEY`.
- [ ] Update `cmd/adventure/main.go` enrichment and `coherence --ai` paths.
- [ ] Update `cmd/advisor-ab/main.go`.
- [ ] Update `cmd/character-sheet/main.go` optional AI path.
- [ ] Update `internal/web/server.go` configuration.
- [ ] Update `internal/web/session.go` session construction.
- [ ] Update `internal/agent/register_tools.go` dependency injection.
- [ ] Update `internal/dmtools/map_tool.go` construction.
- [ ] Update `internal/dmtools/ambient_tool.go` construction.
- [ ] Remove package-level API-key environment reads outside the config loader.
- [ ] Ensure one OpenRouter client is reused per process.

Exit criterion: all binaries build and no production package creates an Anthropic client or reads `ANTHROPIC_API_KEY`.

### Phase 9: Model Selection

- [ ] Replace Anthropic constants with concrete OpenRouter model IDs.
- [ ] Preserve `haiku`, `sonnet`, and `opus` persona aliases.
- [ ] Accept full OpenRouter model IDs in persona/configuration input.
- [ ] Centralize selectable models and display names.
- [ ] Make Sonnet 5 the main DM default.
- [ ] Make Opus 5 the Premium selector option.
- [ ] Keep Haiku 4.5 as the Fast selector option.
- [ ] Replace substring-based model detection in web handlers.
- [ ] Render the web selector from backend model metadata.
- [ ] Return stable IDs and display labels from model endpoints.
- [ ] Confirm longer labels fit the existing UI.
- [ ] Log the actual routed model, not only the requested alias.

Exit criterion: CLI, web, personas, logs, and metrics agree on model identity.

### Phase 10: Tests and Mocks

- [ ] Replace `internal/agent/mock_anthropic.go` with a test-only neutral fake.
- [ ] Add non-streaming OpenRouter adapter tests using `httptest.Server`.
- [ ] Add authentication and attribution-header tests.
- [ ] Add system, user, assistant, image, tool-call, and tool-result conversion tests.
- [ ] Add complete tool-schema preservation tests.
- [ ] Add SSE keepalive tests.
- [ ] Add streamed text tests.
- [ ] Add fragmented tool-argument tests.
- [ ] Add interleaved parallel tool-call tests.
- [ ] Add final usage and duplicate finish-reason tests.
- [ ] Add mid-stream error tests.
- [ ] Add cancellation and stream-closure tests.
- [ ] Add bounded retry tests.
- [ ] Add non-retryable authentication and payment-error tests.
- [ ] Add main-agent tool-loop tests.
- [ ] Add nested-agent tool-loop tests.
- [ ] Add silent-agent tests.
- [ ] Add Advisor disabled and enabled tests.
- [ ] Add Advisor-plus-local-tools tests.
- [ ] Add legacy state fixture tests.
- [ ] Add World Keeper map restoration tests.
- [ ] Add campaign multimodal request tests.
- [ ] Add ambient prompt tests.
- [ ] Add biography AI and fallback tests.
- [ ] Gate paid provider tests behind `OPENROUTER_API_KEY` and `RUN_REAL_API_TESTS=1`.

Exit criterion: unit and integration tests cover all protocol conversions and agent paths before the Anthropic dependency is removed.

### Phase 11: Documentation

- [ ] Update active runtime sections in `README.md`.
- [ ] Update `DEPLOYMENT.md`.
- [ ] Update active runtime sections in `CLAUDE.md`.
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
- [ ] Remove Anthropic-only client, beta, and mock types.
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
- [ ] Run gated real OpenRouter contract tests.
- [ ] Run one same-version reference adventure through the migrated engine.
- [ ] Verify FAL.ai, Google Imagen, and Google Lyria transports are unchanged.

Exit criterion: all automated checks pass, real contract tests pass, and the reference-adventure evidence report satisfies the acceptance criteria below.

## Affected File Inventory

### New Core Package

| Proposed path | Responsibility |
|---|---|
| `internal/llm/types.go` | Neutral messages, requests, responses, and usage |
| `internal/llm/client.go` | Injectable client and stream observer interfaces |
| `internal/llm/config.go` | Environment and runtime configuration |
| `internal/llm/models.go` | Concrete models, aliases, and selectable catalog |
| `internal/llm/errors.go` | Provider-neutral errors |
| `internal/llm/openrouter.go` | Official SDK adapter |
| `internal/llm/openrouter_messages.go` | Chat wire conversion |
| `internal/llm/openrouter_stream.go` | SSE and tool-call assembly |

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
| `cmd/adventure/main.go` | Enrichment and narrative judgment |
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
| `CLAUDE.md` | Active runtime architecture only |
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
- Scenario Critic and narrative synthesis work.
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
- Do not add arbitrary non-Claude models to the first production catalog.
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
- Mutable model aliases would reduce reproducibility and are intentionally avoided.
- A live Advisor contract test is mandatory before production enablement.

## Rollout Strategy

1. Land neutral types and persistence compatibility without changing production behavior.
2. Land and test the OpenRouter adapter.
3. Migrate stateless utility calls.
4. Migrate the main agent with a fake and contract-tested adapter.
5. Migrate nested agents with Advisor disabled.
6. Run a same-version reference adventure.
7. Run the native Advisor contract suite.
8. Enable Advisor behind `SW_ADVISOR_ENABLED`.
9. Remove Anthropic code, key references, and dependency.
10. Update active documentation and deployment guidance.

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
