package gong

import (
	"context"
	"log/slog"

	"github.com/a-h/templ"
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
				if value != nil {
					ctx = context.WithValue(ctx, key, value)
				}
			}
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
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

func ContextLogHandler(keys ...ctxKey) echo.MiddlewareFunc {
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
	keys []ctxKey
}

func (h contextLogHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := []any{}
	for _, key := range h.keys {
		if val := ctx.Value(key); val != nil {
			attrs = append(attrs, key.String(), val)
		}
	}
	r.Add(attrs...)
	return h.Handler.Handle(ctx, r)
}

func NewTemplHandler(component templ.Component) echo.HandlerFunc {
	return func(e echo.Context) error {
		return component.Render(e.Request().Context(), e.Response())
	}
}
