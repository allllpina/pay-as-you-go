package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/repository"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/checkout/session"
)

type PaymentService interface {
	GetUserBalance(ctx context.Context, userID string) (int, error)
	CreateCheckoutSession(ctx context.Context, userID string, packageID string) (string, error)
	ProcessSuccessfulPayment(ctx context.Context, requestID string) error
	ProcessFailedPayment(ctx context.Context, requestID string) error

	DeductTokens(ctx context.Context, userID string, amount int, referenceID string) error
}
type StripeConfig struct {
	SecretKey   string
	BaseURL     string
	FrontendURL string
}

type paymentService struct {
	ledgerRepo       repository.LedgerRepository
	tokenPackageRepo repository.TokenPackageRepository
	depositReqRepo   repository.DepositRequestRepository
	depositStatRepo  repository.DepositStatusRepository
	config           StripeConfig
}

func NewPaymentService(
	ledgerRepo repository.LedgerRepository,
	tokenPackageRepo repository.TokenPackageRepository,
	depositReqRepo repository.DepositRequestRepository,
	depositStatRepo repository.DepositStatusRepository,
	cfg StripeConfig,
) PaymentService {
	stripe.Key = cfg.SecretKey
	return &paymentService{
		ledgerRepo:       ledgerRepo,
		depositReqRepo:   depositReqRepo,
		depositStatRepo:  depositStatRepo,
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

	reqID, err := s.depositReqRepo.Create(ctx, domain.DepositRequest{
		UserID:    userID,
		PackageID: packageID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to add pending status: %w", err)
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
	params.AddMetadata("request_id", reqID)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe session: %w", err)
	}

	err = s.depositReqRepo.UpdateSessionID(ctx, reqID, sess.ID)
	if err != nil {
		log.Printf("Warning: failed to update session ID for request %s: %v", reqID, err)
	}

	return sess.URL, nil
}

func (s *paymentService) ProcessSuccessfulPayment(ctx context.Context, requestID string) error {
	req, err := s.depositReqRepo.GetByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("business error: deposit request not found %s", requestID)
	}

	pkg, err := s.tokenPackageRepo.GetByID(ctx, req.PackageID)
	if err != nil {
		return fmt.Errorf("business error: package not found %s", req.PackageID)
	}

	err = s.depositStatRepo.AddStatus(ctx, requestID, string(domain.StatusSuccess))
	if err != nil {
		return fmt.Errorf("failed to update status to success: %w", err)
	}

	record := domain.LedgerRecord{
		UserID:        req.UserID,
		Amount:        pkg.TokensAmount,
		ReferenceID:   &requestID,
		OperationType: domain.OpTypeStripeDeposit,
		CreatedAt:     time.Now(),
	}

	err = s.ledgerRepo.AddRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to add deposit record to ledger: %w", err)
	}

	return nil
}

func (s *paymentService) ProcessFailedPayment(ctx context.Context, requestID string) error {
	err := s.depositStatRepo.AddStatus(ctx, requestID, string(domain.StatusFailed))
	if err != nil {
		return fmt.Errorf("failed to update status to failed: %w", err)
	}
	return nil
}

func (s *paymentService) DeductTokens(ctx context.Context, userID string, amount int, referenceID string) error {
	balance, err := s.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if balance < amount {
		return domain.ErrInsufficientFunds
	}

	record := domain.LedgerRecord{
		UserID:        userID,
		Amount:        -amount,
		ReferenceID:   &referenceID,
		OperationType: domain.OpTypeServiceUsage,
		CreatedAt:     time.Now(),
	}

	err = s.ledgerRepo.AddRecord(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to add deduction record to ledger: %w", err)
	}

	return nil
}
