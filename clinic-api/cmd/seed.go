package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/internal/payment"
)

// seedData inserts a fixed demo dataset directly into the in-memory
// repositories, bypassing service-layer validation and state
// transitions (e.g. clinic pending->active) so the seeded records can
// be created already in their final state. Only runs when SEED_DATA
// is enabled — never in normal `make run`/`go test` usage.
func seedData(ctx context.Context, clinicRepo clinic.Repository, dentistRepo dentist.Repository, paymentRepo payment.Repository) {
	now := time.Now().UTC()

	seedSorriso(ctx, clinicRepo, dentistRepo, paymentRepo, now)
	seedVida(ctx, clinicRepo, dentistRepo, paymentRepo, now)
	seedAurora(ctx, clinicRepo, now)

	slog.Info("seed data loaded", "clinics", 3)
}

// seedSorriso: active clinic, full profile, predictable email
// (demo@clinic.com) for easy mock-login testing.
func seedSorriso(ctx context.Context, clinicRepo clinic.Repository, dentistRepo dentist.Repository, paymentRepo payment.Repository, now time.Time) {
	const id = "a0000000-0000-0000-0000-000000000001"
	c := clinic.Clinic{
		ID:           id,
		Document:     "11122233344",
		LegalName:    "Clínica Sorriso LTDA",
		TradeName:    "Clínica Sorriso",
		Description:  "Clínica odontológica completa, especializada em ortodontia e implantes, com atendimento humanizado há mais de 10 anos.",
		Address:      &clinic.Address{Street: "Av. Paulista, 1000", City: "São Paulo", State: "SP", ZipCode: "01310-100"},
		Phone:        "(11) 4000-1000",
		Email:        "demo@clinic.com",
		Website:      "www.clinicasorriso.com.br",
		OpeningHours: "Seg a Sex, 8h às 18h",
		Specialties:  []string{"Ortodontia", "Implantodontia", "Estética"},
		Status:       clinic.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	logIfErr("seed clinic sorriso", clinicRepo.Create(ctx, c))

	anaID := "a0000000-0000-0000-0000-000000000011"
	logIfErr("seed dentist ana", dentistRepo.Create(ctx, dentist.Dentist{
		ID: anaID, ClinicID: id, Name: "Dra. Ana Souza", Phone: "(11) 98888-0001", Email: "ana.souza@clinicasorriso.com.br",
		Bio: "Especialista em ortodontia, com foco em harmonização do sorriso.", Specialties: []string{"Ortodontia", "Invisalign"}, YearsOfExperience: 12,
		IsAdministrator: true, IsLegalRepresentative: true,
		CreatedAt: now, UpdatedAt: now,
	}))
	logIfErr("seed dentist bruno", dentistRepo.Create(ctx, dentist.Dentist{
		ID: "a0000000-0000-0000-0000-000000000012", ClinicID: id, Name: "Dr. Bruno Lima", Phone: "(11) 98888-0002", Email: "bruno.lima@clinicasorriso.com.br",
		Bio: "Atua com implantodontia e reabilitação oral.", Specialties: []string{"Implantodontia"}, YearsOfExperience: 7,
		CreatedAt: now, UpdatedAt: now,
	}))

	approvedAt := now.Add(-24 * time.Hour)
	_, _, err := paymentRepo.Create(ctx, payment.Payment{
		ID: "a0000000-0000-0000-0000-000000000021", ClinicID: id, DentistID: &anaID,
		AmountCents: 25000, Status: payment.StatusApproved, PixCode: "00020126SEEDsorriso0001", IdempotencyKey: "seed-sorriso-1",
		CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: approvedAt, ApprovedAt: &approvedAt,
	})
	logIfErr("seed payment sorriso 1", err)

	_, _, err = paymentRepo.Create(ctx, payment.Payment{
		ID: "a0000000-0000-0000-0000-000000000022", ClinicID: id,
		AmountCents: 18000, Status: payment.StatusApproved, PixCode: "00020126SEEDsorriso0002", IdempotencyKey: "seed-sorriso-2",
		CreatedAt: now.Add(-30 * time.Hour), UpdatedAt: approvedAt, ApprovedAt: &approvedAt,
	})
	logIfErr("seed payment sorriso 2", err)

	_, _, err = paymentRepo.Create(ctx, payment.Payment{
		ID: "a0000000-0000-0000-0000-000000000023", ClinicID: id,
		AmountCents: 32000, Status: payment.StatusPending, PixCode: "00020126SEEDsorriso0003", IdempotencyKey: "seed-sorriso-3",
		CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour),
	})
	logIfErr("seed payment sorriso 3", err)
}

