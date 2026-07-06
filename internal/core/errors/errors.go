// Package errors provides a small typed-error toolkit used across the app.
//
// Handlers translate an *Error into an HTTP status + machine readable code,
// while services return domain-meaningful errors regardless of transport.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Kind classifies an error into a coarse category used to derive an HTTP status.
type Kind string

const (
	KindInvalidInput Kind = "invalid_input"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindInternal     Kind = "internal"
)

// Error is an application error carrying a stable code, a human message and a Kind.
type Error struct {
	Kind    Kind   `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	err     error  // wrapped cause (never serialized)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause.
func (e *Error) Unwrap() error { return e.err }

// Wrap attaches a cause to the error, returning a copy.
func (e *Error) Wrap(cause error) *Error {
	clone := *e
	clone.err = cause
	return &clone
}

// WithMessage overrides the human message, returning a copy.
func (e *Error) WithMessage(msg string) *Error {
	clone := *e
	clone.Message = msg
	return &clone
}

// New builds a new *Error.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

// HTTPStatus maps an error to an HTTP status code.
func HTTPStatus(err error) int {
	var appErr *Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case KindInvalidInput:
			return http.StatusBadRequest
		case KindUnauthorized:
			return http.StatusUnauthorized
		case KindForbidden:
			return http.StatusForbidden
		case KindNotFound:
			return http.StatusNotFound
		case KindConflict:
			return http.StatusConflict
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// As is a thin re-export of errors.As for callers that only import this package.
func As(err error, target any) bool { return errors.As(err, target) }

// Is is a thin re-export of errors.Is for callers that only import this package.
func Is(err, target error) bool { return errors.Is(err, target) }

// Common, reusable application errors.
var (
	ErrInternal          = New(KindInternal, "internal_error", "internal server error")
	ErrValidation        = New(KindInvalidInput, "validation_failed", "request validation failed")
	ErrMalformedBody     = New(KindInvalidInput, "malformed_body", "request body is malformed")
	ErrUnauthorized      = New(KindUnauthorized, "unauthorized", "authentication required")
	ErrInvalidToken      = New(KindUnauthorized, "invalid_token", "invalid or expired session token")
	ErrInvalidCredential = New(KindUnauthorized, "invalid_credentials", "invalid nickname or password")

	ErrUserNotFound     = New(KindNotFound, "user_not_found", "user not found")
	ErrNicknameTaken    = New(KindConflict, "nickname_taken", "nickname is already taken")
	ErrPhoneTaken       = New(KindConflict, "phone_taken", "phone number is already registered")
	ErrRecipientMissing = New(KindNotFound, "recipient_not_found", "recipient does not exist")
	ErrSelfMessage      = New(KindInvalidInput, "self_message", "cannot send a message to yourself")
)
