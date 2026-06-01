package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount     = errors.New("amount must be greater than zero")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidCurrency   = errors.New("currency must be a 3-letter ISO 4217 code")
)

type Account struct {
	ID           uuid.UUID
	Name         string
	Currency     string
	BalanceCents int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(name, currency string) (*Account, error) {
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}
	now := time.Now().UTC()
	return &Account{
		ID:           uuid.New(),
		Name:         name,
		Currency:     currency,
		BalanceCents: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (a *Account) FormattedBalance() float64 {
	return float64(a.BalanceCents) / 100
}
