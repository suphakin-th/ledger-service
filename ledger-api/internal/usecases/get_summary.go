package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/ports"
)

type AccountSummary struct {
	AccountID      uuid.UUID `json:"account_id"`
	Name           string    `json:"name"`
	Currency       string    `json:"currency"`
	BalanceCents   int64     `json:"balance_cents"`
	BalanceDisplay float64   `json:"balance"`
	TotalCreditCents int64   `json:"total_credit_cents"`
	TotalDebitCents  int64   `json:"total_debit_cents"`
	TransactionCount int     `json:"transaction_count"`
}

type GetSummaryUseCase struct {
	accounts     ports.AccountRepository
	transactions ports.TransactionRepository
}

func NewGetSummary(accounts ports.AccountRepository, transactions ports.TransactionRepository) *GetSummaryUseCase {
	return &GetSummaryUseCase{accounts: accounts, transactions: transactions}
}

func (uc *GetSummaryUseCase) Execute(ctx context.Context, accountID uuid.UUID) (*AccountSummary, error) {
	a, err := uc.accounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	creditCents, debitCents, err := uc.transactions.SumByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	txList, err := uc.transactions.ListByAccount(ctx, accountID, 0, 0)
	if err != nil {
		return nil, err
	}

	return &AccountSummary{
		AccountID:        a.ID,
		Name:             a.Name,
		Currency:         a.Currency,
		BalanceCents:     a.BalanceCents,
		BalanceDisplay:   a.FormattedBalance(),
		TotalCreditCents: creditCents,
		TotalDebitCents:  debitCents,
		TransactionCount: len(txList),
	}, nil
}
