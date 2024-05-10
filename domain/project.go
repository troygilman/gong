package domain

import (
	"gong/sdk"
	"gong/types"
	"slices"

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
	actors    map[string]types.EventList
	listeners map[string]*actor.PID
	session   int32
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
		p.actors = make(map[string]types.EventList)
		p.listeners = make(map[string]*actor.PID)
	case MsgSdkEvent:
		if err := p.handleEvent(ctx, msg.Event); err != nil {
			return err
		}
	case MsgGetProjectProps:
		props := types.ProjectProps{
			ID:         p.id,
			ActorProps: make([]types.ProjectActorProps, 0, len(p.actors)),
		}
		for pid, events := range p.actors {
			var alive bool
			if len(events) > 0 {
				alive = isActorAlive(events[len(events)-1].MsgType)
			}
			props.ActorProps = append(props.ActorProps, types.ProjectActorProps{
				PID:       pid,
				NumEvents: int32(len(events)),
				Alive:     alive,
			})
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectActorProps:
		events := p.actors[msg.ActorPID]
		var alive bool
		if len(events) > 0 {
			alive = isActorAlive(events[len(events)-1].MsgType)
		}
		props := types.ProjectActorProps{
			PID:       msg.ActorPID,
			NumEvents: int32(len(events)),
			Alive:     alive,
		}
		ctx.Respond(RspGetProjectActorProps{
			Props: props,
		})
	case MsgGetProjectEvents:
		events, ok := p.actors[msg.ActorPID]
		if ok {
			startIdx := max(msg.StartIdx, 0)
			endIdx := min(msg.EndIdx, int32(len(events)))
			ctx.Respond(RspGetProjectEvents{
				Events:         slices.Clone(events[startIdx:endIdx]),
				NumTotalEvents: int32(len(events)),
			})
		} else {
			ctx.Respond(RspGetProjectEvents{
				Events: make(types.EventList, 0),
			})
		}
	case MsgEventsListenStart:
		pid := ctx.Engine().Spawn(newListener(p.id, ctx.Sender()), ActorKindListener)
		p.listeners[msg.ID] = pid
	case MsgEventsListenStop:
		pid, ok := p.listeners[msg.ID]
		if ok {
			ctx.Engine().Poison(pid)
			delete(p.listeners, msg.ID)
		}
	default:
		_ = msg
	}
	return nil
}

func (p *project) load(id uuid.UUID) error {
	p.id = id
	if session, err := p.app.DB().QuerySession(p.id); err == nil {
		p.session = session + 1
	}
	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) handleEvent(ctx *actor.Context, sdkEvent sdk.Event) error {
	event := types.Event{
		ProjectID: p.id,
		Session:   p.session,
		Event:     sdkEvent,
	}

	events, ok := p.actors[event.PID]
	if !ok {
		events = make(types.EventList, 0)
	}

	event.Index = int32(len(events))
	events.Push(event)
	p.actors[event.PID] = events

	for _, pid := range p.listeners {
		ctx.Send(pid, MsgEvent{Event: event})
	}

	if err := p.app.DB().CreateEvent(&event); err != nil {
		return err
	}

	return nil
}

func isActorAlive(eventType string) bool {
	return eventType != ActorStopped && eventType != ActorStoppedEvent
}
