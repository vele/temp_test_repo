package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vele/temp_test_repo/internal/domain"
	"github.com/vele/temp_test_repo/internal/event"
	postgresstorage "github.com/vele/temp_test_repo/internal/storage/postgres"
	"github.com/vele/temp_test_repo/internal/testutil"
)

func TestCreateUser_ValidatesAge(t *testing.T) {
	svc, _, _ := setupService(t)
	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Name:  "John",
		Email: "john@example.com",
		Age:   18,
	})
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestCreateUser_RequiresUniqueEmail(t *testing.T) {
	svc, repo, _ := setupService(t)
	ctx := context.Background()

	err := repo.Create(ctx, &domain.User{
		Name:  "Jane",
		Email: "jane@example.com",
		Age:   30,
	})
	require.NoError(t, err)

	_, err = svc.CreateUser(ctx, CreateUserInput{
		Name:  "John",
		Email: "jane@example.com",
		Age:   25,
	})
	require.ErrorIs(t, err, domain.ErrConflict)
}

func TestDeleteUser_PublishesEvent(t *testing.T) {
	svc, _, publisher := setupService(t)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, CreateUserInput{
		Name:  "Jane",
		Email: "jane@example.com",
		Age:   30,
	})
	require.NoError(t, err)

	err = svc.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	events := publisher.Events()
	require.Len(t, events, 2) // create + delete
	require.Equal(t, event.UserDeleted, events[1].Type)
	require.Equal(t, user.ID, events[1].UserID)
}

func setupService(t *testing.T) (*UserService, *postgresstorage.Repository, *event.InMemoryPublisher) {
	t.Helper()

	dsn := testutil.StartPostgres(t)
	repo := testutil.ConnectRepository(t, dsn)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Truncate(ctx)
		require.NoError(t, repo.Close())
	})

	publisher := event.NewInMemoryPublisher()
	svc := NewUserService(repo, repo, publisher)
	return svc, repo, publisher
}
