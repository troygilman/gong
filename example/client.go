package example

import (
	"gong/sdk"
	"log"
	"log/slog"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func StartClient() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	gongPID := engine.Spawn(sdk.NewGong(sdk.NewDefaultGongConfig("2db8b7bc-5237-4a01-bf7a-d0339da782d9")), "gong")

	pid := engine.Spawn(newClient(), "client", actor.WithID("primary"), actor.WithMiddleware(sdk.GongMiddleware()))
	engine.Send(pid, testMsg{})
	time.Sleep(5 * time.Second)
	engine.Poison(pid)

	time.Sleep(1 * time.Second)
	engine.Send(pid, testMsg{})
	time.Sleep(10 * time.Second)

	engine.Stop(gongPID)
	time.Sleep(time.Second)
}

type testMsg struct {
}

func (m testMsg) Log() (slog.Level, string, []any) {
	return slog.LevelInfo, "test", []any{"id", "jsdjsidjsjdi"}
}

type client struct{}

func newClient() actor.Producer {
	return func() actor.Receiver {
		return client{}
	}
}

func (c client) Receive(ctx *actor.Context) {

}
