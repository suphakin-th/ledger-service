use anyhow::{Context, Result};
use futures_util::StreamExt;
use redis::aio::PubSub;
use redis::AsyncCommands;
use std::sync::Arc;
use tracing::{error, info, warn};

use crate::domain::transaction::{BalanceUpdatedEvent, TransactionCreatedEvent, TransactionType};
use crate::ports::BalanceRepository;

pub struct RedisConsumer {
    pubsub: PubSub,
    publisher: redis::aio::MultiplexedConnection,
    repo: Arc<dyn BalanceRepository>,
}

impl RedisConsumer {
    pub async fn new(redis_url: &str, repo: Arc<dyn BalanceRepository>) -> Result<Self> {
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

        // Destructure to avoid simultaneous mutable borrows: `stream` holds a
        // reference to `pubsub`, so we must separate it from `publisher` and `repo`
        // before entering the loop.
        let Self {
            pubsub,
            mut publisher,
            repo,
        } = self;

        let mut stream = pubsub.on_message();

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

            if let Err(e) = process_event(&mut publisher, repo.as_ref(), &event).await {
                error!(tx_id = %event.transaction_id, "process failed: {e}");
            }
        }
        Ok(())
    }
}

async fn process_event(
    publisher: &mut redis::aio::MultiplexedConnection,
    repo: &dyn BalanceRepository,
    event: &TransactionCreatedEvent,
) -> Result<()> {
    let delta = match event.transaction_type {
        TransactionType::Credit => event.amount_cents,
        TransactionType::Debit => -event.amount_cents,
    };

    let new_balance = repo.apply_delta(event.account_id, delta).await?;
    repo.mark_completed(event.transaction_id).await?;

    let update = BalanceUpdatedEvent::new(event.account_id, new_balance);
    let json = serde_json::to_string(&update)?;
    publisher
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
