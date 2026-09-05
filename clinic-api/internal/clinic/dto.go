package clinic

import "time"

type bankingResponse struct {
	Bank    string `json:"bank"`
	Agency  string `json:"agency"`
	Account string `json:"account"`
}

type addressDTO struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

type clinicResponse struct {
	ID           string           `json:"id"`
	Document     string           `json:"document"`
	LegalName    string           `json:"legal_name"`
	TradeName    string           `json:"trade_name"`
	Banking      *bankingResponse `json:"banking"`
	Description  string           `json:"description"`
	Address      *addressDTO      `json:"address"`
	Phone        string           `json:"phone"`
	Email        string           `json:"email"`
	Website      string           `json:"website"`
	OpeningHours string           `json:"opening_hours"`
	Specialties  []string         `json:"specialties"`
	Status       ClinicStatus     `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

func toClinicResponse(c Clinic) clinicResponse {
	resp := clinicResponse{
		ID:           c.ID,
		Document:     c.Document,
		LegalName:    c.LegalName,
		TradeName:    c.TradeName,
		Description:  c.Description,
		Phone:        c.Phone,
		Email:        c.Email,
		Website:      c.Website,
		OpeningHours: c.OpeningHours,
		Specialties:  c.Specialties,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if resp.Specialties == nil {
		resp.Specialties = []string{}
	}
	if c.Bank != nil && c.Agency != nil && c.Account != nil {
		resp.Banking = &bankingResponse{Bank: *c.Bank, Agency: *c.Agency, Account: *c.Account}
	}
	if c.Address != nil {
		resp.Address = &addressDTO{
			Street:  c.Address.Street,
			City:    c.Address.City,
			State:   c.Address.State,
			ZipCode: c.Address.ZipCode,
		}
	}
	return resp
}

type listResponse struct {
	Items  []clinicResponse `json:"items"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

func toListResponse(result ListResult, params ListParams) listResponse {
	items := make([]clinicResponse, len(result.Items))
	for i, c := range result.Items {
		items[i] = toClinicResponse(c)
	}
	return listResponse{
		Items:  items,
		Total:  result.Total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}
}
