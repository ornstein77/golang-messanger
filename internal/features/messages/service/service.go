// Package service holds the messages feature business logic.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// Repository is the persistence contract for the messages service.
type Repository interface {
	Create(ctx context.Context, m domain.Message) error
	Conversation(ctx context.Context, self, other uuid.UUID, limit, offset int) ([]domain.Message, error)
}

// Service implements the messaging use cases.
type Service struct {
	repo Repository
	now  func() time.Time
}

// New creates a messages Service.
func New(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// Send delivers a message from sender to recipient.
func (s *Service) Send(ctx context.Context, sender, recipient uuid.UUID, body string) (domain.Message, error) {
	if sender == recipient {
		return domain.Message{}, apperrors.ErrSelfMessage
	}

	msg := domain.Message{
		ID:          uuid.New(),
		SenderID:    sender,
		RecipientID: recipient,
		Body:        body,
		CreatedAt:   s.now().UTC(),
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return domain.Message{}, err
	}
	return msg, nil
}

// Conversation returns the message history between the current user and another.
func (s *Service) Conversation(ctx context.Context, self, other uuid.UUID, limit, offset int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.Conversation(ctx, self, other, limit, offset)
}
