package payment

import "time"

// paymentResponse is the JSON representation of a Payment returned by the
// API. Status starts as "pending" and transitions to "approved" on its own
// (see payment.Service.scheduleApproval); poll GET /payments/{id} to observe it.
type paymentResponse struct {
	ID         string     `json:"id" example:"62818880-fffa-4512-bb48-34692f49e3cb"`
	ClinicID   string     `json:"clinic_id" example:"a0000000-0000-0000-0000-000000000001"`
	DentistID  *string    `json:"dentist_id"`
	Amount     int64      `json:"amount" example:"15000"`
	Status     string     `json:"status" example:"pending" enums:"pending,approved"`
	PixCode    string     `json:"pix_code" example:"000201262yZEd13q_BCi_wyx1gxaDLUbQfdU7oY5dm7GMumLL_A"`
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

type listResponse struct {
	Items  []paymentResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func toListResponse(result ListResult, params ListParams) listResponse {
	items := make([]paymentResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = toPaymentResponse(p)
	}
	return listResponse{Items: items, Total: result.Total, Limit: params.Limit, Offset: params.Offset}
}
