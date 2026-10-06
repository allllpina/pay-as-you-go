package repository

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, email string) (string, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, email string) (string, error) {
	query := `
		INSERT INTO user_schema.users (id, email, created_at)
		VALUES (gen_random_uuid(), $1, NOW())
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(ctx, query, email).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, stripe_customer_id, created_at
		FROM user_schema.users
		WHERE id = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.StripeCustomerID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
