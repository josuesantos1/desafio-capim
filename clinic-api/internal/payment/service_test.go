package payment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/internal/payment"
	"github.com/josuesantos1/desafio/mocks"
	"github.com/josuesantos1/desafio/pkg/pix"
)

const (
	testClinicID  = "11111111-1111-1111-1111-111111111111"
	testDentistID = "22222222-2222-2222-2222-222222222222"
)

// fakePixProvider is a hand-written fake — PixProvider has a single
// method, so a mockery-generated mock would be pure boilerplate here.
type fakePixProvider struct {
	response pix.ChargeResponse
	err      error
}

func (f *fakePixProvider) CreateCharge(ctx context.Context, req pix.ChargeRequest) (pix.ChargeResponse, error) {
	return f.response, f.err
}

func newFakePixProvider() *fakePixProvider {
	return &fakePixProvider{response: pix.ChargeResponse{CopyPasteCode: "00020126fake=="}}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name              string
		in                payment.CreateInput
		setupRepo         func(repo *mocks.PaymentRepository)
		setupClinic       func(clinicRepo *mocks.Repository)
		setupDentist      func(dentistRepo *mocks.DentistRepository)
		wantErr           error
		wantValidationErr bool
	}{
		{
			name: "success without dentist",
			in:   payment.CreateInput{ClinicID: testClinicID, Amount: 15000},
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{ID: testClinicID}, nil).Once()
			},
			setupDentist: func(dentistRepo *mocks.DentistRepository) {},
		},
		{
			name: "success with dentist",
			in:   payment.CreateInput{ClinicID: testClinicID, Amount: 15000, DentistID: strPtr(testDentistID)},
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{ID: testClinicID}, nil).Once()
			},
			setupDentist: func(dentistRepo *mocks.DentistRepository) {
				dentistRepo.EXPECT().GetByID(mock.Anything, testClinicID, testDentistID).
					Return(dentist.Dentist{ID: testDentistID, ClinicID: testClinicID}, nil).Once()
			},
		},
		{
			name:      "clinic not found",
			in:        payment.CreateInput{ClinicID: testClinicID, Amount: 15000},
			setupRepo: func(repo *mocks.PaymentRepository) {},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			setupDentist: func(dentistRepo *mocks.DentistRepository) {},
			wantErr:      payment.ErrClinicNotFound,
		},
		{
			name:      "dentist not found",
			in:        payment.CreateInput{ClinicID: testClinicID, Amount: 15000, DentistID: strPtr(testDentistID)},
			setupRepo: func(repo *mocks.PaymentRepository) {},
			setupClinic: func(clinicRepo *mocks.Repository) {
				clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{ID: testClinicID}, nil).Once()
			},
			setupDentist: func(dentistRepo *mocks.DentistRepository) {
				dentistRepo.EXPECT().GetByID(mock.Anything, testClinicID, testDentistID).
					Return(dentist.Dentist{}, dentist.ErrNotFound).Once()
			},
			wantErr: payment.ErrDentistNotFound,
		},
		{
			name:              "validation error",
			in:                payment.CreateInput{Amount: 0},
			setupRepo:         func(repo *mocks.PaymentRepository) {},
			setupClinic:       func(clinicRepo *mocks.Repository) {},
			setupDentist:      func(dentistRepo *mocks.DentistRepository) {},
			wantValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewPaymentRepository(t)
			clinicRepo := mocks.NewRepository(t)
			dentistRepo := mocks.NewDentistRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			tt.setupDentist(dentistRepo)

			// Long delay: this test doesn't assert on the background
			// approval goroutine (that's TestService_Create_SchedulesBackgroundApproval),
			// so it must never fire and touch the mock after the
			// subtest (and its t.Cleanup assertions) have finished.
			svc := payment.NewService(repo, clinicRepo, dentistRepo, newFakePixProvider(),
				payment.WithApprovalDelay(func() time.Duration { return time.Hour }))

			p, err := svc.Create(context.Background(), tt.in)

			switch {
			case tt.wantValidationErr:
				var ve *payment.ValidationError
				require.True(t, errors.As(err, &ve))
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.NotEmpty(t, p.ID)
				assert.Equal(t, payment.StatusPending, p.Status)
				assert.NotEmpty(t, p.PixCode)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	found := payment.Payment{ID: "payment-1", Status: payment.StatusPending}

	tests := []struct {
		name      string
		id        string
		setupRepo func(repo *mocks.PaymentRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   "payment-1",
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().GetByID(mock.Anything, "payment-1").Return(found, nil).Once()
			},
		},
		{
			name: "not found",
			id:   "missing-id",
			setupRepo: func(repo *mocks.PaymentRepository) {
				repo.EXPECT().GetByID(mock.Anything, "missing-id").Return(payment.Payment{}, payment.ErrNotFound).Once()
			},
			wantErr: payment.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewPaymentRepository(t)
			clinicRepo := mocks.NewRepository(t)
			dentistRepo := mocks.NewDentistRepository(t)
			tt.setupRepo(repo)

			svc := payment.NewService(repo, clinicRepo, dentistRepo, newFakePixProvider())

			p, err := svc.Get(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, found.ID, p.ID)
		})
	}
}

func TestService_Create_SchedulesBackgroundApproval(t *testing.T) {
	repo := mocks.NewPaymentRepository(t)
	clinicRepo := mocks.NewRepository(t)
	dentistRepo := mocks.NewDentistRepository(t)

	clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{ID: testClinicID}, nil).Once()
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()

	approved := make(chan struct{})
	repo.EXPECT().Approve(mock.Anything, mock.Anything, mock.AnythingOfType("time.Time")).
		Run(func(ctx context.Context, id string, approvedAt time.Time) { close(approved) }).
		Return(nil).Once()

	svc := payment.NewService(repo, clinicRepo, dentistRepo, newFakePixProvider(),
		payment.WithApprovalDelay(func() time.Duration { return 0 }))

	_, err := svc.Create(context.Background(), payment.CreateInput{ClinicID: testClinicID, Amount: 15000})
	require.NoError(t, err)

	select {
	case <-approved:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for background approval to run")
	}
}

func strPtr(s string) *string { return &s }
