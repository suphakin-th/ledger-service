package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/adapters/postgres"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/usecases"
)

func TestCreateAndGetAccount(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t, ctx)

	repo := postgres.NewAccountRepo(pool)
	uc := usecases.NewCreateAccount(repo)

	a, err := uc.Execute(ctx, usecases.CreateAccountCommand{
		Name:     "Test Account",
		Currency: "THB",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, a.ID)
	assert.Equal(t, "Test Account", a.Name)
	assert.Equal(t, "THB", a.Currency)
	assert.Equal(t, int64(0), a.BalanceCents)

	fetched, err := repo.GetByID(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, a.ID, fetched.ID)
	assert.Equal(t, a.Name, fetched.Name)
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t, ctx)
	repo := postgres.NewAccountRepo(pool)
	uc := usecases.NewCreateAccount(repo)

	for _, name := range []string{"Savings", "Checking", "Investment"} {
		_, err := uc.Execute(ctx, usecases.CreateAccountCommand{Name: name, Currency: "USD"})
		require.NoError(t, err)
	}

	accounts, err := repo.List(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(accounts), 3)
}
