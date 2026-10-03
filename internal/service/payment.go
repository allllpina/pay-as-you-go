package service

import (
	"context"
	"fmt"
	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/repository"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/checkout/session"
)

type PaymentService interface {
	GetUserBalance(ctx context.Context, userID string) (int, error)
	CreateCheckoutSession(ctx context.Context, userID string, priceID string) (string, error)
	AddDeposit(ctx context.Context, userID string, amount int) error
}

type StripeConfig struct {
	SecretKey   string
	BaseURL     string
	FrontendURL string
}

type paymentService struct {
	ledgerRepo repository.LedgerRepository
	config     StripeConfig
}

func NewPaymentService(repo repository.LedgerRepository, cfg StripeConfig) PaymentService {
	stripe.Key = cfg.SecretKey
	return &paymentService{
		ledgerRepo: repo,
		config:     cfg,
	}
}

func (s *paymentService) GetUserBalance(ctx context.Context, userID string) (int, error) {
	balance, err := s.ledgerRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("service layer error getting balance: %w", err)
	}
	return balance, nil
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, userID string, priceID string) (string, error) {
	successURL := fmt.Sprintf("%s/success?session_id={CHECKOUT_SESSION_ID}", s.config.FrontendURL)
	cancelURL := fmt.Sprintf("%s/cancel", s.config.FrontendURL)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(userID),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe session: %w", err)
	}

	return sess.URL, nil
}

func (s *paymentService) AddDeposit(ctx context.Context, userID string, amount int) error {
	record := domain.LedgerRecord{
		UserID:        userID,
		Amount:        amount,
		OperationType: domain.OpTypeStripeDeposit,
	}

	if err := s.ledgerRepo.AddRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to process deposit: %w", err)
	}
	return nil
}
