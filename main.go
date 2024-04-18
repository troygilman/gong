package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func main() {
	slog.SetDefault(slog.Default())

	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Initialized:
			ctx.Engine().Subscribe(ctx.PID())
		case actor.EventLogger:
			level, logMsg, attr := msg.Log()
			level = slog.LevelInfo
			slog.Log(context.Background(), level, logMsg, attr...)
		default:
			log.Printf("%T %+v", msg, msg)
		}
	}, "gong")

	pid := engine.Spawn(newServer(), "server", actor.WithID("primary"))
	engine.Poison(pid)

	time.Sleep(1 * time.Second)
	engine.Send(pid, testMsg{})
	engine.BroadcastEvent(testMsg{})

	time.Sleep(10 * time.Second)
}

type testMsg struct {
}

func (m testMsg) Log() (slog.Level, string, []any) {
	return slog.LevelInfo, "test", []any{"id", "jsdjsidjsjdi"}
}

type server struct{}

func newServer() actor.Producer {
	return func() actor.Receiver {
		return server{}
	}
}

func (s server) Receive(ctx *actor.Context) {

}
