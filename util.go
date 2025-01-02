package gong

import (
	"context"
	"strings"
)

func buildComponentID(ctx context.Context) string {
	gCtx := getContext(ctx)
	id := "gong" + strings.ReplaceAll(gCtx.path, "/", "-")
	if gCtx.kind != "" {
		id += "_" + gCtx.kind
	}
	if gCtx.id != "" {
		id += "_" + gCtx.id
	}
	return id
}

func Method(ctx context.Context) string {
	return getContext(ctx).request.Method
}

func UseLoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	return gCtx.loader.Loader(ctx).(Data)
}
