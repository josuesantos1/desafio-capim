package dentist

import "errors"

var (
	ErrValidation     = errors.New("dentist: validation failed")
	ErrClinicNotFound = errors.New("dentist: clinic not found")

	// ErrLastAdminRequired and ErrLastLegalRepresentativeRequired guard
	// an active clinic's invariant: it must always keep at least one
	// active administrator and one active legal representative. They
	// only ever trigger while the clinic is active (see repository.go)
	// — a pending clinic has no such restriction.
	ErrLastAdminRequired               = errors.New("dentist: last administrator required")
	ErrLastLegalRepresentativeRequired = errors.New("dentist: last legal representative required")
)

// ValidationError wraps ErrValidation carrying the invalid fields and
// their messages, so the handler can build the "fields" object of the
// error response without re-deriving which fields failed.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return ErrValidation.Error()
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}
