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
	Route           gong.Route
	ChildRouteIndex int
	Component       gong.Component
	Request         *http.Request
	Writer          *response_writer.ResponseWriter
	RequestRouteID  string
	CurrentRouteID  string
	ComponentID     string
	Path            string
	Depth           int
	Action          bool
	Link            bool
	ErrorHandler    gong.ErrorHandler
}

func GetContext(ctx context.Context) Context {
	return ctx.Value(contextKey).(Context)
}

func SetContext(ctx context.Context, gCtx Context) context.Context {
	return context.WithValue(ctx, contextKey, gCtx)
}
