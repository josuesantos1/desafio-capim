package payment

import "time"

// Payment status lifecycle: a payment is created as StatusPending and
// transitions to StatusApproved on its own, via a background goroutine
// that simulates the Pix provider's confirmation after a random 2-5s delay.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
)

// Payment is a simulated Pix charge against a clinic, optionally
// attributed to one of its dentists.
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

// CreateInput is the request body for POST /payments. The clinic must be
// active (have an administrator and a legal representative) to receive
// payments; DentistID is optional and, when set, must belong to ClinicID.
type CreateInput struct {
	ClinicID string `json:"clinic_id" example:"a0000000-0000-0000-0000-000000000001"`
	// Amount is in cents (e.g. 15000 = R$ 150,00). Must be greater than zero.
	Amount    int64   `json:"amount" example:"15000"`
	DentistID *string `json:"dentist_id"`
}

type ListParams struct {
	Limit    int
	Offset   int
	ClinicID string
	Status   *string
}

type ListResult struct {
	Items []Payment
	Total int
}
