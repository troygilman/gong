package gong

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
)

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
}

func Bind(ctx context.Context, dest any) error {
	gCtx := getContext(ctx)
	if err := json.NewDecoder(gCtx.request.Body).Decode(dest); err != nil {
		return err
	}
	return nil
}

func GetParam(ctx context.Context, key string) string {
	gCtx := getContext(ctx)
	return gCtx.request.FormValue(key)
}

func buildComponentID(ctx context.Context, id string) string {
	gCtx := getContext(ctx)
	prefix := "gong" + "_" + hash(gCtx.path)
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

func GetRequest(ctx context.Context) *http.Request {
	return getContext(ctx).request
}

func GetLoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	if gCtx.loader == nil {
		return data
	}
	return gCtx.loader.Loader(ctx).(Data)
}

func render(ctx context.Context, gCtx gongContext, w io.Writer, component templ.Component) error {
	ctx = context.WithValue(ctx, contextKey, gCtx)
	return component.Render(ctx, w)
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.Itoa(int(h.Sum32()))
}

type RenderFunc func(ctx context.Context, w io.Writer) error

func (r RenderFunc) Render(ctx context.Context, w io.Writer) error {
	return r(ctx, w)
}
