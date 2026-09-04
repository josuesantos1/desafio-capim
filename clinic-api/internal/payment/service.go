package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/pkg/pix"
)

type clinicGetter interface {
	GetByID(ctx context.Context, id string) (clinic.Clinic, error)
}

type dentistGetter interface {
	GetByID(ctx context.Context, clinicID, id string) (dentist.Dentist, error)
}

type PixProvider interface {
	CreateCharge(ctx context.Context, req pix.ChargeRequest) (pix.ChargeResponse, error)
}

type Service struct {
	repo          Repository
	clinicRepo    clinicGetter
	dentistRepo   dentistGetter
	pixProvider   PixProvider
	approvalDelay func() time.Duration
}

type Option func(*Service)

func WithApprovalDelay(f func() time.Duration) Option {
	return func(s *Service) { s.approvalDelay = f }
}

func NewService(repo Repository, clinicRepo clinicGetter, dentistRepo dentistGetter, pixProvider PixProvider, opts ...Option) *Service {
	s := &Service{
		repo:          repo,
		clinicRepo:    clinicRepo,
		dentistRepo:   dentistRepo,
		pixProvider:   pixProvider,
		approvalDelay: randomApprovalDelay,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func randomApprovalDelay() time.Duration {
	return time.Duration(2+rand.Intn(4)) * time.Second
}

func (s *Service) Create(ctx context.Context, idempotencyKey string, in CreateInput) (Payment, bool, error) {
	if err := validateCreate(in); err != nil {
		return Payment{}, false, err
	}

	c, err := s.clinicRepo.GetByID(ctx, in.ClinicID)
	if err != nil {
		if errors.Is(err, clinic.ErrNotFound) {
			return Payment{}, false, ErrClinicNotFound
		}
		return Payment{}, false, fmt.Errorf("payment: check clinic: %w", err)
	}
	if c.Status != clinic.StatusActive {
		return Payment{}, false, ErrClinicNotActive
	}

	if in.DentistID != nil {
		if _, err := s.dentistRepo.GetByID(ctx, in.ClinicID, *in.DentistID); err != nil {
			if errors.Is(err, dentist.ErrNotFound) {
				return Payment{}, false, ErrDentistNotFound
			}
			return Payment{}, false, fmt.Errorf("payment: check dentist: %w", err)
		}
	}

	id := uuid.NewString()

	charge, err := s.pixProvider.CreateCharge(ctx, pix.ChargeRequest{
		AmountCents: in.Amount,
		ReferenceID: id,
	})
	if err != nil {
		return Payment{}, false, fmt.Errorf("payment: create charge: %w", err)
	}

	now := time.Now().UTC()
	p := Payment{
		ID:             id,
		ClinicID:       in.ClinicID,
		DentistID:      in.DentistID,
		AmountCents:    in.Amount,
		Status:         StatusPending,
		PixCode:        charge.CopyPasteCode,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, created, err := s.repo.Create(ctx, p)
	if err != nil {
		return Payment{}, false, fmt.Errorf("payment: create: %w", err)
	}

	if created {
		s.scheduleApproval(result.ID)
	}

	return result, created, nil
}

func (s *Service) Get(ctx context.Context, id string) (Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Payment{}, fmt.Errorf("payment: get: %w", err)
	}
	return p, nil
}

func (s *Service) scheduleApproval(id string) {
	delay := s.approvalDelay()
	go func() {
		time.Sleep(delay)
		approvedAt := time.Now().UTC()
		if err := s.repo.Approve(context.Background(), id, approvedAt); err != nil {
			slog.Error("failed to approve payment", "payment_id", id, "error", err)
		}
	}()
}
