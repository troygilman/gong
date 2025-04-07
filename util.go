package gong

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"strconv"

	"github.com/a-h/templ"
)

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
}

func buildComponentID(ctx context.Context, id string) string {
	gCtx := getContext(ctx)
	prefix := "gong" + "_" + hash(gCtx.route.Path())
	if gCtx.id != "" {
		prefix += "_" + gCtx.id
	}
	if id != "" {
		prefix += "_" + id
	}
	return prefix
}

func buildOutletID(ctx context.Context) string {
	gCtx := getContext(ctx)
	return "gong" + "_" + hash(gCtx.route.Path()) + "_outlet"
}

func buildHeaders(ctx context.Context, requestType string) string {
	gCtx := getContext(ctx)
	return fmt.Sprintf(`{"%s": "%s", "%s": "%s", "%s": "%s"}`,
		HeaderGongRequestType,
		requestType,
		HeaderGongRoutePath,
		gCtx.route.Path(),
		HeaderGongComponentID,
		gCtx.id,
	)
}

func render(ctx context.Context, gCtx gongContext, w io.Writer, component templ.Component) error {
	if component == nil {
		panic("cannot render nil templ.Component")
	}
	ctx = context.WithValue(ctx, contextKey, gCtx)
	return component.Render(ctx, w)
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.Itoa(int(h.Sum32()))
}
