package core

import (
	"context"
	"errors"
	"fmt"
	"gong/common"
	"gong/internal/persistance"
	"log"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	projectUpdatePropsInterval = 100 * time.Millisecond
)

type msgUpdateProjectInternal struct{}

type project struct {
	CommonActorData
	actors       map[string]common.Actor
	actorUpdates map[string]common.Actor
	eventUpdates []common.Event
	metricUpdate Metric
	repeater     actor.SendRepeater
	listeners    map[ListenerType]map[*actor.PID]struct{}
	metrics      MetricsManager
}

func NewProjectProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			CommonActorData: CommonActorData{
				App: app,
			},
		}
	}
}

func (p *project) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case msgUpdateProjectInternal, MsgUpdateProject:
	default:
		actor_utils.LogActorMsg(aCtx)
	}

	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Initialized:
			p.actors = make(map[string]common.Actor)
			p.actorUpdates = make(map[string]common.Actor)
			p.listeners = make(map[ListenerType]map[*actor.PID]struct{})
			p.metrics = NewMetricsManager(20)
		case actor.Started:
			p.repeater = aCtx.SendRepeat(aCtx.PID(), msgUpdateProjectInternal{}, projectUpdatePropsInterval)
		case actor.Stopped:
			p.repeater.Stop()
		case MsgGetProjectProps:
			props := common.ProjectProps{
				ID:     p.Id,
				Actors: make(map[string]common.Actor, 0),
			}
			for pid, data := range p.actors {
				props.Actors[pid] = data
			}
			log.Println(props)
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

		case MsgUpdateProject:
			for _, actorData := range msg.Actors {
				p.actorUpdates[actorData.PID] = actorData
			}
			p.eventUpdates = append(p.eventUpdates, msg.Events...)
			p.metricUpdate = msg.Metric

		case msgUpdateProjectInternal:
			p.metrics.Process(p.metricUpdate)
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
						for _, event := range p.eventUpdates {
							listenerData[listenerPID] = append(listenerData[listenerPID], event)
						}
					case ListenerTypeMetrics:
						listenerData[listenerPID] = append(listenerData[listenerPID], p.metrics.Group())
					default:
					}
				}
			}

			// push data to every listener concurrently
			for pid, data := range listenerData {
				if len(data) > 0 {
					aCtx.Send(pid, MsgListenData{
						Data: data,
					})
				}
			}

			// reset update state
			for _, actor := range p.actorUpdates {
				p.actors[actor.PID] = actor
				delete(p.actorUpdates, actor.PID)
			}
			p.eventUpdates = []common.Event{}

		case MsgGetProjectEvents:
			_, ok := p.actors[msg.ActorPID]
			if ok {
				query := persistance.EventsQuery{
					ProjectID: p.Id,
					ActorPID:  msg.ActorPID,
					Timestamp: msg.Timestamp,
					NumEvents: msg.NumEvents,
					Filter:    msg.Filter,
				}
				go queryEvents(aCtx.Engine(), aCtx.Sender(), p.App.Database(), query)
			} else {
				aCtx.Respond(fmt.Errorf("actor does not exist"))
			}
		case MsgGetActorSourcesSinks:
			_, ok := p.actors[msg.ActorPID]
			if ok {
				duration, err := msg.Range.Duration()
				if err != nil {
					return err
				}

				query := persistance.ActorSourcesSinksQuery{
					Sources:   msg.Sources,
					ProjectID: p.Id,
					ActorPID:  msg.ActorPID,
					MinTime:   time.Now().Add(-duration),
					MaxTime:   time.Now(),
				}
				go querySourcesSinks(aCtx.Engine(), aCtx.Sender(), p.App.Database(), query)
			} else {
				aCtx.Respond(fmt.Errorf("actor does not exist"))
			}

		case MsgGetSourcesSinksNodeInfo:
			if _, ok := p.actors[msg.ActorPID]; ok {
				duration, err := msg.Range.Duration()
				if err != nil {
					return err
				}

				query := persistance.SourcesSinksNodeInfoQuery{
					Source:    msg.Source,
					ProjectID: p.Id,
					ActorPID:  msg.ActorPID,
					NodePID:   msg.NodePID,
					MinTime:   time.Now().Add(-duration),
					MaxTime:   time.Now(),
				}
				go querySourcesSinksNodeInfo(aCtx.Engine(), aCtx.Sender(), p.App.Database(), query)
			} else {
				aCtx.Respond(fmt.Errorf("actor does not exist"))
			}

		case MsgListenStart:
			listeners, ok := p.listeners[msg.LType]
			if !ok {
				listeners = make(map[*actor.PID]struct{})
			}
			listeners[aCtx.Sender()] = struct{}{}
			p.listeners[msg.LType] = listeners
		case MsgListenStop:
			listeners, ok := p.listeners[msg.LType]
			if ok {
				delete(listeners, aCtx.Sender())
			}
		}
		return nil
	}(); err != nil {
		actor_utils.LogActorError(aCtx, err)
		if aCtx.Sender() != nil {
			aCtx.Respond(err)
		}
	}
}

func (p *project) Load(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	p.Id = id

	actorProps, err := p.App.Database().QueryActors(ctx, id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *project) Create(ctx context.Context, id uuid.UUID) error {
	p.Id = id
	return nil
}

func queryEvents(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.EventsQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := db.QueryEvents(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, RspGetProjectEvents{
		Events: events,
	})
}

func querySourcesSinks(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.ActorSourcesSinksQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pids, err := db.QueryActorSourcesSinks(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, RspGetActorSourcesSinks{
		SourcesSinks: pids,
	})
}

func querySourcesSinksNodeInfo(engine *actor.Engine, sender *actor.PID, db persistance.Database, query persistance.SourcesSinksNodeInfoQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := db.QuerySourcesSinksNodeInfo(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, RspGetSourcesSinksNodeInfo{
		Info: info,
	})
}
