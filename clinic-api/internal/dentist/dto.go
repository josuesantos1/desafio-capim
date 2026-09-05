package dentist

import "time"

type dentistResponse struct {
	ID                    string    `json:"id"`
	ClinicID              string    `json:"clinic_id"`
	Name                  string    `json:"name"`
	Phone                 string    `json:"phone"`
	Email                 string    `json:"email"`
	Bio                   string    `json:"bio"`
	Specialties           []string  `json:"specialties"`
	YearsOfExperience     int       `json:"years_of_experience"`
	IsAdministrator       bool      `json:"is_administrator"`
	IsLegalRepresentative bool      `json:"is_legal_representative"`
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
