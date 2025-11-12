package testutil

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	postgresstorage "github.com/vele/temp_test_repo/internal/storage/postgres"
)

func StartPostgres(t *testing.T) string {
	t.Helper()

	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("users"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		if strings.Contains(err.Error(), "rootless Docker is not supported on Windows") {
			t.Skipf("postgres container unavailable: %v (set TEST_POSTGRES_DSN to use an existing instance)", err)
			return ""
		}
		require.NoError(t, err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	return dsn
}

func ConnectRepository(t *testing.T, dsn string) *postgresstorage.Repository {
	t.Helper()

	const attempts = 20
	var lastErr error
	for i := 0; i < attempts; i++ {
		repo, err := postgresstorage.NewRepository(dsn)
		if err == nil {
			return repo
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * 250 * time.Millisecond)
	}
	require.NoError(t, lastErr)
	return nil
}
