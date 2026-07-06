// Package transport exposes the auth feature over HTTP.
package transport

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"messenger/internal/core/domain"
	"messenger/internal/core/transport/http/middleware"
	"messenger/internal/core/transport/http/request"
	"messenger/internal/core/transport/http/response"
	authservice "messenger/internal/features/auth/service"
)

// Service is the auth use-case contract required by the HTTP handlers.
type Service interface {
	Register(ctx context.Context, in authservice.RegisterInput) (domain.User, error)
	Login(ctx context.Context, nickname, password string) (domain.Session, domain.User, error)
	Logout(ctx context.Context, rawToken string) error
}

// Handler serves the auth endpoints.
type Handler struct {
	svc  Service
	auth middleware.Middleware
	log  *zap.Logger
}

// New creates the auth Handler. The auth middleware is used to protect logout.
func New(svc Service, auth middleware.Middleware, log *zap.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

// Register attaches the auth routes to the mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.Handle("POST /api/v1/auth/logout", h.auth(http.HandlerFunc(h.logout)))
}

// registerRequest is the registration payload.
type registerRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=32,alphanum" example:"john_doe"`
	Password string `json:"password" validate:"required,min=6,max=72" example:"s3cr3t-pass"`
	Phone    string `json:"phone" validate:"required,e164" example:"+14155552671"`
}

// userResponse is the public representation of a user.
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// sessionResponse is returned on successful login.
type sessionResponse struct {
	Token     uuid.UUID    `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userResponse `json:"user"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{ID: u.ID, Nickname: u.Nickname, Phone: u.Phone, CreatedAt: u.CreatedAt}
}

// register godoc
//
//	@Summary		Register a new user
//	@Description	Creates an account with a nickname, password and phone number.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		registerRequest	true	"Registration data"
//	@Success		201		{object}	types.Envelope{data=userResponse}
//	@Failure		400		{object}	types.Envelope{error=types.ErrorBody}
//	@Failure		409		{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/auth/register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !request.Bind(w, r, &req) {
		return
	}

	user, err := h.svc.Register(r.Context(), authservice.RegisterInput{
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		response.Error(w, h.log, err)
		return
	}
	response.Created(w, toUserResponse(user))
}

// loginRequest is the login payload.
type loginRequest struct {
	Nickname string `json:"nickname" validate:"required" example:"john_doe"`
	Password string `json:"password" validate:"required" example:"s3cr3t-pass"`
}

// login godoc
//
//	@Summary		Log in
//	@Description	Verifies credentials and returns a bearer session token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		loginRequest	true	"Login data"
//	@Success		200		{object}	types.Envelope{data=sessionResponse}
//	@Failure		400		{object}	types.Envelope{error=types.ErrorBody}
//	@Failure		401		{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/auth/login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !request.Bind(w, r, &req) {
		return
	}

	session, user, err := h.svc.Login(r.Context(), req.Nickname, req.Password)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}
	response.OK(w, sessionResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		User:      toUserResponse(user),
	})
}

// logout godoc
//
//	@Summary		Log out
//	@Description	Revokes the current session token.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204	"No Content"
//	@Failure		401	{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/auth/logout [post]
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if err := h.svc.Logout(r.Context(), token); err != nil {
		response.Error(w, h.log, err)
		return
	}
	response.NoContent(w)
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
