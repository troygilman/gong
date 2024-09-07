package gong

import (
	"context"

	"github.com/labstack/echo/v4"
)

type Group interface {
	Group(string, ...echo.MiddlewareFunc) *echo.Group
}

func NewGroup(g Group, path string, f func(e *echo.Group), middlewares ...echo.MiddlewareFunc) {
	f(g.Group(path, middlewares...))
}

type Extractor func(echo.Context) (any, any)

func WithContext(extractors ...Extractor) echo.MiddlewareFunc {
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
