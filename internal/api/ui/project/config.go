package project

import (
	"stagebase/internal/api/common"
	"stagebase/internal/api/ui/graph"
	"stagebase/internal/api/ui/project/events"
	"stagebase/internal/api/ui/timefilter"
	"stagebase/internal/core"
	"stagebase/pkg/gong"

	"github.com/labstack/echo/v4"
)

func Configure(g gong.Group) {
	gong.NewGroup(g, "/project/:project_id",
		func(e *echo.Group) {
			e.GET("", gong.NewTemplHandler(projectProducer))
			e.GET("/events", gong.NewTemplHandler(events.EventTableProducer(false)))
			e.GET("/sse", handleConnectSseProject)
		},
		gong.Extractors(
			gong.NewExtractor(common.ExtractProjectId, gong.WithLogging()),
			gong.NewExtractor(extractView, gong.WithLogging()),
			gong.NewExtractor(extractActorPID),
			gong.NewExtractor(extractTab, gong.WithLogging()),
			gong.NewExtractor(extractFlowType),
			gong.NewExtractor(extractFlowNode),
			gong.NewExtractor(events.ExtractPage),
			gong.NewExtractor(events.ExtractFromFixed),
			gong.NewExtractor(events.ExtractToFixed),
			gong.NewExtractor(common.ExtractFilter),
			gong.NewExtractor(timefilter.Extract),
			gong.NewExtractor(graph.Extract),
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
