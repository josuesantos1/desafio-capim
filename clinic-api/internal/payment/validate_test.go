package payment

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreate(t *testing.T) {
	validClinicID := "11111111-1111-1111-1111-111111111111"
	validDentistID := "22222222-2222-2222-2222-222222222222"
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name      string
		in        CreateInput
		wantErr   bool
		wantField string
	}{
		{name: "valid, no dentist", in: CreateInput{ClinicID: validClinicID, Amount: 15000}},
		{name: "valid, with dentist", in: CreateInput{ClinicID: validClinicID, Amount: 15000, DentistID: &validDentistID}},
		{name: "missing clinic_id", in: CreateInput{Amount: 15000}, wantErr: true, wantField: "clinic_id"},
		{name: "invalid clinic_id", in: CreateInput{ClinicID: invalidUUID, Amount: 15000}, wantErr: true, wantField: "clinic_id"},
		{name: "zero amount", in: CreateInput{ClinicID: validClinicID, Amount: 0}, wantErr: true, wantField: "amount"},
		{name: "negative amount", in: CreateInput{ClinicID: validClinicID, Amount: -100}, wantErr: true, wantField: "amount"},
		{name: "invalid dentist_id", in: CreateInput{ClinicID: validClinicID, Amount: 15000, DentistID: &invalidUUID}, wantErr: true, wantField: "dentist_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreate(tt.in)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			var ve *ValidationError
			require.True(t, errors.As(err, &ve))
			assert.Contains(t, ve.Fields, tt.wantField)
		})
	}
}
