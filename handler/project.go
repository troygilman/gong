package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gong/domain"
	"gong/sdk"
	"gong/types"
	"gong/view"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/troygilman0/utils"
)

func GetProject(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID, err := uuid.Parse(ctx.Param("projectID"))
	if err != nil {
		return err
	}

	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	r := app.Engine().Request(projectPID, domain.MsgGetProjectProps{}, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectProps](r)
	if err != nil {
		return err
	}

	return renderView(ctx, view.Project(response.Props, -1))
}

func GetProjectEvents(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID, err := uuid.Parse(ctx.Param("projectID"))
	if err != nil {
		return err
	}

	actorPID := ctx.QueryParam("actor")
	if actorPID == "" {
		return errors.New("actor must be non-empty")
	}

	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	msg := domain.MsgGetProjectEvents{
		ActorPID: actorPID,
	}

	r := app.Engine().Request(projectPID, msg, domain.DefaultTimeout)
	response, err := utils.HandleResponse[domain.RspGetProjectEvents](r)
	if err != nil {
		return err
	}

	return renderView(ctx, view.ProjectActorEventsContainer(projectID, actorPID, response.Events, false))
}

func ConnectProjectSSE(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID, err := uuid.Parse(ctx.Param("projectID"))
	if err != nil {
		return err
	}

	projectPID, err := app.PID(domain.ActorKindProject, projectID.String())
	if err != nil {
		return err
	}

	eventChan := make(chan types.Event, 100)
	app.Engine().Send(projectPID, domain.MsgEventsListenStart{
		ID:     requestID(ctx),
		Events: eventChan,
	})

	defer app.Engine().Send(projectPID, domain.MsgEventsListenStop{
		ID: requestID(ctx),
	})

	flusher, ok := ctx.Response().Writer.(http.Flusher)
	if !ok {
		return errors.New("sse not supported")
	}

	ctx.Response().Header().Set("Content-Type", "text/event-stream")

	timerDuration := 50 * time.Millisecond
	timer := time.NewTimer(timerDuration)
	batch := make(map[string]types.EventList)

listenerLoop:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break listenerLoop
			}

			events := batch[event.Event.PID]
			events.Push(event)
			batch[event.Event.PID] = events

			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			for pid, events := range batch {
				err := func() error {
					var sseType string = types.SseTypeNewEvents
					var numEvents int32
					var actorAlive bool

					if len(events) > 0 {
						numEvents = events[len(events)-1].Index + 1
						actorAlive = events[len(events)-1].MsgType != domain.ActorStopped
						if events[0].Index == 0 {
							sseType = types.SseTypeNewActor
						}
					}

					switch sseType {
					case types.SseTypeNewActor:
						props := domain.ProjectActorProps{
							PID:       pid,
							NumEvents: int32(len(events)),
							Alive:     actorAlive,
						}
						if err := writeServerSentEvent(ctx, types.SseTypeNewActor, view.ProjectActorListItem(props, projectID)); err != nil {
							return err
						}
					case types.SseTypeNewEvents:
						if err := writeServerSentEvent(ctx, types.SseTypeNewEvents+pid, view.ProjectActorEvents(events)); err != nil {
							return err
						}
						if err := writeServerSentEvent(ctx, types.SseTypeUpdateActorNumEvents+pid, view.ProjectActorBadge(pid, numEvents, actorAlive)); err != nil {
							return err
						}
					case types.SseTypeDeleteActor:
						if err := writeServerSentEvent(ctx, types.SseTypeDeleteActor+pid, nil); err != nil {
							return err
						}
					}

					return nil
				}()

				if err != nil {
					return fmt.Errorf("error while writing SSE: %s", err.Error())
				}

				delete(batch, pid)
			}

			if err := writeServerSentEvent(ctx, "ping", nil); err != nil {
				return err
			}

			flusher.Flush()
		}
		timer.Reset(timerDuration)
	}

	return nil
}

var upgrader = websocket.Upgrader{}

func ConnectProject(ctx echo.Context) error {
	app := mustApp(ctx)

	projectID, err := uuid.Parse(ctx.Param("projectID"))
	if err != nil {
		return err
	}

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

		app.Engine().Send(projectPID, domain.MsgEvent{
			Event: event,
		})
	}
	return nil
}
