package service

import (
	"context"
	"fmt"
	"pay-as-you-use/internal/repository"
)

type PaymentService interface {
	GetUserBalance(ctx context.Context, userID string) (int, error)
}

type paymentService struct {
	ledgerRepo repository.LedgerRepository
}

func NewPaymentService(repo repository.LedgerRepository) PaymentService {
	return &paymentService{
		ledgerRepo: repo,
	}
}

func (s *paymentService) GetUserBalance(ctx context.Context, userID string) (int, error) {
	balance, err := s.ledgerRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("service layer error getting balance: %w", err)
	}
	return balance, nil
}
