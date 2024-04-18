package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
)

var (
	gongActorPID *actor.PID = nil
	upgrader                = websocket.Upgrader{}
	serverHost              = "localhost:8000"
)

type gongActor struct {
	conn chan<- []byte
}

func NewGongActor() actor.Producer {
	return func() actor.Receiver {
		return &gongActor{}
	}
}

func (g *gongActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		gongActorPID = ctx.PID()
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		{
			conn := make(chan []byte, 1000)
			g.conn = conn
			go connect(conn)
		}
	case actor.Stopped:
		close(g.conn)
	case actor.EventLogger:
		level, logMsg, attr := msg.Log()
		event := Event{
			MsgType:  fmt.Sprintf("%T", msg),
			Receiver: nil,
			Level:    level,
			Msg:      logMsg,
			Attr:     attr,
		}
		if err := g.handleEvent(event); err != nil {
			log.Println(err)
		}
	case Event:
		if err := g.handleEvent(msg); err != nil {
			log.Println(err)
		}
	}
}

func (g *gongActor) handleEvent(e Event) error {
	log.Printf("%T %+v", e, e)
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(e); err != nil {
		return err
	}
	g.conn <- buff.Bytes()
	return nil
}

func connect(msgs <-chan []byte) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *&serverHost, Path: "/connect"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	close := func() error {
		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return err
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		return nil
	}

	for {
		select {
		case <-done:
			return
		case msg := <-msgs:
			if msg == nil {
				if err := close(); err != nil {
					log.Println(err)
				}
				return
			}
			err := c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			if err := close(); err != nil {
				log.Println(err)
			}
			return
		}
	}
}
