package clinic

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name           string
		in             CreateInput
		wantErr        bool
		wantField      string
		wantNormalized string
	}{
		{
			name:           "valid CPF",
			in:             CreateInput{Document: "123.456.789-00", LegalName: "Clinic Ltda", TradeName: "Clinic"},
			wantNormalized: "12345678900",
		},
		{
			name: "valid CNPJ",
			in:   CreateInput{Document: "12.345.678/0001-95", LegalName: "Clinic Ltda", TradeName: "Clinic"},
		},
		{
			name:      "invalid document length",
			in:        CreateInput{Document: "123", LegalName: "Clinic Ltda", TradeName: "Clinic"},
			wantErr:   true,
			wantField: "document",
		},
		{
			name:      "missing legal_name",
			in:        CreateInput{Document: "12345678900", LegalName: "  ", TradeName: "Clinic"},
			wantErr:   true,
			wantField: "legal_name",
		},
		{
			name:      "missing trade_name",
			in:        CreateInput{Document: "12345678900", LegalName: "Clinic Ltda", TradeName: ""},
			wantErr:   true,
			wantField: "trade_name",
		},
		{
			name:      "partial banking",
			in:        CreateInput{Document: "12345678900", LegalName: "Clinic Ltda", TradeName: "Clinic", Banking: &Banking{Bank: "341", Agency: "", Account: "123"}},
			wantErr:   true,
			wantField: "banking",
		},
		{
			name: "full banking",
			in:   CreateInput{Document: "12345678900", LegalName: "Clinic Ltda", TradeName: "Clinic", Banking: &Banking{Bank: "341", Agency: "0001", Account: "123456-7"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := validateCreate(tt.in)

			if !tt.wantErr {
				require.NoError(t, err)
				if tt.wantNormalized != "" {
					assert.Equal(t, tt.wantNormalized, normalized)
				}
				return
			}

			var ve *ValidationError
			require.True(t, errors.As(err, &ve), "expected *ValidationError, got %v (%T)", err, err)
			assert.Contains(t, ve.Fields, tt.wantField)
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	empty := ""
	legal := "Legal"

	tests := []struct {
		name      string
		in        UpdateInput
		wantErr   bool
		wantField string
	}{
		{name: "empty input is valid", in: UpdateInput{}},
		{name: "valid legal_name", in: UpdateInput{LegalName: &legal}},
		{name: "empty legal_name", in: UpdateInput{LegalName: &empty}, wantErr: true, wantField: "legal_name"},
		{name: "empty trade_name", in: UpdateInput{TradeName: &empty}, wantErr: true, wantField: "trade_name"},
		{name: "partial banking", in: UpdateInput{Banking: &Banking{Bank: "341"}}, wantErr: true, wantField: "banking"},
		{name: "full banking", in: UpdateInput{Banking: &Banking{Bank: "341", Agency: "0001", Account: "123"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdate(tt.in)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			var ve *ValidationError
			require.True(t, errors.As(err, &ve), "expected *ValidationError, got %v (%T)", err, err)
			assert.Contains(t, ve.Fields, tt.wantField)
		})
	}
}
