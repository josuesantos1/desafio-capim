package dentist

import "time"

// Dentist belongs to exactly one clinic (ClinicID). A clinic transitions
// to "active" once it has at least one dentist with IsAdministrator=true
// and one (possibly the same) with IsLegalRepresentative=true.
type Dentist struct {
	ID                    string
	ClinicID              string
	Name                  string
	Phone                 string
	Email                 string
	Bio                   string
	Specialties           []string
	YearsOfExperience     int
	IsAdministrator       bool
	IsLegalRepresentative bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

// CreateInput is the request body for POST /clinics/{clinic_id}/dentists.
type CreateInput struct {
	Name              string   `json:"name" example:"Dra. Ana Souza"`
	Phone             string   `json:"phone" example:"(11) 98888-0001"`
	Email             string   `json:"email" example:"ana.souza@clinicasorriso.com.br"`
	Bio               string   `json:"bio" example:"Especialista em ortodontia"`
	Specialties       []string `json:"specialties" example:"Ortodontia,Invisalign"`
	YearsOfExperience int      `json:"years_of_experience" example:"12"`
}

// UpdateInput is the request body for PUT /clinics/{clinic_id}/dentists/{id}.
// All fields are optional pointers: only non-nil fields are applied.
type UpdateInput struct {
	Name              *string   `json:"name" example:"Dra. Ana Souza"`
	Phone             *string   `json:"phone" example:"(11) 98888-0001"`
	Email             *string   `json:"email" example:"ana.souza@clinicasorriso.com.br"`
	Bio               *string   `json:"bio" example:"Especialista em ortodontia"`
	Specialties       *[]string `json:"specialties"`
	YearsOfExperience *int      `json:"years_of_experience" example:"12"`
}

// RolesInput is the request body for PATCH /clinics/{clinic_id}/dentists/{id}/roles.
// Only the roles present in the payload are changed. An active clinic must
// always keep at least one administrator and one legal representative:
// demoting the last one of either role is rejected (see
// ErrLastAdminRequired / ErrLastLegalRepresentativeRequired).
type RolesInput struct {
	IsAdministrator       *bool `json:"is_administrator,omitempty" example:"true"`
	IsLegalRepresentative *bool `json:"is_legal_representative,omitempty" example:"true"`
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
