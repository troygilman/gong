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

	pid1 := engine.Spawn(newClient(), "client1", actor.WithID("primary"), actor.WithMiddleware(sdk.GongMiddleware()))
	pid2 := engine.Spawn(newClient(), "client2", actor.WithID("primary"), actor.WithMiddleware(sdk.GongMiddleware()))
	for range 100 {
		engine.Send(pid1, testMsg{})
		engine.Send(pid2, testMsg{})
		time.Sleep(100 * time.Millisecond)
	}
	engine.Poison(pid1)
	engine.Poison(pid2)

	time.Sleep(1 * time.Second)
	engine.Send(pid1, testMsg{})
	time.Sleep(1 * time.Second)

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
