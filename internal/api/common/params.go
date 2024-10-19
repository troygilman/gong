package common

import (
	"context"
	"fmt"
	"log/slog"
	"stagebase/internal/core"
	"stagebase/internal/core/model"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ParamRequestId    = "request_id"
	ParamApp          = "app"
	ParamProjectId    = "project_id"
	ParamFilter       = "filter"
	ParamFilterPid    = "filter_pid"
	ParamFilterType   = "filter_type"
	ParamFilterSender = "filter_sender"
	ParamFilterLevel  = "filter_level"
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

func ExtractFilter(ctx echo.Context) (any, any) {
	filter := model.Filter{}

	filterPid := ctx.QueryParam(ParamFilterPid)
	if filterPid != "" {
		filter.PID = &filterPid
	}

	filterType := ctx.QueryParam(ParamFilterType)
	if filterType != "" {
		filter.Type = &filterType
	}

	filterSender := ctx.QueryParam(ParamFilterSender)
	if filterSender != "" {
		filter.Sender = &filterSender
	}

	if level, err := strconv.Atoi(ctx.QueryParam(ParamFilterLevel)); err != nil {
		filter.Level = slog.LevelDebug
	} else {
		filter.Level = slog.Level(level)
	}

	return ParamFilter, filter
}

func MustGetFilter(ctx context.Context) model.Filter {
	return MustGetHelper[model.Filter](ctx, ParamFilter)
}
