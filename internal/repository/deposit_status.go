package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DepositStatusRepository interface {
	AddStatus(ctx context.Context, requestID, status string) error
}

type depositStatusRepo struct {
	db *pgxpool.Pool
}

func NewDepositStatusRepository(db *pgxpool.Pool) DepositStatusRepository {
	return &depositStatusRepo{db: db}
}

func (r *depositStatusRepo) AddStatus(ctx context.Context, requestID, status string) error {
	query := `
			INSERT INTO payment_schema.deposit_status_history (deposit_request_id, status)
			VALUES ($1, $2)
	`
	_, err := r.db.Exec(ctx, query, requestID, status)
	if err != nil {
		return fmt.Errorf("failed to insert deposit status: %w", err)
	}
	return nil
}
