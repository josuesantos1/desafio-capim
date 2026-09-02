package dentist

import "errors"

var (
	ErrValidation     = errors.New("dentist: validation failed")
	ErrClinicNotFound = errors.New("dentist: clinic not found")
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
