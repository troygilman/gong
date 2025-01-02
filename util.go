package gong

import (
	"context"
	"strings"
)

func buildComponentID(ctx context.Context) string {
	return "gong" + strings.ReplaceAll(getContext(ctx).path, "/", "-")
}

func Method(ctx context.Context) string {
	return getContext(ctx).request.Method
}

func UseLoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	return gCtx.loader.Loader(ctx).(Data)
}
