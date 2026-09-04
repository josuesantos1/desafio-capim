// Package pix simulates a client SDK for a Pix payment gateway.
package pix

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

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
	ReferenceID string
}

type ChargeResponse struct {
	CopyPasteCode string
	ExpiresAt     time.Time
}

func (c *Client) CreateCharge(ctx context.Context, req ChargeRequest) (ChargeResponse, error) {
	payload := fmt.Sprintf("PIX|ref=%s|amount=%d|ts=%d", req.ReferenceID, req.AmountCents, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(payload))
	code := "00020126" + base64.RawURLEncoding.EncodeToString(hash[:])

	return ChargeResponse{
		CopyPasteCode: code,
		ExpiresAt:     time.Now().Add(time.Hour),
	}, nil
}
