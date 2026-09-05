package payment

import (
	"context"
	"sync"
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

func newPayment(id, clinicID string, dentistID *string, idempotencyKey string) Payment {
	now := time.Now().UTC()
	return Payment{
		ID: id, ClinicID: clinicID, DentistID: dentistID, AmountCents: 1000,
		Status: StatusPending, PixCode: "code", IdempotencyKey: idempotencyKey,
		CreatedAt: now, UpdatedAt: now,
	}
}

func TestMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success without dentist", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		result, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.True(t, created)
		require.Equal(t, "p-1", result.ID)
	})

	t.Run("success with dentist", func(t *testing.T) {
		repo, _, d := newPaymentRepo(t)
		_, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, &d.ID, "idem-1"))
		require.NoError(t, err)
		require.True(t, created)
	})

	t.Run("clinic not found returns ErrNotFound", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		_, created, err := repo.Create(ctx, newPayment("p-1", "missing-clinic", nil, "idem-1"))
		require.ErrorIs(t, err, ErrNotFound)
		require.False(t, created)
	})

	t.Run("dentist not found returns ErrNotFound", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		missing := "missing-dentist"
		_, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, &missing, "idem-1"))
		require.ErrorIs(t, err, ErrNotFound)
		require.False(t, created)
	})

	t.Run("dentist belonging to another clinic returns ErrNotFound", func(t *testing.T) {
		repo, _, d := newPaymentRepo(t)
		_, created, err := repo.Create(ctx, newPayment("p-1", "other-clinic", &d.ID, "idem-1"))
		require.ErrorIs(t, err, ErrNotFound)
		require.False(t, created)
	})

	t.Run("replay: same key and same payload returns original, created=false", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		original, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.True(t, created)

		replay, created, err := repo.Create(ctx, newPayment("p-2", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.False(t, created)
		require.Equal(t, original.ID, replay.ID)

		_, err = repo.GetByID(ctx, "p-2")
		require.ErrorIs(t, err, ErrNotFound, "the replay attempt's payment must never have been inserted")
	})

	t.Run("conflict: same key, different amount returns ErrIdempotencyKeyConflict", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		_, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.True(t, created)

		conflicting := newPayment("p-2", testClinicIDRepo, nil, "idem-1")
		conflicting.AmountCents = 9999
		_, created, err = repo.Create(ctx, conflicting)
		require.ErrorIs(t, err, ErrIdempotencyKeyConflict)
		require.False(t, created)

		_, err = repo.GetByID(ctx, "p-2")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("conflict: same key, different dentist_id (nil vs set) returns ErrIdempotencyKeyConflict", func(t *testing.T) {
		repo, _, d := newPaymentRepo(t)
		_, created, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.True(t, created)

		_, created, err = repo.Create(ctx, newPayment("p-2", testClinicIDRepo, &d.ID, "idem-1"))
		require.ErrorIs(t, err, ErrIdempotencyKeyConflict)
		require.False(t, created)
	})

	t.Run("concurrent creates with the same new key and same payload: exactly one created", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)

		var wg sync.WaitGroup
		results := make([]bool, 2)
		errs := make([]error, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, results[0], errs[0] = repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-concurrent"))
		}()
		go func() {
			defer wg.Done()
			_, results[1], errs[1] = repo.Create(ctx, newPayment("p-2", testClinicIDRepo, nil, "idem-concurrent"))
		}()
		wg.Wait()

		require.NoError(t, errs[0])
		require.NoError(t, errs[1])
		createdCount := 0
		for _, c := range results {
			if c {
				createdCount++
			}
		}
		require.Equal(t, 1, createdCount, "exactly one of the two concurrent creates must have inserted a new payment")
	})

	t.Run("concurrent creates with the same new key but different payloads: one created, one conflict", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)

		p1 := newPayment("p-1", testClinicIDRepo, nil, "idem-concurrent-conflict")
		p2 := newPayment("p-2", testClinicIDRepo, nil, "idem-concurrent-conflict")
		p2.AmountCents = 9999

		var wg sync.WaitGroup
		created := make([]bool, 2)
		errs := make([]error, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, created[0], errs[0] = repo.Create(ctx, p1)
		}()
		go func() {
			defer wg.Done()
			_, created[1], errs[1] = repo.Create(ctx, p2)
		}()
		wg.Wait()

		createdCount, conflictCount := 0, 0
		for i, c := range created {
			switch {
			case c:
				createdCount++
				require.NoError(t, errs[i])
			case errs[i] != nil:
				require.ErrorIs(t, errs[i], ErrIdempotencyKeyConflict)
				conflictCount++
			default:
				t.Fatalf("unexpected outcome: created=false with no error at index %d", i)
			}
		}
		require.Equal(t, 1, createdCount)
		require.Equal(t, 1, conflictCount)
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo, _, _ := newPaymentRepo(t)
	_, _, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
	require.NoError(t, err)

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
		_, _, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)

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
		_, _, err := repo.Create(ctx, newPayment("p-1", testClinicIDRepo, nil, "idem-1"))
		require.NoError(t, err)
		require.NoError(t, repo.Approve(ctx, "p-1", time.Now().UTC()))

		err = repo.Approve(ctx, "p-1", time.Now().UTC())
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_List(t *testing.T) {
	ctx := context.Background()
	const otherClinicID = "22222222-2222-2222-2222-222222222222"
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	seed := func(repo *memoryRepository) {
		p1 := newPayment("p-1", testClinicIDRepo, nil, "idem-1")
		p1.Status = StatusPending
		p1.CreatedAt = base
		require.NoError(t, repo.store.Insert(p1.ID, p1))

		p2 := newPayment("p-2", testClinicIDRepo, nil, "idem-2")
		p2.Status = StatusApproved
		p2.CreatedAt = base.Add(time.Hour)
		require.NoError(t, repo.store.Insert(p2.ID, p2))

		p3 := newPayment("p-3", testClinicIDRepo, nil, "idem-3")
		p3.Status = StatusPending
		p3.CreatedAt = base.Add(2 * time.Hour)
		require.NoError(t, repo.store.Insert(p3.ID, p3))

		other := newPayment("p-other", otherClinicID, nil, "idem-other")
		other.CreatedAt = base.Add(3 * time.Hour)
		require.NoError(t, repo.store.Insert(other.ID, other))
	}

	t.Run("no filter returns clinic's payments ordered by CreatedAt descending", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, ClinicID: testClinicIDRepo})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Equal(t, []string{"p-3", "p-2", "p-1"}, paymentIDs(result.Items))
	})

	t.Run("pagination slices by limit/offset but keeps total", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 1, Offset: 1, ClinicID: testClinicIDRepo})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Equal(t, []string{"p-2"}, paymentIDs(result.Items))
	})

	t.Run("filter by status", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		seed(repo)

		approved := StatusApproved
		result, err := repo.List(ctx, ListParams{Limit: 20, ClinicID: testClinicIDRepo, Status: &approved})
		require.NoError(t, err)
		require.Equal(t, []string{"p-2"}, paymentIDs(result.Items))
	})

	t.Run("isolates by clinic_id", func(t *testing.T) {
		repo, _, _ := newPaymentRepo(t)
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, ClinicID: otherClinicID})
		require.NoError(t, err)
		require.Equal(t, []string{"p-other"}, paymentIDs(result.Items))
	})
}

func paymentIDs(items []Payment) []string {
	ids := make([]string, len(items))
	for i, p := range items {
		ids[i] = p.ID
	}
	return ids
}
