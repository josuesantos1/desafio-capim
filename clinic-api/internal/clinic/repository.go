package clinic

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/josuesantos1/desafio/pkg/storage"
)

var (
	ErrNotFound       = errors.New("clinic: not found")
	ErrDocumentExists = errors.New("clinic: document already exists")
)

type Repository interface {
	// Create returns ErrDocumentExists if another non-deleted clinic
	// already has the same document.
	Create(ctx context.Context, c Clinic) error
	GetByID(ctx context.Context, id string) (Clinic, error)
	GetByDocument(ctx context.Context, document string) (Clinic, error)
	Update(ctx context.Context, c Clinic) error
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
}

// memoryRepository serializes every write (Create/Update/SoftDelete/
// Activate) under mu, since the uniqueness scan and the soft-delete
// check-then-write need to be atomic — storage.Store only guarantees
// atomicity for a single key operation, not for a compound sequence.
// Reads (GetByID/GetByDocument) rely solely on the Store's own RLock.
type memoryRepository struct {
	mu    sync.Mutex
	store *storage.Store[Clinic]
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{store: storage.New[Clinic]()}
}

func (r *memoryRepository) Create(ctx context.Context, c Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.ContainsFunc(r.store.All(), func(existing Clinic) bool {
		return existing.DeletedAt == nil && existing.Document == c.Document
	}) {
		return ErrDocumentExists
	}
	return r.store.Insert(c.ID, c)
}

func (r *memoryRepository) GetByID(ctx context.Context, id string) (Clinic, error) {
	c, err := r.store.Read(id)
	if err != nil || c.DeletedAt != nil {
		return Clinic{}, ErrNotFound
	}
	return c, nil
}

func (r *memoryRepository) GetByDocument(ctx context.Context, document string) (Clinic, error) {
	for _, c := range r.store.All() {
		if c.DeletedAt == nil && c.Document == document {
			return c, nil
		}
	}
	return Clinic{}, ErrNotFound
}

func (r *memoryRepository) Update(ctx context.Context, c Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.store.Read(c.ID)
	if err != nil || current.DeletedAt != nil {
		return ErrNotFound
	}
	return r.store.Update(c.ID, c)
}

func (r *memoryRepository) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.store.Read(id)
	if err != nil || current.DeletedAt != nil {
		return ErrNotFound
	}
	current.DeletedAt = &deletedAt
	current.UpdatedAt = deletedAt
	return r.store.Update(id, current)
}

// Activate is not part of Repository — it is an extra capability
// exposed by the concrete type and consumed structurally by
// internal/dentist's clinicActivator, mirroring the clinicGetter/
// PixProvider pattern already used in this codebase. dentist owns the
// pending -> active transition logic (it knows when a clinic gained
// its last required admin/legal representative), so it needs write
// access to clinic status without widening clinic.Repository's public
// contract for a single internal caller.
//
// It is fire-and-forget: no error is returned for a clinic that
// doesn't exist, is soft-deleted, or is already active — same as the
// original SQL recompute statement, which never checked rows
// affected.
func (r *memoryRepository) Activate(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.store.Read(id)
	if err != nil || current.DeletedAt != nil || current.Status != StatusPending {
		return nil
	}
	current.Status = StatusActive
	current.UpdatedAt = time.Now().UTC()
	return r.store.Update(id, current)
}
