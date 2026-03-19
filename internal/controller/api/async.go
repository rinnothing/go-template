package api

import (
	"context"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

// PutAsync implements [gen.ServerInterface].
func (s *serverImplementation) PutKeyvalAsync(ctx context.Context, request gen.PutKeyvalAsyncRequestObject) (gen.PutKeyvalAsyncResponseObject, error) {
	ctx = s.addLogger(ctx, "PutKeyvalAsync", request)

	keyval := &model.KeyVal{
		Key: model.Key(request.Body.Key),
		Val: model.Val(request.Body.Value),
	}
	err := validateKeyval(keyval)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.PutKeyvalAsync400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	err = s.async.Put(ctx, keyval)
	if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval put in async queue")
	return gen.PutKeyvalAsync200Response{}, nil
}
