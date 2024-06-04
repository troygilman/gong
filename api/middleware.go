package api

import (
	"errors"
	"gong/common"
	"gong/core"
	"log"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func middlewareLogRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			log.Printf("%s: %s %s", ctx.Request().Method, ctx.Request().RequestURI, requestID(ctx))
			return next(ctx)
		}
	}
}

func middlewareWithApp(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("app", app)
			return next(ctx)
		}
	}
}

func middlewareJWT(app core.App) echo.MiddlewareFunc {
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

func middlewareProjectContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := common.NewContextBuilder(ctx.Request().Context())

			{ // project ID
				projectID, err := uuid.Parse(ctx.Param("projectID"))
				if err != nil {
					return errors.New("projectID is empty")
				}
				cb = cb.With(common.CtxProject, projectID)
			}

			{ // actor PID
				actorPID := ctx.QueryParam(string(common.CtxPID))
				if actorPID != "" {
					cb = cb.With(common.CtxPID, actorPID)
				}
			}

			{ // tab
				tab := ctx.QueryParam(string(common.CtxTab))
				if tab == "" {
					tab = "Events"
				}
				cb = cb.With(common.CtxTab, tab)
			}

			{ // flow chart type
				flowType := ctx.QueryParam(string(common.CtxFlow))
				if flowType == "" {
					flowType = "Sources"
				}
				cb = cb.With(common.CtxFlow, flowType)
			}

			{ // events filter
				filter := common.NewFilter(ctx)
				cb = cb.With(common.CtxFilter, filter)
			}
			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}
