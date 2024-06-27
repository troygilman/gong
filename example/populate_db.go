package example

import (
	"gong/sdk"
	"sync"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func ExamplePopulateDB() error {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		return err
	}

	server1 := "local/server/1"

	config := sdk.NewDefaultGongConfig("2db8b7bc-5237-4a01-bf7a-d0339da782d9")
	config.QueueSize = 100000
	gongPID := engine.Spawn(sdk.NewGong(config), "gong", actor.WithID("primary"))

	now := time.Now()
	day := time.Hour * 24
	for i := range 10000000 {
		e := sdk.Event{
			PID:       "local/client/1",
			Timestamp: now.Add(-time.Duration(i/1000000) * day),
			Sender:    server1,
		}
		engine.Send(gongPID, e)
		e = sdk.Event{
			PID:       "local/client/2",
			Timestamp: now.Add(-time.Duration(i/1000000) * day),
			Sender:    server1,
		}
		engine.Send(gongPID, e)
	}

	time.Sleep(5 * time.Second)
	wg := &sync.WaitGroup{}
	engine.Poison(gongPID, wg)
	wg.Wait()

	return nil
}
