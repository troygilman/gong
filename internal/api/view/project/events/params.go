package events

import (
	"context"
	"stagebase/internal/api/common"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	ParamPage  = "page"
	EventsView = "events"
)

func ExtractPage(e echo.Context) (any, any) {
	page, err := strconv.Atoi(e.QueryParam(ParamPage))
	if err != nil {
		return ParamPage, 0
	}
	return ParamPage, page
}

func MustGetPage(ctx context.Context) int {
	return common.MustGetHelper[int](ctx, ParamPage)
}
