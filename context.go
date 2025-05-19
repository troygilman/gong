package gong

import (
	"context"
	"net/http"

	"github.com/troygilman/gong/internal/response_writer"
)

type contextKeyType int

const contextKey = contextKeyType(0)

type gongContext struct {
	Node            *routeNode
	ChildRouteIndex int
	Component       Component
	Request         *http.Request
	Writer          *response_writer.ResponseWriter
	RouteID         string
	ComponentID     string
	Path            string
	Action          bool
	Link            bool
	ErrorHandler    ErrorHandler
	RenderedPath    string
}

func getContext(ctx context.Context) gongContext {
	return ctx.Value(contextKey).(gongContext)
}

func setContext(ctx context.Context, gCtx gongContext) context.Context {
	return context.WithValue(ctx, contextKey, gCtx)
}
