package util

import (
	"context"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/troygilman/gong"
	"github.com/troygilman/gong/internal/gctx"
)

// Render renders a templ component with the provided Gong context.
// This is an internal utility used by components, routes, and other elements
// to consistently render with the Gong context. It handles error propagation
// and custom error handling if provided.
func Render(ctx context.Context, gCtx gctx.Context, w io.Writer, component templ.Component) error {
	if component == nil {
		panic("cannot render nil templ.Component")
	}
	ctx = gctx.SetContext(ctx, gCtx)
	err := component.Render(ctx, w)
	if err != nil && gCtx.ErrorHandler != nil {
		gCtx.ErrorHandler(ctx, err)
		return nil
	}
	return err
}

// GongHeaders generates the standard set of Gong HTTP headers for a request.
// These headers are used to identify the request type, route ID, and component ID,
// which allows the server to properly handle the request and route it to the
// correct component.
func GongHeaders(ctx context.Context, requestType string) []string {
	gCtx := gctx.GetContext(ctx)
	return []string{
		gong.HeaderGongRequestType,
		requestType,
		gong.HeaderGongRouteID,
		gCtx.CurrentRouteID,
		gong.HeaderGongComponentID,
		gCtx.ComponentID,
	}
}

// BuildHeaders converts a flat array of key-value pairs into a JSON object string.
// The input array should have keys at even indices and values at odd indices.
// This is used to create the header JSON string for HTMX requests in Gong components.
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
