//go:build integration

package dentist_test

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
)

// TestIntegration_ConcurrentDemote_LastAdminGuard proves, against a
// real Postgres (not mocks), that the guard protecting an active
// clinic's last administrator is race-free: two concurrent
// UpdateRoles calls each demoting one of the clinic's two
// administrators must not both succeed.
//
// Run locally against the same throwaway postgres:16-alpine container
// used for manual validation of every feature in this project, with
// migrations 000001-000005 applied, e.g.:
//
//	docker run --rm -d --name clinic-it -e POSTGRES_PASSWORD=postgres -p 5433:5432 postgres:16-alpine
//	# apply migrations/*.up.sql against it, then:
//	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable" \
//	  go test -tags=integration ./internal/dentist/... -run TestIntegration -v
func TestIntegration_ConcurrentDemote_LastAdminGuard(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}

	db, err := sqlx.Connect("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	clinicRepo := clinic.NewPostgresRepository(db)
	clinicSvc := clinic.NewService(clinicRepo)
	dentistRepo := dentist.NewPostgresRepository(db)
	dentistSvc := dentist.NewService(dentistRepo, clinicRepo)

	ctx := context.Background()

	c, err := clinicSvc.Create(ctx, clinic.CreateInput{
		Document:  randomDigits(11),
		LegalName: "Integration Test Clinic",
		TradeName: "IT Clinic",
	})
	require.NoError(t, err)
	require.Equal(t, clinic.StatusPending, c.Status)

	d1, err := dentistSvc.Create(ctx, c.ID, dentist.CreateInput{Name: "Dr. A", Phone: "1", Email: "a@it.test"})
	require.NoError(t, err)
	d2, err := dentistSvc.Create(ctx, c.ID, dentist.CreateInput{Name: "Dr. B", Phone: "2", Email: "b@it.test"})
	require.NoError(t, err)

	// d1 becomes admin + legal representative (activates the clinic);
	// d2 also becomes an administrator, so there are two active admins.
	yes := true
	_, err = dentistSvc.UpdateRoles(ctx, c.ID, d1.ID, dentist.RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &yes})
	require.NoError(t, err)
	_, err = dentistSvc.UpdateRoles(ctx, c.ID, d2.ID, dentist.RolesInput{IsAdministrator: &yes})
	require.NoError(t, err)

	got, err := clinicSvc.Get(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, clinic.StatusActive, got.Status, "clinic must be active before the race")

	// Two concurrent requests, each demoting a different one of the
	// two administrators — at most one may succeed.
	no := false
	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, errs[0] = dentistSvc.UpdateRoles(ctx, c.ID, d1.ID, dentist.RolesInput{IsAdministrator: &no})
	}()
	go func() {
		defer wg.Done()
		_, errs[1] = dentistSvc.UpdateRoles(ctx, c.ID, d2.ID, dentist.RolesInput{IsAdministrator: &no})
	}()
	wg.Wait()

	successes, blocked := 0, 0
	for _, e := range errs {
		switch {
		case e == nil:
			successes++
		case e != nil:
			blocked++
		}
	}
	require.Equal(t, 1, successes, "exactly one concurrent demote must succeed")
	require.Equal(t, 1, blocked, "exactly one concurrent demote must be rejected")

	result, err := dentistSvc.List(ctx, c.ID, dentist.ListParams{Limit: 10, IsAdministrator: &yes})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total, "clinic must end up with exactly one active administrator")
}

func randomDigits(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}
