// Package repository is the persistence layer for the auth feature (users and
// sessions).
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
	pgxutil "messenger/internal/core/repository/postgres/pgx"
)

// Repository stores users and authentication sessions in PostgreSQL.
type Repository struct {
	db pgxutil.Querier
}

// New creates a Repository backed by the given querier (pool or transaction).
func New(db pgxutil.Querier) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user, translating unique-constraint violations into
// domain errors.
func (r *Repository) CreateUser(ctx context.Context, u domain.User) error {
	const q = `
		INSERT INTO users (id, nickname, phone, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, q, u.ID, u.Nickname, u.Phone, u.PasswordHash, u.CreatedAt)
	if err != nil {
		if constraint, ok := pgxutil.UniqueViolation(err); ok {
			switch constraint {
			case "users_nickname_unique":
				return apperrors.ErrNicknameTaken
			case "users_phone_unique":
				return apperrors.ErrPhoneTaken
			}
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetUserByNickname fetches a user by nickname.
func (r *Repository) GetUserByNickname(ctx context.Context, nickname string) (domain.User, error) {
	const q = `
		SELECT id, nickname, phone, password_hash, created_at
		FROM users WHERE nickname = $1`
	return r.scanUser(r.db.QueryRow(ctx, q, nickname))
}

// GetUserByID fetches a user by id.
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const q = `
		SELECT id, nickname, phone, password_hash, created_at
		FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRow(ctx, q, id))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *Repository) scanUser(row rowScanner) (domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Nickname, &u.Phone, &u.PasswordHash, &u.CreatedAt); err != nil {
		if pgxutil.IsNoRows(err) {
			return domain.User{}, apperrors.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}

// CreateSession persists a new session.
func (r *Repository) CreateSession(ctx context.Context, s domain.Session) error {
	const q = `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)`
	if _, err := r.db.Exec(ctx, q, s.Token, s.UserID, s.CreatedAt, s.ExpiresAt); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// GetSession fetches a session by token.
func (r *Repository) GetSession(ctx context.Context, token uuid.UUID) (domain.Session, error) {
	const q = `
		SELECT token, user_id, created_at, expires_at
		FROM sessions WHERE token = $1`
	var s domain.Session
	err := r.db.QueryRow(ctx, q, token).Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if pgxutil.IsNoRows(err) {
			return domain.Session{}, apperrors.ErrInvalidToken
		}
		return domain.Session{}, fmt.Errorf("scan session: %w", err)
	}
	return s, nil
}

// DeleteSession removes a session by token (idempotent).
func (r *Repository) DeleteSession(ctx context.Context, token uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE token = $1`
	if _, err := r.db.Exec(ctx, q, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
