package payment

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/josuesantos1/desafio/pkg/storage"
)

var ErrNotFound = errors.New("payment: not found")

type Repository interface {
	Create(ctx context.Context, p Payment) (result Payment, created bool, err error)
	GetByID(ctx context.Context, id string) (Payment, error)
	Approve(ctx context.Context, id string, approvedAt time.Time) error
	List(ctx context.Context, params ListParams) (ListResult, error)
}

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

func (r *memoryRepository) List(ctx context.Context, params ListParams) (ListResult, error) {
	var items []Payment
	for _, p := range r.store.All() {
		if p.ClinicID != params.ClinicID {
			continue
		}
		if params.Status != nil && p.Status != *params.Status {
			continue
		}
		items = append(items, p)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })

	total := len(items)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)
	return ListResult{Items: items[start:end], Total: total}, nil
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
