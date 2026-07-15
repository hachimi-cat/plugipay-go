package plugipay

import "fmt"

// Error is the single error type returned by every SDK operation. Wraps
// an HTTP status code, an opaque error code from the backend, a message,
// and an optional request id so callers can branch on well-defined
// reasons without matching strings.
//
// Use errors.As to extract:
//
//	var pe *plugipay.Error
//	if errors.As(err, &pe) && pe.Code == "rate_limited" { ... }
type Error struct {
	// HTTP status code. 0 for transport-layer failures (timeout, DNS, etc).
	Status int
	// Stable error code (e.g. "rate_limited", "signature_invalid",
	// "network_error", "timeout", "invalid_response").
	Code string
	// Human-readable message. Safe to surface to operators; not safe to
	// surface to end users without context.
	Message string
	// RequestID from the API envelope's meta block, when present.
	RequestID string
}

func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("plugipay: %s: %s (status=%d, requestId=%s)",
			e.Code, e.Message, e.Status, e.RequestID)
	}
	if e.Status != 0 {
		return fmt.Sprintf("plugipay: %s: %s (status=%d)", e.Code, e.Message, e.Status)
	}
	return fmt.Sprintf("plugipay: %s: %s", e.Code, e.Message)
}

func newErr(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}
