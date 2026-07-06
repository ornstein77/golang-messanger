package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperrors "messenger/internal/core/errors"
	"messenger/internal/core/transport/http/response"
	"messenger/internal/core/transport/http/types"
)

// Authenticator resolves a bearer token into a user id. It is implemented by
// the auth service, keeping the middleware decoupled from feature packages.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (uuid.UUID, error)
}

// Auth validates the Authorization bearer token and injects the user id into
// the request context. Requests without a valid token are rejected with 401.
func Auth(authenticator Authenticator, log *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				response.Error(w, log, apperrors.ErrUnauthorized)
				return
			}

			userID, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				response.Error(w, log, err)
				return
			}

			ctx := types.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}
