package dentist

import "time"

// Dentist is the domain type — no json tags. Serialization for the API
// response format is handled by dto.go.
//
// IsAdministrator and IsLegalRepresentative are independent — a
// dentist can hold both, either, or neither. There is no "dentist"
// role value: both false is the default, common state.
type Dentist struct {
	ID                    string
	ClinicID              string
	Name                  string
	Phone                 string
	Email                 string
	IsAdministrator       bool
	IsLegalRepresentative bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

type CreateInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type UpdateInput struct {
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
	Email *string `json:"email"`
}

// RolesInput is the body of PATCH .../roles. At least one field must
// be present (see validateRolesInput).
type RolesInput struct {
	IsAdministrator       *bool `json:"is_administrator,omitempty"`
	IsLegalRepresentative *bool `json:"is_legal_representative,omitempty"`
}

type ListParams struct {
	Limit                 int
	Offset                int
	IsAdministrator       *bool
	IsLegalRepresentative *bool
}

type ListResult struct {
	Items []Dentist
	Total int
}
