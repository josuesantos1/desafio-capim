package pix_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/pkg/pix"
)

func TestClient_CreateCharge(t *testing.T) {
	tests := []struct {
		name string
		req  pix.ChargeRequest
	}{
		{name: "with reference and amount", req: pix.ChargeRequest{AmountCents: 15000, ReferenceID: "ref-1"}},
		{name: "different amount produces different code", req: pix.ChargeRequest{AmountCents: 999, ReferenceID: "ref-1"}},
	}

	client := pix.NewClient(pix.Config{APIKey: "test-key"})

	var previousCode string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.CreateCharge(context.Background(), tt.req)

			require.NoError(t, err)
			assert.NotEmpty(t, resp.CopyPasteCode)
			assert.True(t, resp.ExpiresAt.After(time.Now()))
			assert.NotEqual(t, previousCode, resp.CopyPasteCode)
			previousCode = resp.CopyPasteCode
		})
	}
}
