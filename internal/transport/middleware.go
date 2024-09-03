package transport

import (
	"errors"
	"log/slog"
	"stagebase/internal"
	"strconv"

	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func withRequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := internal.NewContextBuilder(ctx.Request().Context())
			cb = cb.With(internal.CtxKeyRequestID, ctx.Response().Header().Get(echo.HeaderXRequestID))
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

func withContextLogger(keys ...internal.CtxKey) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			logger := internal.MustLogger(ctx.Request().Context())
			logger = internal.WithContextHandler(slog.Default(), keys...)
			cb := internal.NewContextBuilder(ctx.Request().Context())
			cb = cb.With(internal.CtxKeyLogger, logger)
			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}

func logRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			logger := internal.MustLogger(ctx.Request().Context())
			logger.InfoContext(ctx.Request().Context(), "REQUEST", "method", ctx.Request().Method, "uri", ctx.Request().RequestURI)
			return next(ctx)
		}
	}
}

func withApp(app internal.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := internal.NewContextBuilder(ctx.Request().Context())
			cb = cb.With(internal.CtxKeyApp, app)
			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))

			ctx.Set("app", app)
			return next(ctx)
		}
	}
}

func withJWT(app internal.App) echo.MiddlewareFunc {
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

func withProjectContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cb := internal.NewContextBuilder(ctx.Request().Context())

			{ // project ID
				projectID, err := uuid.Parse(ctx.Param(string(internal.CtxKeyProjectID)))
				if err != nil {
					return errors.New("projectID is empty")
				}
				cb = cb.With(internal.CtxKeyProjectID, projectID)
			}

			{ // actor PID
				actorPID := ctx.QueryParam(internal.ParamPID)
				if actorPID != "" {
					cb = cb.With(internal.CtxKeyPID, actorPID)
				}
			}

			{ // tab
				tab := ctx.QueryParam(internal.ParamTab)
				if tab == "" {
					tab = "Events"
				}
				cb = cb.With(internal.CtxKeyTab, tab)
			}

			{ // flow chart type
				flowType := ctx.QueryParam(internal.ParamFlow)
				if flowType == "" {
					flowType = "Sources"
				}
				cb = cb.With(internal.CtxKeyFlow, flowType)
			}

			{ // flow chart node
				flowNode := ctx.QueryParam(internal.ParamFlowNode)
				if flowNode != "" {
					cb = cb.With(internal.CtxKeyFlowNode, flowNode)
				}
			}

			{ // flow chart range
				flowRange := ctx.QueryParam(internal.ParamFlowRange)
				var r internal.Range
				if err := r.Parse(flowRange); err == nil {
					cb = cb.With(internal.CtxKeyFlowRange, r)
				}
			}

			{ // events filter
				filter := internal.NewFilter(ctx)
				cb = cb.With(internal.CtxKeyFilter, filter)
			}

			{ // time start
				timeStartStr := ctx.QueryParam(internal.ParamTimeStart)
				if timeStartInt64, err := strconv.ParseInt(timeStartStr, 10, 64); err == nil {
					cb = cb.With(internal.CtxKeyTimeStart, time.Unix(0, timeStartInt64))
				}
			}

			{
				pageIdxStr := ctx.QueryParam(internal.ParamPageIndex)
				if pageIdx, err := strconv.Atoi(pageIdxStr); err == nil {
					cb = cb.With(internal.CtxKeyPageIndex, pageIdx)
				}
			}

			ctx.SetRequest(ctx.Request().WithContext(cb.Build()))
			return next(ctx)
		}
	}
}
