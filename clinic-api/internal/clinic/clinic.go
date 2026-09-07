package clinic

import "time"

// ClinicStatus represents the lifecycle state of a clinic.
//
// A clinic starts as StatusPending and transitions to StatusActive
// automatically once it has at least one dentist marked as administrator
// and one marked as legal representative (see internal/dentist). Only an
// active clinic can receive Pix payments.
type ClinicStatus string

const (
	StatusPending ClinicStatus = "pending"
	StatusActive  ClinicStatus = "active"
)

// Address is the clinic's physical location.
type Address struct {
	Street  string `json:"street" example:"Av. Paulista, 1000"`
	City    string `json:"city" example:"São Paulo"`
	State   string `json:"state" example:"SP"`
	ZipCode string `json:"zip_code" example:"01310-100"`
}

// Clinic is the aggregate root for a dental clinic: its registration data,
// banking details, and current activation status.
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

// CreateInput is the request body for POST /clinics.
type CreateInput struct {
	// Document is the CPF/CNPJ. Immutable after creation.
	Document     string   `json:"document" example:"12345678000199"`
	LegalName    string   `json:"legal_name" example:"Clínica Sorriso LTDA"`
	TradeName    string   `json:"trade_name" example:"Clínica Sorriso"`
	Banking      *Banking `json:"banking"`
	Description  string   `json:"description" example:"Clínica odontológica completa"`
	Address      *Address `json:"address"`
	Phone        string   `json:"phone" example:"(11) 4000-1000"`
	Email        string   `json:"email" example:"contato@clinicasorriso.com.br"`
	Website      string   `json:"website" example:"www.clinicasorriso.com.br"`
	OpeningHours string   `json:"opening_hours" example:"Seg a Sex, 8h às 18h"`
	Specialties  []string `json:"specialties" example:"Ortodontia,Implantodontia"`
}

// UpdateInput is the request body for PUT /clinics/{id}. All fields are
// optional pointers: only non-nil fields are applied (partial update).
// Document, once set, cannot be changed (see ErrDocumentImmutable).
type UpdateInput struct {
	LegalName    *string   `json:"legal_name" example:"Clínica Sorriso LTDA"`
	TradeName    *string   `json:"trade_name" example:"Clínica Sorriso"`
	Banking      *Banking  `json:"banking"`
	Document     *string   `json:"document" example:"12345678000199"`
	Description  *string   `json:"description" example:"Clínica odontológica completa"`
	Address      *Address  `json:"address"`
	Phone        *string   `json:"phone" example:"(11) 4000-1000"`
	Email        *string   `json:"email" example:"contato@clinicasorriso.com.br"`
	Website      *string   `json:"website" example:"www.clinicasorriso.com.br"`
	OpeningHours *string   `json:"opening_hours" example:"Seg a Sex, 8h às 18h"`
	Specialties  *[]string `json:"specialties"`
}

// Banking holds the clinic's bank account for receiving Pix settlements.
type Banking struct {
	Bank    string `json:"bank" example:"Banco do Brasil"`
	Agency  string `json:"agency" example:"1234"`
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
