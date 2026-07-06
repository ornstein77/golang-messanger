// Package transport exposes the users feature over HTTP.
package transport

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"messenger/internal/core/domain"
	apperrors "messenger/internal/core/errors"
	"messenger/internal/core/transport/http/middleware"
	"messenger/internal/core/transport/http/response"
	"messenger/internal/core/transport/http/types"
)

// Service is the users use-case contract required by the HTTP handlers.
type Service interface {
	List(ctx context.Context, requester uuid.UUID, limit, offset int) ([]domain.User, error)
	Me(ctx context.Context, id uuid.UUID) (domain.User, error)
}

// Handler serves the users endpoints. All routes require authentication.
type Handler struct {
	svc  Service
	auth middleware.Middleware
	log  *zap.Logger
}

// New creates the users Handler.
func New(svc Service, auth middleware.Middleware, log *zap.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

// Register attaches the users routes to the mux, protected by auth middleware.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/users", h.auth(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/v1/users/me", h.auth(http.HandlerFunc(h.me)))
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{ID: u.ID, Nickname: u.Nickname, Phone: u.Phone, CreatedAt: u.CreatedAt}
}

// list godoc
//
//	@Summary		List users
//	@Description	Lists other registered users so you can start a conversation.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Max results (default 50, max 200)"
//	@Param			offset	query		int	false	"Pagination offset"
//	@Success		200		{object}	types.Envelope{data=[]userResponse}
//	@Failure		401		{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/users [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	requester, ok := types.UserID(r.Context())
	if !ok {
		response.Error(w, h.log, apperrors.ErrUnauthorized)
		return
	}

	limit := atoiDefault(r.URL.Query().Get("limit"), 0)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)

	users, err := h.svc.List(r.Context(), requester, limit, offset)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}
	response.OK(w, out)
}

// me godoc
//
//	@Summary		Current profile
//	@Description	Returns the authenticated user's profile.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	types.Envelope{data=userResponse}
//	@Failure		401	{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/users/me [get]
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, ok := types.UserID(r.Context())
	if !ok {
		response.Error(w, h.log, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.svc.Me(r.Context(), id)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}
	response.OK(w, toUserResponse(user))
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
