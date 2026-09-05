package clinic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newClinic(id, document string) Clinic {
	now := time.Now().UTC()
	return Clinic{
		ID:        id,
		Document:  document,
		LegalName: "Legal Name",
		TradeName: "Trade Name",
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

		got, err := repo.GetByID(ctx, "id-1")
		require.NoError(t, err)
		require.Equal(t, "doc-1", got.Document)
	})

	t.Run("duplicate document among non-deleted returns ErrDocumentExists", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

		err := repo.Create(ctx, newClinic("id-2", "doc-1"))
		require.ErrorIs(t, err, ErrDocumentExists)
	})

	t.Run("document of a soft-deleted clinic does not block creation", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))
		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))

		require.NoError(t, repo.Create(ctx, newClinic("id-2", "doc-1")))
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "missing")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("soft-deleted returns not found", func(t *testing.T) {
		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))
		_, err := repo.GetByID(ctx, "id-1")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_GetByDocument(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

	t.Run("success", func(t *testing.T) {
		got, err := repo.GetByDocument(ctx, "doc-1")
		require.NoError(t, err)
		require.Equal(t, "id-1", got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByDocument(ctx, "missing-doc")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewMemoryRepository()
		c := newClinic("id-1", "doc-1")
		require.NoError(t, repo.Create(ctx, c))

		c.LegalName = "Updated Name"
		require.NoError(t, repo.Update(ctx, c))

		got, err := repo.GetByID(ctx, "id-1")
		require.NoError(t, err)
		require.Equal(t, "Updated Name", got.LegalName)
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewMemoryRepository()
		err := repo.Update(ctx, newClinic("missing", "doc-1"))
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("soft-deleted returns not found", func(t *testing.T) {
		repo := NewMemoryRepository()
		c := newClinic("id-1", "doc-1")
		require.NoError(t, repo.Create(ctx, c))
		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))

		err := repo.Update(ctx, c)
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))

		_, err := repo.GetByID(ctx, "id-1")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewMemoryRepository()
		err := repo.SoftDelete(ctx, "missing", time.Now().UTC())
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("double delete returns not found", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))
		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))

		err := repo.SoftDelete(ctx, "id-1", time.Now().UTC())
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemoryRepository_Activate(t *testing.T) {
	ctx := context.Background()

	t.Run("pending clinic becomes active", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))

		require.NoError(t, repo.Activate(ctx, "id-1"))

		got, err := repo.GetByID(ctx, "id-1")
		require.NoError(t, err)
		require.Equal(t, StatusActive, got.Status)
	})

	t.Run("already active clinic is a no-op", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))
		require.NoError(t, repo.Activate(ctx, "id-1"))

		require.NoError(t, repo.Activate(ctx, "id-1"))

		got, err := repo.GetByID(ctx, "id-1")
		require.NoError(t, err)
		require.Equal(t, StatusActive, got.Status)
	})

	t.Run("missing clinic is a no-op, no error", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Activate(ctx, "missing"))
	})

	t.Run("soft-deleted clinic is a no-op, no error", func(t *testing.T) {
		repo := NewMemoryRepository()
		require.NoError(t, repo.Create(ctx, newClinic("id-1", "doc-1")))
		require.NoError(t, repo.SoftDelete(ctx, "id-1", time.Now().UTC()))

		require.NoError(t, repo.Activate(ctx, "id-1"))
	})
}

func TestMemoryRepository_List(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	seed := func(repo *memoryRepository) {
		c1 := newClinic("id-1", "doc-1")
		c1.TradeName = "Clínica Sorriso"
		c1.LegalName = "Sorriso LTDA"
		c1.Specialties = []string{"Ortodontia"}
		c1.Address = &Address{City: "São Paulo"}
		c1.CreatedAt = base
		require.NoError(t, repo.Create(ctx, c1))

		c2 := newClinic("id-2", "doc-2")
		c2.TradeName = "Odonto Vida"
		c2.LegalName = "Vida LTDA"
		c2.Specialties = []string{"Implantodontia"}
		c2.Address = &Address{City: "Rio de Janeiro"}
		c2.CreatedAt = base.Add(time.Hour)
		require.NoError(t, repo.Create(ctx, c2))

		c3 := newClinic("id-3", "doc-3")
		c3.TradeName = "Sem Endereço"
		c3.CreatedAt = base.Add(2 * time.Hour)
		require.NoError(t, repo.Create(ctx, c3))

		c4 := newClinic("id-4", "doc-4")
		c4.TradeName = "Excluída"
		c4.CreatedAt = base.Add(3 * time.Hour)
		require.NoError(t, repo.Create(ctx, c4))
		require.NoError(t, repo.SoftDelete(ctx, "id-4", base.Add(4*time.Hour)))
	}

	t.Run("no filter returns all non-deleted ordered by CreatedAt descending (newest first)", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Equal(t, []string{"id-3", "id-2", "id-1"}, idsOf(result.Items))
	})

	t.Run("pagination slices by limit/offset but keeps total", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 1, Offset: 1})
		require.NoError(t, err)
		require.Equal(t, 3, result.Total)
		require.Equal(t, []string{"id-2"}, idsOf(result.Items))
	})

	t.Run("search by trade name", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, Query: "sorriso"})
		require.NoError(t, err)
		require.Equal(t, []string{"id-1"}, idsOf(result.Items))
	})

	t.Run("search by legal name", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, Query: "vida"})
		require.NoError(t, err)
		require.Equal(t, []string{"id-2"}, idsOf(result.Items))
	})

	t.Run("search by specialty", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, Query: "implantodontia"})
		require.NoError(t, err)
		require.Equal(t, []string{"id-2"}, idsOf(result.Items))
	})

	t.Run("filter by city", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, City: "paulo"})
		require.NoError(t, err)
		require.Equal(t, []string{"id-1"}, idsOf(result.Items))
	})

	t.Run("clinic without address never matches city filter", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, City: "endereco"})
		require.NoError(t, err)
		require.Empty(t, result.Items)
	})

	t.Run("query and city combined require both to match", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20, Query: "sorriso", City: "rio"})
		require.NoError(t, err)
		require.Empty(t, result.Items)
	})

	t.Run("soft-deleted never appears", func(t *testing.T) {
		repo := NewMemoryRepository()
		seed(repo)

		result, err := repo.List(ctx, ListParams{Limit: 20})
		require.NoError(t, err)
		require.NotContains(t, idsOf(result.Items), "id-4")
		require.Equal(t, 3, result.Total)
	})
}

func idsOf(items []Clinic) []string {
	ids := make([]string, len(items))
	for i, c := range items {
		ids[i] = c.ID
	}
	return ids
}
