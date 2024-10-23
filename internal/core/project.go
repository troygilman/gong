package core

import (
	"context"
	"errors"
	"stagebase/internal/model"
	"stagebase/pkg/dolly"
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
	actors       map[string]model.Actor
	actorUpdates map[string]model.Actor
	eventUpdates []model.Event
	metricUpdate Metric
	repeater     actor.SendRepeater
	listeners    map[ListenerType]map[*actor.PID]struct{}
	metrics      MetricsManager
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

func (p *project) Receive(aCtx *actor.Context) {
	switch aCtx.Message().(type) {
	case msgUpdateProjectInternal, updateProjectMsg:
	default:
		dolly.LogActorMsg(aCtx)
	}

	if err := func() error {
		switch msg := aCtx.Message().(type) {
		case actor.Initialized:
			p.actors = make(map[string]model.Actor)
			p.actorUpdates = make(map[string]model.Actor)
			p.listeners = make(map[ListenerType]map[*actor.PID]struct{})
			p.metrics = NewMetricsManager(5)
		case actor.Started:
			p.repeater = aCtx.SendRepeat(aCtx.PID(), msgUpdateProjectInternal{}, projectUpdatePropsInterval)
		case actor.Stopped:
			p.repeater.Stop()
		case GetProjectPropsMsg:
			props := model.ProjectProps{
				ID:     p.id,
				Actors: make(map[string]model.Actor, 0),
			}
			for pid, data := range p.actors {
				props.Actors[pid] = data
			}
			aCtx.Respond(props)

		case GetActorMsg:
			actorData, ok := p.actors[msg.Pid]
			if ok {
				aCtx.Respond(actorData)
			} else {
				return errors.New("actor pid does not exist")
			}

		case updateProjectMsg:
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
					case ListenerTypeProject:
						for _, actor := range p.actorUpdates {
							_, exists := p.actors[actor.PID]
							if exists {
								listenerData[listenerPID] = append(listenerData[listenerPID], SseNewEventsMsg{
									Actor: actor,
								})
							} else {
								listenerData[listenerPID] = append(listenerData[listenerPID], SseNewActorMsg{
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
					aCtx.Send(pid, listenDataMsg{
						data: data,
					})
				}
			}

			// reset update state
			for _, actor := range p.actorUpdates {
				p.actors[actor.PID] = actor
				delete(p.actorUpdates, actor.PID)
			}
			p.eventUpdates = []model.Event{}

		case listenStartMsg:
			listeners, ok := p.listeners[msg.lType]
			if !ok {
				listeners = make(map[*actor.PID]struct{})
			}
			listeners[aCtx.Sender()] = struct{}{}
			p.listeners[msg.lType] = listeners
		case listenStopMsg:
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
