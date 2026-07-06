// Package service holds the users feature business logic.
package service

import (
	"context"

	"github.com/google/uuid"

	"messenger/internal/core/domain"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// Repository is the persistence contract for the users service.
type Repository interface {
	List(ctx context.Context, exclude uuid.UUID, limit, offset int) ([]domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// Service implements the users use cases.
type Service struct {
	repo Repository
}

// New creates a users Service.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns other users the requester can chat with. limit/offset are
// normalized to safe bounds.
func (s *Service) List(ctx context.Context, requester uuid.UUID, limit, offset int) ([]domain.User, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, requester, limit, offset)
}

// Me returns the current user's profile.
func (s *Service) Me(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
