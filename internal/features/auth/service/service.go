// Package service holds the auth business logic: registration, login, logout
// and session verification.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"messenger/internal/core/config"
	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
)

// Repository is the persistence contract required by the auth service.
type Repository interface {
	CreateUser(ctx context.Context, u domain.User) error
	GetUserByNickname(ctx context.Context, nickname string) (domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	CreateSession(ctx context.Context, s domain.Session) error
	GetSession(ctx context.Context, token uuid.UUID) (domain.Session, error)
	DeleteSession(ctx context.Context, token uuid.UUID) error
}

// Service implements the auth use cases.
type Service struct {
	repo       Repository
	sessionTTL time.Duration
	now        func() time.Time
}

// New creates a new auth Service.
func New(repo Repository, cfg config.Session) *Service {
	return &Service{
		repo:       repo,
		sessionTTL: cfg.TTL,
		now:        time.Now,
	}
}

// RegisterInput carries the data required to register a new user.
type RegisterInput struct {
	Nickname string
	Phone    string
	Password string
}

// Register creates a new user with a securely hashed password.
func (s *Service) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		Nickname:     in.Nickname,
		Phone:        in.Phone,
		PasswordHash: string(hash),
		CreatedAt:    s.now().UTC(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

// Login verifies credentials and issues a new session.
func (s *Service) Login(ctx context.Context, nickname, password string) (domain.Session, domain.User, error) {
	user, err := s.repo.GetUserByNickname(ctx, nickname)
	if err != nil {
		// Mask "user not found" as invalid credentials to avoid user enumeration.
		if apperrors.Is(err, apperrors.ErrUserNotFound) {
			return domain.Session{}, domain.User{}, apperrors.ErrInvalidCredential
		}
		return domain.Session{}, domain.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.Session{}, domain.User{}, apperrors.ErrInvalidCredential
	}

	now := s.now().UTC()
	session := domain.Session{
		Token:     uuid.New(),
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return domain.Session{}, domain.User{}, err
	}
	return session, user, nil
}

// Logout revokes a session token.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	token, err := uuid.Parse(rawToken)
	if err != nil {
		return apperrors.ErrInvalidToken
	}
	return s.repo.DeleteSession(ctx, token)
}

// Authenticate validates a raw bearer token and returns the associated user id.
// It satisfies middleware.Authenticator.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (uuid.UUID, error) {
	token, err := uuid.Parse(rawToken)
	if err != nil {
		return uuid.Nil, apperrors.ErrInvalidToken
	}

	session, err := s.repo.GetSession(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}

	if session.Expired(s.now().UTC()) {
		// Best-effort cleanup; ignore the delete error.
		_ = s.repo.DeleteSession(ctx, token)
		return uuid.Nil, apperrors.ErrInvalidToken
	}
	return session.UserID, nil
}
