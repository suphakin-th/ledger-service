use anyhow::{Context, Result};
use redis::aio::PubSub;
use redis::AsyncCommands;
use tracing::{error, info, warn};

use crate::domain::transaction::{BalanceUpdatedEvent, TransactionCreatedEvent, TransactionType};
use crate::ports::BalanceRepository;

pub struct RedisConsumer {
    pubsub: PubSub,
    publisher: redis::aio::MultiplexedConnection,
    repo: std::sync::Arc<dyn BalanceRepository>,
}

impl RedisConsumer {
    pub async fn new(redis_url: &str, repo: std::sync::Arc<dyn BalanceRepository>) -> Result<Self> {
        let client = redis::Client::open(redis_url)?;
        let pubsub = client.get_async_pubsub().await?;
        let publisher = client.get_multiplexed_async_connection().await?;
        Ok(Self {
            pubsub,
            publisher,
            repo,
        })
    }

    pub async fn run(mut self) -> Result<()> {
        self.pubsub
            .subscribe("transactions.created")
            .await
            .context("subscribe transactions.created")?;

        info!("ledger-processor subscribed to transactions.created");

        use futures_util::StreamExt;
        let mut stream = self.pubsub.on_message();

        while let Some(msg) = stream.next().await {
            let payload: String = match msg.get_payload() {
                Ok(p) => p,
                Err(e) => {
                    error!("read payload: {e}");
                    continue;
                }
            };

            let event: TransactionCreatedEvent = match serde_json::from_str(&payload) {
                Ok(e) => e,
                Err(e) => {
                    warn!("deserialize event: {e}");
                    continue;
                }
            };

            if let Err(e) = self.process(&event).await {
                error!(tx_id = %event.transaction_id, "process failed: {e}");
            }
        }
        Ok(())
    }

    async fn process(&mut self, event: &TransactionCreatedEvent) -> Result<()> {
        let delta = match event.transaction_type {
            TransactionType::Credit => event.amount_cents,
            TransactionType::Debit => -event.amount_cents,
        };

        let new_balance = self.repo.apply_delta(event.account_id, delta).await?;
        self.repo.mark_completed(event.transaction_id).await?;

        let update = BalanceUpdatedEvent::new(event.account_id, new_balance);
        let json = serde_json::to_string(&update)?;
        self.publisher
            .publish::<_, _, ()>("accounts.balance_updated", json)
            .await?;

        info!(
            tx_id = %event.transaction_id,
            account = %event.account_id,
            delta = delta,
            balance = new_balance,
            "balance updated"
        );
        Ok(())
    }
}
