package example

import (
	"fmt"
	"log/slog"
	"stagebase/sdk"
	"sync"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func ExampleClient() error {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		return err
	}

	config := sdk.NewDefaultStagebaseConfig("2db8b7bc-5237-4a01-bf7a-d0339da782d9")
	config.Host = "localhost"
	config.Port = 8080
	config.Scheme = "ws"
	config.QueueSize = 10000
	stagebasePID := engine.Spawn(sdk.NewStagebase(config), "stagebase", actor.WithID("primary"))

	start := time.Now()

	pids := []*actor.PID{}
	for range 10 {
		pid := engine.Spawn(newServer(), "server", actor.WithMiddleware(sdk.StagebaseMiddleware()))
		pids = append(pids, pid)
	}

	time.Sleep(10 * time.Second)

	wg := &sync.WaitGroup{}
	for _, pid := range pids {
		engine.Poison(pid, wg)
	}
	wg.Wait()

	engine.Poison(stagebasePID).Wait()

	fmt.Printf("Finished in %fs\n", -time.Until(start).Seconds())
	return nil
}

type testMsg struct {
}

func (m testMsg) Log() (slog.Level, string, []any) {
	return slog.LevelInfo, "test", []any{"id", "jsdjsidjsjdi", "pid", actor.NewPID("hello", "123")}
}

type server struct {
	pids []*actor.PID
}

func newServer() actor.Producer {
	return func() actor.Receiver {
		return server{}
	}
}

func (s server) Receive(act *actor.Context) {
	switch msg := act.Message().(type) {
	case actor.Started:
		act.Engine().Subscribe(act.PID())
	case actor.ActorStartedEvent:
		act.Send(msg.PID, testMsg{})
	case testMsg:
		act.Send(act.Sender(), testMsg{})
	}
}
