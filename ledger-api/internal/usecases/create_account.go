package usecases

import (
	"context"

	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/account"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/ports"
)

type CreateAccountCommand struct {
	Name     string
	Currency string
}

type CreateAccountUseCase struct {
	accounts ports.AccountRepository
}

func NewCreateAccount(accounts ports.AccountRepository) *CreateAccountUseCase {
	return &CreateAccountUseCase{accounts: accounts}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, cmd CreateAccountCommand) (*account.Account, error) {
	a, err := account.New(cmd.Name, cmd.Currency)
	if err != nil {
		return nil, err
	}
	if err := uc.accounts.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
