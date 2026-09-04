package dentist_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/mocks"
)

const handlerClinicID = "11111111-1111-1111-1111-111111111111"
const handlerDentistID = "22222222-2222-2222-2222-222222222222"

func newTestRouter(t *testing.T) (chi.Router, *mocks.DentistRepository, *mocks.Repository) {
	t.Helper()
	repo := mocks.NewDentistRepository(t)
	clinicRepo := mocks.NewRepository(t)
	svc := dentist.NewService(repo, clinicRepo)
	r := chi.NewRouter()
	dentist.RegisterRoutes(r, svc)
	return r, repo, clinicRepo
}

func expectClinicActiveHandler(clinicRepo *mocks.Repository) {
	clinicRepo.EXPECT().GetByID(mock.Anything, handlerClinicID).Return(clinic.Clinic{ID: handlerClinicID}, nil).Once()
}

func expectClinicNotFoundHandler(clinicRepo *mocks.Repository) {
	clinicRepo.EXPECT().GetByID(mock.Anything, handlerClinicID).Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
}

func TestHandler_CreateDentist(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			body: `{"name":"Dr. A","phone":"123","email":"a@x.com"}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "clinic not found",
			body:        `{"name":"Dr. A","phone":"123","email":"a@x.com"}`,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "CLINIC_NOT_FOUND",
		},
		{
			name:        "validation error",
			body:        `{"name":""}`,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusBadRequest,
			wantError:   "VALIDATION_ERROR",
		},
		{
			name: "email exists",
			body: `{"name":"Dr. A","phone":"123","email":"a@x.com"}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(dentist.ErrEmailExists).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusConflict,
			wantError:   "EMAIL_ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists", handlerClinicID)
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_GetDentist(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			id:   handlerDentistID,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, handlerClinicID, handlerDentistID).
					Return(dentist.Dentist{ID: handlerDentistID, ClinicID: handlerClinicID}, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "clinic not found",
			id:          handlerDentistID,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "CLINIC_NOT_FOUND",
		},
		{
			name: "dentist not found",
			id:   handlerDentistID,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, handlerClinicID, handlerDentistID).
					Return(dentist.Dentist{}, dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "DENTIST_NOT_FOUND",
		},
		{
			name:        "invalid id",
			id:          "not-a-uuid",
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: func(clinicRepo *mocks.Repository) {},
			wantStatus:  http.StatusBadRequest,
			wantError:   "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists/%s", handlerClinicID, tt.id)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_UpdateDentist(t *testing.T) {
	current := dentist.Dentist{ID: handlerDentistID, ClinicID: handlerClinicID, Name: "Old", Phone: "1", Email: "a@x.com"}

	tests := []struct {
		name        string
		body        string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			body: `{"name":"New"}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, handlerClinicID, handlerDentistID).Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
		},
		{
			name: "empty body is no-op",
			body: `{}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, handlerClinicID, handlerDentistID).Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "clinic not found",
			body:        `{}`,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "CLINIC_NOT_FOUND",
		},
		{
			name: "not found",
			body: `{}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, handlerClinicID, handlerDentistID).Return(dentist.Dentist{}, dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "DENTIST_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists/%s", handlerClinicID, handlerDentistID)
			req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_UpdateRoles(t *testing.T) {
	updated := dentist.Dentist{ID: handlerDentistID, ClinicID: handlerClinicID, IsAdministrator: true}

	tests := []struct {
		name        string
		body        string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			body: `{"is_administrator":true}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().UpdateRoles(mock.Anything, handlerClinicID, handlerDentistID, mock.Anything).
					Return(updated, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "validation error — empty body",
			body:        `{}`,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusBadRequest,
			wantError:   "VALIDATION_ERROR",
		},
		{
			name:        "clinic not found",
			body:        `{"is_administrator":true}`,
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "CLINIC_NOT_FOUND",
		},
		{
			name: "last admin required",
			body: `{"is_administrator":false}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().UpdateRoles(mock.Anything, handlerClinicID, handlerDentistID, mock.Anything).
					Return(dentist.Dentist{}, dentist.ErrLastAdminRequired).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusConflict,
			wantError:   "LAST_ADMIN_REQUIRED",
		},
		{
			name: "last legal representative required",
			body: `{"is_legal_representative":false}`,
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().UpdateRoles(mock.Anything, handlerClinicID, handlerDentistID, mock.Anything).
					Return(dentist.Dentist{}, dentist.ErrLastLegalRepresentativeRequired).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusConflict,
			wantError:   "LAST_LEGAL_REPRESENTATIVE_REQUIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists/%s/roles", handlerClinicID, handlerDentistID)
			req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_DeleteDentist(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().SoftDelete(mock.Anything, handlerClinicID, handlerDentistID, mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusNoContent,
		},
		{
			name:        "clinic not found",
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "CLINIC_NOT_FOUND",
		},
		{
			name: "not found",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().SoftDelete(mock.Anything, handlerClinicID, handlerDentistID, mock.AnythingOfType("time.Time")).Return(dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusNotFound,
			wantError:   "DENTIST_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists/%s", handlerClinicID, handlerDentistID)
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_ListDentists(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantBody    string
	}{
		{
			name:  "success",
			query: "",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, handlerClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.Limit == 20 && p.Offset == 0
				})).Return(dentist.ListResult{Items: []dentist.Dentist{}, Total: 0}, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
			wantBody:    `"total":0`,
		},
		{
			name:        "clinic not found",
			query:       "",
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFoundHandler,
			wantStatus:  http.StatusNotFound,
			wantBody:    "CLINIC_NOT_FOUND",
		},
		{
			name:  "clamps limit above max",
			query: "?limit=1000",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, handlerClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.Limit == 100
				})).Return(dentist.ListResult{}, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
			wantBody:    `"limit":100`,
		},
		{
			name:  "filters by is_administrator",
			query: "?is_administrator=true",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, handlerClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.IsAdministrator != nil && *p.IsAdministrator == true && p.IsLegalRepresentative == nil
				})).Return(dentist.ListResult{}, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
			wantBody:    `"total":0`,
		},
		{
			name:  "invalid filter value is treated as absent",
			query: "?is_administrator=not-a-bool",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, handlerClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.IsAdministrator == nil
				})).Return(dentist.ListResult{}, nil).Once()
			},
			setupClinic: expectClinicActiveHandler,
			wantStatus:  http.StatusOK,
			wantBody:    `"total":0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			path := fmt.Sprintf("/clinics/%s/dentists%s", handlerClinicID, tt.query)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			require.Contains(t, rec.Body.String(), tt.wantBody)
		})
	}
}
