package dentist

import (
	"context"
	"errors"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/pkg/storage"
)

var (
	ErrNotFound    = errors.New("dentist: not found")
	ErrEmailExists = errors.New("dentist: email already exists")
)

type Repository interface {
	Create(ctx context.Context, d Dentist) error
	GetByID(ctx context.Context, clinicID, id string) (Dentist, error)
	Update(ctx context.Context, d Dentist) error
	UpdateRoles(ctx context.Context, clinicID, id string, in RolesInput) (Dentist, error)
	SoftDelete(ctx context.Context, clinicID, id string, deletedAt time.Time) error
	List(ctx context.Context, clinicID string, params ListParams) (ListResult, error)
}

type clinicActivator interface {
	clinicGetter
	Activate(ctx context.Context, id string) error
}

type memoryRepository struct {
	mu      sync.Mutex
	store   *storage.Store[Dentist]
	clinics clinicActivator
}

func NewMemoryRepository(clinics clinicActivator) *memoryRepository {
	return &memoryRepository{store: storage.New[Dentist](), clinics: clinics}
}

func (r *memoryRepository) Create(ctx context.Context, d Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := r.clinics.GetByID(ctx, d.ClinicID); err != nil {
		return ErrNotFound
	}
	if slices.ContainsFunc(r.store.All(), func(existing Dentist) bool {
		return existing.DeletedAt == nil && existing.Email == d.Email
	}) {
		return ErrEmailExists
	}
	return r.store.Insert(d.ID, d)
}

func (r *memoryRepository) GetByID(ctx context.Context, clinicID, id string) (Dentist, error) {
	if _, err := r.clinics.GetByID(ctx, clinicID); err != nil {
		return Dentist{}, ErrNotFound
	}
	d, err := r.store.Read(id)
	if err != nil || d.ClinicID != clinicID || d.DeletedAt != nil {
		return Dentist{}, ErrNotFound
	}
	return d, nil
}

func (r *memoryRepository) Update(ctx context.Context, d Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := r.clinics.GetByID(ctx, d.ClinicID); err != nil {
		return ErrNotFound
	}
	current, err := r.store.Read(d.ID)
	if err != nil || current.ClinicID != d.ClinicID || current.DeletedAt != nil {
		return ErrNotFound
	}
	if current.Email != d.Email && slices.ContainsFunc(r.store.All(), func(existing Dentist) bool {
		return existing.ID != d.ID && existing.DeletedAt == nil && existing.Email == d.Email
	}) {
		return ErrEmailExists
	}
	return r.store.Update(d.ID, d)
}

func (r *memoryRepository) UpdateRoles(ctx context.Context, clinicID, id string, in RolesInput) (Dentist, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, err := r.clinics.GetByID(ctx, clinicID)
	if err != nil {
		return Dentist{}, ErrNotFound
	}
	clinicActive := c.Status == clinic.StatusActive

	current, err := r.store.Read(id)
	if err != nil || current.ClinicID != clinicID || current.DeletedAt != nil {
		return Dentist{}, ErrNotFound
	}

	newIsAdmin := current.IsAdministrator
	if in.IsAdministrator != nil {
		newIsAdmin = *in.IsAdministrator
	}
	newIsLegalRep := current.IsLegalRepresentative
	if in.IsLegalRepresentative != nil {
		newIsLegalRep = *in.IsLegalRepresentative
	}

	if clinicActive && !newIsAdmin && current.IsAdministrator &&
		!r.hasOtherActiveWithFlag(clinicID, id, func(d Dentist) bool { return d.IsAdministrator }) {
		return Dentist{}, ErrLastAdminRequired
	}
	if clinicActive && !newIsLegalRep && current.IsLegalRepresentative &&
		!r.hasOtherActiveWithFlag(clinicID, id, func(d Dentist) bool { return d.IsLegalRepresentative }) {
		return Dentist{}, ErrLastLegalRepresentativeRequired
	}

	current.IsAdministrator = newIsAdmin
	current.IsLegalRepresentative = newIsLegalRep
	current.UpdatedAt = time.Now().UTC()
	if err := r.store.Update(id, current); err != nil {
		return Dentist{}, err
	}

	if r.hasActiveWithFlag(clinicID, func(d Dentist) bool { return d.IsAdministrator }) &&
		r.hasActiveWithFlag(clinicID, func(d Dentist) bool { return d.IsLegalRepresentative }) {
		_ = r.clinics.Activate(ctx, clinicID)
	}

	return current, nil
}

func (r *memoryRepository) SoftDelete(ctx context.Context, clinicID, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, err := r.clinics.GetByID(ctx, clinicID)
	if err != nil {
		return ErrNotFound
	}
	clinicActive := c.Status == clinic.StatusActive

	current, err := r.store.Read(id)
	if err != nil || current.ClinicID != clinicID || current.DeletedAt != nil {
		return ErrNotFound
	}

	if clinicActive && current.IsAdministrator &&
		!r.hasOtherActiveWithFlag(clinicID, id, func(d Dentist) bool { return d.IsAdministrator }) {
		return ErrLastAdminRequired
	}
	if clinicActive && current.IsLegalRepresentative &&
		!r.hasOtherActiveWithFlag(clinicID, id, func(d Dentist) bool { return d.IsLegalRepresentative }) {
		return ErrLastLegalRepresentativeRequired
	}

	current.DeletedAt = &deletedAt
	current.UpdatedAt = deletedAt
	return r.store.Update(id, current)
}

func (r *memoryRepository) List(ctx context.Context, clinicID string, params ListParams) (ListResult, error) {
	if _, err := r.clinics.GetByID(ctx, clinicID); err != nil {
		return ListResult{}, ErrNotFound
	}

	var items []Dentist
	for _, d := range r.store.All() {
		if d.ClinicID != clinicID || d.DeletedAt != nil {
			continue
		}
		if params.IsAdministrator != nil && d.IsAdministrator != *params.IsAdministrator {
			continue
		}
		if params.IsLegalRepresentative != nil && d.IsLegalRepresentative != *params.IsLegalRepresentative {
			continue
		}
		items = append(items, d)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })

	total := len(items)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)
	return ListResult{Items: items[start:end], Total: total}, nil
}

func (r *memoryRepository) hasOtherActiveWithFlag(clinicID, excludeID string, flag func(Dentist) bool) bool {
	return slices.ContainsFunc(r.store.All(), func(d Dentist) bool {
		return d.ClinicID == clinicID && d.ID != excludeID && d.DeletedAt == nil && flag(d)
	})
}

func (r *memoryRepository) hasActiveWithFlag(clinicID string, flag func(Dentist) bool) bool {
	return slices.ContainsFunc(r.store.All(), func(d Dentist) bool {
		return d.ClinicID == clinicID && d.DeletedAt == nil && flag(d)
	})
}
