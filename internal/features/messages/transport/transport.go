// Package transport exposes the messages feature over HTTP.
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
	"messenger/internal/core/transport/http/request"
	"messenger/internal/core/transport/http/response"
	"messenger/internal/core/transport/http/types"
)

// Service is the messaging use-case contract required by the HTTP handlers.
type Service interface {
	Send(ctx context.Context, sender, recipient uuid.UUID, body string) (domain.Message, error)
	Conversation(ctx context.Context, self, other uuid.UUID, limit, offset int) ([]domain.Message, error)
}

// Handler serves the messaging endpoints. All routes require authentication.
type Handler struct {
	svc  Service
	auth middleware.Middleware
	log  *zap.Logger
}

// New creates the messages Handler.
func New(svc Service, auth middleware.Middleware, log *zap.Logger) *Handler {
	return &Handler{svc: svc, auth: auth, log: log}
}

// Register attaches the messaging routes to the mux, protected by auth middleware.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/messages", h.auth(http.HandlerFunc(h.send)))
	mux.Handle("GET /api/v1/messages", h.auth(http.HandlerFunc(h.conversation)))
}

type sendRequest struct {
	RecipientID string `json:"recipient_id" validate:"required,uuid" example:"6f9619ff-8b86-d011-b42d-00c04fc964ff"`
	Body        string `json:"body" validate:"required,min=1,max=4000" example:"Hey there!"`
}

type messageResponse struct {
	ID          uuid.UUID `json:"id"`
	SenderID    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

func toMessageResponse(m domain.Message) messageResponse {
	return messageResponse{
		ID:          m.ID,
		SenderID:    m.SenderID,
		RecipientID: m.RecipientID,
		Body:        m.Body,
		CreatedAt:   m.CreatedAt,
	}
}

// send godoc
//
//	@Summary		Send a message
//	@Description	Sends a direct message to another user.
//	@Tags			messages
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		sendRequest	true	"Message payload"
//	@Success		201		{object}	types.Envelope{data=messageResponse}
//	@Failure		400		{object}	types.Envelope{error=types.ErrorBody}
//	@Failure		401		{object}	types.Envelope{error=types.ErrorBody}
//	@Failure		404		{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/messages [post]
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	sender, ok := types.UserID(r.Context())
	if !ok {
		response.Error(w, h.log, apperrors.ErrUnauthorized)
		return
	}

	var req sendRequest
	if !request.Bind(w, r, &req) {
		return
	}

	recipient, err := uuid.Parse(req.RecipientID)
	if err != nil {
		response.ValidationError(w, map[string]string{"recipient_id": "must be a valid UUID"})
		return
	}

	msg, err := h.svc.Send(r.Context(), sender, recipient, req.Body)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}
	response.Created(w, toMessageResponse(msg))
}

// conversation godoc
//
//	@Summary		Conversation history
//	@Description	Returns the messages exchanged with another user, oldest first.
//	@Tags			messages
//	@Produce		json
//	@Security		BearerAuth
//	@Param			with	query		string	true	"Other user's UUID"
//	@Param			limit	query		int		false	"Max results (default 50, max 200)"
//	@Param			offset	query		int		false	"Pagination offset"
//	@Success		200		{object}	types.Envelope{data=[]messageResponse}
//	@Failure		400		{object}	types.Envelope{error=types.ErrorBody}
//	@Failure		401		{object}	types.Envelope{error=types.ErrorBody}
//	@Router			/messages [get]
func (h *Handler) conversation(w http.ResponseWriter, r *http.Request) {
	self, ok := types.UserID(r.Context())
	if !ok {
		response.Error(w, h.log, apperrors.ErrUnauthorized)
		return
	}

	other, err := uuid.Parse(r.URL.Query().Get("with"))
	if err != nil {
		response.ValidationError(w, map[string]string{"with": "query parameter must be a valid UUID"})
		return
	}

	limit := atoiDefault(r.URL.Query().Get("limit"), 0)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)

	messages, err := h.svc.Conversation(r.Context(), self, other, limit, offset)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	out := make([]messageResponse, 0, len(messages))
	for _, m := range messages {
		out = append(out, toMessageResponse(m))
	}
	response.OK(w, out)
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
