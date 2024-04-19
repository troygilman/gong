package domain

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindProject = "Project"
)

type project struct {
	commonActorData
}

func NewProjectProducer() actor.Producer {
	return func() actor.Receiver {
		return &project{}
	}
}

func NewProjectProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *project) Receive(ctx *actor.Context) {
	actor_utils.LogActorMsg(ctx)
	if err := p.receive(ctx); err != nil {
		actor_utils.LogActorError(ctx, err)
		if ctx.Sender() != nil {
			ctx.Respond(err)
		}
	}
}

func (p *project) receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case MsgEvent:
	default:
		_ = msg
	}
	return nil
}

func (p *project) load(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}
