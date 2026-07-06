// Package repository is the persistence layer for the messages feature.
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
	pgxutil "messenger/internal/core/repository/postgres/pgx"
)

// Repository stores and retrieves direct messages.
type Repository struct {
	db pgxutil.Querier
}

// New creates a messages Repository.
func New(db pgxutil.Querier) *Repository {
	return &Repository{db: db}
}

// Create inserts a message. A foreign-key violation (unknown recipient) is
// mapped to a domain error.
func (r *Repository) Create(ctx context.Context, m domain.Message) error {
	const q = `
		INSERT INTO messages (id, sender_id, recipient_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, q, m.ID, m.SenderID, m.RecipientID, m.Body, m.CreatedAt)
	if err != nil {
		if pgxutil.ForeignKeyViolation(err) {
			return apperrors.ErrRecipientMissing
		}
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}

// Conversation returns messages exchanged between two users, oldest first,
// paged by limit/offset.
func (r *Repository) Conversation(ctx context.Context, self, other uuid.UUID, limit, offset int) ([]domain.Message, error) {
	const q = `
		SELECT id, sender_id, recipient_id, body, created_at
		FROM messages
		WHERE (sender_id = $1 AND recipient_id = $2)
		   OR (sender_id = $2 AND recipient_id = $1)
		ORDER BY created_at
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, q, self, other, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query conversation: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0, limit)
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.Body, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}
	return messages, nil
}
