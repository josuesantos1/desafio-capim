package clinic

import "time"

type bankingResponse struct {
	Bank    string `json:"bank"`
	Agency  string `json:"agency"`
	Account string `json:"account"`
}

type clinicResponse struct {
	ID        string           `json:"id"`
	Document  string           `json:"document"`
	LegalName string           `json:"legal_name"`
	TradeName string           `json:"trade_name"`
	Banking   *bankingResponse `json:"banking"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func toClinicResponse(c Clinic) clinicResponse {
	resp := clinicResponse{
		ID:        c.ID,
		Document:  c.Document,
		LegalName: c.LegalName,
		TradeName: c.TradeName,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Bank != nil && c.Agency != nil && c.Account != nil {
		resp.Banking = &bankingResponse{Bank: *c.Bank, Agency: *c.Agency, Account: *c.Account}
	}
	return resp
}

type errorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
