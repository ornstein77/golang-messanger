// Package types holds transport-level shared types: the JSON response envelope
// and request-context helpers for the authenticated user.
package types

import (
	"context"

	"github.com/google/uuid"
)

// ErrorBody is the machine-readable error payload returned to clients.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Envelope is the uniform JSON response wrapper. Exactly one of Data or Error
// is set for any given response.
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type contextKey string

const userIDKey contextKey = "auth.user_id"

// WithUserID stores the authenticated user id in the context.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserID extracts the authenticated user id from the context.
func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
