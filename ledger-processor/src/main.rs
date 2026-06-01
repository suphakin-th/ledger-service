mod adapters;
mod domain;
mod ports;

use std::sync::Arc;

use anyhow::Result;
use sqlx::postgres::PgPoolOptions;
use tracing::info;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

use adapters::{postgres_repo::PostgresBalanceRepo, redis_consumer::RedisConsumer};

#[tokio::main]
async fn main() -> Result<()> {
    dotenvy::dotenv().ok();

    tracing_subscriber::registry()
        .with(EnvFilter::from_default_env().add_directive("info".parse()?))
        .with(tracing_subscriber::fmt::layer().json())
        .init();

    let db_url = std::env::var("DATABASE_URL").expect("DATABASE_URL required");
    let redis_url = std::env::var("REDIS_URL").expect("REDIS_URL required");

    let pool = PgPoolOptions::new()
        .max_connections(10)
        .connect(&db_url)
        .await?;

    info!("ledger-processor connected to postgres");

    let repo: Arc<dyn ports::BalanceRepository> = Arc::new(PostgresBalanceRepo::new(pool));
    let consumer = RedisConsumer::new(&redis_url, repo).await?;

    info!("ledger-processor starting event loop");
    consumer.run().await
}
