package dentist

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
)

const testClinicID = "11111111-1111-1111-1111-111111111111"

// newDentistRepo returns a dentist.memoryRepository wired to a real
// clinic.memoryRepository (satisfies clinicActivator structurally),
// with testClinicID already created and pending. The returned func
// activates the clinic when called with "active".
func newDentistRepo(t *testing.T) (*memoryRepository, func(status string)) {
	t.Helper()
	clinics := clinic.NewMemoryRepository()
	now := time.Now().UTC()
	require.NoError(t, clinics.Create(context.Background(), clinic.Clinic{
		ID: testClinicID, Document: "doc-1", LegalName: "L", TradeName: "T",
		Status: clinic.StatusPending, CreatedAt: now, UpdatedAt: now,
	}))

	activate := func(status string) {
		if status == "active" {
			require.NoError(t, clinics.Activate(context.Background(), testClinicID))
		}
	}

	return NewMemoryRepository(clinics), activate
}

func newDentist(id, email string) Dentist {
	now := time.Now().UTC()
	return Dentist{
		ID: id, ClinicID: testClinicID, Name: "Dr. X", Phone: "123", Email: email,
		CreatedAt: now, UpdatedAt: now,
	}
}

func TestMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
	})

	t.Run("clinic not found returns ErrNotFound", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		d := newDentist("d-1", "a@test.com")
		d.ClinicID = "missing-clinic"
		err := repo.Create(ctx, d)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("duplicate email returns ErrEmailExists", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))

		err := repo.Create(ctx, newDentist("d-2", "a@test.com"))
		require.ErrorIs(t, err, ErrEmailExists)
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo, _ := newDentistRepo(t)
	require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))

	t.Run("success", func(t *testing.T) {
		got, err := repo.GetByID(ctx, testClinicID, "d-1")
		require.NoError(t, err)
		require.Equal(t, "a@test.com", got.Email)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, testClinicID, "missing")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("wrong clinic returns not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "other-clinic", "d-1")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		d := newDentist("d-1", "a@test.com")
		require.NoError(t, repo.Create(ctx, d))

		d.Name = "Dr. Updated"
		require.NoError(t, repo.Update(ctx, d))

		got, err := repo.GetByID(ctx, testClinicID, "d-1")
		require.NoError(t, err)
		require.Equal(t, "Dr. Updated", got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		err := repo.Update(ctx, newDentist("missing", "a@test.com"))
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("email changed to one already in use returns ErrEmailExists", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		d2 := newDentist("d-2", "b@test.com")
		require.NoError(t, repo.Create(ctx, d2))

		d2.Email = "a@test.com"
		err := repo.Update(ctx, d2)
		require.ErrorIs(t, err, ErrEmailExists)
	})
}

func TestMemoryRepository_UpdateRoles(t *testing.T) {
	ctx := context.Background()
	yes, no := true, false

	t.Run("pending clinic allows demoting sole admin", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes})
		require.NoError(t, err)

		_, err = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &no})
		require.NoError(t, err, "clinic still pending, guard doesn't apply")
	})

	t.Run("active clinic blocks demoting the last administrator", func(t *testing.T) {
		repo, activate := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
		require.NoError(t, err)
		activate("active")

		_, err = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &no})
		require.ErrorIs(t, err, ErrLastAdminRequired)
	})

	t.Run("active clinic allows demoting admin when another active admin exists", func(t *testing.T) {
		repo, activate := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		require.NoError(t, repo.Create(ctx, newDentist("d-2", "b@test.com")))
		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
		require.NoError(t, err)
		_, err = repo.UpdateRoles(ctx, testClinicID, "d-2", RolesInput{IsAdministrator: &yes})
		require.NoError(t, err)
		activate("active")

		_, err = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &no})
		require.NoError(t, err)
	})

	t.Run("admin guard is checked before legal representative guard", func(t *testing.T) {
		repo, activate := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
		require.NoError(t, err)
		activate("active")

		_, err = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &no, IsLegalRepresentative: &no})
		require.ErrorIs(t, err, ErrLastAdminRequired)
	})

	t.Run("clinic activates once it has both an admin and a legal representative", func(t *testing.T) {
		clinics := clinic.NewMemoryRepository()
		now := time.Now().UTC()
		require.NoError(t, clinics.Create(ctx, clinic.Clinic{
			ID: testClinicID, Document: "doc-1", LegalName: "L", TradeName: "T",
			Status: clinic.StatusPending, CreatedAt: now, UpdatedAt: now,
		}))
		repo := NewMemoryRepository(clinics)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))

		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes})
		require.NoError(t, err)
		got, err := clinics.GetByID(ctx, testClinicID)
		require.NoError(t, err)
		require.Equal(t, clinic.StatusPending, got.Status, "still missing a legal representative")

		_, err = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsLegalRepresentative: &yes})
		require.NoError(t, err)
		got, err = clinics.GetByID(ctx, testClinicID)
		require.NoError(t, err)
		require.Equal(t, clinic.StatusActive, got.Status)
	})

	t.Run("clinic not found returns not found", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		_, err := repo.UpdateRoles(ctx, "missing-clinic", "d-1", RolesInput{IsAdministrator: &yes})
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	yes := true

	t.Run("success", func(t *testing.T) {
		repo, _ := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))

		require.NoError(t, repo.SoftDelete(ctx, testClinicID, "d-1", time.Now().UTC()))

		_, err := repo.GetByID(ctx, testClinicID, "d-1")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("active clinic blocks deleting the last administrator", func(t *testing.T) {
		repo, activate := newDentistRepo(t)
		require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
		_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
		require.NoError(t, err)
		activate("active")

		err = repo.SoftDelete(ctx, testClinicID, "d-1", time.Now().UTC())
		require.ErrorIs(t, err, ErrLastAdminRequired)
	})
}

