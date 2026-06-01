package transaction

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidType = errors.New("transaction type must be credit or debit")

type Type string

const (
	Credit Type = "credit"
	Debit  Type = "debit"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Transaction struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	Type        Type
	AmountCents int64
	Currency    string
	Description string
	Status      Status
	CreatedAt   time.Time
}

func New(accountID uuid.UUID, txType Type, amountCents int64, currency, description string) (*Transaction, error) {
	if txType != Credit && txType != Debit {
		return nil, ErrInvalidType
	}
	if amountCents <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	return &Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Type:        txType,
		AmountCents: amountCents,
		Currency:    currency,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (t *Transaction) FormattedAmount() float64 {
	return float64(t.AmountCents) / 100
}
