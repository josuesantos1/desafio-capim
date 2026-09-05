package clinic

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/josuesantos1/desafio/pkg/storage"
)

var (
	ErrNotFound       = errors.New("clinic: not found")
	ErrDocumentExists = errors.New("clinic: document already exists")
)

type Repository interface {
	Create(ctx context.Context, c Clinic) error
	GetByID(ctx context.Context, id string) (Clinic, error)
	GetByDocument(ctx context.Context, document string) (Clinic, error)
	Update(ctx context.Context, c Clinic) error
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
	List(ctx context.Context, params ListParams) (ListResult, error)
}

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

func (r *memoryRepository) List(ctx context.Context, params ListParams) (ListResult, error) {
	q := strings.ToLower(params.Query)
	city := strings.ToLower(params.City)

	var items []Clinic
	for _, c := range r.store.All() {
		if c.DeletedAt != nil {
			continue
		}
		if q != "" && !matchesQuery(c, q) {
			continue
		}
		if city != "" && (c.Address == nil || !strings.Contains(strings.ToLower(c.Address.City), city)) {
			continue
		}
		items = append(items, c)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })

	total := len(items)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)
	return ListResult{Items: items[start:end], Total: total}, nil
}

func matchesQuery(c Clinic, q string) bool {
	if strings.Contains(strings.ToLower(c.TradeName), q) || strings.Contains(strings.ToLower(c.LegalName), q) {
		return true
	}
	for _, s := range c.Specialties {
		if strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}

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
