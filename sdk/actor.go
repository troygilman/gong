package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
)

var (
	gongPID    *actor.PID = nil
	serverHost            = "localhost:8000"
)

type Event struct {
	MsgType  string
	Receiver *actor.PID
	Level    slog.Level
	Msg      string
	Attr     []any
}

type GongConfig struct {
	QueueSize uint
	Logger    *slog.Logger
}

func NewDefaultGongConfig() GongConfig {
	return GongConfig{
		QueueSize: 1000,
		Logger:    slog.Default(),
	}
}

type Gong struct {
	config GongConfig
	events chan<- Event
}

func NewGong(config GongConfig) actor.Producer {
	return func() actor.Receiver {
		return &Gong{
			config: config,
		}
	}
}

func (g *Gong) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		gongPID = ctx.PID()
	case actor.Started:
		ctx.Engine().Subscribe(ctx.PID())
		{
			events := make(chan Event, g.config.QueueSize)
			g.events = events
			go connect(events, func(err error) {
				if g.config.Logger != nil {
					g.config.Logger.Error("Gong: " + err.Error())
				}
			})
		}
	case actor.Stopped:
		close(g.events)
	case actor.EventLogger:
		level, logMsg, attr := msg.Log()
		event := Event{
			MsgType:  fmt.Sprintf("%T", msg),
			Receiver: nil,
			Level:    level,
			Msg:      logMsg,
			Attr:     attr,
		}
		if err := g.pushEvent(event); err != nil {
			log.Println(err)
		}
	case Event:
		if err := g.pushEvent(msg); err != nil {
			log.Println(err)
		}
	}
}

func (g *Gong) pushEvent(e Event) error {
	log.Printf("%T %+v", e, e)
	g.events <- e
	return nil
}

func connect(events <-chan Event, logErr func(err error)) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *&serverHost, Path: "/connect"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logErr(fmt.Errorf("dial: %s", err.Error()))
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if err != websocket.ErrCloseSent {
					logErr(fmt.Errorf("read: %s", err.Error()))
				}
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
		case event, ok := <-events:
			if !ok {
				if err := close(); err != nil {
					logErr(fmt.Errorf("close: %s", err.Error()))
				}
				return
			}
			var buff bytes.Buffer
			if err := json.NewEncoder(&buff).Encode(&event); err != nil {
				logErr(fmt.Errorf("encode: %s", err.Error()))
				return
			}
			if err := c.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				logErr(fmt.Errorf("write: %s", err.Error()))
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			if err := close(); err != nil {
				logErr(fmt.Errorf("close: %s", err.Error()))
			}
			return
		}
	}
}
