package dentist

import "time"

// Dentist is the domain type — no json tags. Serialization for the API
// response format is handled by dto.go.
type Dentist struct {
	ID        string
	ClinicID  string
	Name      string
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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

type ListParams struct {
	Limit  int
	Offset int
}

type ListResult struct {
	Items []Dentist
	Total int
}
