package db

import (
	"context"
	"database/sql"
)

type contextKey string

const txKey contextKey = "db_tx"

func txFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sql.Tx)
	return tx, ok
}

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, txKey, tx)
	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
