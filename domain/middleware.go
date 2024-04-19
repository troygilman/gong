package domain

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

func ActorMiddlewareLoad(id uuid.UUID) func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case actor.Initialized:
				a, ok := ctx.Receiver().(persistentActor)
				if ok {
					if err := a.load(id); err != nil {
						ctx.Engine().Poison(ctx.PID())
					}
				}
			}
			next(ctx)
		}
	}
}

func ActorMiddlewareCreate(id uuid.UUID) func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case actor.Initialized:
				a, ok := ctx.Receiver().(persistentActor)
				if ok {
					if err := a.create(id); err != nil {
						ctx.Engine().Poison(ctx.PID())
					}
				}
			}
			next(ctx)
		}
	}
}
