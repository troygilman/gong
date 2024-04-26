package domain

import (
	"gong/sdk"
	"log"
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
	actors    map[string]EventList
	listeners map[string]chan<- sdk.Event
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
		p.actors = make(map[string]EventList)
		p.listeners = make(map[string]chan<- sdk.Event)
	case MsgEvent:
		p.handleEvent(msg.Event)
	case MsgGetProjectProps:
		props := ProjectProps{
			ID:         p.id,
			ActorProps: make([]ProjectActorProps, 0, len(p.actors)),
		}
		for pid, events := range p.actors {
			props.ActorProps = append(props.ActorProps, ProjectActorProps{
				PID:       pid,
				NumEvents: len(events),
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
	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) handleEvent(event sdk.Event) {
	log.Printf("%+v", event)

	for _, c := range p.listeners {
		select {
		case c <- event:
		default:
		}
	}

	switch event.MsgType {
	case ActorInitialized:
		events := make(EventList, 0)
		p.actors[event.PID] = events
	case ActorStopped:
		delete(p.actors, event.PID)
		return
	}

	events, ok := p.actors[event.PID]
	if ok {
		events.Push(event)
		p.actors[event.PID] = events
		log.Println(events)
	}
}
