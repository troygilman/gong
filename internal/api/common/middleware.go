package common

import (
	"context"
	"stagebase/internal/core"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func WithRequestIdMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			ctx = context.WithValue(ctx, ParamRequestId, e.Response().Header().Get(echo.HeaderXRequestID))
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
}

func WithHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			e.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			e.Response().Header().Set("Pragma", "no-cache")
			e.Response().Header().Set("Expires", "0")
			return next(e)
		}
	}
}

func logRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			logger := gong.Logger(ctx)
			logger.InfoContext(ctx, "REQUEST", "method", e.Request().Method, "uri", e.Request().RequestURI)
			return next(e)
		}
	}
}

func WithApp(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			ctx = context.WithValue(ctx, ParamApp, app)
			e.SetRequest(e.Request().WithContext(ctx))
			return next(e)
		}
	}
}

func WithJWT(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			tokenCookie, err := e.Cookie("auth-token")
			if err != nil {
				return err
			}
			claims, err := app.JWT(tokenCookie.Value)
			if err != nil {
				return err
			}
			e.Set("claims", claims)
			return next(e)
		}
	}
}
