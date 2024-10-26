package v1

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"stagebase/internal/api/common"
	"stagebase/internal/core"
	"stagebase/pkg/gong"
	"stagebase/sdk"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func Configure(g gong.Group) {
	gong.NewGroup(g, "/project/:project_id",
		func(e *echo.Group) {
			e.GET("/connect", handleConnectProject)
		},
		gong.Extractors(
			gong.NewExtractor(common.ExtractProjectId),
		),
		// projectMiddleware(),
		// logRequest(),
	)
}

func handleConnectProject(e echo.Context) error {
	ctx := e.Request().Context()

	var (
		app       core.App
		projectId uuid.UUID
	)

	if err := gong.Bind(
		ctx,
		gong.NewBinding(common.ParamApp, &app),
		gong.NewBinding(common.ParamProjectId, &projectId),
	); err != nil {
		return err
	}

	upgrader := websocket.Upgrader{}
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

		processorPID, err := app.PID(core.KindProcessor, projectId.String())
		if err != nil {
			return err
		}

		app.Send(processorPID, core.ProcessEventsMsg{
			Events: events,
		})
	}
}
