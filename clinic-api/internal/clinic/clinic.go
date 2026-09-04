package clinic

import "time"

type ClinicStatus string

const (
	StatusPending ClinicStatus = "pending"
	StatusActive  ClinicStatus = "active"
)

type Clinic struct {
	ID        string
	Document  string
	LegalName string
	TradeName string
	Bank      *string
	Agency    *string
	Account   *string
	Status    ClinicStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CreateInput struct {
	Document  string   `json:"document"`
	LegalName string   `json:"legal_name"`
	TradeName string   `json:"trade_name"`
	Banking   *Banking `json:"banking"`
}

type UpdateInput struct {
	LegalName *string  `json:"legal_name"`
	TradeName *string  `json:"trade_name"`
	Banking   *Banking `json:"banking"`
	Document  *string  `json:"document"`
}

type Banking struct {
	Bank    string `json:"bank"`
	Agency  string `json:"agency"`
	Account string `json:"account"`
}
