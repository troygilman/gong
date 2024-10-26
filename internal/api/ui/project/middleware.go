package project

import (
	"context"
	"stagebase/internal/api/common"
	"stagebase/internal/core"
	"stagebase/internal/model"
	"stagebase/pkg/dolly"
	"stagebase/pkg/gong"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func defaultActorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			ctx := e.Request().Context()
			if _, ok := ctx.Value(paramPid).(string); !ok {
				var (
					app       core.App
					projectId uuid.UUID
				)

				if err := gong.Bind(
					ctx,
					gong.NewBinding(common.ParamApp, &app),
					gong.NewBinding(common.ParamProjectId, &projectId),
				); err != nil {
					return err
				}

				pid, err := app.PID(core.KindProject, projectId.String())
				if err != nil {
					return err
				}
				props, err := dolly.Result[model.ProjectProps](app.Request(pid, core.GetProjectPropsMsg{}))
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
