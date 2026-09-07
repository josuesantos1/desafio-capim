package dentist

import "time"

// dentistResponse is the JSON representation of a Dentist returned by the API.
type dentistResponse struct {
	ID                    string    `json:"id" example:"a0000000-0000-0000-0000-000000000011"`
	ClinicID              string    `json:"clinic_id" example:"a0000000-0000-0000-0000-000000000001"`
	Name                  string    `json:"name" example:"Dra. Ana Souza"`
	Phone                 string    `json:"phone" example:"(11) 98888-0001"`
	Email                 string    `json:"email" example:"ana.souza@clinicasorriso.com.br"`
	Bio                   string    `json:"bio" example:"Especialista em ortodontia"`
	Specialties           []string  `json:"specialties"`
	YearsOfExperience     int       `json:"years_of_experience" example:"12"`
	IsAdministrator       bool      `json:"is_administrator" example:"true"`
	IsLegalRepresentative bool      `json:"is_legal_representative" example:"true"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func toDentistResponse(d Dentist) dentistResponse {
	resp := dentistResponse{
		ID:                    d.ID,
		ClinicID:              d.ClinicID,
		Name:                  d.Name,
		Phone:                 d.Phone,
		Email:                 d.Email,
		Bio:                   d.Bio,
		Specialties:           d.Specialties,
		YearsOfExperience:     d.YearsOfExperience,
		IsAdministrator:       d.IsAdministrator,
		IsLegalRepresentative: d.IsLegalRepresentative,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
	if resp.Specialties == nil {
		resp.Specialties = []string{}
	}
	return resp
}

type listResponse struct {
	Items  []dentistResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func toListResponse(result ListResult, params ListParams) listResponse {
	items := make([]dentistResponse, len(result.Items))
	for i, d := range result.Items {
		items[i] = toDentistResponse(d)
	}
	return listResponse{
		Items:  items,
		Total:  result.Total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}
}
