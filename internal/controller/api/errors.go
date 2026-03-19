package api

import (
	"context"
	"fmt"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

func newError(code gen.ErrorResponseCode, message string) *gen.ErrorResponse {
	return &gen.ErrorResponse{
		Code:    code,
		Message: message,
	}
}

func notFound(key model.Key) gen.NotFoundJSONResponse {
	return gen.NotFoundJSONResponse(gen.ErrorResponse{
		Code:    gen.NOTFOUND,
		Message: fmt.Sprintf("keyval with key %s not found", key),
	})
}

func badRequest(err error) gen.BadRequestJSONResponse {
	return gen.BadRequestJSONResponse(gen.ErrorResponse{
		Code:    gen.BADREQUEST,
		Message: fmt.Sprintf("bad request: %s", err),
	})
}

func logBadRequest(ctx context.Context, err error) {
	logger.InfoCtx(ctx, "Validation error", zap.Error(err))
}

func logNotFound(ctx context.Context, key model.Key) {
	logger.InfoCtx(ctx, "Keyval not found", zap.Any("key", key))
}

func logInternalError(ctx context.Context, err error) {
	logger.ErrorCtx(ctx, "Internal server error", zap.Error(err))
}
