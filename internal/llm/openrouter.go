package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	oropenrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/retry"
	"github.com/OpenRouterTeam/go-sdk/types/stream"
)

// attributionClient injects OpenRouter attribution headers on every request.
// The generated SDK fails to populate these globals for Chat Completions, so
// the adapter owns them at the transport level.
type attributionClient struct {
	inner   oropenrouter.HTTPClient
	referer string
	title   string
}

func (c *attributionClient) Do(req *http.Request) (*http.Response, error) {
	if c.referer != "" {
		req.Header.Set("HTTP-Referer", c.referer)
	}
	if c.title != "" {
		req.Header.Set("X-Title", c.title)
	}
	return c.inner.Do(req)
}

// defaultRetry bounds the SDK's retry window (the SDK default retries 5XX for
// up to one hour). Only 5XX provider responses are retried by the SDK, always
// before any streamed output is observable. 429 is deliberately not retried.
func defaultRetry() retry.Config {
	return retry.Config{
		Strategy: "backoff",
		Backoff: &retry.BackoffStrategy{
			InitialInterval: 500,
			MaxInterval:     2000,
			Exponent:        1.5,
			MaxElapsedTime:  10000,
		},
		RetryConnectionErrors: true,
	}
}

// OpenRouterClient implements Client with the official OpenRouter Go SDK.
type OpenRouterClient struct {
	sdk *oropenrouter.OpenRouter
}

// NewOpenRouterClient builds the shared OpenRouter client for a process.
func NewOpenRouterClient(cfg Config) (*OpenRouterClient, error) {
	return newOpenRouterClient(cfg, nil, "")
}

// newOpenRouterClient is the internal constructor used by tests to point the
// SDK at a stub server and override the retry policy.
func newOpenRouterClient(cfg Config, retries *retry.Config, serverURL string) (*OpenRouterClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm: OpenRouter API key is required")
	}

	opts := []oropenrouter.SDKOption{
		oropenrouter.WithSecurity(cfg.APIKey),
	}
	if retries == nil {
		r := defaultRetry()
		retries = &r
	}
	opts = append(opts, oropenrouter.WithRetryConfig(*retries))
	if serverURL != "" {
		opts = append(opts, oropenrouter.WithServerURL(serverURL))
	}
	// The SDK does not propagate WithHTTPReferer/WithXTitle globals to the
	// Chat operation, so attribution headers are injected by the HTTP client.
	if cfg.HTTPReferer != "" || cfg.AppName != "" {
		opts = append(opts, oropenrouter.WithClient(&attributionClient{
			inner:   http.DefaultClient,
			referer: cfg.HTTPReferer,
			title:   cfg.AppName,
		}))
	}

	return &OpenRouterClient{sdk: oropenrouter.New(opts...)}, nil
}

// Complete performs a non-streaming Chat Completions request.
func (c *OpenRouterClient) Complete(ctx context.Context, req Request) (Response, error) {
	chatReq, err := buildChatRequest(req, false)
	if err != nil {
		return Response{}, err
	}

	res, err := c.sdk.Chat.Send(ctx, chatReq, nil)
	if err != nil {
		return Response{}, NormalizeSendError(err)
	}
	if res.ChatResult == nil {
		return Response{}, &Error{Kind: ErrKindProvider, Message: "unexpected empty chat result"}
	}

	return chatResultToResponse(res.ChatResult)
}

// Stream performs a streaming Chat Completions request. Text deltas are
// forwarded to obs as they arrive; the returned Response carries the
// accumulated assistant message. No retry happens after the stream opens.
func (c *OpenRouterClient) Stream(ctx context.Context, req Request, obs StreamObserver) (Response, error) {
	chatReq, err := buildChatRequest(req, true)
	if err != nil {
		return Response{}, err
	}

	res, err := c.sdk.Chat.Send(ctx, chatReq, nil)
	if err != nil {
		return Response{}, NormalizeSendError(err)
	}
	if res.EventStream == nil {
		return Response{}, &Error{Kind: ErrKindProvider, Message: "expected streaming response"}
	}

	return consumeStream(ctx, res.EventStream, obs)
}

