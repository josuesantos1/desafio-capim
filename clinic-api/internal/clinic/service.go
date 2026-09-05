package clinic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Clinic, error) {
	normalizedDocument, err := validateCreate(in)
	if err != nil {
		return Clinic{}, err
	}

	now := time.Now().UTC()
	c := Clinic{
		ID:           uuid.NewString(),
		Document:     normalizedDocument,
		LegalName:    in.LegalName,
		TradeName:    in.TradeName,
		Description:  in.Description,
		Address:      in.Address,
		Phone:        in.Phone,
		Email:        in.Email,
		Website:      in.Website,
		OpeningHours: in.OpeningHours,
		Specialties:  in.Specialties,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if in.Banking != nil {
		c.Bank = &in.Banking.Bank
		c.Agency = &in.Banking.Agency
		c.Account = &in.Banking.Account
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return Clinic{}, fmt.Errorf("clinic: create: %w", err)
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (Clinic, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Clinic{}, fmt.Errorf("clinic: get: %w", err)
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (Clinic, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Clinic{}, fmt.Errorf("clinic: update: %w", err)
	}

	if in.Document != nil && normalizeDocument(*in.Document) != current.Document {
		return Clinic{}, ErrDocumentImmutable
	}

	if err := validateUpdate(in); err != nil {
		return Clinic{}, err
	}

	if in.LegalName != nil {
		current.LegalName = *in.LegalName
	}
	if in.TradeName != nil {
		current.TradeName = *in.TradeName
	}
	if in.Banking != nil {
		current.Bank = &in.Banking.Bank
		current.Agency = &in.Banking.Agency
		current.Account = &in.Banking.Account
	}
	if in.Description != nil {
		current.Description = *in.Description
	}
	if in.Address != nil {
		current.Address = in.Address
	}
	if in.Phone != nil {
		current.Phone = *in.Phone
	}
	if in.Email != nil {
		current.Email = *in.Email
	}
	if in.Website != nil {
		current.Website = *in.Website
	}
	if in.OpeningHours != nil {
		current.OpeningHours = *in.OpeningHours
	}
	if in.Specialties != nil {
		current.Specialties = *in.Specialties
	}
	current.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, current); err != nil {
		return Clinic{}, fmt.Errorf("clinic: update: %w", err)
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id, time.Now().UTC()); err != nil {
		return fmt.Errorf("clinic: delete: %w", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, params ListParams) (ListResult, error) {
	params.Limit = clampLimit(params.Limit)
	params.Offset = clampOffset(params.Offset)

	result, err := s.repo.List(ctx, params)
	if err != nil {
		return ListResult{}, fmt.Errorf("clinic: list: %w", err)
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
