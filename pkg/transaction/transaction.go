package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

type Transactor interface {
	DoAtomically(context.Context, func(context.Context) error) error
	CreateLocalTransaction(ctx context.Context) (pgx.Tx, error)
	FinishLocalTransaction(ctx context.Context, tx pgx.Tx, txErr *error)
}

var _ Transactor = &impl{}

type impl struct {
	db *pgxpool.Pool
}

func NewTransactor(db *pgxpool.Pool) *impl {
	return &impl{db: db}
}

type globalTransaction struct {
	pgx.Tx
}

func (t *impl) DoAtomically(ctx context.Context, f func(context.Context) error) (txError error) {
	txCtx, tx, err := injectTx(ctx, t.db, true)
	if err != nil {
		return fmt.Errorf("cannot inject transaction: %w", err)
	}

	defer func() {
		if txError != nil {
			if err = tx.Rollback(txCtx); err != nil {
				logger.ErrorCtx(ctx, "cannot rollback transaction", zap.Error(err))
			}
			return
		}

		if insErr := tx.Commit(txCtx); insErr != nil {
			logger.ErrorCtx(ctx, "cannot commit transaction", zap.Error(insErr))
		}
	}()

	err = f(txCtx)
	if err != nil {
		return err
	}

	return nil
}

func (t *impl) CreateLocalTransaction(ctx context.Context) (pgx.Tx, error) {
	_, tx, err := injectTx(ctx, t.db, false)
	return tx, err
}

func (t *impl) FinishLocalTransaction(ctx context.Context, tx pgx.Tx, txErr *error) {
	if _, isGlobal := tx.(globalTransaction); isGlobal {
		return
	}

	if txErr != nil {
		_ = tx.Rollback(ctx)
		return
	}

	_ = tx.Commit(ctx)
}

type keyType struct{}

var ErrTxNotFound = errors.New("tx not found in context")

func ExtractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(keyType{}).(pgx.Tx)
	if !ok {
		return nil, ErrTxNotFound
	}

	return tx, nil
}

func injectTx(ctx context.Context, pool *pgxpool.Pool, global bool) (context.Context, pgx.Tx, error) {
	if tx, err := ExtractTx(ctx); err == nil {
		return ctx, tx, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	if global {
		tx = globalTransaction{tx}
	}

	return context.WithValue(ctx, keyType{}, tx), tx, nil
}
