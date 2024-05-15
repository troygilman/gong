package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"gong/domain"
	"gong/sdk"
	"gong/types"
	"gong/view"
	"log"
	"net/http"
	"strconv"

	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/troygilman0/utils"
)

func GetProject(ctx echo.Context) error {
	app := mustApp(ctx)
	projectID := mustProjectID(ctx)

	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	r := app.Engine().Request(projectPID, domain.MsgGetProjectProps{}, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectProps](r)
	if err != nil {
		return err
	}

	return renderView(
		ctx,
		ctx.Request().Context(),
		view.Project(response.Props),
	)
}

func GetProjectActor(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID := mustProjectID(ctx)
	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	actorPID, ok := actorPID(ctx)
	if !ok {
		return errors.New("actorPID is empty")
	}

	msg := domain.MsgGetProjectActorProps{
		ActorPID: actorPID,
	}

	r := app.Engine().Request(projectPID, msg, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectActorProps](r)
	if err != nil {
		return err
	}

	return renderView(
		ctx,
		ctx.Request().Context(),
		view.ProjectActor(response.Props),
	)
}

func GetProjectEvents(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID := mustProjectID(ctx)
	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	actorPID, ok := actorPID(ctx)
	if !ok {
		return errors.New("actorPID is empty")
	}

	eventID, err := ulid.Parse(ctx.QueryParam("id"))
	if err != nil {
		eventID = ulid.ULID{}
	}

	n, err := strconv.Atoi(ctx.QueryParam("n"))
	if err != nil {
		return errors.New("'n' must be non-empty")
	}

	msg := domain.MsgGetProjectEvents{
		ActorPID:  actorPID,
		EventID:   eventID,
		NumEvents: n,
	}

	r := app.Engine().Request(projectPID, msg, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectEvents](r)
	if err != nil {
		return err
	}

	appendTop := n > 0
	moreTop := true
	moreBottom := true

	if appendTop {
		moreTop = len(response.Events) > 0
	} else {
		moreBottom = len(response.Events) > 0
	}

	return renderView(
		ctx,
		ctx.Request().Context(),
		view.ProjectActorEvents(response.Events, appendTop, moreTop, moreBottom),
	)
}

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
	done := make(chan error)

	pid := app.Engine().SpawnFunc(func(actorCtx *actor.Context) {
		switch msg := actorCtx.Message().(type) {
		case actor.Started:
			actorCtx.Send(projectPID, domain.MsgEventsListenStart{ID: requestID(ctx)})
		case actor.Stopped:
			actorCtx.Send(projectPID, domain.MsgEventsListenStop{ID: requestID(ctx)})
		case domain.MsgSse:
			if msg.Msgs == nil {
				done <- nil
			}
			for _, msg := range msg.Msgs {
				switch msg := msg.(type) {
				case domain.MsgSsePing:
					if err := writeServerSentEvent(
						ctx.Request().Context(),
						writer,
						"ping",
						nil,
					); err != nil {
						done <- err
					}
				case domain.MsgSseNewActor:
					if err := writeServerSentEvent(
						ctx.Request().Context(),
						writer,
						types.SseTypeNewActor,
						view.ProjectActorListItem(msg.ActorProps, false),
					); err != nil {
						done <- err
					}
				case domain.MsgSseNewEvents:
					if err := writeServerSentEvent(
						ctx.Request().Context(),
						writer,
						types.SseTypeUpdateActorNumEvents+msg.ActorPID,
						view.ProjectActorBadge(msg.ActorPID, msg.NumEvents, msg.ActorAlive),
					); err != nil {
						done <- err
					}
				}
			}
			flusher.Flush()
		}
	}, "sse-handler", actor.WithID(requestID(ctx)))

	err = <-done
	app.Engine().Poison(pid)
	return err
}

var upgrader = websocket.Upgrader{}

func ConnectProject(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID := mustProjectID(ctx)
	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	c, err := upgrader.Upgrade(ctx.Response().Writer, ctx.Request(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var event sdk.Event
		if err := json.NewDecoder(bytes.NewReader(message)).Decode(&event); err != nil {
			log.Println(err)
			break
		}

		app.Engine().Send(projectPID, domain.MsgSdkEvent{
			Event: event,
		})
	}
	return nil
}
