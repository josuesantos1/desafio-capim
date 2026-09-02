package clinic

import "time"

// Clinic is the domain type — no json tags. Serialization for the API
// response format is handled by dto.go, not by this type.
type Clinic struct {
	ID        string
	Document  string
	LegalName string
	TradeName string
	Bank      *string
	Agency    *string
	Account   *string
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
	// Document is present only to detect an attempted change; it is
	// never applied to Clinic.Document.
	Document *string `json:"document"`
}

type Banking struct {
	Bank    string `json:"bank"`
	Agency  string `json:"agency"`
	Account string `json:"account"`
}
