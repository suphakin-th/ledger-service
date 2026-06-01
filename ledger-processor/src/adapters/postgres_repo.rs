use anyhow::Result;
use async_trait::async_trait;
use sqlx::PgPool;
use uuid::Uuid;

use crate::ports::BalanceRepository;

pub struct PostgresBalanceRepo {
    pool: PgPool,
}

impl PostgresBalanceRepo {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[derive(sqlx::FromRow)]
struct BalanceRow {
    balance_cents: i64,
}

#[async_trait]
impl BalanceRepository for PostgresBalanceRepo {
    async fn apply_delta(&self, account_id: Uuid, delta_cents: i64) -> Result<i64> {
        let row = sqlx::query_as::<_, BalanceRow>(
            "UPDATE accounts
             SET balance_cents = balance_cents + $1,
                 updated_at    = NOW()
             WHERE id = $2
             RETURNING balance_cents",
        )
        .bind(delta_cents)
        .bind(account_id)
        .fetch_one(&self.pool)
        .await?;

        Ok(row.balance_cents)
    }

    async fn mark_completed(&self, transaction_id: Uuid) -> Result<()> {
        sqlx::query("UPDATE transactions SET status = 'completed' WHERE id = $1")
            .bind(transaction_id)
            .execute(&self.pool)
            .await?;
        Ok(())
    }
}
