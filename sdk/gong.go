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

const (
	serverHost            = "localhost:8080"
	actorInitializedEvent = "actor.ActorInitializedEvent"
	actorStartedEvent     = "actor.ActorStartedEvent"
	actorStoppedEvent     = "actor.ActorStoppedEvent"
)

var (
	gongPID *actor.PID = nil
)

type Event struct {
	PID       string
	Sender    string
	Timestamp time.Time
	MsgType   string
	Level     slog.Level
	Msg       string
	Attr      map[string]any
}

func GongMiddleware() actor.MiddlewareFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {

			sender := ""
			if ctx.Sender() != nil {
				sender = ctx.Sender().String()
			}

			switch msg := ctx.Message().(type) {
			case actor.EventLogger:
				level, logMsg, attr := msg.Log()
				ctx.Send(gongPID, Event{
					PID:       ctx.PID().String(),
					Sender:    sender,
					Timestamp: time.Now(),
					MsgType:   fmt.Sprintf("%T", msg),
					Level:     level,
					Msg:       logMsg,
					Attr:      createAttrMap(attr),
				})
			case actor.Initialized, actor.Started, actor.Stopped:
				ctx.Send(gongPID, Event{
					PID:       ctx.PID().String(),
					Sender:    sender,
					Timestamp: time.Now(),
					MsgType:   fmt.Sprintf("%T", msg),
					Level:     slog.LevelInfo,
					Attr:      make(map[string]any),
				})
			}

			next(ctx)
		}
	}
}

type GongConfig struct {
	ProjectID string
	QueueSize uint
	Logger    *slog.Logger
}

func NewDefaultGongConfig(projectID string) GongConfig {
	return GongConfig{
		ProjectID: projectID,
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
		ctx.Engine().Subscribe(ctx.PID())
	case actor.Started:
		events := make(chan Event, g.config.QueueSize)
		g.events = events
		go connect(g.config.ProjectID, events, func(err error) {
			if g.config.Logger != nil {
				g.config.Logger.Error("Gong: " + err.Error())
			}
		})
	case actor.Stopped:
		close(g.events)
	case actor.EventLogger:
		level, logMsg, attr := msg.Log()
		event := Event{
			MsgType: fmt.Sprintf("%T", msg),
			Level:   level,
			Msg:     logMsg,
			Attr:    createAttrMap(attr),
		}

		if ctx.Sender() != nil {
			event.Sender = ctx.Sender().String()
		}

		switch event.MsgType {
		case actorInitializedEvent, actorStartedEvent, actorStoppedEvent:
			pid, ok := event.Attr["pid"].(*actor.PID)
			if !ok {
				break
			}
			event.PID = pid.String()
		}

		switch msg := msg.(type) {
		case actor.ActorInitializedEvent:
			event.Timestamp = msg.Timestamp
		case actor.ActorStartedEvent:
			event.Timestamp = msg.Timestamp
		case actor.ActorStoppedEvent:
			event.Timestamp = msg.Timestamp
		case actor.ActorRestartedEvent:
			event.Timestamp = msg.Timestamp
		case actor.ActorMaxRestartsExceededEvent:
			event.Timestamp = msg.Timestamp
		default:
			event.Timestamp = time.Now()
		}

		// if err := g.pushEvent(event); err != nil {
		// 	log.Println(err)
		// }
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

func connect(projectID string, events <-chan Event, logErr func(err error)) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: serverHost, Path: fmt.Sprintf("project/%s/connect", projectID)}
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

func createAttrMap(attr []any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(attr); i += 2 {
		var val any = nil
		if i+1 < len(attr) {
			val = attr[i+1]
		}
		m[fmt.Sprintf("%+v", attr[i])] = val
	}
	return m
}

func parsePIDString(a any) (string, bool) {
	m, ok := a.(map[string]any)
	if !ok {
		return "", false
	}
	address, ok := m["address"].(string)
	if !ok {
		return "", false
	}
	id, ok := m["ID"].(string)
	if !ok {
		return "", false
	}
	return actor.NewPID(address, id).String(), true
}
