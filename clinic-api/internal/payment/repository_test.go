package payment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
)

const testClinicIDRepo = "11111111-1111-1111-1111-111111111111"

func newPaymentRepo(t *testing.T) (*memoryRepository, *clinic.Clinic, *dentist.Dentist) {
	t.Helper()
	ctx := context.Background()

	clinics := clinic.NewMemoryRepository()
	now := time.Now().UTC()
	c := clinic.Clinic{
		ID: testClinicIDRepo, Document: "doc-1", LegalName: "L", TradeName: "T",
		Status: clinic.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, clinics.Create(ctx, c))

	dentists := dentist.NewMemoryRepository(clinics)
	d := dentist.Dentist{
		ID: "dentist-1", ClinicID: testClinicIDRepo, Name: "Dr. X", Phone: "1", Email: "x@test.com",
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, dentists.Create(ctx, d))

	return NewMemoryRepository(clinics, dentists), &c, &d
}

func newPayment(id, clinicID string, dentistID *string) Payment {
	now := time.Now().UTC()
	return Payment{
		ID: id, ClinicID: clinicID, DentistID: dentistID, AmountCents: 1000,
		Status: StatusPending, PixCode: "code", CreatedAt: now, UpdatedAt: now,
	}
}

func TestMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success without dentist", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		require.NoError(t, repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil)))
	})

	t.Run("success with dentist", func(t *testing.T) {
		repo, _, d := newPaymentRepo(t)
		require.NoError(t, repo.Create(ctx, newPayment("p-1", testClinicIDRepo, &d.ID)))
	})

	t.Run("clinic not found returns ErrNotFound", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		err := repo.Create(ctx, newPayment("p-1", "missing-clinic", nil))
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("dentist not found returns ErrNotFound", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		missing := "missing-dentist"
		err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, &missing))
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("dentist belonging to another clinic returns ErrNotFound", func(t *testing.T) {
		repo, _, d := newPaymentRepo(t)
		err := repo.Create(ctx, newPayment("p-1", "other-clinic", &d.ID))
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo, _, _ := newPaymentRepo(t)
	require.NoError(t, repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil)))

	t.Run("success", func(t *testing.T) {
		got, err := repo.GetByID(ctx, "p-1")
		require.NoError(t, err)
		require.Equal(t, StatusPending, got.Status)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "missing")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_Approve(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		require.NoError(t, repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil)))

		require.NoError(t, repo.Approve(ctx, "p-1", time.Now().UTC()))

		got, err := repo.GetByID(ctx, "p-1")
		require.NoError(t, err)
		require.Equal(t, StatusApproved, got.Status)
		require.NotNil(t, got.ApprovedAt)
	})

	t.Run("not found", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		err := repo.Approve(ctx, "missing", time.Now().UTC())
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("double approve returns ErrNotFound (idempotency guard)", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		require.NoError(t, repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil)))
		require.NoError(t, repo.Approve(ctx, "p-1", time.Now().UTC()))

		err := repo.Approve(ctx, "p-1", time.Now().UTC())
		require.ErrorIs(t, err, ErrNotFound)
	})
}
