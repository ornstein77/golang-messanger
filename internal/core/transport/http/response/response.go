// Package response centralizes writing JSON responses and errors.
package response

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	apperrors "messenger/internal/core/errors"
	"messenger/internal/core/transport/http/types"
)

// JSON writes data as a JSON envelope with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, types.Envelope{Data: data})
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// OK writes a 200 response.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error maps err to an HTTP status and writes the error envelope. Unexpected
// (non-*apperrors.Error) errors are logged and masked as internal errors.
func Error(w http.ResponseWriter, log *zap.Logger, err error) {
	status := apperrors.HTTPStatus(err)

	var appErr *apperrors.Error
	if !apperrors.As(err, &appErr) {
		appErr = apperrors.ErrInternal
		if log != nil {
			log.Error("unhandled error", zap.Error(err))
		}
	} else if status >= http.StatusInternalServerError && log != nil {
		log.Error("internal error", zap.Error(err))
	}

	write(w, status, types.Envelope{Error: &types.ErrorBody{
		Code:    appErr.Code,
		Message: appErr.Message,
	}})
}

// ValidationError writes a 400 response including per-field messages.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	write(w, http.StatusBadRequest, types.Envelope{Error: &types.ErrorBody{
		Code:    apperrors.ErrValidation.Code,
		Message: apperrors.ErrValidation.Message,
		Fields:  fields,
	}})
}

func write(w http.ResponseWriter, status int, env types.Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(env)
}
