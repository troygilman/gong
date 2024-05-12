package view

import (
	"context"
	"fmt"
	"gong/types"
	"net/url"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func contextToValues(ctx context.Context) url.Values {
	values := url.Values{}
	if pid, ok := ctx.Value(types.CtxPID).(string); ok {
		values.Set("pid", pid)
	}
	if tab, ok := ctx.Value(types.CtxTab).(string); ok {
		values.Set("tab", tab)
	}
	if filter, ok := ctx.Value(types.CtxFilter).(types.EventsFilter); ok {
		filter.SetValues(values)
	}
	return values
}

func formatProjectUrlWithParams(ctx context.Context) string {
	values := contextToValues(ctx)
	url := fmt.Sprintf("/project/%s?%s", ctx.Value(types.CtxProject).(uuid.UUID).String(), values.Encode())
	return string(templ.URL(url))
}

func formatProjectUrl(ctx context.Context, path string) string {
	url := fmt.Sprintf("/project/%s%s", ctx.Value(types.CtxProject).(uuid.UUID).String(), path)
	return string(templ.URL(url))
}

const includes = "#actor-pid, #tab, #events-filter"
