// Package dentist_test — integration tests hitting the real router,
// Service, and in-memory repositories together (no mocks, including a
// real clinic.MemoryRepository dependency). This file has no build tag
// and requires no external infrastructure — unlike the historical
// internal/dentist/integration_test.go it replaces (removed during the
// in-memory storage migration), which was gated by //go:build
// integration and required a real Postgres.
package dentist_test

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
	"github.com/josuesantos1/desafio/internal/dentist"
)

const (
	integrationClinicID  = "11111111-1111-1111-1111-111111111111"
	integrationDentistID = "22222222-2222-2222-2222-222222222222"
)

func newIntegrationRouter(t *testing.T) (chi.Router, clinic.Repository, dentist.Repository) {
	t.Helper()
	clinicRepo := clinic.NewMemoryRepository()
	dentistRepo := dentist.NewMemoryRepository(clinicRepo)
	svc := dentist.NewService(dentistRepo, clinicRepo)
	r := chi.NewRouter()
	dentist.RegisterRoutes(r, svc)
	return r, clinicRepo, dentistRepo
}

func mustCreateClinic(t *testing.T, repo clinic.Repository, id string) clinic.Clinic {
	t.Helper()
	now := time.Now().UTC()
	c := clinic.Clinic{
		ID: id, Document: "12345678900", LegalName: "Legal", TradeName: "Trade",
		Status: clinic.StatusPending, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(t.Context(), c))
	return c
}

func mustCreateDentist(t *testing.T, repo dentist.Repository, clinicID, id, email string) dentist.Dentist {
	t.Helper()
	now := time.Now().UTC()
	d := dentist.Dentist{
		ID: id, ClinicID: clinicID, Name: "Dr. X", Phone: "123", Email: email,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(t.Context(), d))
	return d
}

