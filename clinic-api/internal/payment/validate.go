package payment

import "github.com/google/uuid"

func validateCreate(in CreateInput) error {
	fields := map[string]string{}

	if in.ClinicID == "" {
		fields["clinic_id"] = "must not be empty"
	} else if _, err := uuid.Parse(in.ClinicID); err != nil {
		fields["clinic_id"] = "must be a valid UUID"
	}

	if in.Amount <= 0 {
		fields["amount"] = "must be greater than zero"
	}

	if in.DentistID != nil {
		if _, err := uuid.Parse(*in.DentistID); err != nil {
			fields["dentist_id"] = "must be a valid UUID"
		}
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
