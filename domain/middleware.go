package domain

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

func ActorMiddlewareLoad(id uuid.UUID) func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case actor.Started:
				a, ok := ctx.Receiver().(persistentActor)
				if ok {
					if err := a.load(ctx.Context(), id); err != nil {
						actor_utils.LogActorError(ctx, err)
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
			case actor.Started:
				a, ok := ctx.Receiver().(persistentActor)
				if ok {
					if err := a.create(ctx.Context(), id); err != nil {
						actor_utils.LogActorError(ctx, err)
						ctx.Engine().Poison(ctx.PID())
					}
				}
			}
			next(ctx)
		}
	}
}
