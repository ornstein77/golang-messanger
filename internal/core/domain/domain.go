// Package domain contains the core business entities shared across features.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is a registered participant of the messenger.
type User struct {
	ID           uuid.UUID `json:"id"`
	Nickname     string    `json:"nickname"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session is an authentication token bound to a user.
type Session struct {
	Token     uuid.UUID `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Expired reports whether the session is no longer valid at time t.
func (s Session) Expired(t time.Time) bool {
	return !s.ExpiresAt.After(t)
}

// Message is a direct message sent from one user to another.
type Message struct {
	ID          uuid.UUID `json:"id"`
	SenderID    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}
