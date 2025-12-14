package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TxFunc func(tx pgx.Tx) error

func (db *DB) WithTx(ctx context.Context, fn TxFunc) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
