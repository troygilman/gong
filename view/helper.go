package view

import (
	"context"
	"fmt"
	"gong/types"
	"net/url"

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
	if filter, ok := ctx.Value(types.CtxFilter).(EventsFilter); ok {
		filter.setValues(values)
	}
	return values
}

func formatUrl(ctx context.Context) string {
	values := contextToValues(ctx)
	return fmt.Sprintf("/project/%s?%s", ctx.Value(types.CtxProject).(uuid.UUID).String(), values.Encode())
}

func formatActorUrl(ctx context.Context) string {
	values := contextToValues(ctx)
	return fmt.Sprintf("/project/%s/actor?%s", ctx.Value(types.CtxProject).(uuid.UUID).String(), values.Encode())
}
