package dentist

import "errors"

var (
	ErrValidation     = errors.New("dentist: validation failed")
	ErrClinicNotFound = errors.New("dentist: clinic not found")

	ErrLastAdminRequired               = errors.New("dentist: last administrator required")
	ErrLastLegalRepresentativeRequired = errors.New("dentist: last legal representative required")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return ErrValidation.Error()
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}
