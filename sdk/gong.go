package sdk

import (
	"bytes"
	"context"
	"encoding/gob"
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
	Type      string
	Level     slog.Level
	Msg       string
	Attr      string
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
				attrStr, err := encodeAttrMap(createAttrMap(attr))
				if err != nil {
					attrStr = ""
				}
				ctx.Send(gongPID, &Event{
					PID:       ctx.PID().String(),
					Sender:    sender,
					Timestamp: time.Now(),
					Type:      fmt.Sprintf("%T", msg),
					Level:     level,
					Msg:       logMsg,
					Attr:      attrStr,
				})
			case actor.Initialized, actor.Started, actor.Stopped:
				ctx.Send(gongPID, &Event{
					PID:       ctx.PID().String(),
					Sender:    sender,
					Timestamp: time.Now(),
					Type:      fmt.Sprintf("%T", msg),
					Level:     slog.LevelInfo,
					Attr:      "",
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
	events chan<- *Event
	done   <-chan struct{}
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
		events := make(chan *Event, g.config.QueueSize)
		g.events = events

		done := make(chan struct{})
		g.done = done

		go connect(g.config.ProjectID, events, func(err error) {
			if g.config.Logger != nil {
				g.config.Logger.Error("Gong: " + err.Error())
			}
		}, done)
	case actor.Stopped:
		fmt.Println("Stopping Gong")
		close(g.events)
		<-g.done
	case actor.EventLogger:
		level, logMsg, attr := msg.Log()
		attrMap := createAttrMap(attr)
		attrStr, err := encodeAttrMap(attrMap)
		if err != nil {
			attrStr = ""
		}

		event := &Event{
			Type:  fmt.Sprintf("%T", msg),
			Level: level,
			Msg:   logMsg,
			Attr:  attrStr,
		}

		if ctx.Sender() != nil {
			event.Sender = ctx.Sender().String()
		}

		switch event.Type {
		case actorInitializedEvent, actorStartedEvent, actorStoppedEvent:
			pid, ok := attrMap["pid"].(*actor.PID)
			if !ok {
				break
			}
			event.PID = pid.String()
		}

		switch msg := msg.(type) {
		case actor.ActorInitializedEvent:
			event.PID = msg.PID.String()
			event.Timestamp = msg.Timestamp
		case actor.ActorStartedEvent:
			event.PID = msg.PID.String()
			event.Timestamp = msg.Timestamp
		case actor.ActorStoppedEvent:
			event.PID = msg.PID.String()
			event.Timestamp = msg.Timestamp
		case actor.ActorRestartedEvent:
			event.PID = msg.PID.String()
			event.Timestamp = msg.Timestamp
		case actor.ActorMaxRestartsExceededEvent:
			event.PID = msg.PID.String()
			event.Timestamp = msg.Timestamp
		default:
			event.Timestamp = time.Now()
		}

		if err := g.pushEvent(event); err != nil {
			log.Println(err)
		}
	case *Event:
		if err := g.pushEvent(msg); err != nil {
			log.Println(err)
		}
	}
}

func (g *Gong) pushEvent(e *Event) error {
	select {
	case g.events <- e:
	default:
		slog.Warn("Cant push event", "event", e)
	}
	return nil
}

func connect(projectID string, events <-chan *Event, logErr func(err error), done chan<- struct{}) {
	defer close(done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var c *websocket.Conn
	{
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		u := url.URL{Scheme: "ws", Host: serverHost, Path: fmt.Sprintf("project/%s/connect", projectID)}
		c, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
		cancel()
		if err != nil {
			logErr(fmt.Errorf("dial: %s", err.Error()))
			return
		}
	}
	defer c.Close()
	fmt.Println("Gong connected")

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
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
		case <-readerDone:
		case <-time.After(time.Second):
		}
		return nil
	}

	send := func(e []Event) error {
		if len(e) > 0 {
			var buff bytes.Buffer
			if err := gob.NewEncoder(&buff).Encode(&e); err != nil {
				return err
			}
			if err := c.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				return err
			}
		}
		return nil
	}

	batchSize := 1000
	batch := make([]Event, batchSize)
	numEvents := 0
	for {
		select {
		case <-readerDone:
			return
		case <-interrupt:
			log.Println("interrupt")
			if err := close(); err != nil {
				logErr(fmt.Errorf("close: %s", err.Error()))
			}
			return
		case event, more := <-events:
			if !more {
				if err := send(batch[:numEvents]); err != nil {
					logErr(fmt.Errorf("send: %s", err.Error()))
					return
				}
				if err := close(); err != nil {
					logErr(fmt.Errorf("close: %s", err.Error()))
				}
				return
			}
			batch[numEvents] = *event
			numEvents++
			if numEvents >= batchSize {
				if err := send(batch[:numEvents]); err != nil {
					logErr(fmt.Errorf("send: %s", err.Error()))
					return
				}
				numEvents = 0
			}
		default:
			if err := send(batch[:numEvents]); err != nil {
				logErr(fmt.Errorf("send: %s", err.Error()))
				return
			}
			numEvents = 0
			time.Sleep(time.Millisecond)
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

func encodeAttrMap(attr map[string]any) (string, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(attr); err != nil {
		return "", err
	}
	return buffer.String(), nil

}
