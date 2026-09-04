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
