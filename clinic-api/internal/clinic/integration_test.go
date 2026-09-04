// Package clinic_test — integration tests hitting the real router,
// Service, and in-memory repository together (no mocks). This file
// has no build tag and requires no external infrastructure — unlike
// the historical internal/dentist/integration_test.go (removed during
// the in-memory storage migration), which was gated by //go:build
// integration and required a real Postgres.
package clinic_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
)

func decodeJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(body, v))
}

func newIntegrationRouter(t *testing.T) (chi.Router, clinic.Repository) {
	t.Helper()
	repo := clinic.NewMemoryRepository()
	svc := clinic.NewService(repo)
	r := chi.NewRouter()
	clinic.RegisterRoutes(r, svc)
	return r, repo
}

func mustCreateClinic(t *testing.T, repo clinic.Repository, id, document string) clinic.Clinic {
	t.Helper()
	now := time.Now().UTC()
	c := clinic.Clinic{
		ID: id, Document: document, LegalName: "Legal", TradeName: "Trade",
		Status: clinic.StatusPending, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(t.Context(), c))
	return c
}

func TestIntegration_CreateClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name          string
		setup         func(t *testing.T, repo clinic.Repository)
		body          string
		wantStatus    int
		wantContains  string
		checkPersists bool
	}{
		{
			name:          "success",
			body:          `{"document":"12345678900","legal_name":"Legal","trade_name":"Trade"}`,
			wantStatus:    http.StatusCreated,
			wantContains:  `"status":"pending"`,
			checkPersists: true,
		},
		{
			name:         "validation error",
			body:         `{"document":"123"}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "VALIDATION_ERROR",
		},
		{
			name: "duplicate document",
			setup: func(t *testing.T, repo clinic.Repository) {
				mustCreateClinic(t, repo, id, "12345678900")
			},
			body:         `{"document":"12345678900","legal_name":"Legal","trade_name":"Trade"}`,
			wantStatus:   http.StatusConflict,
			wantContains: "DOCUMENT_ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			req := httptest.NewRequest(http.MethodPost, "/clinics", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)

			if tt.checkPersists {
				var created struct {
					ID string `json:"id"`
				}
				decodeJSON(t, rec.Body.Bytes(), &created)

				got, err := repo.GetByID(t.Context(), created.ID)
				require.NoError(t, err)
				require.Equal(t, "12345678900", got.Document)
			}
		})
	}
}

func TestIntegration_GetClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name         string
		setup        func(t *testing.T, repo clinic.Repository)
		path         string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "success",
			setup:        func(t *testing.T, repo clinic.Repository) { mustCreateClinic(t, repo, id, "12345678900") },
			path:         "/clinics/" + id,
			wantStatus:   http.StatusOK,
			wantContains: "12345678900",
		},
		{
			name:         "not found",
			path:         "/clinics/" + id,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name:         "invalid id",
			path:         "/clinics/not-a-uuid",
			wantStatus:   http.StatusBadRequest,
			wantContains: "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_UpdateClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name            string
		setup           func(t *testing.T, repo clinic.Repository)
		body            string
		wantStatus      int
		wantContains    string
		wantGetContains string
	}{
		{
			name:            "success reflected on subsequent GET",
			setup:           func(t *testing.T, repo clinic.Repository) { mustCreateClinic(t, repo, id, "12345678900") },
			body:            `{"trade_name":"New Trade"}`,
			wantStatus:      http.StatusOK,
			wantGetContains: "New Trade",
		},
		{
			name:       "empty body is no-op",
			setup:      func(t *testing.T, repo clinic.Repository) { mustCreateClinic(t, repo, id, "12345678900") },
			body:       `{}`,
			wantStatus: http.StatusOK,
		},
		{
			name:         "document immutable",
			setup:        func(t *testing.T, repo clinic.Repository) { mustCreateClinic(t, repo, id, "12345678900") },
			body:         `{"document":"98765432100"}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "DOCUMENT_IMMUTABLE",
		},
		{
			name:         "not found",
			body:         `{}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			req := httptest.NewRequest(http.MethodPut, "/clinics/"+id, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantContains != "" {
				require.Contains(t, rec.Body.String(), tt.wantContains)
			}

			if tt.wantGetContains != "" {
				getReq := httptest.NewRequest(http.MethodGet, "/clinics/"+id, nil)
				getRec := httptest.NewRecorder()
				r.ServeHTTP(getRec, getReq)
				require.Contains(t, getRec.Body.String(), tt.wantGetContains)
			}
		})
	}
}

func TestIntegration_DeleteClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name            string
		setup           func(t *testing.T, repo clinic.Repository)
		wantStatus      int
		wantContains    string
		wantGetNotFound bool
	}{
		{
			name:            "success then GET returns not found",
			setup:           func(t *testing.T, repo clinic.Repository) { mustCreateClinic(t, repo, id, "12345678900") },
			wantStatus:      http.StatusNoContent,
			wantGetNotFound: true,
		},
		{
			name:         "not found",
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			req := httptest.NewRequest(http.MethodDelete, "/clinics/"+id, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantContains != "" {
				require.Contains(t, rec.Body.String(), tt.wantContains)
			}

			if tt.wantGetNotFound {
				getReq := httptest.NewRequest(http.MethodGet, "/clinics/"+id, nil)
				getRec := httptest.NewRecorder()
				r.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusNotFound, getRec.Code)
				require.Contains(t, getRec.Body.String(), "CLINIC_NOT_FOUND")
			}
		})
	}
}
