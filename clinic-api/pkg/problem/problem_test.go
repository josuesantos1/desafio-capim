package problem_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/pkg/problem"
)

func TestWrite_NoFields(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.Write(rec, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", nil)

	require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	require.Equal(t, http.StatusNotFound, rec.Code)

	var got problem.Details
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, problem.Details{
		Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound,
		Detail: "clinic not found", Code: "CLINIC_NOT_FOUND",
	}, got)
	require.NotContains(t, rec.Body.String(), `"errors"`)
}

func TestWrite_WithFields_DeterministicOrder(t *testing.T) {
	fields := map[string]string{
		"trade_name": "must not be empty",
		"document":   "must be 11 or 14 digits",
		"legal_name": "must not be empty",
	}

	for range 5 {
		rec := httptest.NewRecorder()
		problem.Write(rec, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", fields)

		var got problem.Details
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, []problem.FieldError{
			{Field: "document", Detail: "must be 11 or 14 digits"},
			{Field: "legal_name", Detail: "must not be empty"},
			{Field: "trade_name", Detail: "must not be empty"},
		}, got.Errors)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.WriteJSON(rec, http.StatusCreated, map[string]string{"id": "abc"})

	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{"id":"abc"}`, rec.Body.String())
}
