package core

import (
	"context"
	"errors"
	"gong/common"
	"gong/persistance"
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
	actors    map[string]common.Actor
	updates   map[string]common.Actor
	repeater  actor.SendRepeater
	listeners map[string]chan<- []any
}

func newProjectProducerWithApp(app App) actor.Producer {
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
		p.actors = make(map[string]common.Actor)
		p.updates = make(map[string]common.Actor)
		p.listeners = make(map[string]chan<- []any)
	case actor.Started:
		p.repeater = aCtx.SendRepeat(aCtx.PID(), msgProjectUpdateState{}, projectUpdatePropsInterval)
	case actor.Stopped:
		p.repeater.Stop()
	case msgGetProjectProps:
		props := common.ProjectProps{
			ID:     p.id,
			Actors: make(map[string]common.Actor, 0),
		}
		for pid, data := range p.actors {
			props.Actors[pid] = data
		}
		aCtx.Respond(rspGetProjectProps{
			Props: props,
		})
	case msgGetProjectActorProps:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			aCtx.Respond(rspGetProjectActor{
				Actor: actorData,
			})
		} else {
			return errors.New("actor pid does not exist")
		}
	case msgUpdateProjectActors:
		for _, actorData := range msg.Actors {
			p.updates[actorData.PID] = actorData
		}

	case msgProjectUpdateState:
		msgs := make([]any, 0)

		for pid, data := range p.updates {
			_, exists := p.actors[pid]
			if exists {
				msgs = append(msgs, common.SseNewEvents{
					Actor: data,
				})
			} else {
				msgs = append(msgs, common.SseNewActor{
					Actor: data,
				})
			}
			p.actors[pid] = data
			delete(p.updates, pid)
		}

		// attempt to push msgs to every listener concurrently
		for _, c := range p.listeners {
			go func(c chan<- []any, msgs []any) {
				select {
				case c <- msgs:
				default:
				}
			}(c, msgs)
		}

	case msgGetProjectEvents:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.EventsQuery{
				ProjectID: p.id,
				ActorPID:  msg.ActorPID,
				EventID:   msg.EventID,
				NumEvents: msg.NumEvents,
				Filter:    msg.Filter,
			}
			events, err := p.app.database().QueryEvents(ctx, query)
			if err != nil {
				return err
			}
			aCtx.Respond(rspGetProjectEvents{
				Events: events,
				Actor:  actorData,
			})
		} else {
			aCtx.Respond(rspGetProjectEvents{
				Events: make(common.EventList, 0),
			})
		}
	case msgGetActorSourcesSinks:
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
				sourcesSinks, err = p.app.database().QueryActorSources(ctx, query)
			case "Sinks":
				sourcesSinks, err = p.app.database().QueryActorSinks(ctx, query)
			}
			if err != nil {
				return err
			}
			aCtx.Respond(rspGetActorSourcesSinks{
				SourcesSinks: sourcesSinks,
			})
		}
	case msgListenStart:
		p.listeners[msg.id] = msg.c
	case msgListenStop:
		if c, ok := p.listeners[msg.id]; ok {
			close(c)
			delete(p.listeners, msg.id)
		}
	}
	return nil
}

func (p *project) load(ctx context.Context, id uuid.UUID) error {
	p.id = id

	actorProps, err := p.app.database().QueryActors(ctx, id)
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
