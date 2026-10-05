package repository

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DepositRequestRepository interface {
	Create(ctx context.Context, req domain.DepositRequest) (string, error)
	UpdateSessionID(ctx context.Context, requestID, sessionID string) error
	GetByID(ctx context.Context, id string) (*domain.DepositRequest, error)
}

type depositRequestRepo struct {
	db *pgxpool.Pool
}

func NewDepositRequestRepository(db *pgxpool.Pool) DepositRequestRepository {
	return &depositRequestRepo{db: db}
}

func (r *depositRequestRepo) Create(ctx context.Context, req domain.DepositRequest) (string, error) {
	query := `
			INSERT INTO payment_schema.deposit_requests (user_id, package_id)
			VALUES ($1, $2)
			RETURNING id
		`
	var id string
	err := r.db.QueryRow(ctx, query, req.UserID, req.PackageID).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create deposit request: %w", err)
	}
	return id, nil
}

func (r *depositRequestRepo) UpdateSessionID(ctx context.Context, requestID, sessionID string) error {
	query := `
			UPDATE payment_schema.deposit_requests
			SET stripe_session_id = $1
			WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, sessionID, requestID)
	if err != nil {
		return fmt.Errorf("failed to update stripe session id: %w", err)
	}
	return nil
}

func (r *depositRequestRepo) GetByID(ctx context.Context, id string) (*domain.DepositRequest, error) {
	query := `
			SELECT id, user_id, stripe_session_id, package_id, created_at
			FROM payment_schema.deposit_requests
			WHERE id = $1
	`
	var req domain.DepositRequest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.UserID,
		&req.StripeSessionID,
		&req.PackageID,
		&req.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit request: %w", err)
	}
	return &req, nil
}
