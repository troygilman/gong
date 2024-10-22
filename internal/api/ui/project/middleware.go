package project

import (
	"context"
	"stagebase/internal/api/common"
	"stagebase/internal/core"
	"stagebase/internal/model"

	"github.com/labstack/echo/v4"
)

func defaultActorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			if _, ok := ctx.Value(paramPid).(string); !ok {
				app := common.MustGetApp(ctx)
				projectID := common.MustGetProjectId(ctx)
				pid, err := app.PID(core.KindProject, projectID.String())
				if err != nil {
					return err
				}
				props, err := core.Request[model.ProjectProps](app, pid, core.GetProjectPropsMsg{})
				if err != nil {
					return err
				}
				for _, actor := range props.Actors {
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
