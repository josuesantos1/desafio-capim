package payment_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/payment"
	"github.com/josuesantos1/desafio/mocks"
)

func newTestRouter(t *testing.T) (chi.Router, *mocks.PaymentRepository, *mocks.Repository, *mocks.DentistRepository) {
	t.Helper()
	repo := mocks.NewPaymentRepository(t)
	clinicRepo := mocks.NewRepository(t)
	dentistRepo := mocks.NewDentistRepository(t)
	svc := payment.NewService(repo, clinicRepo, dentistRepo, newFakePixProvider(),
		payment.WithApprovalDelay(func() time.Duration { return time.Hour })) // effectively disabled for handler tests
	r := chi.NewRouter()
	payment.RegisterRoutes(r, svc)
	return r, repo, clinicRepo, dentistRepo
}

func TestHandler_CreatePayment(t *testing.T) {
	const clinicID = "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name        string
		body        string
		setupRepo   func(repo *mocks.PaymentRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantStatus  int
		wantError   string
	}{
		{
			name: "success",
			body: `{"clinic_id":"` + clinicID + `","amount":15000}`,
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, clinicID).Return(clinic.Clinic{ID: clinicID}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:      "clinic not found",
			body:      `{"clinic_id":"` + clinicID + `","amount":15000}`,
			setupRepo: func(repo *mocks.PaymentRepository) {},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, clinicID).Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "CLINIC_NOT_FOUND",
		},
		{
			name:        "validation error",
			body:        `{"amount":0}`,
			setupRepo:   func(repo *mocks.PaymentRepository) {},
			setupClinic: func(clinicRepo *mocks.Repository) {},
			wantStatus:  http.StatusBadRequest,
			wantError:   "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, clinicRepo, _ := newTestRouter(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)

			req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}

func TestHandler_GetPayment(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupRepo  func(repo *mocks.PaymentRepository)
		wantStatus int
		wantError  string
	}{
		{
			name: "success",
			id:   "11111111-1111-1111-1111-111111111111",
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().GetByID(mock.Anything, "11111111-1111-1111-1111-111111111111").
					Return(payment.Payment{ID: "11111111-1111-1111-1111-111111111111", Status: payment.StatusPending}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "22222222-2222-2222-2222-222222222222",
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().GetByID(mock.Anything, "22222222-2222-2222-2222-222222222222").
					Return(payment.Payment{}, payment.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
			wantError:  "PAYMENT_NOT_FOUND",
		},
		{
			name:       "invalid id",
			id:         "not-a-uuid",
			setupRepo:  func(repo *mocks.PaymentRepository) {},
			wantStatus: http.StatusBadRequest,
			wantError:  "INVALID_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, repo, _, _ := newTestRouter(t)
			tt.setupRepo(repo)

			req := httptest.NewRequest(http.MethodGet, "/payments/"+tt.id, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code, "body=%s", rec.Body.String())
			if tt.wantError != "" {
				require.Contains(t, rec.Body.String(), tt.wantError)
			}
		})
	}
}
