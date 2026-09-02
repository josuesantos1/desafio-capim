package payment

import "time"

type paymentResponse struct {
	ID         string     `json:"id"`
	ClinicID   string     `json:"clinic_id"`
	DentistID  *string    `json:"dentist_id"`
	Amount     int64      `json:"amount"`
	Status     string     `json:"status"`
	PixCode    string     `json:"pix_code"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ApprovedAt *time.Time `json:"approved_at"`
}

func toPaymentResponse(p Payment) paymentResponse {
	return paymentResponse{
		ID:         p.ID,
		ClinicID:   p.ClinicID,
		DentistID:  p.DentistID,
		Amount:     p.AmountCents,
		Status:     p.Status,
		PixCode:    p.PixCode,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		ApprovedAt: p.ApprovedAt,
	}
}

type errorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
