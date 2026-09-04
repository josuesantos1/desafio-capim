package payment

import "errors"

var (
	ErrValidation      = errors.New("payment: validation failed")
	ErrClinicNotFound  = errors.New("payment: clinic not found")
	ErrDentistNotFound = errors.New("payment: dentist not found")
	ErrClinicNotActive = errors.New("payment: clinic not active")

	// ErrIdempotencyKeyConflict indicates the Idempotency-Key was
	// already used by a payment whose business fields (ClinicID,
	// AmountCents, DentistID) differ from the current request.
	ErrIdempotencyKeyConflict = errors.New("payment: idempotency key reused with a different payload")
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
