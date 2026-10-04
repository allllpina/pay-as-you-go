package repository

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenPackageRepository interface {
	GetByStripePriceID(ctx context.Context, priceID string) (*domain.TokenPackage, error)
	GetByID(ctx context.Context, id string) (*domain.TokenPackage, error)
}

type tokenPackageRepo struct {
	db *pgxpool.Pool
}

func NewTokenPackageRepository(db *pgxpool.Pool) TokenPackageRepository {
	return &tokenPackageRepo{db: db}
}

func (r *tokenPackageRepo) GetByStripePriceID(ctx context.Context, priceID string) (*domain.TokenPackage, error) {
	query := `
        SELECT id, stripe_price_id, tokens_amount, price_cents, currency, is_active
        FROM payment_schema.token_packages
        WHERE stripe_price_id = $1 AND is_active = true
    `

	var pkg domain.TokenPackage
	err := r.db.QueryRow(ctx, query, priceID).Scan(
		&pkg.ID,
		&pkg.StripePriceID,
		&pkg.TokensAmount,
		&pkg.PriceCents,
		&pkg.Currency,
		&pkg.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get token package: %w", err)
	}

	return &pkg, nil
}

func (r *tokenPackageRepo) GetByID(ctx context.Context, id string) (*domain.TokenPackage, error) {
	query := `
        SELECT id, stripe_price_id, tokens_amount, price_cents, currency, is_active
        FROM payment_schema.token_packages
        WHERE id = $1 AND is_active = true
    `
	var pkg domain.TokenPackage
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pkg.ID,
		&pkg.StripePriceID,
		&pkg.TokensAmount,
		&pkg.PriceCents,
		&pkg.Currency,
		&pkg.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token package by id: %w", err)
	}

	return &pkg, nil
}
