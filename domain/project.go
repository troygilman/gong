package domain

import (
	"errors"
	"gong/types"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorKindProject           = "project"
	ActorInitialized           = "actor.Initialized"
	ActorStarted               = "actor.Started"
	ActorStopped               = "actor.Stopped"
	ActorStoppedEvent          = "actor.ActorStoppedEvent"
	projectUpdatePropsInterval = 100 * time.Millisecond
)

type project struct {
	commonActorData
	actors    map[string]types.Actor
	updates   map[string]types.Actor
	repeater  actor.SendRepeater
	listeners map[*actor.PID]struct{}
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
		p.actors = make(map[string]types.Actor)
		p.updates = make(map[string]types.Actor)
		p.listeners = make(map[*actor.PID]struct{})
	case actor.Started:
		p.repeater = ctx.SendRepeat(ctx.PID(), msgProjectUpdateState{}, projectUpdatePropsInterval)
	case actor.Stopped:
		p.repeater.Stop()
	case MsgGetProjectProps:
		props := types.ProjectProps{
			ID:     p.id,
			Actors: make(map[string]types.Actor, 0),
		}
		for pid, data := range p.actors {
			props.Actors[pid] = data
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectActorProps:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			ctx.Respond(RspGetProjectActor{
				Actor: actorData,
			})
		} else {
			return errors.New("actor pid does not exist")
		}
	case MsgUpdateProjectActors:
		for _, actorData := range msg.Actors {
			p.updates[actorData.PID] = actorData
		}

	case msgProjectUpdateState:
		msgs := make([]any, 0)

		for pid, data := range p.updates {
			_, exists := p.actors[pid]
			if exists {
				msgs = append(msgs, MsgSseNewEvents{
					Actor: data,
				})
			} else {
				msgs = append(msgs, MsgSseNewActor{
					Actor: data,
				})
			}
			p.actors[pid] = data
			delete(p.updates, pid)
		}

		for pid := range p.listeners {
			ctx.Send(pid, MsgSse{Msgs: msgs})
		}

	case MsgGetProjectEvents:
		actorData, ok := p.actors[msg.ActorPID]
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
				Actor:  actorData,
			})
		} else {
			ctx.Respond(RspGetProjectEvents{
				Events: make(types.EventList, 0),
			})
		}
	case MsgEventsListenStart:
		p.listeners[ctx.Sender()] = struct{}{}
	case MsgEventsListenStop:
		delete(p.listeners, ctx.Sender())
	}
	return nil
}

func (p *project) load(id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.DB().QueryActors(id)
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
