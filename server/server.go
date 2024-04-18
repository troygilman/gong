package server

import (
	"bytes"
	"encoding/json"
	"gong/sdk"
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{}
)

func Start() error {
	e := echo.New()
	e.GET("/connect", func(ctx echo.Context) error {
		log.Println("/connect")
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
			log.Printf("recv: %+v", event)
		}
		return nil
	})
	return e.Start(":8000")
}
