package domain

import (
	"errors"
	"gong/types"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindProject  = "project"
	ActorInitialized  = "actor.Initialized"
	ActorStarted      = "actor.Started"
	ActorStopped      = "actor.Stopped"
	ActorStoppedEvent = "actor.ActorStoppedEvent"
)

type project struct {
	commonActorData
	actors    map[string]types.ProjectActorProps
	listeners map[string]*actor.PID
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
	case actor.Initialized:
		p.actors = make(map[string]types.ProjectActorProps)
		p.listeners = make(map[string]*actor.PID)
	case MsgGetProjectProps:
		props := types.ProjectProps{
			ID:         p.id,
			ActorProps: make(map[string]types.ProjectActorProps, 0),
		}
		for pid, actorProps := range p.actors {
			props.ActorProps[pid] = actorProps
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectActorProps:
		props, ok := p.actors[msg.ActorPID]
		if ok {
			ctx.Respond(RspGetProjectActorProps{
				Props: props,
			})
		} else {
			return errors.New("actor pid does not exist")
		}
	case MsgUpdateProjectActorProps:
		msgs := make([]any, 0)

		for _, props := range msg.Props {
			_, exists := p.actors[props.PID]
			if exists {
				msgs = append(msgs, MsgSseNewEvents{
					ActorProps: props,
				})
			} else {
				msgs = append(msgs, MsgSseNewActor{
					ActorProps: props,
				})
			}
			p.actors[props.PID] = props
		}

		msgs = append(msgs, MsgSsePing{})

		for _, pid := range p.listeners {
			ctx.Send(pid, MsgSse{Msgs: msgs})
		}

	case MsgGetProjectEvents:
		actorProps, ok := p.actors[msg.ActorPID]
		if ok {
			query := eventsQuery{
				projectID: p.id,
				actorPID:  msg.ActorPID,
				eventID:   msg.EventID,
				numEvents: msg.NumEvents,
				filter:    msg.Filter,
			}
			events, err := p.app.DB().QueryEvents(query)
			if err != nil {
				return err
			}
			ctx.Respond(RspGetProjectEvents{
				Events: events,
				Props:  actorProps,
			})
		} else {
			ctx.Respond(RspGetProjectEvents{
				Events: make(types.EventList, 0),
			})
		}
	case MsgEventsListenStart:
		//pid := ctx.Engine().Spawn(newListener(p.id, ctx.Sender()), ActorKindListener)
		p.listeners[msg.ID] = ctx.Sender()
	case MsgEventsListenStop:
		delete(p.listeners, msg.ID)
		// pid, ok := p.listeners[msg.ID]
		// if ok {
		// 	ctx.Engine().Poison(pid)
		// 	delete(p.listeners, msg.ID)
		// }
	default:
		_ = msg
	}
	return nil
}

func (p *project) load(id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.DB().QueryActorProps(id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}

func isActorAlive(eventType string) bool {
	return eventType != ActorStopped && eventType != ActorStoppedEvent
}
