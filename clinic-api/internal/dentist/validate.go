package dentist

import "strings"

func validateCreate(in CreateInput) error {
	fields := map[string]string{}

	if strings.TrimSpace(in.Name) == "" {
		fields["name"] = "must not be empty"
	}
	if strings.TrimSpace(in.Phone) == "" {
		fields["phone"] = "must not be empty"
	}
	if strings.TrimSpace(in.Email) == "" {
		fields["email"] = "must not be empty"
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func validateRolesInput(in RolesInput) error {
	if in.IsAdministrator == nil && in.IsLegalRepresentative == nil {
		return &ValidationError{Fields: map[string]string{
			"roles": "at least one of is_administrator, is_legal_representative must be present",
		}}
	}
	return nil
}

func validateUpdate(in UpdateInput) error {
	fields := map[string]string{}

	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		fields["name"] = "must not be empty"
	}
	if in.Phone != nil && strings.TrimSpace(*in.Phone) == "" {
		fields["phone"] = "must not be empty"
	}
	if in.Email != nil && strings.TrimSpace(*in.Email) == "" {
		fields["email"] = "must not be empty"
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
