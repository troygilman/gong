package core

import (
	"context"
	"errors"
	"fmt"
	"gong/internal/common"
	"gong/pkg/dolly"
	"log"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

const (
	projectUpdatePropsInterval = 1000 * time.Millisecond
)

type msgUpdateProjectInternal struct{}

type project struct {
	commonActorData
	actors       map[string]common.Actor
	actorUpdates map[string]common.Actor
	eventUpdates []common.Event
	metricUpdate Metric
	repeater     actor.SendRepeater
	listeners    map[common.ListenerType]map[*actor.PID]struct{}
	metrics      MetricsManager
}

func NewProjectProducerWithApp(app common.App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *project) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case msgUpdateProjectInternal, msgUpdateProject:
	default:
		dolly.LogActorMsg(aCtx)
	}

	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Initialized:
			p.actors = make(map[string]common.Actor)
			p.actorUpdates = make(map[string]common.Actor)
			p.listeners = make(map[common.ListenerType]map[*actor.PID]struct{})
			p.metrics = NewMetricsManager(5)
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
			log.Println(props)
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
			p.metricUpdate = msg.metric

		case msgUpdateProjectInternal:
			p.metrics.Process(p.metricUpdate)
			listenerData := make(map[*actor.PID][]any)

			// build data for every listener
			for listenerType, listeners := range p.listeners {
				for listenerPID := range listeners {
					switch listenerType {
					case common.ListenerTypeProject:
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
					case common.ListenerTypeEvents:
						for _, event := range p.eventUpdates {
							listenerData[listenerPID] = append(listenerData[listenerPID], event)
						}
					case common.ListenerTypeMetrics:
						listenerData[listenerPID] = append(listenerData[listenerPID], p.metrics.Group())
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
				query := common.EventsQuery{
					ProjectID: p.id,
					ActorPID:  msg.actorPID,
					Timestamp: msg.timestamp,
					NumEvents: msg.numEvents,
					Filter:    msg.filter,
				}
				go queryEvents(aCtx.Engine(), aCtx.Sender(), p.app.Database(), query)
			} else {
				aCtx.Respond(fmt.Errorf("actor does not exist"))
			}
		case msgGetActorSourcesSinks:
			_, ok := p.actors[msg.actorPID]
			if ok {
				duration, err := msg.rng.Duration()
				if err != nil {
					return err
				}

				query := common.ActorSourcesSinksQuery{
					Sources:   msg.sources,
					ProjectID: p.id,
					ActorPID:  msg.actorPID,
					MinTime:   time.Now().Add(-duration),
					MaxTime:   time.Now(),
				}
				go querySourcesSinks(aCtx.Engine(), aCtx.Sender(), p.app.Database(), query)
			} else {
				aCtx.Respond(fmt.Errorf("actor does not exist"))
			}

		case msgGetSourcesSinksNodeInfo:
			if _, ok := p.actors[msg.actorPID]; ok {
				duration, err := msg.rng.Duration()
				if err != nil {
					return err
				}

				query := common.SourcesSinksNodeInfoQuery{
					Source:    msg.source,
					ProjectID: p.id,
					ActorPID:  msg.actorPID,
					NodePID:   msg.nodePID,
					MinTime:   time.Now().Add(-duration),
					MaxTime:   time.Now(),
				}
				go querySourcesSinksNodeInfo(aCtx.Engine(), aCtx.Sender(), p.app.Database(), query)
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
	}(); err != nil {
		dolly.LogActorError(aCtx, err)
		if aCtx.Sender() != nil {
			aCtx.Respond(err)
		}
	}
}

func (p *project) Load(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	p.id = id

	actorProps, err := p.app.Database().QueryActors(ctx, id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *project) Create(ctx context.Context, id uuid.UUID) error {
	p.id = id
	return nil
}

func queryEvents(engine *actor.Engine, sender *actor.PID, db common.Database, query common.EventsQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func querySourcesSinks(engine *actor.Engine, sender *actor.PID, db common.Database, query common.ActorSourcesSinksQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func querySourcesSinksNodeInfo(engine *actor.Engine, sender *actor.PID, db common.Database, query common.SourcesSinksNodeInfoQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := db.QuerySourcesSinksNodeInfo(ctx, query)
	if err != nil {
		engine.Send(sender, err)
		return
	}

	engine.Send(sender, rspGetSourcesSinksNodeInfo{
		info: info,
	})
}
