package domain

import (
	"context"
	"errors"
	"gong/domain/persistance"
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

func NewProjectProducerWithApp(app *App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *project) Receive(aCtx *actor.Context) {
	actor_utils.LogActorMsg(aCtx)

	ctx, cancel := context.WithTimeout(aCtx.Context(), 1*time.Second)
	defer cancel()

	if err := p.receive(ctx, aCtx); err != nil {
		actor_utils.LogActorError(aCtx, err)
		if aCtx.Sender() != nil {
			aCtx.Respond(err)
		}
	}
}

func (p *project) receive(ctx context.Context, aCtx *actor.Context) error {
	switch msg := aCtx.Message().(type) {
	case actor.Initialized:
		p.actors = make(map[string]types.Actor)
		p.updates = make(map[string]types.Actor)
		p.listeners = make(map[*actor.PID]struct{})
	case actor.Started:
		p.repeater = aCtx.SendRepeat(aCtx.PID(), msgProjectUpdateState{}, projectUpdatePropsInterval)
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
		aCtx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectActorProps:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			aCtx.Respond(RspGetProjectActor{
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
			aCtx.Send(pid, MsgSse{Msgs: msgs})
		}

	case MsgGetProjectEvents:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.EventsQuery{
				ProjectID: p.id,
				ActorPID:  msg.ActorPID,
				EventID:   msg.EventID,
				NumEvents: msg.NumEvents,
				Filter:    msg.Filter,
			}
			events, err := p.app.db.QueryEvents(ctx, query)
			if err != nil {
				return err
			}
			aCtx.Respond(RspGetProjectEvents{
				Events: events,
				Actor:  actorData,
			})
		} else {
			aCtx.Respond(RspGetProjectEvents{
				Events: make(types.EventList, 0),
			})
		}
	case MsgGetActorSourcesSinks:
		_, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.ActorSourcesSinksQuery{
				ProjectID: p.id,
				ActorPID:  msg.ActorPID,
				MinTime:   time.Now().Add(-24 * time.Hour),
				MaxTime:   time.Now(),
			}
			var sourcesSinks []string
			var err error
			switch msg.Flow {
			case "Sources":
				sourcesSinks, err = p.app.db.QueryActorSources(ctx, query)
			case "Sinks":
				sourcesSinks, err = p.app.db.QueryActorSinks(ctx, query)
			}
			if err != nil {
				return err
			}
			aCtx.Respond(RspGetActorSourcesSinks{
				SourcesSinks: sourcesSinks,
			})
		}
	case MsgEventsListenStart:
		p.listeners[aCtx.Sender()] = struct{}{}
	case MsgEventsListenStop:
		delete(p.listeners, aCtx.Sender())
	}
	return nil
}

func (p *project) load(ctx context.Context, id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.db.QueryActors(ctx, id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *project) create(ctx context.Context, id uuid.UUID) error {
	p.id = id
	return nil
}

func isActorAlive(eventType string) bool {
	return eventType != ActorStopped && eventType != ActorStoppedEvent
}
