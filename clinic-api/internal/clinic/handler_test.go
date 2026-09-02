package clinic_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/mocks"
)

func newTestRouter(t *testing.T) (chi.Router, *mocks.Repository) {
	t.Helper()
	repo := mocks.NewRepository(t)
	svc := clinic.NewService(repo)
	r := chi.NewRouter()
	clinic.RegisterRoutes(r, svc)
	return r, repo
}

func TestHandler_CreateClinic(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupRepo  func(repo *mocks.Repository)
		wantStatus int
		wantError  string
	}{
		{
			name: "success",
			body: `{"document":"12345678900","legal_name":"Legal","trade_name":"Trade"}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "validation error",
			body:       `{"document":"123"}`,
			setupRepo:  func(repo *mocks.Repository) {},
			wantStatus: http.StatusBadRequest,
			wantError:  "VALIDATION_ERROR",
		},
		{
			name: "duplicate document",
			body: `{"document":"12345678900","legal_name":"Legal","trade_name":"Trade"}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(clinic.ErrDocumentExists).Once()
			},
			wantStatus: http.StatusConflict,
			wantError:  "DOCUMENT_ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newTestRouter(t)
			tt.setupRepo(repo)

			req := httptest.NewRequest(http.MethodPost, "/clinics", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_GetClinic(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupRepo  func(repo *mocks.Repository)
		wantStatus int
		wantError  string
	}{
		{
			name: "success",
			id:   "00000000-0000-0000-0000-000000000001",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000001").
					Return(clinic.Clinic{ID: "00000000-0000-0000-0000-000000000001"}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "00000000-0000-0000-0000-000000000000",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000000").
					Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "CLINIC_NOT_FOUND",
		},
		{
			name:       "invalid id",
			id:         "not-a-uuid",
			setupRepo:  func(repo *mocks.Repository) {},
			wantStatus: http.StatusBadRequest,
			wantError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newTestRouter(t)
			tt.setupRepo(repo)

			req := httptest.NewRequest(http.MethodGet, "/clinics/"+tt.id, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_UpdateClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"
	current := clinic.Clinic{ID: id, Document: "12345678900", LegalName: "Legal", TradeName: "Trade"}

	tests := []struct {
		name       string
		body       string
		setupRepo  func(repo *mocks.Repository)
		wantStatus int
		wantError  string
	}{
		{
			name: "success",
			body: `{"trade_name":"New Trade"}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty body is no-op",
			body: `{}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "document immutable",
			body: `{"document":"98765432100"}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(current, nil).Once()
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "DOCUMENT_IMMUTABLE",
		},
		{
			name: "not found",
			body: `{}`,
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newTestRouter(t)
			tt.setupRepo(repo)

			req := httptest.NewRequest(http.MethodPut, "/clinics/"+id, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_DeleteClinic(t *testing.T) {
	const id = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name       string
		setupRepo  func(repo *mocks.Repository)
		wantStatus int
		wantError  string
	}{
		{
			name: "success",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().SoftDelete(mock.Anything, id, mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().SoftDelete(mock.Anything, id, mock.AnythingOfType("time.Time")).Return(clinic.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "CLINIC_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo := newTestRouter(t)
			tt.setupRepo(repo)

			req := httptest.NewRequest(http.MethodDelete, "/clinics/"+id, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}
