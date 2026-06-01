package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS accounts (
    id            UUID        PRIMARY KEY,
    name          TEXT        NOT NULL,
    currency      CHAR(3)     NOT NULL,
    balance_cents BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id            UUID        PRIMARY KEY,
    account_id    UUID        NOT NULL REFERENCES accounts(id),
    type          TEXT        NOT NULL CHECK (type IN ('credit', 'debit')),
    amount_cents  BIGINT      NOT NULL CHECK (amount_cents > 0),
    currency      CHAR(3)     NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    status        TEXT        NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
`
