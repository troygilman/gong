package transport

import (
	"errors"
	"gong/internal/common"
	"strconv"

	"log"
	"time"

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

func middlewareWithApp(app common.App) echo.MiddlewareFunc {
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

func middlewareJWT(app common.App) echo.MiddlewareFunc {
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
				projectID, err := uuid.Parse(ctx.Param(string(common.CtxKeyProjectID)))
				if err != nil {
					return errors.New("projectID is empty")
				}
				cb = cb.With(common.CtxKeyProjectID, projectID)
			}

			{ // actor PID
				actorPID := ctx.QueryParam(common.ParamPID)
				if actorPID != "" {
					cb = cb.With(common.CtxKeyPID, actorPID)
				}
			}

			{ // tab
				tab := ctx.QueryParam(common.ParamTab)
				if tab == "" {
					tab = "Events"
				}
				cb = cb.With(common.CtxKeyTab, tab)
			}

			{ // flow chart type
				flowType := ctx.QueryParam(common.ParamFlow)
				if flowType == "" {
					flowType = "Sources"
				}
				cb = cb.With(common.CtxKeyFlow, flowType)
			}

			{ // flow chart node
				flowNode := ctx.QueryParam(common.ParamFlowNode)
				if flowNode != "" {
					cb = cb.With(common.CtxKeyFlowNode, flowNode)
				}
			}

			{ // flow chart range
				flowRange := ctx.QueryParam(common.ParamFlowRange)
				var r common.Range
				if err := r.Parse(flowRange); err == nil {
					cb = cb.With(common.CtxKeyFlowRange, r)
				}
			}

			{ // events filter
				filter := common.NewFilter(ctx)
				cb = cb.With(common.CtxKeyFilter, filter)
			}

			{ // time start
				timeStartStr := ctx.QueryParam(common.ParamTimeStart)
				if timeStartInt64, err := strconv.ParseInt(timeStartStr, 10, 64); err == nil {
					cb = cb.With(common.CtxKeyTimeStart, time.Unix(0, timeStartInt64))
				}
			}

			{
				pageIdxStr := ctx.QueryParam(common.ParamPageIndex)
				if pageIdx, err := strconv.Atoi(pageIdxStr); err == nil {
					cb = cb.With(common.CtxKeyPageIndex, pageIdx)
				}
			}

			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}
