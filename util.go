package gong

import (
	"context"
	"fmt"
	"strings"
)

func buildComponentID(ctx context.Context, id string) string {
	gCtx := getContext(ctx)
	prefix := "gong" + strings.ReplaceAll(gCtx.path, "/", "-")
	if gCtx.kind != "" {
		prefix += "_" + gCtx.kind
	}
	if id != "" {
		prefix += "_" + id
	}
	return prefix
}

func buildHeaders(ctx context.Context) string {
	gCtx := getContext(ctx)
	return fmt.Sprintf(`{"%s": "true", "%s": "%s"}`, GongActionHeader, GongKindHeader, gCtx.kind)
}

func Method(ctx context.Context) string {
	return getContext(ctx).request.Method
}

func UseLoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	return gCtx.loader.Loader(ctx).(Data)
}
