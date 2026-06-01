package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/adapters/postgres"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/transaction"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/ports"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/usecases"
)

type noopBus struct{}

func (n *noopBus) Publish(_ context.Context, _ string, _ any) error { return nil }

var _ ports.EventBus = (*noopBus)(nil)

func TestCreateTransaction(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t, ctx)

	accountRepo := postgres.NewAccountRepo(pool)
	txRepo := postgres.NewTransactionRepo(pool)
	createAccount := usecases.NewCreateAccount(accountRepo)
	createTx := usecases.NewCreateTransaction(accountRepo, txRepo, &noopBus{})

	account, err := createAccount.Execute(ctx, usecases.CreateAccountCommand{Name: "wallet", Currency: "THB"})
	require.NoError(t, err)

	tx, err := createTx.Execute(ctx, usecases.CreateTransactionCommand{
		AccountID:   account.ID,
		Type:        transaction.Credit,
		AmountCents: 10000,
		Currency:    "THB",
		Description: "initial deposit",
	})
	require.NoError(t, err)
	assert.Equal(t, transaction.StatusPending, tx.Status)
	assert.Equal(t, int64(10000), tx.AmountCents)

	txs, err := txRepo.ListByAccount(ctx, account.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, tx.ID, txs[0].ID)
}

func TestTransactionSumByAccount(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t, ctx)

	accountRepo := postgres.NewAccountRepo(pool)
	txRepo := postgres.NewTransactionRepo(pool)
	createAccount := usecases.NewCreateAccount(accountRepo)
	createTx := usecases.NewCreateTransaction(accountRepo, txRepo, &noopBus{})

	account, err := createAccount.Execute(ctx, usecases.CreateAccountCommand{Name: "sum-test", Currency: "USD"})
	require.NoError(t, err)

	// SumByAccount only counts 'completed' — pending transactions should not appear
	_, err = createTx.Execute(ctx, usecases.CreateTransactionCommand{
		AccountID: account.ID, Type: transaction.Credit, AmountCents: 5000, Currency: "USD",
	})
	require.NoError(t, err)

	creditCents, debitCents, err := txRepo.SumByAccount(ctx, account.ID)
	require.NoError(t, err)
	// pending transactions excluded from sum
	assert.Equal(t, int64(0), creditCents)
	assert.Equal(t, int64(0), debitCents)
}
