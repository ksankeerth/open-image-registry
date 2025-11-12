package store

import (
	"context"
	"database/sql"
)

type TxBeginner interface {
	Begin(ctx context.Context) (*sql.Tx, error)
}

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type ctxKey int

const txKey ctxKey = 0

// WithTxContext stores transaction in context.
// Used by the Service layer to wrap store operations in a transaction.
func WithTxContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext retrieves transaction from context.
// Used by the concrete store implementations
// to check if an active transaction is available.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sql.Tx)
	return tx, ok
}
