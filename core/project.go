package core

import (
	"context"
	"errors"
	"fmt"
	"gong/common"
	"gong/persistance"
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
	actors    map[string]common.Actor
	updates   map[string]common.Actor
	repeater  actor.SendRepeater
	listeners map[ListenerType]map[*actor.PID]struct{}
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
		p.listeners = make(map[ListenerType]map[*actor.PID]struct{})
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

		if len(msgs) > 0 {
			// attempt to push msgs to every listener concurrently
			listeners := p.listeners[ListenerTypeProject]
			for pid := range listeners {
				aCtx.Send(pid, msgListenData{
					data: msgs,
				})
			}
		}

	case msgGetProjectEvents:
		_, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.EventsQuery{
				ProjectID: p.id,
				ActorPID:  msg.ActorPID,
				EventID:   msg.EventID,
				NumEvents: msg.NumEvents,
				Filter:    msg.Filter,
			}
			go queryEvents(aCtx.Engine(), aCtx.Sender(), p.app.database(), query)
		} else {
			aCtx.Respond(fmt.Errorf("actor does not exist"))
		}
	case msgGetActorSourcesSinks:
		_, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.ActorSourcesSinksQuery{
				Sources:   msg.Sources,
				ProjectID: p.id,
				ActorPID:  msg.ActorPID,
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
	ctx := context.Background()
	events, err := db.QueryEvents(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, rspGetProjectEvents{
		Events: events,
	})
}

func querySourcesSinks(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.ActorSourcesSinksQuery) {
	pids, err := db.QueryActorSourcesSinks(context.Background(), query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, rspGetActorSourcesSinks{
		SourcesSinks: pids,
	})
}
