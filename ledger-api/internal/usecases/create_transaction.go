package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/transaction"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/ports"
)

type CreateTransactionCommand struct {
	AccountID   uuid.UUID
	Type        transaction.Type
	AmountCents int64
	Currency    string
	Description string
}

type CreateTransactionUseCase struct {
	accounts     ports.AccountRepository
	transactions ports.TransactionRepository
	events       ports.EventBus
}

func NewCreateTransaction(
	accounts ports.AccountRepository,
	transactions ports.TransactionRepository,
	events ports.EventBus,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		accounts:     accounts,
		transactions: transactions,
		events:       events,
	}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, cmd CreateTransactionCommand) (*transaction.Transaction, error) {
	if _, err := uc.accounts.GetByID(ctx, cmd.AccountID); err != nil {
		return nil, err
	}

	tx, err := transaction.New(cmd.AccountID, cmd.Type, cmd.AmountCents, cmd.Currency, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := uc.transactions.Create(ctx, tx); err != nil {
		return nil, err
	}

	event := transaction.CreatedEvent{
		EventType:   transaction.EventTransactionCreated,
		TxID:        tx.ID,
		AccountID:   tx.AccountID,
		TxType:      tx.Type,
		AmountCents: tx.AmountCents,
		Currency:    tx.Currency,
		OccurredAt:  time.Now().UTC(),
	}
	_ = uc.events.Publish(ctx, "transactions.created", event)

	return tx, nil
}
