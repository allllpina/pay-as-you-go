package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pay-as-you-use/internal/domain"
	"time"
)

type PaymentClient interface {
	DeductTokens(ctx context.Context, userID string, amount int, referenceID string) error
}

type paymentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPaymentClient(baseURL string) PaymentClient {
	return &paymentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *paymentClient) DeductTokens(ctx context.Context, userID string, amount int, referenceID string) error {
	reqBody := map[string]interface{}{
		"user_id":      userID,
		"amount":       amount,
		"reference_id": referenceID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/deduct", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error calling payment service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPaymentRequired {
		return domain.ErrInsufficientFunds
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}
	return nil
}
