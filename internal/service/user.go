package service

import (
	"context"
	"fmt"

	"pay-as-you-use/internal/domain"
	"pay-as-you-use/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, email string) (string, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, email string) (string, error) {
	// TODO in the future: to create Stripe customer and save his ID as well
	id, err := s.userRepo.Create(ctx, email)
	if err != nil {
		return "", fmt.Errorf("service failed to create user: %w", err)
	}
	return id, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