// toolAccum accumulates one streamed tool call across fragments.
type toolAccum struct {
	id   string
	name string
	args strings.Builder
}

// consumeStream assembles the SSE stream into a neutral response:
//   - checks each chunk for an embedded mid-stream error,
//   - forwards text deltas immediately,
//   - assembles interleaved parallel tool calls by index,
//   - captures the final usage chunk (which repeats the terminal finish reason),
//   - rejects incomplete or malformed tool calls.
func consumeStream(ctx context.Context, es *stream.EventStream[components.ChatStreamingResponse], obs StreamObserver) (Response, error) {
	var resp Response
	resp.Message.Role = RoleAssistant

	accs := map[int64]*toolAccum{}
	var order []int64
	getAcc := func(index int64) *toolAccum {
		if acc, ok := accs[index]; ok {
			return acc
		}
		acc := &toolAccum{}
		accs[index] = acc
		order = append(order, index)
		return acc
	}

	for es.Next() {
		ev := es.Value()
		if ev == nil {
			continue
		}
		chunk := ev.Data

		if chunk.Error != nil {
			closeStream(es)
			return resp, &Error{
				Kind:      ErrKindMidStream,
				Status:    int(chunk.Error.Code),
				Message:   chunk.Error.Message,
				Retryable: false,
			}
		}

		if resp.ID == "" {
			resp.ID = chunk.ID
		}
		if resp.Model == "" {
			resp.Model = chunk.Model
		}

		for _, choice := range chunk.Choices {
			if choice.Index != 0 {
				continue
			}

			if s, ok := choice.Delta.Content.Get(); ok && s != nil && *s != "" {
				if obs != nil {
					obs.OnTextDelta(*s)
				}
				resp.Message.Text += *s
			}

			for _, tc := range choice.Delta.ToolCalls {
				acc := getAcc(tc.Index)
				if tc.ID != nil && *tc.ID != "" {
					acc.id = *tc.ID
				}
				if tc.Function != nil {
					if tc.Function.Name != nil && *tc.Function.Name != "" {
						acc.name = *tc.Function.Name
					}
					if tc.Function.Arguments != nil {
						acc.args.WriteString(*tc.Function.Arguments)
					}
				}
			}

			if choice.FinishReason != nil {
				resp.FinishReason = mapFinishReason(choice.FinishReason)
			}
		}

		if chunk.Usage != nil {
			resp.Usage = usageFromChat(chunk.Usage)
		}
	}

	if err := es.Err(); err != nil {
		closeStream(es)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return resp, wrapContextError(ctxErr)
		}
		return resp, wrapContextError(err)
	}
	closeStream(es)

	calls, err := finalizeToolCalls(order, accs)
	if err != nil {
		return resp, err
	}
	resp.Message.ToolCalls = calls

	if len(resp.Message.ToolCalls) > 0 && resp.FinishReason == FinishNone {
		resp.FinishReason = FinishToolCalls
	}

	return resp, nil
}

// finalizeToolCalls assembles each accumulated tool call exactly once,
// rejecting incomplete calls or malformed argument JSON.
func finalizeToolCalls(order []int64, accs map[int64]*toolAccum) ([]ToolCall, error) {
	var calls []ToolCall
	for _, index := range order {
		acc := accs[index]
		if acc.id == "" || acc.name == "" {
			return nil, &Error{
				Kind:    ErrKindMalformed,
				Message: fmt.Sprintf("incomplete streamed tool call at index %d", index),
			}
		}
		args := acc.args.String()
		if args == "" {
			args = "{}"
		}
		if !json.Valid([]byte(args)) {
			return nil, &Error{
				Kind:    ErrKindMalformed,
				Message: fmt.Sprintf("malformed streamed tool arguments for %s", acc.name),
			}
		}
		calls = append(calls, ToolCall{ID: acc.id, Name: acc.name, Arguments: json.RawMessage(args)})
	}
	return calls, nil
}

func closeStream(es *stream.EventStream[components.ChatStreamingResponse]) {
	_ = es.Close()
}