// mustActivateClinic promotes clinicID to active via a single real
// PATCH .../roles call granting dentistID both flags — the real
// pending -> active transition (dentist.memoryRepository.UpdateRoles
// -> clinic.memoryRepository.Activate), exercised through the router
// under test, not reimplemented. dentistID becomes the clinic's sole
// administrator and legal representative.
func mustActivateClinic(t *testing.T, r chi.Router, clinicID, dentistID string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/clinics/"+clinicID+"/dentists/"+dentistID+"/roles",
		bytes.NewBufferString(`{"is_administrator":true,"is_legal_representative":true}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
}

func decodeJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(body, v))
}

// setupFunc seeds fixtures directly on the real repositories (and, for
// mustActivateClinic, through the router under test) before a case's
// request is issued.
type setupFunc func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository)

func TestIntegration_CreateDentist(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		body         string
		wantStatus   int
		wantContains string
	}{
		{
			name: "success",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			body:         `{"name":"Dr. X","phone":"123","email":"x@test.com"}`,
			wantStatus:   http.StatusCreated,
			wantContains: "x@test.com",
		},
		{
			name: "validation error",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			body:         `{"name":"Dr. X","phone":"123","email":""}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "VALIDATION_ERROR",
		},
		{
			name:         "clinic not found",
			body:         `{"name":"Dr. X","phone":"123","email":"x@test.com"}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name: "duplicate email",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			body:         `{"name":"Dr. Y","phone":"456","email":"x@test.com"}`,
			wantStatus:   http.StatusConflict,
			wantContains: "EMAIL_ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodPost, "/clinics/"+integrationClinicID+"/dentists", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_GetDentist(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		wantStatus   int
		wantContains string
	}{
		{
			name: "success",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			wantStatus:   http.StatusOK,
			wantContains: "x@test.com",
		},
		{
			name: "dentist not found",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			wantStatus:   http.StatusNotFound,
			wantContains: "DENTIST_NOT_FOUND",
		},
		{
			name:         "clinic not found",
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodGet, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_UpdateDentist(t *testing.T) {
	const otherDentistID = "33333333-3333-3333-3333-333333333333"

	tests := []struct {
		name            string
		setup           setupFunc
		body            string
		wantStatus      int
		wantContains    string
		wantGetContains string
	}{
		{
			name: "success reflected on subsequent GET",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			body:            `{"name":"Dr. Updated"}`,
			wantStatus:      http.StatusOK,
			wantGetContains: "Dr. Updated",
		},
		{
			name: "not found",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			body:         `{"name":"Dr. Updated"}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "DENTIST_NOT_FOUND",
		},
		{
			name:         "clinic not found",
			body:         `{"name":"Dr. Updated"}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name: "duplicate email",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, otherDentistID, "taken@test.com")
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			body:         `{"email":"taken@test.com"}`,
			wantStatus:   http.StatusConflict,
			wantContains: "EMAIL_ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodPut, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantContains != "" {
				require.Contains(t, rec.Body.String(), tt.wantContains)
			}

			if tt.wantGetContains != "" {
				getReq := httptest.NewRequest(http.MethodGet, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID, nil)
				getRec := httptest.NewRecorder()
				r.ServeHTTP(getRec, getReq)
				require.Contains(t, getRec.Body.String(), tt.wantGetContains)
			}
		})
	}
}

func TestIntegration_UpdateDentistRoles(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		body         string
		wantStatus   int
		wantContains string
	}{
		{
			name: "success while clinic is pending",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			body:         `{"is_administrator":true}`,
			wantStatus:   http.StatusOK,
			wantContains: `"is_administrator":true`,
		},
		{
			name: "last administrator required once clinic is active",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
				mustActivateClinic(t, r, integrationClinicID, integrationDentistID)
			},
			body:         `{"is_administrator":false}`,
			wantStatus:   http.StatusConflict,
			wantContains: "LAST_ADMIN_REQUIRED",
		},
		{
			name:         "clinic not found",
			body:         `{"is_administrator":true}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name: "empty body validation error",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			body:         `{}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodPatch, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID+"/roles", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_DeleteDentist(t *testing.T) {
	tests := []struct {
		name            string
		setup           setupFunc
		wantStatus      int
		wantContains    string
		wantGetNotFound bool
	}{
		{
			name: "success then GET returns not found",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
			},
			wantStatus:      http.StatusNoContent,
			wantGetNotFound: true,
		},
		{
			name: "last administrator of active clinic cannot be deleted",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "x@test.com")
				mustActivateClinic(t, r, integrationClinicID, integrationDentistID)
			},
			wantStatus:   http.StatusConflict,
			wantContains: "LAST_ADMIN_REQUIRED",
		},
		{
			name:         "clinic not found",
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodDelete, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantContains != "" {
				require.Contains(t, rec.Body.String(), tt.wantContains)
			}

			if tt.wantGetNotFound {
				getReq := httptest.NewRequest(http.MethodGet, "/clinics/"+integrationClinicID+"/dentists/"+integrationDentistID, nil)
				getRec := httptest.NewRecorder()
				r.ServeHTTP(getRec, getReq)
				require.Equal(t, http.StatusNotFound, getRec.Code)
			}
		})
	}
}

func TestIntegration_ListDentists(t *testing.T) {
	const otherDentistID = "33333333-3333-3333-3333-333333333333"

	tests := []struct {
		name         string
		setup        setupFunc
		path         string
		wantStatus   int
		wantContains string
		wantTotal    *int
	}{
		{
			name: "filters by is_administrator",
			setup: func(t *testing.T, r chi.Router, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
				mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID, "a@test.com")
				mustCreateDentist(t, dentistRepo, integrationClinicID, otherDentistID, "b@test.com")
				mustActivateClinic(t, r, integrationClinicID, integrationDentistID)
			},
			path:       "/clinics/" + integrationClinicID + "/dentists?is_administrator=true",
			wantStatus: http.StatusOK,
			wantTotal:  intPtr(1),
		},
		{
			name:         "clinic not found",
			path:         "/clinics/" + integrationClinicID + "/dentists",
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, r, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantContains != "" {
				require.Contains(t, rec.Body.String(), tt.wantContains)
			}
			if tt.wantTotal != nil {
				var got struct {
					Total int `json:"total"`
				}
				decodeJSON(t, rec.Body.Bytes(), &got)
				require.Equal(t, *tt.wantTotal, got.Total)
			}
		})
	}
}

func intPtr(v int) *int { return &v }
