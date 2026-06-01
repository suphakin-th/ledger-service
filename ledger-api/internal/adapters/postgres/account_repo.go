package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/account"
)

type AccountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{pool: pool}
}

func (r *AccountRepo) Create(ctx context.Context, a *account.Account) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO accounts (id, name, currency, balance_cents, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		a.ID, a.Name, a.Currency, a.BalanceCents, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *AccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, currency, balance_cents, created_at, updated_at
		 FROM accounts WHERE id = $1`, id,
	)
	a := &account.Account{}
	if err := row.Scan(&a.ID, &a.Name, &a.Currency, &a.BalanceCents, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return a, nil
}

func (r *AccountRepo) List(ctx context.Context) ([]*account.Account, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, currency, balance_cents, created_at, updated_at FROM accounts ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return func() ([]*account.Account, error) {
		var accounts []*account.Account
		for rows.Next() {
			a := &account.Account{}
			if err := rows.Scan(&a.ID, &a.Name, &a.Currency, &a.BalanceCents, &a.CreatedAt, &a.UpdatedAt); err != nil {
				return nil, err
			}
			accounts = append(accounts, a)
		}
		return accounts, rows.Err()
	}()
}
