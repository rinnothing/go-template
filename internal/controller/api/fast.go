package api

import (
	"context"
	"errors"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

// GetFast implements [gen.ServerInterface].
func (s *serverImplementation) GetKeyvalFast(ctx context.Context, request gen.GetKeyvalFastRequestObject) (gen.GetKeyvalFastResponseObject, error) {
	ctx = s.addLogger(ctx, "GetKeyvalFast", request)

	key := model.Key(request.Params.Key)
	err := validateKey(key)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.GetKeyvalFast400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	keyval, err := s.fast.Get(ctx, key)
	if errors.Is(err, model.ErrNotFound) {
		logNotFound(ctx, key)
		return gen.GetKeyvalFast404JSONResponse{NotFoundJSONResponse: notFound(key)}, nil
	} else if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval found in cache", zap.Any("keyval", keyval))
	return gen.GetKeyvalFast200JSONResponse{
		Key:   string(keyval.Key),
		Value: string(keyval.Val),
	}, nil
}
