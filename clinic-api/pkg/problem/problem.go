// Package problem implements RFC 9457 Problem Details for HTTP APIs.
package problem

import (
	"encoding/json"
	"net/http"
	"sort"
)

type Details struct {
	Type   string       `json:"type"`
	Title  string       `json:"title"`
	Status int          `json:"status"`
	Detail string       `json:"detail,omitempty"`
	Code   string       `json:"code,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field  string `json:"field"`
	Detail string `json:"detail"`
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
	json.NewEncoder(w).Encode(d)
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
