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
	// Create is idempotency-aware: p.IdempotencyKey must be set. If a
	// non-conflicting payment with the same IdempotencyKey already
	// exists (same ClinicID, AmountCents, DentistID), Create returns
	// that existing payment and created=false — no new payment is
	// inserted. If an existing payment with the same IdempotencyKey
	// has different business fields, Create returns
	// ErrIdempotencyKeyConflict. Otherwise p is inserted and Create
	// returns (p, true, nil). The existence check and the insert
	// happen atomically under the same lock.
	Create(ctx context.Context, p Payment) (result Payment, created bool, err error)
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

func (r *memoryRepository) Create(ctx context.Context, p Payment) (Payment, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.store.All() {
		if existing.IdempotencyKey != p.IdempotencyKey {
			continue
		}
		if existing.ClinicID == p.ClinicID &&
			existing.AmountCents == p.AmountCents &&
			sameDentistID(existing.DentistID, p.DentistID) {
			return existing, false, nil
		}
		return Payment{}, false, ErrIdempotencyKeyConflict
	}

	if _, err := r.clinics.GetByID(ctx, p.ClinicID); err != nil {
		return Payment{}, false, ErrNotFound
	}
	if p.DentistID != nil {
		if _, err := r.dentists.GetByID(ctx, p.ClinicID, *p.DentistID); err != nil {
			return Payment{}, false, ErrNotFound
		}
	}
	if err := r.store.Insert(p.ID, p); err != nil {
		return Payment{}, false, err
	}
	return p, true, nil
}

// sameDentistID compares two possibly-nil dentist ids by value, never
// dereferencing without a nil check first.
func sameDentistID(a, b *string) bool {
	return (a == nil) == (b == nil) && (a == nil || *a == *b)
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
