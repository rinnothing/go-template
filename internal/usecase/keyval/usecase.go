package keyval

import (
	"context"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

type Usecase interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
	Add(ctx context.Context, keyval *model.KeyVal) error
	Update(ctx context.Context, keyval *model.KeyVal) error
	Remove(ctx context.Context, key model.Key) error
}

type Cache interface {
	Set(ctx context.Context, keyval *model.KeyVal) error
	Remove(ctx context.Context, key model.Key) error
}

type Repository interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
	Add(ctx context.Context, keyval *model.KeyVal) error
	Update(ctx context.Context, keyval *model.KeyVal) error
	Remove(ctx context.Context, key model.Key) error
}

var _ Usecase = &impl{}

type impl struct {
	repo  Repository
	cache Cache
}

func New(repo Repository, cache Cache) *impl {
	return &impl{
		repo:  repo,
		cache: cache,
	}
}

// Add implements [Usecase].
func (i *impl) Add(ctx context.Context, keyval *model.KeyVal) error {
	logger.DebugCtx(ctx, "adding keyval", zap.Any("keyval", keyval))

	err := i.repo.Add(ctx, keyval)
	if err != nil {
		logger.DebugCtx(ctx, "error adding keyval to database", zap.Error(err))
		return err
	}

	err = i.cache.Set(ctx, keyval)
	if err != nil {
		logger.DebugCtx(ctx, "error adding keyval to cache", zap.Error(err))
		// cache is not important to me, so I do nothing about errors
	}

	return nil
}

// Get implements [Usecase].
func (i *impl) Get(ctx context.Context, key model.Key) (*model.KeyVal, error) {
	logger.DebugCtx(ctx, "getting keyval", zap.Any("key", key))

	val, err := i.repo.Get(ctx, key)
	if err != nil {
		logger.DebugCtx(ctx, "error querying keyval from database", zap.Error(err))
		return nil, err
	}
	return val, nil
}

// Remove implements [Usecase].
func (i *impl) Remove(ctx context.Context, key model.Key) error {
	logger.DebugCtx(ctx, "removing keyval", zap.Any("key", key))

	err := i.repo.Remove(ctx, key)
	if err != nil {
		logger.DebugCtx(ctx, "error removing keyval from database", zap.Error(err))
		return err
	}

	err = i.cache.Remove(ctx, key)
	if err != nil {
		logger.DebugCtx(ctx, "error removing keyval from cache", zap.Error(err))
	}

	return nil
}

// Update implements [Usecase].
func (i *impl) Update(ctx context.Context, keyval *model.KeyVal) error {
	logger.DebugCtx(ctx, "updating keyval", zap.Any("keyval", keyval))

	err := i.repo.Update(ctx, keyval)
	if err != nil {
		logger.DebugCtx(ctx, "error updating keyval in database", zap.Error(err))
		return err
	}

	err = i.cache.Set(ctx, keyval)
	if err != nil {
		logger.DebugCtx(ctx, "error updating keyval in cache", zap.Error(err))
	}

	return nil
}