// seedVida: second active clinic, full profile, different email — to
// demonstrate the "minhas clínicas" ownership filter isolating owners.
func seedVida(ctx context.Context, clinicRepo clinic.Repository, dentistRepo dentist.Repository, paymentRepo payment.Repository, now time.Time) {
	const id = "a0000000-0000-0000-0000-000000000002"
	c := clinic.Clinic{
		ID:           id,
		Document:     "55566677788",
		LegalName:    "Clínica Vida Odontologia LTDA",
		TradeName:    "Clínica Vida",
		Description:  "Estrutura completa para toda a família, do check-up de rotina a procedimentos especializados.",
		Address:      &clinic.Address{Street: "Rua das Flores, 250", City: "Curitiba", State: "PR", ZipCode: "80010-000"},
		Phone:        "(41) 3222-5000",
		Email:        "contato@vida.com.br",
		Website:      "www.clinicavida.com.br",
		OpeningHours: "Seg a Sáb, 9h às 19h",
		Specialties:  []string{"Odontopediatria", "Clínica Geral"},
		Status:       clinic.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	logIfErr("seed clinic vida", clinicRepo.Create(ctx, c))

	logIfErr("seed dentist carla", dentistRepo.Create(ctx, dentist.Dentist{
		ID: "a0000000-0000-0000-0000-000000000031", ClinicID: id, Name: "Dra. Carla Mendes", Phone: "(41) 98888-0003", Email: "carla.mendes@clinicavida.com.br",
		Bio: "Dedicada à odontopediatria e ao cuidado infantil.", Specialties: []string{"Odontopediatria"}, YearsOfExperience: 9,
		IsAdministrator: true, IsLegalRepresentative: true,
		CreatedAt: now, UpdatedAt: now,
	}))
	logIfErr("seed dentist diego", dentistRepo.Create(ctx, dentist.Dentist{
		ID: "a0000000-0000-0000-0000-000000000032", ClinicID: id, Name: "Dr. Diego Alves", Phone: "(41) 98888-0004", Email: "diego.alves@clinicavida.com.br",
		Bio: "Clínico geral com atendimento humanizado.", Specialties: []string{"Clínica Geral"}, YearsOfExperience: 4,
		CreatedAt: now, UpdatedAt: now,
	}))

	_, _, err := paymentRepo.Create(ctx, payment.Payment{
		ID: "a0000000-0000-0000-0000-000000000041", ClinicID: id,
		AmountCents: 12000, Status: payment.StatusPending, PixCode: "00020126SEEDvida0001", IdempotencyKey: "seed-vida-1",
		CreatedAt: now.Add(-5 * time.Hour), UpdatedAt: now.Add(-5 * time.Hour),
	})
	logIfErr("seed payment vida 1", err)

	_, _, err = paymentRepo.Create(ctx, payment.Payment{
		ID: "a0000000-0000-0000-0000-000000000042", ClinicID: id,
		AmountCents: 9000, Status: payment.StatusPending, PixCode: "00020126SEEDvida0002", IdempotencyKey: "seed-vida-2",
		CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour),
	})
	logIfErr("seed payment vida 2", err)
}

// seedAurora: pending clinic, no profile filled and no dentists yet —
// demonstrates the "pending" badge and empty-profile placeholders.
func seedAurora(ctx context.Context, clinicRepo clinic.Repository, now time.Time) {
	logIfErr("seed clinic aurora", clinicRepo.Create(ctx, clinic.Clinic{
		ID: "a0000000-0000-0000-0000-000000000003", Document: "99988877766",
		LegalName: "Clínica Aurora LTDA", TradeName: "Clínica Aurora", Email: "aurora@example.com",
		Status: clinic.StatusPending, CreatedAt: now, UpdatedAt: now,
	}))
}

func logIfErr(action string, err error) {
	if err != nil {
		slog.Error("seed step failed", "action", action, "error", err)
	}
}
