package clinic

import "errors"

var (
	ErrValidation        = errors.New("clinic: validation failed")
	ErrDocumentImmutable = errors.New("clinic: document is immutable")
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
