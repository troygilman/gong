package core

import (
	"context"
	"errors"
	"fmt"
	"gong/common"
	"gong/persistance"
	"slices"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	ActorInitialized           = "actor.Initialized"
	ActorStarted               = "actor.Started"
	ActorStopped               = "actor.Stopped"
	ActorStoppedEvent          = "actor.ActorStoppedEvent"
	projectUpdatePropsInterval = 100 * time.Millisecond
)

type project struct {
	commonActorData
	actors       map[string]common.Actor
	actorUpdates map[string]common.Actor
	eventUpdates []common.Event
	repeater     actor.SendRepeater
	listeners    map[ListenerType]map[*actor.PID]struct{}
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
		p.actorUpdates = make(map[string]common.Actor)
		p.listeners = make(map[ListenerType]map[*actor.PID]struct{})
	case actor.Started:
		p.repeater = aCtx.SendRepeat(aCtx.PID(), msgUpdateProjectInternal{}, projectUpdatePropsInterval)
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
			props: props,
		})

	case msgGetProjectActorProps:
		actorData, ok := p.actors[msg.actorPID]
		if ok {
			aCtx.Respond(rspGetProjectActor{
				actor: actorData,
			})
		} else {
			return errors.New("actor pid does not exist")
		}

	case msgUpdateProject:
		for _, actorData := range msg.actors {
			p.actorUpdates[actorData.PID] = actorData
		}
		p.eventUpdates = append(p.eventUpdates, msg.events...)

	case msgUpdateProjectInternal:
		listenerData := make(map[*actor.PID][]any)

		// build data for every listener
		for listenerType, listeners := range p.listeners {
			for listenerPID := range listeners {
				switch listenerType {
				case ListenerTypeProject:
					for _, actor := range p.actorUpdates {
						_, exists := p.actors[actor.PID]
						if exists {
							listenerData[listenerPID] = append(listenerData[listenerPID], common.SseNewEvents{
								Actor: actor,
							})
						} else {
							listenerData[listenerPID] = append(listenerData[listenerPID], common.SseNewActor{
								Actor: actor,
							})
						}
					}
				case ListenerTypeEvents:
					listenerData[listenerPID] = append(listenerData[listenerPID], slices.Clone(p.eventUpdates))
				default:
				}
			}
		}

		// push data to every listener concurrently
		for pid, data := range listenerData {
			if len(data) > 0 {
				aCtx.Send(pid, msgListenData{
					data: data,
				})
			}
		}

		// reset update state
		for _, actor := range p.actorUpdates {
			p.actors[actor.PID] = actor
			delete(p.actorUpdates, actor.PID)
		}
		p.eventUpdates = []common.Event{}

	case msgGetProjectEvents:
		_, ok := p.actors[msg.actorPID]
		if ok {
			query := persistance.EventsQuery{
				ProjectID: p.id,
				ActorPID:  msg.actorPID,
				EventID:   msg.eventID,
				NumEvents: msg.numEvents,
				Filter:    msg.filter,
			}
			go queryEvents(aCtx.Engine(), aCtx.Sender(), p.app.database(), query)
		} else {
			aCtx.Respond(fmt.Errorf("actor does not exist"))
		}
	case msgGetActorSourcesSinks:
		_, ok := p.actors[msg.actorPID]
		if ok {
			query := persistance.ActorSourcesSinksQuery{
				Sources:   msg.sources,
				ProjectID: p.id,
				ActorPID:  msg.actorPID,
				MinTime:   time.Now().Add(-24 * time.Hour),
				MaxTime:   time.Now(),
			}
			go querySourcesSinks(aCtx.Engine(), aCtx.Sender(), p.app.database(), query)
		} else {
			aCtx.Respond(fmt.Errorf("actor does not exist"))
		}
	case msgListenStart:
		listeners, ok := p.listeners[msg.lType]
		if !ok {
			listeners = make(map[*actor.PID]struct{})
		}
		listeners[aCtx.Sender()] = struct{}{}
		p.listeners[msg.lType] = listeners
	case msgListenStop:
		listeners, ok := p.listeners[msg.lType]
		if ok {
			delete(listeners, aCtx.Sender())
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

func queryEvents(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.EventsQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, err := db.QueryEvents(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, rspGetProjectEvents{
		events: events,
	})
}

func querySourcesSinks(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.ActorSourcesSinksQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pids, err := db.QueryActorSourcesSinks(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, rspGetActorSourcesSinks{
		sourcesSinks: pids,
	})
}
