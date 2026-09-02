package payment

import "errors"

var (
	ErrValidation      = errors.New("payment: validation failed")
	ErrClinicNotFound  = errors.New("payment: clinic not found")
	ErrDentistNotFound = errors.New("payment: dentist not found")
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
