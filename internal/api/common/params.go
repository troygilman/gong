package common

import (
	"context"
	"fmt"
	"stagebase/internal/core"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ParamRequestId = "request_id"
	ParamApp       = "app"
	ParamProjectId = "project_id"
)

func GetHelper[T any](ctx context.Context, key any) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

func MustGetHelper[T any](ctx context.Context, key any) (val T) {
	if val, ok := GetHelper[T](ctx, key); ok {
		return val
	} else {
		panic(fmt.Sprintf("value %s of type %T does not exist in context", key, val))
	}
}

func MustGetApp(ctx context.Context) core.App {
	return MustGetHelper[core.App](ctx, ParamApp)
}

func ExtractProjectId(e echo.Context) (any, any) {
	projectID, err := uuid.Parse(e.Param(ParamProjectId))
	if err != nil {
		return nil, nil
	}
	return ParamProjectId, projectID
}

func MustGetProjectId(ctx context.Context) uuid.UUID {
	return MustGetHelper[uuid.UUID](ctx, ParamProjectId)
}
