package api

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/internal/usecase/async"
	"github.com/rinnothing/avito-pr/internal/usecase/fast"
	"github.com/rinnothing/avito-pr/internal/usecase/keyval"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

var _ gen.StrictServerInterface = &serverImplementation{}

type serverImplementation struct {
	l *zap.Logger

	keyval keyval.Usecase
	fast   fast.Usecase
	async  async.Usecase
}

func New(keyval keyval.Usecase, fast fast.Usecase, async async.Usecase, l *zap.Logger) *serverImplementation {
	return &serverImplementation{l: l, keyval: keyval, fast: fast, async: async}
}

func (s *serverImplementation) withLogger(ctx context.Context) context.Context {
	return logger.NewContext(ctx, s.l)
}

func (s *serverImplementation) addLogger(ctx context.Context, methodName string, request any) context.Context {
	id := uuid.New()
	ctx = context.WithValue(ctx, "uuid", id)
	ctx = logger.NewContext(ctx, s.l.With(zap.String("method", methodName), zap.String("uuid", id.String())))

	logger.InfoCtx(ctx, "Processing method request", zap.Any("request", request))
	return ctx
}

func validateKey(key model.Key) error {
	if key == "" {
		return fmt.Errorf("key can't be empty")
	}
	return nil
}

func validateVal(val model.Val) error {
	if val == "" {
		return fmt.Errorf("val can't be empty")
	}
	return nil
}

func validateKeyval(keyval *model.KeyVal) error {
	err := validateKey(keyval.Key)
	if err != nil {
		return err
	}
	return validateVal(keyval.Val)
}
