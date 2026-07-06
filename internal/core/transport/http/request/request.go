// Package request decodes and validates incoming JSON request bodies.
package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"messenger/internal/core/transport/http/response"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var validate = validator.New(validator.WithRequiredStructEnabled())

// Bind decodes the JSON request body into dst and validates it using struct
// tags. On failure it writes the appropriate error response and returns false;
// on success it returns true.
func Bind(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			response.ValidationError(w, map[string]string{"body": "request body must not be empty"})
			return false
		}
		response.ValidationError(w, map[string]string{"body": "request body is not valid JSON"})
		return false
	}

	if err := validate.Struct(dst); err != nil {
		var invalid *validator.InvalidValidationError
		if errors.As(err, &invalid) {
			response.ValidationError(w, map[string]string{"body": "invalid request"})
			return false
		}
		response.ValidationError(w, fieldErrors(err))
		return false
	}

	return true
}

func fieldErrors(err error) map[string]string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string]string{"body": "invalid request"}
	}
	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		field := strings.ToLower(fe.Field())
		out[field] = describe(fe)
	}
	return out
}

func describe(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "e164":
		return "must be a valid phone number in E.164 format (e.g. +14155552671)"
	case "alphanum":
		return "must contain only letters and digits"
	case "uuid":
		return "must be a valid UUID"
	default:
		return "failed validation: " + fe.Tag()
	}
}
