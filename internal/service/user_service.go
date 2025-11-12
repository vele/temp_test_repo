package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vele/temp_test_repo/internal/domain"
	"github.com/vele/temp_test_repo/internal/event"
	"github.com/vele/temp_test_repo/internal/repository"
)

type UserService struct {
	users     repository.UserRepository
	files     repository.FileRepository
	publisher event.Publisher
}

func NewUserService(users repository.UserRepository, files repository.FileRepository, publisher event.Publisher) *UserService {
	return &UserService{
		users:     users,
		files:     files,
		publisher: publisher,
	}
}

type CreateUserInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required"`
}

type UpdateUserInput struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Age   *int    `json:"age"`
}

type FileInput struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id uint) (domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.User{}, err
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return *user, nil
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (domain.User, error) {
	if err := validateUserInput(input.Name, input.Email, input.Age); err != nil {
		return domain.User{}, err
	}

	existing, err := s.users.GetByEmail(ctx, strings.ToLower(input.Email))
	if err != nil {
		return domain.User{}, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return domain.User{}, domain.ErrConflict
	}

	user := domain.User{
		Name:  strings.TrimSpace(input.Name),
		Email: strings.ToLower(strings.TrimSpace(input.Email)),
		Age:   input.Age,
	}
	if err := s.users.Create(ctx, &user); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	evt := event.Event{
		Type:       event.UserCreated,
		UserID:     user.ID,
		Payload:    user,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.publisher.Publish(ctx, evt); err != nil {
		return domain.User{}, fmt.Errorf("publish user created: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, input UpdateUserInput) (domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	if input.Name != nil {
		user.Name = strings.TrimSpace(*input.Name)
	}
	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if err := validateEmail(email); err != nil {
			return domain.User{}, err
		}
		existing, err := s.users.GetByEmail(ctx, email)
		if err != nil {
			return domain.User{}, fmt.Errorf("check email: %w", err)
		}
		if existing != nil && existing.ID != user.ID {
			return domain.User{}, domain.ErrConflict
		}
		user.Email = email
	}
	if input.Age != nil {
		if *input.Age <= 18 {
			return domain.User{}, domain.ErrInvalidInput
		}
		user.Age = *input.Age
	}

	if err := s.users.Update(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	evt := event.Event{
		Type:       event.UserUpdated,
		UserID:     user.ID,
		Payload:    user,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.publisher.Publish(ctx, evt); err != nil {
		return domain.User{}, fmt.Errorf("publish user updated: %w", err)
	}
	return *user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	if err := s.users.Delete(ctx, id); err != nil {
		return err
	}
	evt := event.Event{
		Type:       event.UserDeleted,
		UserID:     id,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("publish user deleted: %w", err)
	}
	return nil
}

func (s *UserService) ListFiles(ctx context.Context, userID uint) ([]domain.File, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.files.ListByUser(ctx, userID)
}

func (s *UserService) AddFile(ctx context.Context, userID uint, input FileInput) (domain.File, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return domain.File{}, err
	}

	file := domain.File{
		UserID: userID,
		Name:   strings.TrimSpace(input.Name),
		Path:   strings.TrimSpace(input.Path),
	}
	if file.Name == "" || file.Path == "" {
		return domain.File{}, domain.ErrInvalidInput
	}

	if err := s.files.Add(ctx, &file); err != nil {
		return domain.File{}, fmt.Errorf("add file: %w", err)
	}
	return file, nil
}

func (s *UserService) DeleteFiles(ctx context.Context, userID uint) error {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return err
	}
	return s.files.DeleteByUser(ctx, userID)
}

func validateUserInput(name, email string, age int) error {
	if strings.TrimSpace(name) == "" {
		return domain.ErrInvalidInput
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if age <= 18 {
		return domain.ErrInvalidInput
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return domain.ErrInvalidInput
	}
	return nil
}
