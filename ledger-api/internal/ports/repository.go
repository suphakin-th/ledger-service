package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/account"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/transaction"
)

type AccountRepository interface {
	Create(ctx context.Context, a *account.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error)
	List(ctx context.Context) ([]*account.Account, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, t *transaction.Transaction) error
	ListByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error)
	SumByAccount(ctx context.Context, accountID uuid.UUID) (creditCents, debitCents int64, err error)
}
