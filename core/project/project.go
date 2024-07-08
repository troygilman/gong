package project

import (
	"context"
	"errors"
	"fmt"
	gong_common "gong/common"
	"gong/persistance"
	"log"
	"time"

	"gong/core/common"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	actor_utils "github.com/troygilman0/utils"
)

const (
	projectUpdatePropsInterval = 100 * time.Millisecond
)

type msgUpdateProjectInternal struct{}

type project struct {
	common.CommonActorData
	actors       map[string]gong_common.Actor
	actorUpdates map[string]gong_common.Actor
	eventUpdates []gong_common.Event
	metricUpdate common.Metric
	repeater     actor.SendRepeater
	listeners    map[common.ListenerType]map[*actor.PID]struct{}
	metrics      common.MetricsManager
}

func NewProjectProducerWithApp(app common.App) actor.Producer {
	return func() actor.Receiver {
		return &project{
			CommonActorData: common.CommonActorData{
				App: app,
			},
		}
	}
}

func (p *project) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case msgUpdateProjectInternal, common.MsgUpdateProject:
	default:
		actor_utils.LogActorMsg(aCtx)
	}

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
		p.actors = make(map[string]gong_common.Actor)
		p.actorUpdates = make(map[string]gong_common.Actor)
		p.listeners = make(map[common.ListenerType]map[*actor.PID]struct{})
		p.metrics = common.NewMetricsManager(20)
	case actor.Started:
		p.repeater = aCtx.SendRepeat(aCtx.PID(), msgUpdateProjectInternal{}, projectUpdatePropsInterval)
	case actor.Stopped:
		p.repeater.Stop()
	case common.MsgGetProjectProps:
		props := gong_common.ProjectProps{
			ID:     p.Id,
			Actors: make(map[string]gong_common.Actor, 0),
		}
		for pid, data := range p.actors {
			props.Actors[pid] = data
		}
		log.Println(props)
		aCtx.Respond(common.RspGetProjectProps{
			Props: props,
		})

	case common.MsgGetProjectActorProps:
		actorData, ok := p.actors[msg.ActorPID]
		if ok {
			aCtx.Respond(common.RspGetProjectActor{
				Actor: actorData,
			})
		} else {
			return errors.New("actor pid does not exist")
		}

	case common.MsgUpdateProject:
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
				case common.ListenerTypeProject:
					for _, actor := range p.actorUpdates {
						_, exists := p.actors[actor.PID]
						if exists {
							listenerData[listenerPID] = append(listenerData[listenerPID], gong_common.SseNewEvents{
								Actor: actor,
							})
						} else {
							listenerData[listenerPID] = append(listenerData[listenerPID], gong_common.SseNewActor{
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
				aCtx.Send(pid, common.MsgListenData{
					Data: data,
				})
			}
		}

		// reset update state
		for _, actor := range p.actorUpdates {
			p.actors[actor.PID] = actor
			delete(p.actorUpdates, actor.PID)
		}
		p.eventUpdates = []gong_common.Event{}

	case common.MsgGetProjectEvents:
		_, ok := p.actors[msg.ActorPID]
		if ok {
			query := persistance.EventsQuery{
				ProjectID: p.Id,
				ActorPID:  msg.ActorPID,
				// EventID:   msg.EventID,
				Timestamp: msg.Timestamp,
				NumEvents: msg.NumEvents,
				Filter:    msg.Filter,
			}
			go queryEvents(aCtx.Engine(), aCtx.Sender(), p.App.Database(), query)
		} else {
			aCtx.Respond(fmt.Errorf("actor does not exist"))
		}
	case common.MsgGetActorSourcesSinks:
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

	case common.MsgGetSourcesSinksNodeInfo:
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

	case common.MsgListenStart:
		listeners, ok := p.listeners[msg.LType]
		if !ok {
			listeners = make(map[*actor.PID]struct{})
		}
		listeners[aCtx.Sender()] = struct{}{}
		p.listeners[msg.LType] = listeners
	case common.MsgListenStop:
		listeners, ok := p.listeners[msg.LType]
		if ok {
			delete(listeners, aCtx.Sender())
		}
	}
	return nil
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

	engine.Send(sender, common.RspGetProjectEvents{
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

	engine.Send(sender, common.RspGetActorSourcesSinks{
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

	engine.Send(sender, common.RspGetSourcesSinksNodeInfo{
		Info: info,
	})
}
