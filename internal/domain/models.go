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

type DepositRequest struct {
	ID              string
	UserID          string
	StripeSessionID *string
	PackageID       string
	CreatedAt       time.Time
}

type DepositStatusHistory struct {
	ID               string
	DepositRequestID string
	Status           DepositStatus
	CreatedAt        time.Time
}

type User struct {
	ID               string
	Email            string
	StripeCustomerID *string
	CreatedAt        time.Time
}

type UsageLog struct {
	ID         string
	UserID     string
	TokensCost int
	ActionTupe string
	Status     UsageStatus
	CreatedAt  time.Time
}
