package handler

import (
	"bytes"
	"encoding/json"
	"gong/domain"
	"gong/sdk"
	"gong/view"
	"log"

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

	return renderView(ctx, view.Project(response.Props))
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
