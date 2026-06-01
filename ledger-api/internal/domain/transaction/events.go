package transaction

import (
	"time"

	"github.com/google/uuid"
)

const EventTransactionCreated = "transaction.created"

type CreatedEvent struct {
	EventType   string    `json:"event_type"`
	TxID        uuid.UUID `json:"transaction_id"`
	AccountID   uuid.UUID `json:"account_id"`
	TxType      Type      `json:"transaction_type"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func NewCreatedEvent(tx *Transaction) CreatedEvent {
	return CreatedEvent{
		EventType:   EventTransactionCreated,
		TxID:        tx.ID,
		AccountID:   tx.AccountID,
		TxType:      tx.Type,
		AmountCents: tx.AmountCents,
		Currency:    tx.Currency,
		OccurredAt:  tx.CreatedAt,
	}
}
