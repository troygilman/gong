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

	gongPID := engine.Spawn(sdk.NewGong(sdk.NewDefaultGongConfig()), "gong")

	pid := engine.Spawn(newClient(), "client", actor.WithID("primary"), actor.WithMiddleware(sdk.GongMiddleware()))
	engine.Send(pid, testMsg{})
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
