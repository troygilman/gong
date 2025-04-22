package gong

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
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

func buildHeaders(headers []string) string {
	builder := &strings.Builder{}
	builder.WriteString("{")
	i := 0
	for i+1 < len(headers) {
		builder.WriteString(`"`)
		builder.WriteString(headers[i])
		builder.WriteString(`": "`)
		builder.WriteString(headers[i+1])
		builder.WriteString(`"`)
		if i < len(headers)-2 {
			builder.WriteString(", ")
		}
		i = i + 2
	}
	builder.WriteString("}")
	return builder.String()
}

func render(ctx context.Context, gCtx gongContext, w io.Writer, component templ.Component) error {
	if component == nil {
		panic("cannot render nil templ.Component")
	}
	ctx = context.WithValue(ctx, contextKey, gCtx)
	return component.Render(ctx, w)
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
