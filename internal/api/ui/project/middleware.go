package project

import (
	"context"
	"stagebase/internal/api/common"

	"github.com/labstack/echo/v4"
)

func defaultActorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			if _, ok := ctx.Value(paramPid).(string); !ok {
				app := common.MustGetApp(ctx)
				projectID := common.MustGetProjectId(ctx)
				projectProps, err := app.ProjectProps(e.Request().Context(), projectID)
				if err != nil {
					return err
				}
				for _, actor := range projectProps.Actors {
					ctx := e.Request().Context()
					ctx = context.WithValue(ctx, paramPid, actor.PID)
					e.SetRequest(e.Request().WithContext(ctx))
					break
				}
			}
			return next(e)
		}
	}
}
