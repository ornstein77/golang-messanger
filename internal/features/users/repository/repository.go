// Package repository is the persistence layer for the users feature.
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
	pgxutil "messenger/internal/core/repository/postgres/pgx"
)

// Repository reads user records from PostgreSQL.
type Repository struct {
	db pgxutil.Querier
}

// New creates a users Repository.
func New(db pgxutil.Querier) *Repository {
	return &Repository{db: db}
}

// List returns users ordered by nickname, excluding the given user id, paged by
// limit/offset.
func (r *Repository) List(ctx context.Context, exclude uuid.UUID, limit, offset int) ([]domain.User, error) {
	const q = `
		SELECT id, nickname, phone, password_hash, created_at
		FROM users
		WHERE id <> $1
		ORDER BY nickname
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, exclude, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, limit)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Nickname, &u.Phone, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

// GetByID fetches a single user by id.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const q = `
		SELECT id, nickname, phone, password_hash, created_at
		FROM users WHERE id = $1`

	var u domain.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Nickname, &u.Phone, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if pgxutil.IsNoRows(err) {
			return domain.User{}, apperrors.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
