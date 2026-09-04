package dentist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/josuesantos1/desafio/internal/clinic"
)

// clinicGetter is the minimal consumer-side interface dentist needs
// from internal/clinic — only existence/activity lookup, not the full
// clinic.Repository. clinic.Repository already satisfies this
// interface structurally, so callers (including tests, via the
// existing clinic mock) pass a clinic.Repository value without any
// adapter.
type clinicGetter interface {
	GetByID(ctx context.Context, id string) (clinic.Clinic, error)
}

type Service struct {
	repo       Repository
	clinicRepo clinicGetter
}

func NewService(repo Repository, clinicRepo clinicGetter) *Service {
	return &Service{repo: repo, clinicRepo: clinicRepo}
}

func (s *Service) Create(ctx context.Context, clinicID string, in CreateInput) (Dentist, error) {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return Dentist{}, err
	}
	if err := validateCreate(in); err != nil {
		return Dentist{}, err
	}

	now := time.Now().UTC()
	d := Dentist{
		ID:        uuid.NewString(),
		ClinicID:  clinicID,
		Name:      in.Name,
		Phone:     in.Phone,
		Email:     in.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return Dentist{}, fmt.Errorf("dentist: create: %w", err)
	}
	return d, nil
}

func (s *Service) Get(ctx context.Context, clinicID, id string) (Dentist, error) {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return Dentist{}, err
	}

	d, err := s.repo.GetByID(ctx, clinicID, id)
	if err != nil {
		return Dentist{}, fmt.Errorf("dentist: get: %w", err)
	}
	return d, nil
}

func (s *Service) Update(ctx context.Context, clinicID, id string, in UpdateInput) (Dentist, error) {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return Dentist{}, err
	}

	current, err := s.repo.GetByID(ctx, clinicID, id)
	if err != nil {
		return Dentist{}, fmt.Errorf("dentist: update: %w", err)
	}

	if err := validateUpdate(in); err != nil {
		return Dentist{}, err
	}

	if in.Name != nil {
		current.Name = *in.Name
	}
	if in.Phone != nil {
		current.Phone = *in.Phone
	}
	if in.Email != nil {
		current.Email = *in.Email
	}
	current.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, current); err != nil {
		return Dentist{}, fmt.Errorf("dentist: update: %w", err)
	}
	return current, nil
}

func (s *Service) UpdateRoles(ctx context.Context, clinicID, id string, in RolesInput) (Dentist, error) {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return Dentist{}, err
	}
	if err := validateRolesInput(in); err != nil {
		return Dentist{}, err
	}

	d, err := s.repo.UpdateRoles(ctx, clinicID, id, in)
	if err != nil {
		return Dentist{}, fmt.Errorf("dentist: update roles: %w", err)
	}
	return d, nil
}

func (s *Service) Delete(ctx context.Context, clinicID, id string) error {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, clinicID, id, time.Now().UTC()); err != nil {
		return fmt.Errorf("dentist: delete: %w", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, clinicID string, params ListParams) (ListResult, error) {
	if err := s.checkClinicActive(ctx, clinicID); err != nil {
		return ListResult{}, err
	}

	params.Limit = clampLimit(params.Limit)
	params.Offset = clampOffset(params.Offset)

	result, err := s.repo.List(ctx, clinicID, params)
	if err != nil {
		return ListResult{}, fmt.Errorf("dentist: list: %w", err)
	}
	return result, nil
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func clampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func (s *Service) checkClinicActive(ctx context.Context, clinicID string) error {
	if _, err := s.clinicRepo.GetByID(ctx, clinicID); err != nil {
		if errors.Is(err, clinic.ErrNotFound) {
			return ErrClinicNotFound
		}
		return fmt.Errorf("dentist: check clinic: %w", err)
	}
	return nil
}
