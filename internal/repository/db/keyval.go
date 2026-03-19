package db

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/async"
	"github.com/rinnothing/avito-pr/internal/usecase/fast"
	"github.com/rinnothing/avito-pr/internal/usecase/keyval"
)

var _ keyval.Repository = &postgresRepository{}
var _ fast.Repository = &postgresRepository{}
var _ async.Repository = &postgresRepository{}

// Add implements [keyval.Repository].
func (p *postgresRepository) Add(ctx context.Context, keyval *model.KeyVal) (txErr error) {
	tx, err := p.t.CreateLocalTransaction(ctx)
	if err != nil {
		return err
	}
	defer p.t.FinishLocalTransaction(ctx, tx, &txErr)

	const queryAdd = `
INSERT INTO keyvals (key, val)
VALUES ($1, $2)
`

	_, err = tx.Exec(ctx, queryAdd, keyval.Key, keyval.Val)
	if unErr, ok := convertUniqueViolation(err); ok {
		return unErr
	} else if err != nil {
		return err
	}

	return nil
}

// Get implements [keyval.Repository].
func (p *postgresRepository) Get(ctx context.Context, key model.Key) (keyvalRet *model.KeyVal, txErr error) {
	tx, err := p.t.CreateLocalTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer p.t.FinishLocalTransaction(ctx, tx, &txErr)

	const queryGet = `
SELECT key, val FROM keyvals
WHERE key = $1
`

	var keyval model.KeyVal
	err = tx.QueryRow(ctx, queryGet, key).Scan(&keyval.Key, &keyval.Val)
	if nfErr, ok := convertNotFound(err); ok {
		return nil, nfErr
	} else if err != nil {
		return nil, err
	}

	return &keyval, nil
}

// Remove implements [keyval.Repository].
func (p *postgresRepository) Remove(ctx context.Context, key model.Key) (txErr error) {
	tx, err := p.t.CreateLocalTransaction(ctx)
	if err != nil {
		return err
	}
	defer p.t.FinishLocalTransaction(ctx, tx, &txErr)

	const queryRemove = `
DELETE FROM keyvals
WHERE key = $1
RETURNING key
`
	var keyRemoved string
	err = tx.QueryRow(ctx, queryRemove, key).Scan(&keyRemoved)
	if nfErr, ok := convertNotFound(err); ok {
		return nfErr
	} else if err != nil {
		return err
	}

	return nil
}

// Update implements [keyval.Repository].
func (p *postgresRepository) Update(ctx context.Context, keyval *model.KeyVal) (txErr error) {
	tx, err := p.t.CreateLocalTransaction(ctx)
	if err != nil {
		return err
	}
	defer p.t.FinishLocalTransaction(ctx, tx, &txErr)

	const queryUpdate = `
UPDATE keyvals
SET val = $2
WHERE key = $1
RETURNING key
`
	var keyUpdated string
	err = tx.QueryRow(ctx, queryUpdate, keyval.Key, keyval.Val).Scan(&keyUpdated)
	if nfErr, ok := convertNotFound(err); ok {
		return nfErr
	} else if err != nil {
		return err
	}

	return nil
}
