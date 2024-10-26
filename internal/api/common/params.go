package common

import (
	"log/slog"
	"stagebase/internal/model"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ParamRequestId      = "request_id"
	ParamApp            = "app"
	ParamProjectId      = "project_id"
	ParamFilter         = "filter"
	ParamFilterPid      = "filter_pid"
	ParamFilterType     = "filter_type"
	ParamFilterSender   = "filter_sender"
	ParamFilterReceiver = "filter_receiver"
	ParamFilterLevel    = "filter_level"
)

func ExtractProjectId(e echo.Context) (any, any) {
	projectID, err := uuid.Parse(e.Param(ParamProjectId))
	if err != nil {
		return nil, nil
	}
	return ParamProjectId, projectID
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

	filterReceiver := ctx.QueryParam(ParamFilterReceiver)
	if filterReceiver != "" {
		filter.Receiver = &filterReceiver
	}

	if level, err := strconv.Atoi(ctx.QueryParam(ParamFilterLevel)); err != nil {
		filter.Level = slog.LevelDebug
	} else {
		filter.Level = slog.Level(level)
	}

	return ParamFilter, filter
}
