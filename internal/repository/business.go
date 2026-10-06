package repository

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessRepository interface {
	CreatePendingLog(ctx context.Context, userID string, cost int, actionType string) (string, error)
	UpdateLogStatus(ctx context.Context, logID string, status domain.UsageStatus) error
}

type businessRepo struct {
	db *pgxpool.Pool
}

func NewBusinessRepository(db *pgxpool.Pool) BusinessRepository {
	return &businessRepo{db: db}
}

func (r *businessRepo) CreatePendingLog(ctx context.Context, userID string, cost int, actionType string) (string, error) {
	query := `
			INSERT INTO business_schema.usage_logs (user_id, tokens_cost, action_type, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, userID, cost, actionType, domain.UsageStatusPending).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create pending log: %w", err)
	}
	return id, nil
}

func (r *businessRepo) UpdateLogStatus(ctx context.Context, logID string, status domain.UsageStatus) error {
	query := `
			UPDATE business_schema.usage_logs
			SET status = $1
			WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, logID)
	if err != nil {
		return fmt.Errorf("failed to update log status: %w", err)
	}
	return nil
}
