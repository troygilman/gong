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

type Extractor func(echo.Context) (any, any)

func ContextExtractors(extractors ...Extractor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			for _, extractor := range extractors {
				key, value := extractor(e)
				if key != nil {
					ctx = context.WithValue(ctx, key, value)
				}
			}
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
}

var LogHandlerKey = "log_handler"

func Logger(ctx context.Context) *slog.Logger {
	handler, ok := ctx.Value(LogHandlerKey).(slog.Handler)
	if !ok {
		return slog.Default()
	}
	return slog.New(handler)
}

func ContextLogHandler(keys ...any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			handler, ok := ctx.Value(LogHandlerKey).(slog.Handler)
			if !ok {
				handler = slog.Default().Handler()
			}
			handler = contextLogHandler{
				Handler: handler,
				keys:    keys,
			}
			ctx = context.WithValue(ctx, LogHandlerKey, handler)
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
}

type contextLogHandler struct {
	slog.Handler
	keys []any
}

func (h contextLogHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.observe(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h contextLogHandler) observe(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}
	for _, key := range h.keys {
		if key, ok := key.(string); ok {
			if val := ctx.Value(key); val != nil {
				attrs = append(attrs, slog.Any(key, val))
			}
		}
	}
	return attrs
}
