package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/async"
	"github.com/rinnothing/avito-pr/internal/usecase/fast"
	"github.com/rinnothing/avito-pr/internal/usecase/keyval"
)

var _ async.Cache = &redisCache{}
var _ fast.Cache = &redisCache{}
var _ keyval.Cache = &redisCache{}

// Get implements [fast.Cache].
func (r *redisCache) Get(ctx context.Context, key model.Key) (*model.KeyVal, error) {
	res, err := r.c.Get(ctx, string(key)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %w", model.ErrNotFound, err)
	} else if err != nil {
		return nil, err
	}

	return &model.KeyVal{
		Key: key,
		Val: model.Val(res),
	}, nil
}

// Set implements [async.Cache].
func (r *redisCache) Set(ctx context.Context, keyval *model.KeyVal) error {
	err := r.c.Set(ctx, string(keyval.Key), string(keyval.Val), r.expirationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

// Remove implements [keyval.Cache].
func (r *redisCache) Remove(ctx context.Context, key model.Key) error {
	num, err := r.c.Del(ctx, string(key)).Result()
	if err != nil {
		return err
	}
	if num == 0 {
		return model.ErrNotFound
	}

	return nil
}
