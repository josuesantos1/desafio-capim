package payment

import "errors"

var (
	ErrValidation      = errors.New("payment: validation failed")
	ErrClinicNotFound  = errors.New("payment: clinic not found")
	ErrDentistNotFound = errors.New("payment: dentist not found")
	ErrClinicNotActive = errors.New("payment: clinic not active")

	ErrIdempotencyKeyConflict = errors.New("payment: idempotency key reused with a different payload")
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
