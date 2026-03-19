package async

import (
	"context"
	"errors"
	"runtime"
	"sync"

	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"github.com/rinnothing/avito-pr/pkg/transaction"

	"go.uber.org/zap"
)

type Usecase interface {
	Put(ctx context.Context, keyval *model.KeyVal) error
}

type Cache interface {
	Set(ctx context.Context, keyval *model.KeyVal) error
}

type Repository interface {
	Get(ctx context.Context, key model.Key) (*model.KeyVal, error)
	Add(ctx context.Context, keyval *model.KeyVal) error
	Update(ctx context.Context, keyval *model.KeyVal) error
}

type Queue interface {
	PutTask(ctx context.Context, keyval *model.KeyVal) error
	ExtractTask(ctx context.Context) (*model.KeyVal, error)
}

var _ Usecase = &impl{}

type impl struct {
	repo  Repository
	cache Cache
	tr    transaction.Transactor
	queue Queue

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func New(ctx context.Context, repo Repository, cache Cache, tr transaction.Transactor, queue Queue, workersNum int) *impl {
	if workersNum < 1 {
		workersNum = runtime.GOMAXPROCS(0)
	}

	ctx, cancel := context.WithCancel(ctx)

	uc := &impl{
		repo:   repo,
		cache:  cache,
		tr:     tr,
		queue:  queue,
		cancel: cancel,
	}

	for range workersNum {
		uc.wg.Go(func() {
			for {
				task, err := queue.ExtractTask(ctx)
				if errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					logger.ErrorCtx(ctx, "error getting value from queue")
					continue
				}

				err = uc.putSync(ctx, task)
				if errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					logger.ErrorCtx(ctx, "error putting value synchronously")
					continue
				}
			}
		})
	}

	return uc
}

func (i *impl) Close(ctx context.Context) {
	i.cancel()
	i.wg.Wait()
}

// Put implements [Usecase].
func (i *impl) Put(ctx context.Context, keyval *model.KeyVal) error {
	logger.DebugCtx(ctx, "putting keyval aynchroniously", zap.Any("keyval", keyval))

	err := i.queue.PutTask(ctx, keyval)
	if err != nil {
		logger.DebugCtx(ctx, "error putting task in message broker", zap.Error(err))
		return err
	}

	return nil
}

// putSync implements [Usecase].
func (i *impl) putSync(ctx context.Context, keyval *model.KeyVal) error {
	logger.DebugCtx(ctx, "putting keyval synchronously", zap.Any("keyval", keyval))

	return i.tr.DoAtomically(ctx, func(ctx context.Context) error {
		_, err := i.repo.Get(ctx, keyval.Key)
		if errors.Is(err, model.ErrNotFound) {
			err := i.repo.Add(ctx, keyval)
			if err != nil {
				logger.DebugCtx(ctx, "error adding keyval to database", zap.Error(err))
				return err
			}

			return nil
		} else if err != nil {
			logger.DebugCtx(ctx, "error getting keyval from database", zap.Error(err))
			return err
		}

		err = i.repo.Update(ctx, keyval)
		if err != nil {
			logger.DebugCtx(ctx, "error updating keyval in database", zap.Error(err))
			return err
		}

		err = i.cache.Set(ctx, keyval)
		if err != nil {
			logger.DebugCtx(ctx, "error putting keyval into cache", zap.Error(err))
		}

		return nil
	})
}
