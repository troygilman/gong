package example

import (
	"fmt"
	"gong/sdk"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func StartClient() {
	start := time.Now()

	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}

	config := sdk.NewDefaultGongConfig("2db8b7bc-5237-4a01-bf7a-d0339da782d9")
	config.QueueSize = 10000
	gongPID := engine.Spawn(sdk.NewGong(config), "gong", actor.WithID("primary"))
	time.Sleep(time.Second)

	wg := &sync.WaitGroup{}
	pid1 := engine.Spawn(newClient(wg), "client", actor.WithID("1"), actor.WithMiddleware(sdk.GongMiddleware()))
	pid2 := engine.Spawn(newClient(wg), "client", actor.WithID("2"), actor.WithMiddleware(sdk.GongMiddleware()))
	engine.Spawn(newServer(pid1, pid2), "server", actor.WithID("primary"), actor.WithMiddleware(sdk.GongMiddleware()))

	wg.Wait()
	fmt.Println("Clients stopped")

	engine.Poison(gongPID, wg)
	wg.Wait()

	fmt.Printf("Finished in %fs\n", -time.Until(start).Seconds())
}

type testMsg struct {
}

func (m testMsg) Log() (slog.Level, string, []any) {
	return slog.LevelInfo, "test", []any{"id", "jsdjsidjsjdi"}
}

type server struct {
	pids []*actor.PID
}

func newServer(pid1 *actor.PID, pid2 *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return server{
			pids: []*actor.PID{pid1, pid2},
		}
	}
}

func (s server) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Started:
		for range 1000000 {
			ctx.Send(s.pids[0], testMsg{})
			ctx.Send(s.pids[1], testMsg{})
			//time.Sleep(1 * time.Millisecond)
		}
		fmt.Println("Finished sending msgs")
		ctx.Engine().Poison(s.pids[0])
		ctx.Engine().Poison(s.pids[1])
		// nums := []int{}
		// num := nums[0]
		// _ = num
	}
}

type client struct {
	wg *sync.WaitGroup
}

func newClient(wg *sync.WaitGroup) actor.Producer {
	return func() actor.Receiver {
		return client{
			wg: wg,
		}
	}
}

func (c client) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Started:
		c.wg.Add(1)
	case actor.Stopped:
		fmt.Println("Stopping client")
		c.wg.Done()
	}
}
