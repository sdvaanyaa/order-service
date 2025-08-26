package pgdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Transactor interface {
	WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error
}

type Transaction struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewTransactor(client *Client) *Transaction {
	return &Transaction{
		pool: client.conn,
		log:  client.log,
	}
}

func (t *Transaction) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		t.log.Error("failed to begin transaction", slog.Any("error", err))
		return fmt.Errorf("begin transaction: %w", err)
	}

	t.log.Debug("transaction began")

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		t.log.Error("transaction failed", slog.Any("error", err))
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			t.log.Error("failed to rollback transaction", slog.Any("error", rollbackErr))
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		t.log.Error("failed to commit transaction", slog.Any("error", err))
		return fmt.Errorf("commit transaction: %w", err)
	}

	t.log.Debug("transaction committed")
	return nil
}
