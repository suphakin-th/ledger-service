package account

import (
	"time"

	"github.com/google/uuid"
)

const EventBalanceUpdated = "account.balance_updated"

type BalanceUpdatedEvent struct {
	EventType      string    `json:"event_type"`
	AccountID      uuid.UUID `json:"account_id"`
	NewBalanceCents int64    `json:"new_balance_cents"`
	OccurredAt     time.Time `json:"occurred_at"`
}
