package domain

import (
	"time"
)

type LedgerRecord struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	Amount        int           `json:"amount"`
	ReferenceID   *string       `json:"reference_id"`
	OperationType OperationType `json:"operation_type"`
	CreatedAt     time.Time     `json:"created_at"`
}

type TokenPackage struct {
	ID            string `json:"id"`
	StripePriceID string `json:"stripe_price_id"`
	TokensAmount  int    `json:"tokens_amount"`
	PriceCents    int    `json:"price_cents"`
	Currency      string `json:"currency"`
	IsActive      bool   `json:"is_active"`
}
