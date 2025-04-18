package gong

import (
	"context"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
}

func buildComponentID(ctx context.Context, id string) string {
	gCtx := getContext(ctx)
	prefix := "gong_" + gCtx.route.ID()
	if gCtx.componentID != "" {
		prefix += "_" + gCtx.componentID
	}
	if id != "" {
		prefix += "_" + id
	}
	return prefix
}

func buildOutletID(ctx context.Context) string {
	gCtx := getContext(ctx)
	return "gong_" + gCtx.route.ID() + "_outlet"
}

func gongHeaders(ctx context.Context, requestType string) []string {
	gCtx := getContext(ctx)
	return []string{
		HeaderGongRequestType,
		requestType,
		HeaderGongRouteID,
		gCtx.route.ID(),
		HeaderGongComponentID,
		gCtx.componentID,
	}
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

func buildRealPath(route Route, request *http.Request) string {
	routePathSplit := strings.Split(route.FullPath(), "/")
	requestPathSplit := strings.Split(request.URL.EscapedPath(), "/")
	for i, routePathFragment := range routePathSplit {
		if i >= len(requestPathSplit) {
			continue
		}
		requestPathFragment := requestPathSplit[i]
		if routePathFragment == requestPathFragment {
			continue
		}
		if strings.HasPrefix(routePathFragment, "{") && strings.HasSuffix(routePathFragment, "}") {
			routePathSplit[i] = requestPathFragment
		}
	}
	return strings.Join(routePathSplit, "/")
}

func findComponent(id string) {

}
