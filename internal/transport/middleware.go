package transport

import (
	"context"
	"stagebase/internal/common"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func withRequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := common.NewContextBuilder(ctx.Request().Context())
			cb = cb.With(common.CtxKeyRequestID, ctx.Response().Header().Get(echo.HeaderXRequestID))
			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}

func withHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			ctx.Response().Header().Set("Pragma", "no-cache")
			ctx.Response().Header().Set("Expires", "0")
			return next(ctx)
		}
	}
}

func logRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			logger := gong.Logger(ctx.Request().Context())
			logger.InfoContext(ctx.Request().Context(), "REQUEST", "method", ctx.Request().Method, "uri", ctx.Request().RequestURI)
			return next(ctx)
		}
	}
}

func withApp(app common.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := common.NewContextBuilder(ctx.Request().Context())
			cb = cb.With(common.CtxKeyApp, app)
			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))

			ctx.Set("app", app)
			return next(ctx)
		}
	}
}

func withJWT(app common.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tokenCookie, err := ctx.Cookie("auth-token")
			if err != nil {
				return err
			}
			claims, err := app.JWT(tokenCookie.Value)
			if err != nil {
				return err
			}
			ctx.Set("claims", claims)
			return next(ctx)
		}
	}
}

func projectMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			if _, ok := actorPID(e); !ok {
				app := mustApp(e)
				projectID := mustProjectID(e)
				projectProps, err := app.ProjectProps(e.Request().Context(), projectID)
				if err != nil {
					return err
				}
				for _, actor := range projectProps.Actors {
					ctx := e.Request().Context()
					ctx = context.WithValue(ctx, common.CtxKeyPID, actor.PID)
					e.SetRequest(e.Request().WithContext(ctx))
					break
				}
			}
			return next(e)
		}
	}
}
