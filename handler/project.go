package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"gong/domain"
	"gong/sdk"
	"gong/view"
	"log"
	"net/http"
	"slices"
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

const (
	sseTypeNew    = "sse-new"
	sseTypeUpdate = "sse-update"
	sseTypeDelete = "sse-delete"
)

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

	eventChan := make(chan sdk.Event, 100)
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
	batch := make(map[string][]sdk.Event)

listenerLoop:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break listenerLoop
			}

			batch[event.PID] = append(batch[event.PID], event)
			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			for pid, events := range batch {
				slices.Reverse(events)

				var sseType string
			eventLoop:
				for _, event := range events {
					switch event.MsgType {
					case domain.ActorInitialized:
						sseType = sseTypeNew
						break eventLoop
					case domain.ActorStopped:
						sseType = sseTypeDelete
						break eventLoop
					default:
						sseType = sseTypeUpdate
					}
				}

				switch sseType {
				case sseTypeNew:
					err = writeServerSentEvent(ctx, sseTypeNew, view.ProjectActor(domain.ProjectActorProps{
						PID:       pid,
						NumEvents: len(events),
					}, projectID))
				case sseTypeUpdate:
					err = writeServerSentEvent(ctx, sseTypeUpdate+"-"+pid, view.ProjectActorEvents(events))
				case sseTypeDelete:
					err = writeServerSentEvent(ctx, sseTypeDelete+"-"+pid, nil)
				}

				if err != nil {
					return err
				}

				flusher.Flush()
				delete(batch, pid)
			}
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
