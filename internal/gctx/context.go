package gctx

import (
	"context"
	"net/http"

	"github.com/troygilman/gong"
	"github.com/troygilman/gong/internal/response_writer"
)

type contextKeyType int

const contextKey = contextKeyType(0)

type Context struct {
	Route       gong.Route
	ChildRoute  gong.Route
	Component   gong.Component
	Request     *http.Request
	Writer      *response_writer.ResponseWriter
	RouteID     string
	ComponentID string
	Path        string
	Action      bool
	Link        bool
}

func GetContext(ctx context.Context) Context {
	return ctx.Value(contextKey).(Context)
}

func SetContext(ctx context.Context, gCtx Context) context.Context {
	return context.WithValue(ctx, contextKey, gCtx)
}
