package handler

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"gong/domain"
	"gong/sdk"
	"gong/types"
	"gong/view"
	"net/http"
	"strconv"
	"time"

	"github.com/anthdm/hollywood/actor"
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

const (
	ssePingInterval = 200 * time.Millisecond
)

type msgSsePing struct{}

func ConnectProjectSSE(ctx echo.Context) error {
	flusher, ok := ctx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}
	ctx.Response().Header().Set("Content-Type", "text/event-stream")

	app := mustApp(ctx)

	projectID := mustProjectID(ctx)
	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	writer := ctx.Response().Writer
	done := make(chan struct{})
	var repeater actor.SendRepeater

	app.Engine().SpawnFunc(func(actorCtx *actor.Context) {
		switch msg := actorCtx.Message().(type) {
		case actor.Started:
			actorCtx.Send(projectPID, domain.MsgEventsListenStart{})
			repeater = actorCtx.SendRepeat(actorCtx.PID(), msgSsePing{}, ssePingInterval)
		case actor.Stopped:
			actorCtx.Send(projectPID, domain.MsgEventsListenStop{})
			repeater.Stop()
			done <- struct{}{}
		case msgSsePing:
			if err := writeServerSentEvent(ctx.Request().Context(), writer, "ping", nil); err != nil {
				actorCtx.Engine().Stop(actorCtx.PID())
				return
			}
			flusher.Flush()
		case domain.MsgSse:
			if msg.Msgs == nil {
				actorCtx.Engine().Stop(actorCtx.PID())
				return
			}
			if err := func() error {
				for _, msg := range msg.Msgs {
					switch msg := msg.(type) {
					case domain.MsgSseNewActor:
						if err := writeServerSentEvent(
							ctx.Request().Context(),
							writer,
							types.SseTypeNewActor,
							view.ProjectActorListItem(msg.Actor, false),
						); err != nil {
							return err
						}
					case domain.MsgSseNewEvents:
						if err := writeServerSentEvent(
							ctx.Request().Context(),
							writer,
							types.SseTypeUpdateActorNumEvents+msg.Actor.PID,
							view.ProjectActorBadge(msg.Actor),
						); err != nil {
							return err
						}
					}
				}
				return nil
			}(); err != nil {
				actorCtx.Engine().Stop(actorCtx.PID())
				return
			}
			flusher.Flush()
		}
	}, "sse-handler", actor.WithID(requestID(ctx)))

	<-done
	return nil
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
