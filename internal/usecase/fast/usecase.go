package fast

import (
	"context"
	"errors"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

type Usecase interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
}

type Repository interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
}

type Cache interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
	Set(ctx context.Context, keyval *model.KeyVal) error
}

var _ Usecase = &impl{}

type impl struct {
	repo  Repository
	cache Cache
}

func New(cache Cache, repo Repository) *impl {
	return &impl{
		cache: cache,
		repo:  repo,
	}
}

// Get implements [Usecase].
func (i *impl) Get(ctx context.Context, key model.Key) (*model.KeyVal, error) {
	logger.DebugCtx(ctx, "getting keyval fast", zap.Any("key", key))

	val, err := i.cache.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		logger.DebugCtx(ctx, "error getting keyval from cache", zap.Error(err))
		return nil, err
	}

	val, err = i.repo.Get(ctx, key)
	if err != nil {
		logger.DebugCtx(ctx, "error getting keyval from database", zap.Error(err))
		return nil, err
	}

	err = i.cache.Set(ctx, val)
	if err != nil {
		logger.DebugCtx(ctx, "error putting keyval into cache", zap.Error(err))
	}

	return val, nil
}
