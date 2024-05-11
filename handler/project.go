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
		view.Project(response.Props, -1),
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
		view.ActorContent(projectID, response.Props),
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

	startIdx, err := strconv.Atoi(ctx.QueryParam("start"))
	if err != nil {
		return errors.New("'start' must be non-empty")
	}

	endIdx, err := strconv.Atoi(ctx.QueryParam("end"))
	if err != nil {
		return errors.New("'end' must be non-empty")
	}

	n, err := strconv.Atoi(ctx.QueryParam("n"))
	if err != nil {
		return errors.New("'n' must be non-empty")
	}
	appendEnd := n > 0

	msg := domain.MsgGetProjectEvents{
		ActorPID: actorPID,
	}

	if appendEnd {
		msg.StartIdx = int32(endIdx)
		msg.EndIdx = msg.StartIdx + int32(n)
	} else {
		msg.EndIdx = int32(startIdx)
		msg.StartIdx = msg.EndIdx + int32(n)
	}

	r := app.Engine().Request(projectPID, msg, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectEvents](r)
	if err != nil {
		return err
	}

	const maxEvents = 100

	newStartIdx := startIdx
	newEndIdx := endIdx
	deleteIdxs := []int32{}

	if appendEnd {
		newEndIdx += len(response.Events)
		overflow := max(newEndIdx-newStartIdx-maxEvents, 0)
		newStartIdx += overflow
		for idx := range overflow {
			deleteIdxs = append(deleteIdxs, int32(startIdx+idx))
		}
	} else {
		newStartIdx -= len(response.Events)
		overflow := max(newEndIdx-newStartIdx-maxEvents, 0)
		newEndIdx -= overflow
		for idx := range overflow {
			deleteIdxs = append(deleteIdxs, int32(endIdx-idx-1))
		}
	}

	moreEvents := int32(newEndIdx) < response.NumTotalEvents-1

	return renderView(
		ctx,
		ctx.Request().Context(),
		view.ProjectActorEvents(projectID, actorPID, response.Events, deleteIdxs, int32(newStartIdx), int32(newEndIdx), appendEnd, moreEvents),
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
						view.ProjectActorListItem(msg.ActorProps, projectID),
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
