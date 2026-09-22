package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/OpenRouterTeam/go-sdk/models/sdkerrors"
)

// ErrorKind classifies normalized provider failures so callers can make
// retry/report decisions without touching provider SDK types.
type ErrorKind string

const (
	ErrKindAuth           ErrorKind = "authentication"
	ErrKindPayment        ErrorKind = "payment"
	ErrKindPermission     ErrorKind = "permission"
	ErrKindInvalidRequest ErrorKind = "invalid_request"
	ErrKindRateLimit      ErrorKind = "rate_limit"
	ErrKindOverloaded     ErrorKind = "overloaded"
	ErrKindUnavailable    ErrorKind = "unavailable"
	ErrKindTimeout        ErrorKind = "timeout"
	ErrKindCanceled       ErrorKind = "canceled"
	ErrKindContextLength  ErrorKind = "context_length"
	ErrKindMalformed      ErrorKind = "malformed_stream"
	ErrKindMidStream      ErrorKind = "mid_stream"
	ErrKindProvider       ErrorKind = "provider"
)

// Error is a normalized provider failure.
type Error struct {
	Kind    ErrorKind
	Status  int
	Message string
	// Retryable reports whether the failure class may be retried. Never retry
	// after observable streamed output, regardless of this flag.
	Retryable bool
}

func (e *Error) Error() string {
	return fmt.Sprintf("llm: %s (kind=%s status=%d)", e.Message, e.Kind, e.Status)
}

// normalizeStatus maps an HTTP status to a normalized error kind.
func normalizeStatus(status int, message string) *Error {
	switch status {
	case 401:
		return &Error{Kind: ErrKindAuth, Status: status, Message: message}
	case 402:
		return &Error{Kind: ErrKindPayment, Status: status, Message: message}
	case 403:
		return &Error{Kind: ErrKindPermission, Status: status, Message: message}
	case 400, 404, 413, 422:
		return &Error{Kind: ErrKindInvalidRequest, Status: status, Message: message}
	case 408, 524:
		return &Error{Kind: ErrKindTimeout, Status: status, Message: message, Retryable: true}
	case 429:
		return &Error{Kind: ErrKindRateLimit, Status: status, Message: message, Retryable: true}
	case 503:
		return &Error{Kind: ErrKindUnavailable, Status: status, Message: message, Retryable: true}
	case 529:
		return &Error{Kind: ErrKindOverloaded, Status: status, Message: message, Retryable: true}
	case 500, 502, 504:
		return &Error{Kind: ErrKindProvider, Status: status, Message: message, Retryable: true}
	default:
		kind := ErrKindProvider
		retryable := status >= 500
		return &Error{Kind: kind, Status: status, Message: message, Retryable: retryable}
	}
}

// NormalizeSendError converts an error returned by the OpenRouter SDK into a
// normalized *Error. Typed per-status errors carry no shared interface, so
// concrete types are matched directly; anything else falls back to
// sdkerrors.APIError or a provider error.
func NormalizeSendError(err error) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *sdkerrors.UnauthorizedResponseError:
		return &Error{Kind: ErrKindAuth, Status: 401, Message: e.Error()}
	case *sdkerrors.PaymentRequiredResponseError:
		return &Error{Kind: ErrKindPayment, Status: 402, Message: e.Error()}
	case *sdkerrors.ForbiddenResponseError:
		return &Error{Kind: ErrKindPermission, Status: 403, Message: e.Error()}
	case *sdkerrors.BadRequestResponseError,
		*sdkerrors.NotFoundResponseError,
		*sdkerrors.PayloadTooLargeResponseError,
		*sdkerrors.UnprocessableEntityResponseError:
		return &Error{Kind: ErrKindInvalidRequest, Status: 400, Message: e.Error()}
	case *sdkerrors.RequestTimeoutResponseError,
		*sdkerrors.EdgeNetworkTimeoutResponseError,
		*sdkerrors.GatewayTimeoutResponseError:
		return &Error{Kind: ErrKindTimeout, Status: 408, Message: e.Error(), Retryable: true}
	case *sdkerrors.TooManyRequestsResponseError:
		return &Error{Kind: ErrKindRateLimit, Status: 429, Message: e.Error(), Retryable: true}
	case *sdkerrors.ServiceUnavailableResponseError:
		return &Error{Kind: ErrKindUnavailable, Status: 503, Message: e.Error(), Retryable: true}
	case *sdkerrors.ProviderOverloadedResponseError:
		return &Error{Kind: ErrKindOverloaded, Status: 529, Message: e.Error(), Retryable: true}
	case *sdkerrors.BadGatewayResponseError,
		*sdkerrors.InternalServerResponseError:
		return &Error{Kind: ErrKindProvider, Status: 500, Message: e.Error(), Retryable: true}
	}

	var apiErr *sdkerrors.APIError
	if errors.As(err, &apiErr) {
		return normalizeStatus(apiErr.StatusCode, apiErr.Message)
	}

	return wrapContextError(err)
}

// wrapContextError distinguishes cancellation and deadlines from provider
// faults so output handlers never report cancellation as a model error.
func wrapContextError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return &Error{Kind: ErrKindCanceled, Message: "request canceled"}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{Kind: ErrKindTimeout, Message: "request deadline exceeded", Retryable: true}
	}
	return &Error{Kind: ErrKindProvider, Message: err.Error()}
}
