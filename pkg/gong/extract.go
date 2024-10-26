package gong

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type Extractor func(echo.Context, extractContext) (extractContext, any, any)

type ExtractFunc func(echo.Context) (any, any)

type ExtractorOption func(extractContext, any) extractContext

type extractContext struct {
	logHandler slog.Handler
}

func NewExtractor(f ExtractFunc, opts ...ExtractorOption) Extractor {
	return func(e echo.Context, ctx extractContext) (extractContext, any, any) {
		key, value := f(e)
		if value != nil {
			for _, opt := range opts {
				ctx = opt(ctx, key)
			}
		}
		return ctx, key, value
	}
}

func newExtractContext(ctx context.Context) extractContext {
	logHandler, ok := ctx.Value(LogHandlerKey).(slog.Handler)
	if !ok {
		logHandler = slog.Default().Handler()
	}
	return extractContext{
		logHandler: logHandler,
	}
}

func Extractors(extractors ...Extractor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			extractCtx := newExtractContext(ctx)
			for _, extractor := range extractors {
				var (
					key   any
					value any
				)
				extractCtx, key, value = extractor(e, extractCtx)
				if value != nil {
					ctx = context.WithValue(ctx, key, value)
				}
			}
			ctx = context.WithValue(ctx, LogHandlerKey, extractCtx.logHandler)
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
}

func WithLogging() ExtractorOption {
	return func(ctx extractContext, key any) extractContext {
		ctx.logHandler = contextLogHandler{
			Handler: ctx.logHandler,
			key:     key,
		}
		return ctx
	}
}
