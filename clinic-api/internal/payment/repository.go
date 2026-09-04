package payment

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/josuesantos1/desafio/pkg/storage"
)

var ErrNotFound = errors.New("payment: not found")

type Repository interface {
	Create(ctx context.Context, p Payment) error
	GetByID(ctx context.Context, id string) (Payment, error)
	// Approve transitions status "pending" -> "approved". Returns
	// ErrNotFound if the payment does not exist or is not "pending"
	// (already approved — idempotency guard against duplicate runs).
	Approve(ctx context.Context, id string, approvedAt time.Time) error
}

// memoryRepository serializes Create/Approve under mu — the guarded
// compound sequences (clinic/dentist existence check, pending->approved
// check-then-write) need atomicity that storage.Store alone only gives
// per single-key operation. GetByID relies solely on the Store's own
// RLock.
type memoryRepository struct {
	mu       sync.Mutex
	store    *storage.Store[Payment]
	clinics  clinicGetter
	dentists dentistGetter
}

func NewMemoryRepository(clinics clinicGetter, dentists dentistGetter) *memoryRepository {
	return &memoryRepository{store: storage.New[Payment](), clinics: clinics, dentists: dentists}
}

func (r *memoryRepository) Create(ctx context.Context, p Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := r.clinics.GetByID(ctx, p.ClinicID); err != nil {
		return ErrNotFound
	}
	if p.DentistID != nil {
		if _, err := r.dentists.GetByID(ctx, p.ClinicID, *p.DentistID); err != nil {
			return ErrNotFound
		}
	}
	return r.store.Insert(p.ID, p)
}

func (r *memoryRepository) GetByID(ctx context.Context, id string) (Payment, error) {
	p, err := r.store.Read(id)
	if err != nil {
		return Payment{}, ErrNotFound
	}
	return p, nil
}

func (r *memoryRepository) Approve(ctx context.Context, id string, approvedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.store.Read(id)
	if err != nil || current.Status != StatusPending {
		return ErrNotFound
	}
	current.Status = StatusApproved
	current.ApprovedAt = &approvedAt
	current.UpdatedAt = approvedAt
	return r.store.Update(id, current)
}
