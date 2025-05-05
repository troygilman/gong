package util

import (
	"context"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/internal/gctx"
)

func Render(ctx context.Context, gCtx gctx.Context, w io.Writer, component templ.Component) error {
	if component == nil {
		panic("cannot render nil templ.Component")
	}
	ctx = gctx.SetContext(ctx, gCtx)
	return component.Render(ctx, w)
}

func GongHeaders(ctx context.Context, requestType string) []string {
	gCtx := gctx.GetContext(ctx)
	return []string{
		gong.HeaderGongRequestType,
		requestType,
		gong.HeaderGongRouteID,
		gCtx.Route.ID(),
		gong.HeaderGongComponentID,
		gCtx.ComponentID,
	}
}

func BuildHeaders(headers []string) string {
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
