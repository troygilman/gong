package gong

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// Render renders a templ component with the provided Gong context.
// This is an internal utility used by components, routes, and other elements
// to consistently render with the Gong context. It handles error propagation
// and custom error handling if provided.
func render(ctx context.Context, gCtx gongContext, w io.Writer, component templ.Component) error {
	if component == nil {
		panic("cannot render nil templ.Component")
	}
	ctx = setContext(ctx, gCtx)
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
func gongHeaders(ctx context.Context, requestType string) []string {
	gCtx := getContext(ctx)
	return []string{
		HeaderGongRequestType,
		requestType,
		HeaderGongRouteID,
		gCtx.Node.id,
		HeaderGongComponentID,
		gCtx.ComponentID,
	}
}
