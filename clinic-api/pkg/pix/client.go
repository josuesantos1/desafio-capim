// Package pix simulates a client SDK for a Pix payment gateway. It
// makes no network calls — it mirrors the shape of a real gateway
// integration (Config, Client, typed request/response) so that
// internal/payment can depend on it the same way it would depend on
// a real third-party SDK.
package pix

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// Config mimics what a real Pix gateway SDK would require to
// authenticate. It is unused by the simulation itself (no network
// call is made), but keeps the package shaped like a real
// integration.
type Config struct {
	APIKey string
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
}

type ChargeRequest struct {
	AmountCents int64
	// ReferenceID is the caller's own identifier for this charge (the
	// Payment's UUID) — real Pix gateways require an idempotency/
	// reference key from the merchant.
	ReferenceID string
}

type ChargeResponse struct {
	CopyPasteCode string
	ExpiresAt     time.Time
}

// CreateCharge simulates a synchronous "charge created" response from
// a real Pix gateway: it returns a copy-paste code immediately.
// Confirmation of payment is asynchronous in real Pix integrations
// too (arrives via webhook) — here it is simulated by the caller
// (internal/payment), not by this SDK, since there is no real webhook
// to receive.
func (c *Client) CreateCharge(ctx context.Context, req ChargeRequest) (ChargeResponse, error) {
	payload := fmt.Sprintf("PIX|ref=%s|amount=%d|ts=%d", req.ReferenceID, req.AmountCents, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(payload))
	code := "00020126" + base64.RawURLEncoding.EncodeToString(hash[:])

	return ChargeResponse{
		CopyPasteCode: code,
		ExpiresAt:     time.Now().Add(time.Hour),
	}, nil
}
