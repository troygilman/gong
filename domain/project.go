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
	ActorKindProject = "Project"
	ActorInitialized = "actor.Initialized"
	ActorStarted     = "actor.Started"
	ActorStopped     = "actor.Stopped"
)

type project struct {
	commonActorData
	actors    map[string]types.EventList
	listeners map[string]chan<- types.Event
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
		p.listeners = make(map[string]chan<- types.Event)
	case MsgEvent:
		if err := p.handleEvent(msg.Event); err != nil {
			return err
		}
	case MsgGetProjectProps:
		props := ProjectProps{
			ID:         p.id,
			ActorProps: make([]ProjectActorProps, 0, len(p.actors)),
		}
		for pid, events := range p.actors {
			var alive bool
			if len(events) > 0 {
				alive = events[len(events)-1].MsgType != ActorStopped
			}
			props.ActorProps = append(props.ActorProps, ProjectActorProps{
				PID:       pid,
				NumEvents: int32(len(events)),
				Alive:     alive,
			})
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectEvents:
		events := slices.Clone(p.actors[msg.ActorPID])
		ctx.Respond(RspGetProjectEvents{
			Events: events,
		})
	case MsgEventsListenStart:
		p.listeners[msg.ID] = msg.Events
	case MsgEventsListenStop:
		delete(p.listeners, msg.ID)
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

func (p *project) handleEvent(sdkEvent sdk.Event) error {
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

	for _, c := range p.listeners {
		select {
		case c <- event:
		default:
		}
	}

	if err := p.app.DB().CreateEvent(&event); err != nil {
		return err
	}

	return nil
}
