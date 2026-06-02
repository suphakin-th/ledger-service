[![CI](https://github.com/suphakin-th/ledger-service/actions/workflows/ci.yml/badge.svg)](https://github.com/suphakin-th/ledger-service/actions/workflows/ci.yml)

# ledger-service

![Go](https://img.shields.io/badge/Go-Gin-00ADD8?logo=go&logoColor=white)
![Rust](https://img.shields.io/badge/Rust-Tokio-000000?logo=rust&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Pub%2FSub-DC382D?logo=redis&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-5C4EE5)

Event-driven financial ledger built as two independent microservices: a Go REST API and a Rust async processor communicating via Redis pub/sub. Each service owns its layer of the stack.

---

## Architecture

```
Client
  |
  | HTTP
  v
+------------------+          +---------------------+
|   ledger-api     |  Redis   |  ledger-processor   |
|   (Go / Gin)     | pub/sub  |  (Rust / Tokio)     |
|                  +--------> |                     |
|  Hexagonal arch  |          |  Async worker pool  |
|  DDD domain      |          |  Balance updater    |
|  REST endpoints  |          |  Transaction finalizer
+--------+---------+          +----------+----------+
         |                               |
         |  PostgreSQL (shared schema)   |
         +---------------+---------------+
                         |
                  postgres:5432
```

**Flow:**
1. Client calls `POST /api/v1/accounts/:id/transactions`
2. `ledger-api` validates, persists the transaction as `pending`, publishes `transaction.created` to Redis
3. `ledger-processor` consumes the event, applies the balance delta to PostgreSQL using `UPDATE ... RETURNING`, marks the transaction `completed`, publishes `account.balance_updated`
4. Balance is eventually consistent — the API returns immediately; the processor runs concurrently

---

## Services

### ledger-api (Go)

REST API implementing **Hexagonal Architecture** (Ports & Adapters):

```
ledger-api/
  cmd/api/main.go                   entry point, wiring
  internal/
    domain/
      account/entity.go             pure Account entity, no I/O
      account/events.go             BalanceUpdatedEvent
      transaction/entity.go         pure Transaction entity, domain validation
      transaction/events.go         TransactionCreatedEvent
    ports/
      repository.go                 AccountRepository, TransactionRepository interfaces
      event_bus.go                  EventBus interface
    usecases/
      create_account.go             CreateAccountUseCase
      create_transaction.go         CreateTransactionUseCase
      get_summary.go                GetSummaryUseCase
    adapters/
      postgres/                     pgx v5 concrete repository implementations
      redis/                        Redis pub/sub EventBus
      http/                         Gin router, handlers, Prometheus middleware
  tests/integration/                testcontainers-go: real PostgreSQL per test run
```

The domain layer has **zero framework imports** — pure Go structs and interfaces. Use cases depend only on port interfaces, never on adapters directly.

### ledger-processor (Rust)

Async event consumer using **Tokio** multi-threaded runtime:

```
ledger-processor/
  src/
    domain/transaction.rs    TransactionCreatedEvent, BalanceUpdatedEvent (Serde)
    ports/mod.rs             BalanceRepository trait (async_trait)
    adapters/
      postgres_repo.rs       sqlx PgPool, UPDATE ... RETURNING balance
      redis_consumer.rs      redis-rs async PubSub consumer loop
    main.rs                  runtime init, dependency wiring
```

Rust's ownership model makes the balance update provably thread-safe — the `Arc<dyn BalanceRepository>` is shared across tasks without a mutex because all mutation goes through atomic PostgreSQL `UPDATE` operations.

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Accounts

| Method | Path | Body | Response |
|---|---|---|---|
| `POST` | `/accounts` | `{"name":"Savings","currency":"THB"}` | `201` Account object |
| `GET` | `/accounts/:id/summary` | — | `200` Summary object |

**POST /accounts — sample:**
```json
// request
{ "name": "Savings", "currency": "THB" }

// response 201
{
  "id": "018e1234-...",
  "name": "Savings",
  "currency": "THB",
  "balance_cents": 0,
  "created_at": "2026-06-02T00:00:00Z"
}
```

### Transactions

| Method | Path | Body | Response |
|---|---|---|---|
| `POST` | `/accounts/:id/transactions` | see below | `201` Transaction object |

**POST /accounts/:id/transactions — sample:**
```json
// request
{
  "type": "credit",
  "amount_cents": 150000,
  "currency": "THB",
  "description": "salary June"
}

// response 201
{
  "id": "018e5678-...",
  "account_id": "018e1234-...",
  "type": "credit",
  "amount_cents": 150000,
  "currency": "THB",
  "description": "salary June",
  "status": "pending",
  "created_at": "2026-06-02T00:01:00Z"
}
```

### Summary

**GET /accounts/:id/summary — sample:**
```json
{
  "account_id": "018e1234-...",
  "name": "Savings",
  "currency": "THB",
  "balance_cents": 150000,
  "balance": 1500.00,
  "total_credit_cents": 150000,
  "total_debit_cents": 0,
  "transaction_count": 1
}
```

### Health & Metrics

| Path | Purpose |
|---|---|
| `GET /healthz` | Liveness probe — always `{"status":"ok"}` |
| `GET /metrics` | Prometheus metrics endpoint |

---

## Local Setup

```bash
# Start everything (postgres + redis + both services)
make dev

# Or manually
docker compose up --build
```

All services are healthy-checked. The API is ready when you see:
```
ledger-api  | {"level":"INFO","msg":"ledger-api starting","port":"8080"}
ledger-processor | {"level":"INFO","msg":"ledger-processor subscribed to transactions.created"}
```

---

## Testing

```bash
make test          # both services
make test-go       # Go integration tests (testcontainers spins real PostgreSQL)
make test-rust     # Rust unit tests
```

Go integration tests use [testcontainers-go](https://golang.testcontainers.org/) — each test suite spins a fresh `postgres:16-alpine` container, runs migrations, executes against real SQL, and tears down on completion. No mocks for the database layer.

```bash
make lint          # golangci-lint + cargo clippy
make build         # compile both services
```

---

## Concurrency & Performance Notes

### Go — worker pool pattern

The `CreateTransactionUseCase` publishes to Redis in a fire-and-forget call (`_ = uc.events.Publish(...)`). The HTTP handler returns immediately after persisting the transaction. The event is processed by the Rust side concurrently. Under high throughput this means the API is never blocked waiting for balance recalculation — latency is bounded by the DB write, not by the full balance update cycle.

Prometheus histograms on every endpoint (`http_request_duration_seconds`) make it trivial to observe p99 latency regressions during load testing.

### Rust — ownership guarantees thread safety

`RedisConsumer` holds `Arc<dyn BalanceRepository>`. Multiple Tokio tasks could hold a clone of this Arc simultaneously. The implementation is safe without a `Mutex` because balance mutation uses a single atomic `UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2 RETURNING balance_cents` — PostgreSQL row-level locking guarantees correctness even under concurrent updates to different accounts.

The `#[async_trait]` bound on `BalanceRepository` enforces that all implementors are `Send + Sync`, so the compiler rejects any implementation that would introduce data races at compile time rather than runtime.

---

## Design Decisions

| Decision | Rationale |
|---|---|
| Amounts stored as `BIGINT` cents | Floating-point arithmetic on money causes rounding errors. Integer cents are exact. |
| Transactions start as `pending` | The API and processor are decoupled. Status transitions to `completed` only after the balance update is committed. This gives an audit trail and allows replay on processor failure. |
| Separate services vs one monolith | Go owns the request path (low latency, ergonomic HTTP). Rust owns the event path (high throughput, provable thread safety). Each language does what it does best. |
| Hexagonal architecture in Go | Swapping the PostgreSQL adapter for an in-memory one (e.g., for testing) requires changing one line of wiring in `main.go`. Nothing in the domain or use-case layer changes. |
| `testcontainers-go` over mocks for DB tests | Mocked repositories hide SQL bugs. Real containers found two production bugs during development (see Known Issues below). |

---

## Known Issues Fixed

| Bug | Root cause | Fix |
|---|---|---|
| `SumByAccount` returned non-zero for pending transactions | Filter clause missing `AND status = 'completed'` | Added status filter in `transaction_repo.go` |
| Race condition on concurrent debit to same account | Application-level balance check before subtract | Removed application-level check; balance delta applied atomically in SQL via `UPDATE ... SET balance_cents = balance_cents + $1` |
