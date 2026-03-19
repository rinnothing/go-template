package controller

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/config"
	"github.com/rinnothing/avito-pr/db"
	"github.com/rinnothing/avito-pr/internal/adapter/queue"
	"github.com/rinnothing/avito-pr/internal/controller/api"
	"github.com/rinnothing/avito-pr/internal/repository/cache"
	dbRepo "github.com/rinnothing/avito-pr/internal/repository/db"
	"github.com/rinnothing/avito-pr/internal/usecase/async"
	"github.com/rinnothing/avito-pr/internal/usecase/fast"
	"github.com/rinnothing/avito-pr/internal/usecase/keyval"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

type Server struct {
	cancel context.CancelFunc
}

func (s *Server) Run(lg *zap.Logger, cfg *config.Config) {
	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctx := logger.NewContext(sigCtx, lg)

	dbPool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		lg.Error("can not create pgxpool", zap.Error(err))
		return
	}
	defer dbPool.Close()

	db.SetupPostgres(dbPool, lg)

	transactor := transaction.NewTransactor(dbPool)
	repo := dbRepo.NewPostgresRepository(dbPool, transactor)

	redisClient := redis.NewClient(&redis.Options{
		Addr:           cfg.Redis.Addr,
		DB:             cfg.Redis.DB,
		MaxActiveConns: cfg.Redis.MaxActiveConns,
	})
	defer redisClient.Close()

	cacheRepo := cache.New(redisClient, cfg.Redis.ExpirationTime)

	keyvalUsecase := keyval.New(repo, cacheRepo)
	fastUsecase := fast.New(cacheRepo, repo)

	kafkaReader := kafka.Reader{}
	defer kafkaReader.Close()

	kafkaWriter := kafka.Writer{}
	defer kafkaWriter.Close()

	queue := queue.New(&kafkaReader, &kafkaWriter)
	asyncUsecase := async.New(ctx, repo, cacheRepo, transactor, queue, cfg.Kafka.WorkersNum)

	srv := api.New(keyvalUsecase, fastUsecase, asyncUsecase, lg)

	e := echo.New()
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			lg.Error("failed with panic", zap.String("path", c.Path()), zap.Error(err), zap.ByteString("stack", stack))
			return err
		},
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			lg.Debug("incoming request", zap.String("uri", v.URI), zap.Any("values", v.FormValues))
			return nil
		},
	}))

	e.IPExtractor = echo.ExtractIPDirect()
	hndlr := gen.NewStrictHandler(srv, nil)
	gen.RegisterHandlers(e, hndlr)

	go func() {
		if err := e.Start(net.JoinHostPort("0.0.0.0", cfg.HTTP.Port)); !errors.Is(err, http.ErrServerClosed) {
			lg.Fatal("server died", zap.Error(err))
		}
	}()

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := e.Shutdown(stopCtx); err != nil {
		lg.Fatal("server shutdown failed", zap.Error(err))
		return
	}

	lg.Info("server shutdown")
}

func (s *Server) Stop() {
	s.cancel()
}
