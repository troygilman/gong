package handler

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"gong/sdk"
	"gong/types"
	"gong/view"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

func GetProject(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)
	projectID := mustProjectID(eCtx)

	switch eCtx.Request().Header.Get("View") {
	case types.ProjectViewActor:
		actorPID, ok := actorPID(eCtx)
		if !ok {
			return errors.New("actorPID is empty")
		}

		actor, err := app.Actor(ctx, projectID, actorPID)
		if err != nil {
			return err
		}

		return view.ProjectActor(actor).Render(ctx, eCtx.Response())

	case types.ProjectViewFlowChart:
		return view.ActorFlowChartWidget().Render(ctx, eCtx.Response())
	default:
		props, err := app.ProjectProps(ctx, projectID)
		if err != nil {
			return err
		}

		return view.Project(props).Render(ctx, eCtx.Response())
	}
}

func GetProjectEvents(eCtx echo.Context) error {
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

	filter, ok := eCtx.Request().Context().Value(types.CtxFilter).(types.EventsFilter)
	if !ok {
		return errors.New("filter is empty")
	}

	events, err := app.Events(ctx, projectID, actorPID, eventID, n, filter)
	if err != nil {
		events = make(types.EventList, 0)
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

func ConnectProjectSSE(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)

	flusher, ok := eCtx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	eCtx.Response().Header().Set("Content-Type", "text/event-stream")

	requestID := requestID(eCtx)

	projectID := mustProjectID(eCtx)

	sseChan := make(chan []any, 100)

	if err := app.ListenStart(ctx, projectID, requestID, sseChan); err != nil {
		return err
	}

	defer func() {
		if err := app.ListenStop(ctx, projectID, requestID); err != nil {
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
					case types.SseNewActor:
						if err := writeServerSentEvent(
							ctx,
							eCtx.Response(),
							types.SseTypeNewActor,
							view.ProjectActorListItem(sse.Actor, false),
						); err != nil {
							return err
						}
					case types.SseNewEvents:
						if err := writeServerSentEvent(
							ctx,
							eCtx.Response(),
							types.SseTypeUpdateActorNumEvents+sse.Actor.PID,
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

var upgrader = websocket.Upgrader{}

func ConnectProject(eCtx echo.Context) error {
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

func GetActorChart(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	app := mustApp(eCtx)
	projectID := mustProjectID(eCtx)

	actorPID, ok := actorPID(eCtx)
	if !ok {
		return errors.New("actorPID is missing")
	}

	flowType, ok := eCtx.Request().Context().Value(types.CtxFlow).(string)
	if !ok {
		return errors.New("flow is missing")
	}

	pids, err := app.SourcesSinks(ctx, projectID, actorPID, flowType)
	if err != nil {
		return err
	}

	//response.SourcesSinks = append(response.SourcesSinks, "test1", "test2")
	return view.ActorFlowChart(pids).Render(ctx, eCtx.Response())
}

func GetActorChartTooltip(ctx echo.Context) error {
	return view.ChartTooltip().Render(ctx.Request().Context(), ctx.Response())
}
