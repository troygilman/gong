package transport

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"stagebase/internal/api/view/graph"
	"stagebase/internal/api/view/timefilter"
	"stagebase/internal/common"
	"stagebase/internal/core"
	adminView "stagebase/internal/view/admin"
	projectView "stagebase/internal/view/project"
	"stagebase/pkg/gong"
	"stagebase/sdk"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func bindProjectHandlers(g gong.Group) {
	gong.NewGroup(g, "/project/:projectID",
		func(e *echo.Group) {
			e.GET("", handleGetProject)
			// e.GET("/events", gong.NewTemplHandler(projectView.EventTableComponent(false)))
			e.GET("/sse/project", handleConnectSseProject)
			e.GET("/sse/events", handleConnectSseEvents)
			e.GET("/graph/diagram", handleGetActorChart)
			e.GET("/connect", handleConnectProject)
			e.GET("/admin", handleGetProjectAdmin)
			e.GET("/admin/sse", handleConnectSseProjectAdmin)
		},
		// gong.ContextLogHandler(
		// 	common.CtxKeyFilter,
		// 	common.CtxKeyPID,
		// ),
		gong.ContextExtractors(
			extractProjectID,
			extractActorPID,
			extractTab,
			extractFilter,
			extractFlowType,
			extractFlowNode,
			extractFlowRange,
			extractPageIndex,
			extractTimeStart,
			timefilter.Extract,
			graph.Extract,
		),
		projectMiddleware(),
		logRequest(),
	)
}

func handleGetProject(e echo.Context) error {
	ctx := e.Request().Context()
	switch e.Request().Header.Get("View") {
	case common.ProjectViewActor:
		return projectView.ProjectActorComponent(ctx).Render(ctx, e.Response())
	case common.ProjectViewFlowChart:
		return projectView.FlowTabContent().Render(ctx, e.Response())
	case common.ProjectViewFlowChartNodeInfo:
		return projectView.FlowChartNodeInfoComponent(ctx, false).Render(ctx, e.Response())
	default:
		return projectView.ProjectComponent(ctx).Render(ctx, e.Response())
	}
}

func handleConnectSseProject(e echo.Context) error {
	ctx := e.Request().Context()
	return sse(e, common.ListenerTypeProject, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case common.SseNewActor:
				if err := writeServerSentEvent(
					ctx,
					e.Response(),
					common.SseTypeNewActor,
					projectView.ProjectActorListItem(sse.Actor, false),
				); err != nil {
					return err
				}
			case common.SseNewEvents:
				if err := writeServerSentEvent(
					ctx,
					e.Response(),
					common.SseTypeUpdateActorNumEvents+sse.Actor.PID,
					projectView.ProjectActorBadge(sse.Actor),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func handleConnectSseEvents(e echo.Context) error {
	ctx := e.Request().Context()
	return sse(e, common.ListenerTypeEvents, func(sses []any) error {
		eventCount := 0
		for _, sse := range sses {
			switch sse := sse.(type) {
			case common.Event:
				eventCount++
			default:
				_ = sse
			}
		}

		if eventCount > 0 {
			if err := writeServerSentEvent(
				ctx,
				e.Response(),
				common.SseTypeNewEvents,
				nil,
			); err != nil {
				return err
			}
		}

		return nil
	})
}

var upgrader = websocket.Upgrader{}

func handleConnectProject(e echo.Context) error {
	ctx := e.Request().Context()
	app := mustApp(e)

	projectID := mustProjectID(e)

	c, err := upgrader.Upgrade(e.Response().Writer, e.Request(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %s", err)
		}

		var events []sdk.Event
		if err := gob.NewDecoder(bytes.NewReader(message)).Decode(&events); err != nil {
			return fmt.Errorf("decode: %s", err)
		}

		if err := app.ProcessEvents(ctx, projectID, events); err != nil {
			return err
		}
	}
}

func handleGetActorChart(e echo.Context) error {
	ctx := e.Request().Context()
	app := mustApp(e)
	projectID := mustProjectID(e)

	actorPID, ok := actorPID(e)
	if !ok {
		return errors.New("actorPID is missing")
	}

	flowType, ok := e.Request().Context().Value(common.CtxKeyFlow).(string)
	if !ok {
		return errors.New("flow is missing")
	}

	timeFilter := timefilter.MustGetTimeFilter(ctx)

	sources, err := isSources(flowType)
	if err != nil {
		return err
	}

	nodes, err := app.Database().QueryActorSourcesSinks(ctx, common.ActorSourcesSinksQuery{
		Sources:   sources,
		ProjectID: projectID,
		ActorPID:  actorPID,
		MinTime:   timeFilter.From,
		MaxTime:   time.Now(),
	})
	if err != nil {
		return err
	}
	_ = nodes

	return graph.Component(false, "").Render(ctx, e.Response())
}

func isSources(flowType string) (bool, error) {
	switch flowType {
	case "Sources":
		return true, nil
	case "Sinks":
		return false, nil
	default:
		return false, errors.New("invalid flow type")
	}
}

func handleGetProjectAdmin(e echo.Context) error {
	return adminView.Admin().Render(e.Request().Context(), e.Response())
}

func handleConnectSseProjectAdmin(e echo.Context) error {
	ctx := e.Request().Context()
	return sse(e, common.ListenerTypeMetrics, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case core.MetricGroup:
				if err := writeServerSentEvent(
					ctx,
					e.Response(),
					common.SseTypeUpdateMetricLatency,
					adminView.Metric(sse.Avg.Latency, common.SseTypeUpdateMetricLatency),
				); err != nil {
					return err
				}
				if err := writeServerSentEvent(
					ctx,
					e.Response(),
					common.SseTypeUpdateMetricThroughput,
					adminView.Metric(sse.Avg.Throughput, common.SseTypeUpdateMetricThroughput),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
