package payment

import "time"

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
)

// Payment is the domain type — no json tags. Serialization is
// handled by dto.go.
type Payment struct {
	ID             string
	ClinicID       string
	DentistID      *string
	AmountCents    int64
	Status         string
	PixCode        string
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ApprovedAt     *time.Time
}

type CreateInput struct {
	ClinicID  string  `json:"clinic_id"`
	Amount    int64   `json:"amount"`
	DentistID *string `json:"dentist_id"`
}
