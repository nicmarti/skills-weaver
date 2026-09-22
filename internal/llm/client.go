package llm

import "context"

// StreamObserver receives streamed output as it arrives.
type StreamObserver interface {
	// OnTextDelta receives each incremental text fragment as it streams.
	OnTextDelta(text string)
}

// Client is the provider-neutral LLM boundary used by agents and utilities.
//
// Implementations own all provider-specific conversion, streaming assembly,
// error normalization, and retry behavior. Requests carry logical model IDs
// resolved through this package's model catalog.
type Client interface {
	// Complete performs a non-streaming request.
	Complete(ctx context.Context, req Request) (Response, error)
	// Stream performs a streaming request, forwarding text deltas to obs and
	// returning the accumulated assistant message.
	Stream(ctx context.Context, req Request, obs StreamObserver) (Response, error)
}
