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

// clinicGetter/dentistGetter mirror the minimal consumer-side
// interfaces already established in internal/dentist — satisfied
// structurally by clinic.Repository / dentist.Repository without an
// adapter.
type clinicGetter interface {
	GetByID(ctx context.Context, id string) (clinic.Clinic, error)
}

type dentistGetter interface {
	GetByID(ctx context.Context, clinicID, id string) (dentist.Dentist, error)
}

// PixProvider is the minimal interface Service needs from a Pix
// gateway — satisfied structurally by *pix.Client (pkg/pix).
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

// Option configures optional Service behavior — currently only used
// to override the approval delay in tests, so they don't need to
// wait 2-5 real seconds for the background approval goroutine.
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

func (s *Service) Create(ctx context.Context, in CreateInput) (Payment, error) {
	if err := validateCreate(in); err != nil {
		return Payment{}, err
	}

	c, err := s.clinicRepo.GetByID(ctx, in.ClinicID)
	if err != nil {
		if errors.Is(err, clinic.ErrNotFound) {
			return Payment{}, ErrClinicNotFound
		}
		return Payment{}, fmt.Errorf("payment: check clinic: %w", err)
	}
	if c.Status != clinic.StatusActive {
		return Payment{}, ErrClinicNotActive
	}

	if in.DentistID != nil {
		if _, err := s.dentistRepo.GetByID(ctx, in.ClinicID, *in.DentistID); err != nil {
			if errors.Is(err, dentist.ErrNotFound) {
				return Payment{}, ErrDentistNotFound
			}
			return Payment{}, fmt.Errorf("payment: check dentist: %w", err)
		}
	}

	id := uuid.NewString()

	charge, err := s.pixProvider.CreateCharge(ctx, pix.ChargeRequest{
		AmountCents: in.Amount,
		ReferenceID: id,
	})
	if err != nil {
		return Payment{}, fmt.Errorf("payment: create charge: %w", err)
	}

	now := time.Now().UTC()
	p := Payment{
		ID:          id,
		ClinicID:    in.ClinicID,
		DentistID:   in.DentistID,
		AmountCents: in.Amount,
		Status:      StatusPending,
		PixCode:     charge.CopyPasteCode,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return Payment{}, fmt.Errorf("payment: create: %w", err)
	}

	s.scheduleApproval(p.ID)

	return p, nil
}

func (s *Service) Get(ctx context.Context, id string) (Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Payment{}, fmt.Errorf("payment: get: %w", err)
	}
	return p, nil
}

// scheduleApproval simulates, in the background, the asynchronous
// confirmation a real Pix webhook would deliver. It uses
// context.Background() deliberately: the HTTP request context that
// triggered Create is already gone by the time this runs.
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
