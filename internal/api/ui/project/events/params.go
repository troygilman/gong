package events

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	ParamPage      = "page"
	FromFixedParam = "from_fixed"
	ToFixedParam   = "to_fixed"
	EventsView     = "events"
)

func ExtractPage(e echo.Context) (any, any) {
	page, err := strconv.Atoi(e.QueryParam(ParamPage))
	if err != nil {
		return ParamPage, 0
	}
	return ParamPage, page
}

func ExtractFromFixed(e echo.Context) (any, any) {
	fromNanos, err := strconv.Atoi(e.QueryParam(FromFixedParam))
	if err != nil {
		return nil, nil
	}
	return FromFixedParam, time.Unix(0, int64(fromNanos))
}

func ExtractToFixed(e echo.Context) (any, any) {
	toNanos, err := strconv.Atoi(e.QueryParam(ToFixedParam))
	if err != nil {
		return nil, nil
	}
	return ToFixedParam, time.Unix(0, int64(toNanos))
}
