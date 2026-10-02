package repository

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerRepository interface {
	GetBalance(ctx context.Context, userID string) (int, error)
	AddRecord(ctx context.Context, record domain.LedgerRecord) error
}

type ledgerRepo struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) LedgerRepository {
	return &ledgerRepo{db: db}
}

func (r *ledgerRepo) GetBalance(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_schema.ledger
		WHERE user_id = $1
	`
	var balance int
	err := r.db.QueryRow(ctx, query, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance for user %s: %w", userID, err)
	}
	return balance, nil
}

func (r *ledgerRepo) AddRecord(ctx context.Context, record domain.LedgerRecord) error {
	query := `
		INSERT INTO payment_schema.ledger
		(user_id, amount, reference_id, operation_type)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		record.UserID,
		record.Amount,
		record.ReferenceID,
		record.OperationType,
	)
	if err != nil {
		return fmt.Errorf("failed to add ledger record: %w", err)
	}
	return nil
}
