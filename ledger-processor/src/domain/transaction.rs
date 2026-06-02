use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum TransactionType {
    Credit,
    Debit,
}

// Serde ignores unknown fields by default, so removing currency and
// occurred_at does not break deserialization of existing Redis messages.
#[derive(Debug, Clone, Deserialize)]
pub struct TransactionCreatedEvent {
    pub transaction_id: Uuid,
    pub account_id: Uuid,
    pub transaction_type: TransactionType,
    pub amount_cents: i64,
}

#[derive(Debug, Clone, Serialize)]
pub struct BalanceUpdatedEvent {
    pub event_type: &'static str,
    pub account_id: Uuid,
    pub new_balance_cents: i64,
    pub occurred_at: DateTime<Utc>,
}

impl BalanceUpdatedEvent {
    pub fn new(account_id: Uuid, new_balance_cents: i64) -> Self {
        Self {
            event_type: "account.balance_updated",
            account_id,
            new_balance_cents,
            occurred_at: Utc::now(),
        }
    }
}
