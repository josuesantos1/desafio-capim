package dentist

import "time"

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

type CreateInput struct {
	Name              string   `json:"name"`
	Phone             string   `json:"phone"`
	Email             string   `json:"email"`
	Bio               string   `json:"bio"`
	Specialties       []string `json:"specialties"`
	YearsOfExperience int      `json:"years_of_experience"`
}

type UpdateInput struct {
	Name              *string   `json:"name"`
	Phone             *string   `json:"phone"`
	Email             *string   `json:"email"`
	Bio               *string   `json:"bio"`
	Specialties       *[]string `json:"specialties"`
	YearsOfExperience *int      `json:"years_of_experience"`
}

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
