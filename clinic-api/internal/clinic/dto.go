package clinic

import "time"

type bankingResponse struct {
	Bank    string `json:"bank" example:"Banco do Brasil"`
	Agency  string `json:"agency" example:"1234"`
	Account string `json:"account" example:"56789-0"`
}

type addressDTO struct {
	Street  string `json:"street" example:"Av. Paulista, 1000"`
	City    string `json:"city" example:"São Paulo"`
	State   string `json:"state" example:"SP"`
	ZipCode string `json:"zip_code" example:"01310-100"`
}

// clinicResponse is the JSON representation of a Clinic returned by the API.
type clinicResponse struct {
	ID           string           `json:"id" example:"a0000000-0000-0000-0000-000000000001"`
	Document     string           `json:"document" example:"12345678000199"`
	LegalName    string           `json:"legal_name" example:"Clínica Sorriso LTDA"`
	TradeName    string           `json:"trade_name" example:"Clínica Sorriso"`
	Banking      *bankingResponse `json:"banking"`
	Description  string           `json:"description" example:"Clínica odontológica completa"`
	Address      *addressDTO      `json:"address"`
	Phone        string           `json:"phone" example:"(11) 4000-1000"`
	Email        string           `json:"email" example:"contato@clinicasorriso.com.br"`
	Website      string           `json:"website" example:"www.clinicasorriso.com.br"`
	OpeningHours string           `json:"opening_hours" example:"Seg a Sex, 8h às 18h"`
	Specialties  []string         `json:"specialties"`
	Status       ClinicStatus     `json:"status" example:"active" enums:"pending,active"`
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
