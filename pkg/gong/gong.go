package gong

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type Group interface {
	Group(string, ...echo.MiddlewareFunc) *echo.Group
}

func NewGroup(g Group, path string, f func(e *echo.Group), middlewares ...echo.MiddlewareFunc) {
	f(g.Group(path, middlewares...))
}

var LogHandlerKey = "log_handler"

type ctxKey interface {
	String() string
}

func Logger(ctx context.Context) *slog.Logger {
	handler, ok := ctx.Value(LogHandlerKey).(slog.Handler)
	if !ok {
		return slog.Default()
	}
	return slog.New(handler)
}

type contextLogHandler struct {
	slog.Handler
	key any
}

func (h contextLogHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := []any{}
	if val := ctx.Value(h.key); val != nil {
		attrs = append(attrs, h.key, val)
	}
	r.Add(attrs...)
	return h.Handler.Handle(ctx, r)
}
