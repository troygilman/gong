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
			cb := types.NewContextBuilder(ctx.Request().Context())

			projectID, err := uuid.Parse(ctx.Param("projectID"))
			if err != nil {
				return errors.New("projectID is empty")
			}
			cb = cb.With(types.CtxProject, projectID)

			actorPID := ctx.QueryParam("pid")
			if actorPID != "" {
				cb = cb.With(types.CtxPID, actorPID)
			}

			tab := ctx.QueryParam("tab")
			if tab == "" {
				tab = "Events"
			}
			cb = cb.With(types.CtxTab, tab)

			filter := types.EventsFilter{}
			if filterType := ctx.QueryParam("filter_type"); filterType != "" {
				filter.Type = &filterType
			}
			if filterSender := ctx.QueryParam("filter_sender"); filterSender != "" {
				filter.Sender = &filterSender
			}
			cb = cb.With(types.CtxFilter, filter)

			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}
