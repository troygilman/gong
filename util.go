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

func FormValue(ctx context.Context, key string) string {
	return Request(ctx).FormValue(key)
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
		GongRequestHeader,
		requestType,
		GongRouteHeader,
		gCtx.route.Path(),
		GongIdHeader,
		gCtx.id,
	)
}

func Request(ctx context.Context) *http.Request {
	return getContext(ctx).request
}

func LoaderData[Data any](ctx context.Context) (data Data) {
	gCtx := getContext(ctx)
	if gCtx.loader == nil {
		return data
	}
	return gCtx.loader.Loader(ctx).(Data)
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

type RenderFunc func(ctx context.Context, w io.Writer) error

func (r RenderFunc) Render(ctx context.Context, w io.Writer) error {
	return r(ctx, w)
}

func Redirect(ctx context.Context, path string) error {
	gCtx := getContext(ctx)
	gCtx.writer.Reset()
	http.Redirect(gCtx.writer, gCtx.request, path, http.StatusSeeOther)
	return nil
}
