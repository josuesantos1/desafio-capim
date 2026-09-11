// Package payment_test — integration tests hitting the real router,
// Service, and in-memory repositories together (no mocks, including
// real clinic/dentist repository dependencies). This file has no
// build tag and requires no external infrastructure — unlike the
// historical internal/dentist/integration_test.go (removed during the
// in-memory storage migration), which was gated by //go:build
// integration and required a real Postgres.
//
// newFakePixProvider is defined in service_test.go (same package) —
// reused here, not redefined. Every service in this file is built
// with WithApprovalDelay(1h) so the real background approval
// goroutine never fires during the test run, avoiding flakiness and
// goroutine leaks under `go test -race`.
package payment_test

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
	"github.com/josuesantos1/desafio/internal/payment"
)

const (
	integrationClinicID  = "11111111-1111-1111-1111-111111111111"
	integrationDentistID = "22222222-2222-2222-2222-222222222222"
	integrationIdemKey   = "integration-idem-key"
)

func newIntegrationRouter(t *testing.T) (chi.Router, clinic.Repository, dentist.Repository, payment.Repository) {
	t.Helper()
	clinicRepo := clinic.NewMemoryRepository()
	dentistRepo := dentist.NewMemoryRepository(clinicRepo)
	paymentRepo := payment.NewMemoryRepository(clinicRepo, dentistRepo)
	svc := payment.NewService(paymentRepo, clinicRepo, dentistRepo, newFakePixProvider(),
		payment.WithApprovalDelay(func() time.Duration { return time.Hour }))
	r := chi.NewRouter()
	payment.RegisterRoutes(r, svc)
	return r, clinicRepo, dentistRepo, paymentRepo
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

func mustCreateDentist(t *testing.T, repo dentist.Repository, clinicID, id string) dentist.Dentist {
	t.Helper()
	now := time.Now().UTC()
	d := dentist.Dentist{
		ID: id, ClinicID: clinicID, Name: "Dr. X", Phone: "123", Email: id + "@test.com",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(t.Context(), d))
	return d
}

// mustActivateClinic promotes clinicID to active by granting dentistID
// both role flags via a real repository call — the real pending ->
// active transition, not reimplemented here.
func mustActivateClinic(t *testing.T, dentistRepo dentist.Repository, clinicID, dentistID string) {
	t.Helper()
	yes := true
	_, err := dentistRepo.UpdateRoles(t.Context(), clinicID, dentistID, dentist.RolesInput{
		IsAdministrator: &yes, IsLegalRepresentative: &yes,
	})
	require.NoError(t, err)
}

func decodeJSONBody(body []byte, v any) error {
	return json.Unmarshal(body, v)
}

func postPayment(r chi.Router, body, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(body))
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// setupFunc seeds fixtures directly on the real repositories before a
// case's request is issued.
type setupFunc func(t *testing.T, clinicRepo clinic.Repository, dentistRepo dentist.Repository)

func activeClinicWithDentist(t *testing.T, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
	mustCreateClinic(t, clinicRepo, integrationClinicID)
	mustCreateDentist(t, dentistRepo, integrationClinicID, integrationDentistID)
	mustActivateClinic(t, dentistRepo, integrationClinicID, integrationDentistID)
}

func TestIntegration_CreatePayment(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		body         string
		skipHeader   bool
		wantStatus   int
		wantContains string
	}{
		{
			name:         "success without dentist",
			setup:        activeClinicWithDentist,
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000}`,
			wantStatus:   http.StatusCreated,
			wantContains: `"status":"pending"`,
		},
		{
			name:         "success with dentist",
			setup:        activeClinicWithDentist,
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000,"dentist_id":"` + integrationDentistID + `"}`,
			wantStatus:   http.StatusCreated,
			wantContains: `"status":"pending"`,
		},
		{
			name:         "missing idempotency key header",
			setup:        activeClinicWithDentist,
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000}`,
			skipHeader:   true,
			wantStatus:   http.StatusBadRequest,
			wantContains: "IDEMPOTENCY_KEY_REQUIRED",
		},
		{
			name:         "validation error",
			body:         `{"clinic_id":"","amount":0}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "VALIDATION_ERROR",
		},
		{
			name:         "clinic not found",
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name: "clinic not active",
			setup: func(t *testing.T, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000}`,
			wantStatus:   http.StatusConflict,
			wantContains: "CLINIC_NOT_ACTIVE",
		},
		{
			name:         "dentist not found",
			setup:        activeClinicWithDentist,
			body:         `{"clinic_id":"` + integrationClinicID + `","amount":1000,"dentist_id":"33333333-3333-3333-3333-333333333333"}`,
			wantStatus:   http.StatusNotFound,
			wantContains: "DENTIST_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, clinicRepo, dentistRepo)
			}

			key := integrationIdemKey
			if tt.skipHeader {
				key = ""
			}
			rec := postPayment(r, tt.body, key)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_CreatePayment_Replay(t *testing.T) {
	r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
	activeClinicWithDentist(t, clinicRepo, dentistRepo)
	body := `{"clinic_id":"` + integrationClinicID + `","amount":1000}`

	first := postPayment(r, body, integrationIdemKey)
	require.Equal(t, http.StatusCreated, first.Code, "body=%s", first.Body.String())
	var firstPayment struct {
		ID string `json:"id"`
	}
	require.NoError(t, decodeJSONBody(first.Body.Bytes(), &firstPayment))

	second := postPayment(r, body, integrationIdemKey)
	require.Equal(t, http.StatusOK, second.Code, "body=%s", second.Body.String())
	var secondPayment struct {
		ID string `json:"id"`
	}
	require.NoError(t, decodeJSONBody(second.Body.Bytes(), &secondPayment))
	require.Equal(t, firstPayment.ID, secondPayment.ID, "replay must return the original payment, not a new one")
}

func TestIntegration_CreatePayment_Conflict(t *testing.T) {
	r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
	activeClinicWithDentist(t, clinicRepo, dentistRepo)

	first := postPayment(r, `{"clinic_id":"`+integrationClinicID+`","amount":1000}`, integrationIdemKey)
	require.Equal(t, http.StatusCreated, first.Code, "body=%s", first.Body.String())

	second := postPayment(r, `{"clinic_id":"`+integrationClinicID+`","amount":9999}`, integrationIdemKey)
	require.Equal(t, http.StatusConflict, second.Code, "body=%s", second.Body.String())
	require.Contains(t, second.Body.String(), "IDEMPOTENCY_KEY_CONFLICT")
}

func TestIntegration_GetPayment(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		createFirst  bool
		path         string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "success",
			setup:        activeClinicWithDentist,
			createFirst:  true,
			wantStatus:   http.StatusOK,
			wantContains: `"status":"pending"`,
		},
		{
			name:         "not found",
			path:         "/payments/00000000-0000-0000-0000-000000000001",
			wantStatus:   http.StatusNotFound,
			wantContains: "PAYMENT_NOT_FOUND",
		},
		{
			name:         "invalid id",
			path:         "/payments/not-a-uuid",
			wantStatus:   http.StatusBadRequest,
			wantContains: "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, clinicRepo, dentistRepo)
			}

			path := tt.path
			if tt.createFirst {
				createRec := postPayment(r, `{"clinic_id":"`+integrationClinicID+`","amount":1000}`, integrationIdemKey)
				require.Equal(t, http.StatusCreated, createRec.Code, "body=%s", createRec.Body.String())

				var created struct {
					ID string `json:"id"`
				}
				require.NoError(t, decodeJSONBody(createRec.Body.Bytes(), &created))
				path = "/payments/" + created.ID
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}

func TestIntegration_ListPayments(t *testing.T) {
	tests := []struct {
		name         string
		setup        setupFunc
		query        string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "clinic_id missing",
			query:        "",
			wantStatus:   http.StatusBadRequest,
			wantContains: "VALIDATION_ERROR",
		},
		{
			name:         "clinic_id malformed",
			query:        "?clinic_id=not-a-uuid",
			wantStatus:   http.StatusBadRequest,
			wantContains: "INVALID_ID",
		},
		{
			name:         "clinic not found",
			query:        "?clinic_id=" + integrationClinicID,
			wantStatus:   http.StatusNotFound,
			wantContains: "CLINIC_NOT_FOUND",
		},
		{
			name: "pending clinic without payments returns empty list, not an error",
			setup: func(t *testing.T, clinicRepo clinic.Repository, dentistRepo dentist.Repository) {
				mustCreateClinic(t, clinicRepo, integrationClinicID)
			},
			query:        "?clinic_id=" + integrationClinicID,
			wantStatus:   http.StatusOK,
			wantContains: `"items":[],"total":0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
			if tt.setup != nil {
				tt.setup(t, clinicRepo, dentistRepo)
			}

			req := httptest.NewRequest(http.MethodGet, "/payments"+tt.query, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}

	t.Run("success with status filter and pagination", func(t *testing.T) {
		r, clinicRepo, dentistRepo, _ := newIntegrationRouter(t)
		activeClinicWithDentist(t, clinicRepo, dentistRepo)

		rec1 := postPayment(r, `{"clinic_id":"`+integrationClinicID+`","amount":1000}`, "idem-1")
		require.Equal(t, http.StatusCreated, rec1.Code, "body=%s", rec1.Body.String())
		rec2 := postPayment(r, `{"clinic_id":"`+integrationClinicID+`","amount":2000}`, "idem-2")
		require.Equal(t, http.StatusCreated, rec2.Code, "body=%s", rec2.Body.String())

		req := httptest.NewRequest(http.MethodGet, "/payments?clinic_id="+integrationClinicID+"&status=pending", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
		require.Contains(t, rec.Body.String(), `"total":2`)

		reqPage := httptest.NewRequest(http.MethodGet, "/payments?clinic_id="+integrationClinicID+"&limit=1", nil)
		recPage := httptest.NewRecorder()
		r.ServeHTTP(recPage, reqPage)
		require.Equal(t, http.StatusOK, recPage.Code, "body=%s", recPage.Body.String())
		require.Contains(t, recPage.Body.String(), `"limit":1`)
	})
}
