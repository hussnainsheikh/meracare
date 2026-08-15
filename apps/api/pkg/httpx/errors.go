// Package httpx holds the transport-level helpers shared by every API module:
// the error envelope, JSON encoding, and the base middleware stack.
package httpx

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes returned to clients. These are part of the public API contract —
// keep them in sync with packages/contracts/src/errors.ts.
const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeRateLimited      = "RATE_LIMITED"
	CodeInternal         = "INTERNAL"
	CodeUnavailable      = "UNAVAILABLE"
)

// ErrorResponse is the body of every non-2xx response.
//
//	{"error": {"code": "FORBIDDEN", "message": "..."}}
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries the client-facing failure description. It must never
// contain database errors, stack traces, or health data.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// APIError is an error that carries an HTTP status and a client-safe message.
// Handlers return it; Respond turns it into an ErrorResponse.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details map[string]string
	// cause is logged server-side and never sent to the client.
	cause error
}

func (e *APIError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the internal cause to errors.Is/errors.As without leaking it
// to clients.
func (e *APIError) Unwrap() error { return e.cause }

// WithCause attaches an internal error for server-side logging.
func (e *APIError) WithCause(cause error) *APIError {
	clone := *e
	clone.cause = cause
	return &clone
}

// WithDetails attaches field-level validation details.
func (e *APIError) WithDetails(details map[string]string) *APIError {
	clone := *e
	clone.Details = details
	return &clone
}

// Constructors for the failures the API actually produces.

func ErrBadRequest(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: CodeBadRequest, Message: message}
}

func ErrValidation(message string, details map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidationFailed,
		Message: message,
		Details: details,
	}
}

func ErrUnauthenticated(message string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: message}
}

func ErrForbidden(message string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: CodeForbidden, Message: message}
}

func ErrNotFound(message string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: CodeNotFound, Message: message}
}

func ErrConflict(message string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: CodeConflict, Message: message}
}

// ErrInternal returns the only message clients ever see for unexpected
// failures. The real cause is logged, not returned.
func ErrInternal(cause error) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "Something went wrong. Please try again.",
		cause:   cause,
	}
}

func ErrUnavailable(message string) *APIError {
	return &APIError{Status: http.StatusServiceUnavailable, Code: CodeUnavailable, Message: message}
}

// AsAPIError converts any error into an APIError, defaulting to an internal
// error so that unexpected failures never leak details.
func AsAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return ErrInternal(err)
}
