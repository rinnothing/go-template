package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
	"go.uber.org/zap"
)

// GetKeyval implements [gen.ServerInterface].
func (s *serverImplementation) GetKeyval(ctx context.Context, request gen.GetKeyvalRequestObject) (gen.GetKeyvalResponseObject, error) {
	ctx = s.addLogger(ctx, "GetKeyval", request)

	key := model.Key(request.Params.Key)
	err := validateKey(key)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.GetKeyval400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	keyval, err := s.keyval.Get(ctx, key)

	if errors.Is(err, model.ErrNotFound) {
		logNotFound(ctx, key)
		return gen.GetKeyval404JSONResponse{NotFoundJSONResponse: notFound(key)}, nil
	} else if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval found", zap.Any("keyval", keyval))
	return gen.GetKeyval200JSONResponse{
		Key:   string(keyval.Key),
		Value: string(keyval.Val),
	}, nil
}

// PostKeyval implements [gen.ServerInterface].
func (s *serverImplementation) CreateKeyval(ctx context.Context, request gen.CreateKeyvalRequestObject) (gen.CreateKeyvalResponseObject, error) {
	ctx = s.addLogger(ctx, "CreateKeyval", request)

	keyval := &model.KeyVal{
		Key: model.Key(request.Body.Key),
		Val: model.Val(request.Body.Value),
	}
	err := validateKeyval(keyval)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.CreateKeyval400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	err = s.keyval.Add(ctx, keyval)
	if errors.Is(err, model.ErrAlreadyExists) {
		logger.InfoCtx(ctx, "keyval with such key already exists", zap.Any("key", keyval.Key))
		return gen.CreateKeyval409JSONResponse{
			Code:    gen.KEYVALEXISTS,
			Message: fmt.Sprintf("keyval with key %s already exists", keyval.Key),
		}, nil
	} else if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval added")
	return gen.CreateKeyval201JSONResponse{
		Key:   string(keyval.Key),
		Value: string(keyval.Val),
	}, nil
}

// PutKeyval implements [gen.ServerInterface].
func (s *serverImplementation) UpdateKeyval(ctx context.Context, request gen.UpdateKeyvalRequestObject) (gen.UpdateKeyvalResponseObject, error) {
	ctx = s.addLogger(ctx, "UpdateKeyval", request)

	keyval := &model.KeyVal{
		Key: model.Key(request.Body.Key),
		Val: model.Val(request.Body.Value),
	}
	err := validateKeyval(keyval)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.UpdateKeyval400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	err = s.keyval.Update(ctx, keyval)
	if errors.Is(err, model.ErrNotFound) {
		logNotFound(ctx, keyval.Key)
		return gen.UpdateKeyval404JSONResponse{NotFoundJSONResponse: notFound(keyval.Key)}, nil
	} else if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval updated")
	return gen.UpdateKeyval200JSONResponse{
		Key:   string(keyval.Key),
		Value: string(keyval.Val),
	}, nil
}

// DeleteKeyval implements [gen.ServerInterface].
func (s *serverImplementation) DeleteKeyval(ctx context.Context, request gen.DeleteKeyvalRequestObject) (gen.DeleteKeyvalResponseObject, error) {
	ctx = s.addLogger(ctx, "DeleteKeyval", request)

	key := model.Key(request.Params.Key)
	err := validateKey(key)
	if err != nil {
		logBadRequest(ctx, err)
		return gen.DeleteKeyval400JSONResponse{BadRequestJSONResponse: badRequest(err)}, nil
	}

	err = s.keyval.Remove(ctx, key)
	if errors.Is(err, model.ErrNotFound) {
		logNotFound(ctx, key)
		return gen.DeleteKeyval404JSONResponse{NotFoundJSONResponse: notFound(key)}, nil
	} else if err != nil {
		logInternalError(ctx, err)
		return nil, err
	}

	logger.InfoCtx(ctx, "keyval deleted")
	return gen.DeleteKeyval200Response{}, nil
}
