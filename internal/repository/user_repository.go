package repository

import (
	"context"

	"github.com/vele/temp_test_repo/internal/domain"
)

type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error
}

type FileRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]domain.File, error)
	Add(ctx context.Context, file *domain.File) error
	DeleteByUser(ctx context.Context, userID uint) error
}
