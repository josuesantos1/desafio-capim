package dentist

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name      string
		in        CreateInput
		wantErr   bool
		wantField string
	}{
		{name: "valid", in: CreateInput{Name: "Dr. A", Phone: "123", Email: "a@x.com"}},
		{name: "missing name", in: CreateInput{Name: " ", Phone: "123", Email: "a@x.com"}, wantErr: true, wantField: "name"},
		{name: "missing phone", in: CreateInput{Name: "Dr. A", Phone: "", Email: "a@x.com"}, wantErr: true, wantField: "phone"},
		{name: "missing email", in: CreateInput{Name: "Dr. A", Phone: "123", Email: ""}, wantErr: true, wantField: "email"},
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

func TestValidateRolesInput(t *testing.T) {
	yes := true
	no := false

	tests := []struct {
		name    string
		in      RolesInput
		wantErr bool
	}{
		{name: "both absent is invalid", in: RolesInput{}, wantErr: true},
		{name: "only is_administrator present", in: RolesInput{IsAdministrator: &yes}},
		{name: "only is_legal_representative present", in: RolesInput{IsLegalRepresentative: &no}},
		{name: "both present", in: RolesInput{IsAdministrator: &yes, IsLegalRepresentative: &no}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRolesInput(tt.in)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			var ve *ValidationError
			require.True(t, errors.As(err, &ve))
			assert.Contains(t, ve.Fields, "roles")
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	empty := ""
	valid := "value"

	tests := []struct {
		name      string
		in        UpdateInput
		wantErr   bool
		wantField string
	}{
		{name: "empty input is valid", in: UpdateInput{}},
		{name: "valid name", in: UpdateInput{Name: &valid}},
		{name: "empty name", in: UpdateInput{Name: &empty}, wantErr: true, wantField: "name"},
		{name: "empty phone", in: UpdateInput{Phone: &empty}, wantErr: true, wantField: "phone"},
		{name: "empty email", in: UpdateInput{Email: &empty}, wantErr: true, wantField: "email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdate(tt.in)

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
