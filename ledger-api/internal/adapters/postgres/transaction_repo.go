package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/transaction"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

func (r *TransactionRepo) Create(ctx context.Context, t *transaction.Transaction) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO transactions (id, account_id, type, amount_cents, currency, description, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.ID, t.AccountID, t.Type, t.AmountCents, t.Currency, t.Description, t.Status, t.CreatedAt,
	)
	return err
}

func (r *TransactionRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	query := `SELECT id, account_id, type, amount_cents, currency, description, status, created_at
	          FROM transactions WHERE account_id = $1 ORDER BY created_at DESC`
	args := []any{accountID}

	if limit > 0 {
		query += ` LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*transaction.Transaction
	for rows.Next() {
		t := &transaction.Transaction{}
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Type, &t.AmountCents, &t.Currency, &t.Description, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, rows.Err()
}

func (r *TransactionRepo) SumByAccount(ctx context.Context, accountID uuid.UUID) (creditCents, debitCents int64, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount_cents) FILTER (WHERE type = 'credit'), 0),
			COALESCE(SUM(amount_cents) FILTER (WHERE type = 'debit'),  0)
		FROM transactions
		WHERE account_id = $1 AND status = 'completed'
	`, accountID)
	err = row.Scan(&creditCents, &debitCents)
	return
}
