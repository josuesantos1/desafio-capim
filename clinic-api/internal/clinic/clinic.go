package clinic

import "time"

type ClinicStatus string

const (
	StatusPending ClinicStatus = "pending"
	StatusActive  ClinicStatus = "active"
)

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

type Clinic struct {
	ID           string
	Document     string
	LegalName    string
	TradeName    string
	Bank         *string
	Agency       *string
	Account      *string
	Description  string
	Address      *Address
	Phone        string
	Email        string
	Website      string
	OpeningHours string
	Specialties  []string
	Status       ClinicStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type CreateInput struct {
	Document     string   `json:"document"`
	LegalName    string   `json:"legal_name"`
	TradeName    string   `json:"trade_name"`
	Banking      *Banking `json:"banking"`
	Description  string   `json:"description"`
	Address      *Address `json:"address"`
	Phone        string   `json:"phone"`
	Email        string   `json:"email"`
	Website      string   `json:"website"`
	OpeningHours string   `json:"opening_hours"`
	Specialties  []string `json:"specialties"`
}

type UpdateInput struct {
	LegalName    *string   `json:"legal_name"`
	TradeName    *string   `json:"trade_name"`
	Banking      *Banking  `json:"banking"`
	Document     *string   `json:"document"`
	Description  *string   `json:"description"`
	Address      *Address  `json:"address"`
	Phone        *string   `json:"phone"`
	Email        *string   `json:"email"`
	Website      *string   `json:"website"`
	OpeningHours *string   `json:"opening_hours"`
	Specialties  *[]string `json:"specialties"`
}

type Banking struct {
	Bank    string `json:"bank"`
	Agency  string `json:"agency"`
	Account string `json:"account"`
}

type ListParams struct {
	Limit  int
	Offset int
	Query  string
	City   string
}

type ListResult struct {
	Items []Clinic
	Total int
}
