package service

import (
	"context"
	"fmt"
	"time"

	"pay-as-you-use/internal/client"
	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/repository"
)

type BusinessService interface {
	ExecuteAction(ctx context.Context, userID string, actionType string, cost int) error
}

type businessService struct {
	businessRepo  repository.BusinessRepository
	paymentClient client.PaymentClient
}

func NewBusinessService(repo repository.BusinessRepository, pClient client.PaymentClient) BusinessService {
	return &businessService{
		businessRepo:  repo,
		paymentClient: pClient,
	}
}

func (s *businessService) ExecuteAction(ctx context.Context, userID string, actionType string, cost int) error {
	logID, err := s.businessRepo.CreatePendingLog(ctx, userID, cost, actionType)
	if err != nil {
		return fmt.Errorf("failed to initiate action: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = s.paymentClient.DeductTokens(timeoutCtx, userID, cost, logID)

	if err != nil {
		_ = s.businessRepo.UpdateLogStatus(context.Background(), logID, domain.UsageStatusFailed)
		return fmt.Errorf("action blocked: %w", err)
	}

	err = s.businessRepo.UpdateLogStatus(ctx, logID, domain.UsageStatusSuccess)
	if err != nil {
		return fmt.Errorf("tokens deducted but failed to update status to success (log %s): %w", logID, err)
	}

	return nil
}
