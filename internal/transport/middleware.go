package transport

import (
	"context"
	"stagebase/internal/common"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func withRequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			cb := common.NewContextBuilder(e.Request().Context())
			cb = cb.With(common.CtxKeyRequestID, e.Response().Header().Get(echo.HeaderXRequestID))
			e.SetRequest(e.Request().WithContext(cb.Build()))
			return next(e)
		}
	}
}

func withHeaders() echo.MiddlewareFunc {
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

func withApp(app common.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			cb := common.NewContextBuilder(e.Request().Context())
			cb = cb.With(common.CtxKeyApp, app)
			e.SetRequest(e.Request().WithContext(cb.Build()))

			e.Set("app", app)
			return next(e)
		}
	}
}

func withJWT(app common.App) echo.MiddlewareFunc {
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
