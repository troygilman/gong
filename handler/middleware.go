package handler

import (
	"errors"
	"gong/domain"
	"gong/types"
	"log"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func MiddlewareLogRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			log.Printf("%s: %s %s", ctx.Request().Method, ctx.Request().RequestURI, requestID(ctx))
			return next(ctx)
		}
	}
}

func MiddlewareWithApp(app domain.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("app", app)
			return next(ctx)
		}
	}
}

func MiddlewareJWT(app domain.App) echo.MiddlewareFunc {
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

func MiddlewareProjectContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := buildContext(ctx)

			projectID, err := uuid.Parse(ctx.Param("projectID"))
			if err != nil {
				return errors.New("projectID is empty")
			}
			cb = cb.with(types.CtxProject, projectID)

			actorPID := ctx.QueryParam("pid")
			if actorPID != "" {
				cb = cb.with(types.CtxPID, actorPID)
			}

			tab := ctx.QueryParam("tab")
			if tab == "" {
				tab = "Events"
			}
			cb = cb.with(types.CtxTab, tab)

			ctx.SetRequest(ctx.Request().WithContext(cb.build()))
			return next(ctx)
		}
	}
}
