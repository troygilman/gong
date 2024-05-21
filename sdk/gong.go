package sdk

import (
	"bytes"
	"context"
	"encoding/gob"
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
	Attrs     string
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
		events := make(chan Event, g.config.QueueSize)
		g.events = events

		done := make(chan struct{})
		g.done = done

		var err error
		var conn *websocket.Conn
		for conn == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			u := url.URL{Scheme: "ws", Host: serverHost, Path: fmt.Sprintf("project/%s/connect", g.config.ProjectID)}
			conn, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
			if err != nil {
				fmt.Printf("dial: %s\n", err.Error())
			}
			cancel()
		}

		fmt.Println("Gong connected")
		go connect(conn, events, func(err error) {
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

		event := Event{
			Type:  fmt.Sprintf("%T", msg),
			Level: level,
			Msg:   logMsg,
			Attrs: attrStr,
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
	select {
	case g.events <- e:
	default:
		//slog.Warn("Cant push event", "event", e)
	}
	return nil
}

func connect(conn *websocket.Conn, events <-chan Event, logErr func(err error), done chan<- struct{}) {
	defer conn.Close()
	defer close(done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			_, message, err := conn.ReadMessage()
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
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return err
		}
		select {
		case <-readerDone:
		case <-time.After(time.Second):
		}
		return nil
	}

	var buff bytes.Buffer
	send := func(e []Event) error {
		if len(e) > 0 {
			if err := gob.NewEncoder(&buff).Encode(&e); err != nil {
				return err
			}
			if err := conn.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				return err
			}
			buff.Reset()
		}
		return nil
	}

	batchSize := 10000
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
			batch[numEvents] = event
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
