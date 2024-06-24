package api

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"gong/common"
	core_common "gong/core/common"
	"gong/sdk"
	"gong/view"
	view_overview "gong/view/overview"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

func bindProjectHandlers(e groupable) {
	newGroup(e, "/project/:projectID", func(e *echo.Group) {
		e.GET("", handleGetProject)
		e.GET("/events", handleGetProjectEvents)
		e.GET("/sse/project", handleConnectSseProject)
		e.GET("/sse/events", handleConnectSseEvents)
		e.GET("/chart", handleGetActorChart)
		e.GET("/connect", handleConnectProject)
	}, middlewareProjectContext())
}

func handleGetProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)
	projectID := mustProjectID(eCtx)

	switch eCtx.Request().Header.Get("View") {
	case common.ProjectViewActor:
		actorPID, ok := actorPID(eCtx)
		if !ok {
			return errors.New("actorPID is empty")
		}

		actor, err := app.Actor(ctx, projectID, actorPID)
		if err != nil {
			return err
		}

		return view.ProjectActor(actor).Render(ctx, eCtx.Response())

	case common.ProjectViewFlowChart:
		return view_overview.FlowChartWidget().Render(ctx, eCtx.Response())
	case common.ProjectViewFlowChartNodeInfo:
		actorPID, ok := actorPID(eCtx)
		if !ok {
			return errors.New("pid is empty")
		}

		nodePID, ok := ctx.Value(common.CtxFlowNode).(string)
		if !ok {
			return errors.New("flow_node is empty")
		}

		flowType, ok := ctx.Value(common.CtxFlow).(string)
		if !ok {
			return errors.New("flow is missing")
		}

		source, err := isSources(flowType)
		if err != nil {
			return err
		}

		nodeInfo, err := app.SourcesSinksNodeInfo(ctx, projectID, actorPID, nodePID, source)
		if err != nil {
			return err
		}

		if nodeInfo.MsgCount > 0 {
			return view_overview.FlowChartNodeInfo(nodeInfo).Render(ctx, eCtx.Response())
		} else {
			return view_overview.FlowChartNodeInfoEmpty().Render(ctx, eCtx.Response())
		}
	default:
		props, err := app.ProjectProps(ctx, projectID)
		if err != nil {
			return err
		}

		return view.Project(props).Render(ctx, eCtx.Response())
	}
}

func handleGetProjectEvents(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	projectID := mustProjectID(eCtx)

	actorPID, ok := actorPID(eCtx)
	if !ok {
		return errors.New("actorPID is empty")
	}

	eventID, err := ulid.Parse(eCtx.QueryParam("id"))
	if err != nil {
		eventID = ulid.ULID{}
	}

	n, err := strconv.Atoi(eCtx.QueryParam("n"))
	if err != nil {
		return errors.New("'n' must be non-empty")
	}

	filter, ok := filter(eCtx)
	if !ok {
		return errors.New("filter is empty")
	}

	events, err := app.Events(ctx, projectID, actorPID, eventID, n, filter)
	if err != nil {
		events = make(common.EventList, 0)
	}

	appendTop := n > 0
	moreTop := true
	moreBottom := true

	if appendTop {
		moreTop = len(events) > 0
	} else {
		moreBottom = len(events) > 0
	}

	return view.ProjectActorEvents(events, appendTop, moreTop, moreBottom).Render(ctx, eCtx.Response())
}

func handleConnectSseProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	flusher, ok := eCtx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	eCtx.Response().Header().Set("Content-Type", "text/event-stream")

	requestID := requestID(eCtx)

	projectID := mustProjectID(eCtx)

	sseChan, err := app.ListenStart(ctx, projectID, requestID, core_common.ListenerTypeProject, common.Filter{})
	if err != nil {
		return err
	}

	defer func() {
		if err := app.ListenStop(ctx, requestID); err != nil {
			log.Println(err)
		}
		// flush the channel until closed
		for range sseChan {
		}
	}()

	for {
		select {
		case sses := <-sseChan:
			for _, sse := range sses {
				if err := func() error {
					switch sse := sse.(type) {
					case common.SseNewActor:
						if err := writeServerSentEvent(
							ctx,
							eCtx.Response(),
							common.SseTypeNewActor,
							view.ProjectActorListItem(sse.Actor, false),
						); err != nil {
							return err
						}
					case common.SseNewEvents:
						if err := writeServerSentEvent(
							ctx,
							eCtx.Response(),
							common.SseTypeUpdateActorNumEvents+sse.Actor.PID,
							view.ProjectActorBadge(sse.Actor),
						); err != nil {
							return err
						}
					}
					return nil
				}(); err != nil {
					return err
				}
				flusher.Flush()
			}
		}
	}
}

func handleConnectSseEvents(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	flusher, ok := eCtx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	eCtx.Response().Header().Set("Content-Type", "text/event-stream")

	requestID := requestID(eCtx)

	projectID := mustProjectID(eCtx)

	filter, ok := filter(eCtx)
	if !ok {
		panic("filter is required")
	}

	sseChan, err := app.ListenStart(ctx, projectID, requestID, core_common.ListenerTypeEvents, filter)
	if err != nil {
		return err
	}

	defer func() {
		if err := app.ListenStop(ctx, requestID); err != nil {
			log.Println(err)
		}
		// flush the channel until closed
		for range sseChan {
		}
	}()

	for {
		select {
		case sses := <-sseChan:
			if err := func() error {
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
						eCtx.Response(),
						common.SseTypeNewEvents,
						nil,
					); err != nil {
						return err
					}
				}

				return nil
			}(); err != nil {
				return err
			}
			flusher.Flush()
		}
	}
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

	flowType, ok := eCtx.Request().Context().Value(common.CtxFlow).(string)
	if !ok {
		return errors.New("flow is missing")
	}

	sources, err := isSources(flowType)
	if err != nil {
		return err
	}

	pids, err := app.SourcesSinks(ctx, projectID, actorPID, sources)
	if err != nil {
		return err
	}

	//response.SourcesSinks = append(response.SourcesSinks, "test1", "test2")
	return view_overview.FlowChart(pids, flowType).Render(ctx, eCtx.Response())
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