func TestMemoryRepository_List(t *testing.T) {
	ctx := context.Background()
	yes := true

	repo, _ := newDentistRepo(t)
	for i, email := range []string{"a@test.com", "b@test.com", "c@test.com"} {
		d := newDentist("d-"+string(rune('1'+i)), email)
		require.NoError(t, repo.Create(ctx, d))
	}
	_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes})
	require.NoError(t, err)

	t.Run("lists all non-deleted, ordered by CreatedAt", func(t *testing.T) {
		result, err := repo.List(ctx, testClinicID, ListParams{Limit: 10})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Len(t, result.Items, 3)
	})

	t.Run("filters by IsAdministrator", func(t *testing.T) {
		result, err := repo.List(ctx, testClinicID, ListParams{Limit: 10, IsAdministrator: &yes})
		require.NoError(t, err)
		require.Equal(t, 1, result.Total)
		require.Equal(t, "d-1", result.Items[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := repo.List(ctx, testClinicID, ListParams{Limit: 2, Offset: 2})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Len(t, result.Items, 1)
	})

	t.Run("clinic not found returns not found", func(t *testing.T) {
		_, err := repo.List(ctx, "missing-clinic", ListParams{Limit: 10})
		require.ErrorIs(t, err, ErrNotFound)
	})
}

// TestMemoryRepository_ConcurrentDemote_LastAdminGuard replaces the
// former Postgres integration test (internal/dentist/integration_test.go):
// two concurrent UpdateRoles calls, each demoting one of a clinic's two
// active administrators, must not both succeed. No real database is
// needed anymore — run with `go test -race` to also confirm the guard
// itself is race-free.
func TestMemoryRepository_ConcurrentDemote_LastAdminGuard(t *testing.T) {
	ctx := context.Background()
	yes, no := true, false

	repo, activate := newDentistRepo(t)
	require.NoError(t, repo.Create(ctx, newDentist("d-1", "a@test.com")))
	require.NoError(t, repo.Create(ctx, newDentist("d-2", "b@test.com")))

	_, err := repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
	require.NoError(t, err)
	_, err = repo.UpdateRoles(ctx, testClinicID, "d-2", RolesInput{IsAdministrator: &yes})
	require.NoError(t, err)
	activate("active")

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, errs[0] = repo.UpdateRoles(ctx, testClinicID, "d-1", RolesInput{IsAdministrator: &no})
	}()
	go func() {
		defer wg.Done()
		_, errs[1] = repo.UpdateRoles(ctx, testClinicID, "d-2", RolesInput{IsAdministrator: &no})
	}()
	wg.Wait()

	successes, blocked := 0, 0
	for _, e := range errs {
		if e == nil {
			successes++
		} else {
			blocked++
		}
	}
	require.Equal(t, 1, successes, "exactly one concurrent demote must succeed")
	require.Equal(t, 1, blocked, "exactly one concurrent demote must be rejected")

	result, err := repo.List(ctx, testClinicID, ListParams{Limit: 10, IsAdministrator: &yes})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total, "clinic must end up with exactly one active administrator")
}
