package project

import (
	"stagebase/internal/api/common"
	"stagebase/internal/api/view/graph"
	"stagebase/internal/api/view/timefilter"
	"stagebase/internal/core"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func Configure(g gong.Group) {
	gong.NewGroup(g, "/project/:project_id",
		func(e *echo.Group) {
			e.GET("", gong.NewTemplHandler(projectProducer))
			e.GET("/events", gong.NewTemplHandler(eventTableProducer(false)))
			e.GET("/sse", handleConnectSseProject)
		},
		gong.ContextExtractors(
			common.ExtractProjectId,
			extractView,
			extractActorPID,
			extractTab,
			extractFlowType,
			extractFlowNode,
			extractPage,
			common.ExtractFilter,
			timefilter.Extract,
			graph.Extract,
		),
		gong.ContextLogHandler(
			common.ParamProjectId,
			paramView,
		),
		defaultActorMiddleware(),
		common.LogRequestMiddleware(),
	)
}

func handleConnectSseProject(e echo.Context) error {
	ctx := e.Request().Context()
	return common.HandleServerSentEvents(e, core.ListenerTypeProject, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case core.SseNewActorMsg:
				if err := common.WriteServerSentEvent(
					ctx,
					e.Response(),
					sseTypeNewActor,
					ProjectActorListItem(sse.Actor, false),
				); err != nil {
					return err
				}
			case core.SseNewEventsMsg:
				if err := common.WriteServerSentEvent(
					ctx,
					e.Response(),
					sseTypeUpdateActorNumEvents+sse.Actor.PID,
					ProjectActorBadge(sse.Actor),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
