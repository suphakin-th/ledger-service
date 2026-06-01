pub mod repository;

use anyhow::Result;
use async_trait::async_trait;
use uuid::Uuid;

#[async_trait]
pub trait BalanceRepository: Send + Sync {
    async fn apply_delta(&self, account_id: Uuid, delta_cents: i64) -> Result<i64>;
    async fn mark_completed(&self, transaction_id: Uuid) -> Result<()>;
}
