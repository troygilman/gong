package core

import (
	"stagebase/pkg/dolly"
	"sync"

	"github.com/anthdm/hollywood/actor"
)

type controller struct {
	actors map[*actor.PID]struct{}
}

func newController() actor.Producer {
	return func() actor.Receiver {
		return &controller{}
	}
}

func (c *controller) Receive(ctx *actor.Context) {
	dolly.LogActorMsg(ctx)
	err := func() error {
		switch msg := ctx.Message().(type) {
		case actor.Initialized:
			c.actors = make(map[*actor.PID]struct{})
		case actor.Started:
			ctx.Engine().Subscribe(ctx.PID())
		case actor.ActorStartedEvent:
			if msg.PID != ctx.PID() {
				c.actors[msg.PID] = struct{}{}
			}
		case actor.Stopped:
			wg := &sync.WaitGroup{}
			for pid := range c.actors {
				ctx.Engine().Poison(pid, wg)
			}
			wg.Wait()
		}
		return nil
	}()
	_ = err
}
