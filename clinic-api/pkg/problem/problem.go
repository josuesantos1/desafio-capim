// Package problem implements RFC 9457 Problem Details for HTTP APIs.
package problem

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
)

// Details is the RFC 9457 (application/problem+json) error body returned by
// every failed request in this API. Code is stable and safe to switch on
// programmatically; Detail is human-readable and may change.
type Details struct {
	Type   string       `json:"type" example:"about:blank"`
	Title  string       `json:"title" example:"Bad Request"`
	Status int          `json:"status" example:"400"`
	Detail string       `json:"detail,omitempty" example:"request validation failed"`
	Code   string       `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Errors []FieldError `json:"errors,omitempty"`
}

// FieldError describes one invalid field, present when Details.Code is
// "VALIDATION_ERROR".
type FieldError struct {
	Field  string `json:"field" example:"clinic_id"`
	Detail string `json:"detail" example:"is required"`
}

func Write(w http.ResponseWriter, status int, code, detail string, fields map[string]string) {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var errs []FieldError
	for _, k := range keys {
		errs = append(errs, FieldError{Field: k, Detail: fields[k]})
	}
	d := Details{
		Type: "about:blank", Title: http.StatusText(status), Status: status,
		Detail: detail, Code: code, Errors: errs,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(d); err != nil {
		slog.Error("failed to encode problem details response", "error", err)
	}
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode json response", "error", err)
	}
}
