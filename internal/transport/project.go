package transport

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"stagebase/internal"
	"stagebase/internal/core"
	view_admin "stagebase/internal/view/admin"
	view_project "stagebase/internal/view/project"
	"stagebase/sdk"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func bindProjectHandlers(e groupable) {
	newGroup(e, "/project/:projectID", func(e *echo.Group) {
		e.GET("", handleGetProject)
		e.GET("/events", handleGetProjectEvents)
		e.GET("/sse/project", handleConnectSseProject)
		e.GET("/sse/events", handleConnectSseEvents)
		e.GET("/chart", handleGetActorChart)
		e.GET("/connect", handleConnectProject)
		e.GET("/admin", handleGetProjectAdmin)
		e.GET("/admin/sse", handleConnectSseProjectAdmin)
	},
		withProjectContext(),
		withContextLogger(internal.CtxKeyFilter),
		logRequest(),
	)
}

func handleGetProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)
	projectID := mustProjectID(eCtx)

	switch eCtx.Request().Header.Get("View") {
	case internal.ProjectViewActor:
		actorPID, ok := actorPID(eCtx)
		if !ok {
			return errors.New("actorPID is empty")
		}

		actor, err := app.Actor(ctx, projectID, actorPID)
		if err != nil {
			return err
		}

		return view_project.ProjectActor(actor).Render(ctx, eCtx.Response())

	case internal.ProjectViewFlowChart:
		return view_project.FlowTabContent().Render(ctx, eCtx.Response())
	case internal.ProjectViewFlowChartNodeInfo:
		actorPID, ok := actorPID(eCtx)
		if !ok {
			return errors.New("pid is empty")
		}

		nodePID, ok := ctx.Value(internal.CtxKeyFlowNode).(string)
		if !ok {
			return errors.New("flow_node is empty")
		}

		flowType, ok := ctx.Value(internal.CtxKeyFlow).(string)
		if !ok {
			return errors.New("flow is missing")
		}

		source, err := isSources(flowType)
		if err != nil {
			return err
		}

		flowRange, ok := ctx.Value(internal.CtxKeyFlowRange).(internal.Range)
		if !ok {
			return errors.New("flow_range is missing")
		}

		nodeInfo, err := app.SourcesSinksNodeInfo(ctx, projectID, actorPID, nodePID, source, flowRange)
		if err != nil {
			return err
		}

		if nodeInfo.MsgCount > 0 {
			return view_project.FlowChartNodeInfoContent(nodeInfo).Render(ctx, eCtx.Response())
		} else {
			return view_project.FlowChartNodeInfoContentEmpty("").Render(ctx, eCtx.Response())
		}
	default:
		props, err := app.ProjectProps(ctx, projectID)
		if err != nil {
			return err
		}

		return view_project.Project(props).Render(ctx, eCtx.Response())
	}
}

func handleGetProjectEvents(eCtx echo.Context) error {
	return view_project.EventTableComponent(eCtx.Request().Context(), false).Render(eCtx.Request().Context(), eCtx.Response())
}

func handleConnectSseProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	return sse(eCtx, internal.ListenerTypeProject, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case internal.SseNewActor:
				if err := writeServerSentEvent(
					ctx,
					eCtx.Response(),
					internal.SseTypeNewActor,
					view_project.ProjectActorListItem(sse.Actor, false),
				); err != nil {
					return err
				}
			case internal.SseNewEvents:
				if err := writeServerSentEvent(
					ctx,
					eCtx.Response(),
					internal.SseTypeUpdateActorNumEvents+sse.Actor.PID,
					view_project.ProjectActorBadge(sse.Actor),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func handleConnectSseEvents(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	return sse(eCtx, internal.ListenerTypeEvents, func(sses []any) error {
		eventCount := 0
		for _, sse := range sses {
			switch sse := sse.(type) {
			case internal.Event:
				eventCount++
			default:
				_ = sse
			}
		}

		if eventCount > 0 {
			if err := writeServerSentEvent(
				ctx,
				eCtx.Response(),
				internal.SseTypeNewEvents,
				nil,
			); err != nil {
				return err
			}
		}

		return nil
	})
}

var upgrader = websocket.Upgrader{}

func handleConnectProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	projectID := mustProjectID(eCtx)

	c, err := upgrader.Upgrade(eCtx.Response().Writer, eCtx.Request(), nil)
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

func handleGetActorChart(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)
	projectID := mustProjectID(eCtx)

	actorPID, ok := actorPID(eCtx)
	if !ok {
		return errors.New("actorPID is missing")
	}

	flowType, ok := eCtx.Request().Context().Value(internal.CtxKeyFlow).(string)
	if !ok {
		return errors.New("flow is missing")
	}

	flowRange, ok := eCtx.Request().Context().Value(internal.CtxKeyFlowRange).(internal.Range)
	if !ok {
		return errors.New("flow range is missing")
	}

	sources, err := isSources(flowType)
	if err != nil {
		return err
	}

	pids, err := app.SourcesSinks(ctx, projectID, actorPID, sources, flowRange)
	if err != nil {
		return err
	}

	return view_project.FlowChart(pids, flowType).Render(ctx, eCtx.Response())
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

func handleGetProjectAdmin(eCtx echo.Context) error {
	return view_admin.Admin().Render(eCtx.Request().Context(), eCtx.Response())
}

func handleConnectSseProjectAdmin(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	return sse(eCtx, internal.ListenerTypeMetrics, func(sses []any) error {
		for _, sse := range sses {
			switch sse := sse.(type) {
			case core.MetricGroup:
				if err := writeServerSentEvent(
					ctx,
					eCtx.Response(),
					internal.SseTypeUpdateMetricLatency,
					view_admin.Metric(sse.Avg.Latency, internal.SseTypeUpdateMetricLatency),
				); err != nil {
					return err
				}
				if err := writeServerSentEvent(
					ctx,
					eCtx.Response(),
					internal.SseTypeUpdateMetricThroughput,
					view_admin.Metric(sse.Avg.Throughput, internal.SseTypeUpdateMetricThroughput),
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
