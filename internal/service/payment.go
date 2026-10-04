package service

import (
	"context"
	"fmt"
	"log"
	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/repository"
	"time"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/checkout/session"
)

type PaymentService interface {
	GetUserBalance(ctx context.Context, userID string) (int, error)
	CreateCheckoutSession(ctx context.Context, userID string, packageID string) (string, error)
	ProcessSuccessfulPayment(ctx context.Context, userID string, packageID string) error
}

type StripeConfig struct {
	SecretKey   string
	BaseURL     string
	FrontendURL string
}

type paymentService struct {
	ledgerRepo       repository.LedgerRepository
	tokenPackageRepo repository.TokenPackageRepository
	config           StripeConfig
}

func NewPaymentService(ledgerRepo repository.LedgerRepository, tokenPackageRepo repository.TokenPackageRepository, cfg StripeConfig) PaymentService {
	stripe.Key = cfg.SecretKey
	return &paymentService{
		ledgerRepo:       ledgerRepo,
		tokenPackageRepo: tokenPackageRepo,
		config:           cfg,
	}
}

func (s *paymentService) GetUserBalance(ctx context.Context, userID string) (int, error) {
	balance, err := s.ledgerRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("service layer error getting balance: %w", err)
	}
	return balance, nil
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, userID string, packageID string) (string, error) {
	successURL := fmt.Sprintf("%s/success?session_id={CHECKOUT_SESSION_ID}", s.config.FrontendURL)
	cancelURL := fmt.Sprintf("%s/cancel", s.config.FrontendURL)

	pkg, err := s.tokenPackageRepo.GetByID(ctx, packageID)
	if err != nil {
		return "", fmt.Errorf("token package not found: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(pkg.Currency),
					UnitAmount: stripe.Int64(int64(pkg.PriceCents)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("%d Tokens Package", pkg.TokensAmount)),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(userID),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
	}
	params.AddMetadata("package_id", pkg.ID)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe session: %w", err)
	}

	return sess.URL, nil
}

func (s *paymentService) ProcessSuccessfulPayment(ctx context.Context, userID string, packageID string) error {
	pkg, err := s.tokenPackageRepo.GetByID(ctx, packageID)
	if err != nil {
		log.Printf("CRITICAL: Payment succeeded but token package not found! UserID: %s, packageID: %s. Error: %v", userID, packageID, err)
		return fmt.Errorf("business error: token package not found for package %s", packageID)
	}

	record := domain.LedgerRecord{
		UserID:        userID,
		Amount:        pkg.TokensAmount,
		OperationType: domain.OpTypeStripeDeposit,
		CreatedAt:     time.Now(),
	}

	if err := s.ledgerRepo.AddRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to process deposit: %w", err)
	}
	return nil
}
