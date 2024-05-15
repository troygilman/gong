package domain

import (
	"errors"
	"gong/sdk"
	"gong/types"
	"math/rand"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
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
	actors    map[string]types.ProjectActorProps
	listeners map[string]*actor.PID
	entropy   ulid.MonotonicReader
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
		p.actors = make(map[string]types.ProjectActorProps)
		p.listeners = make(map[string]*actor.PID)
		p.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	case MsgSdkEvent:
		if err := p.handleEvent(ctx, msg.Event); err != nil {
			return err
		}
	case MsgGetProjectProps:
		props := types.ProjectProps{
			ID:         p.id,
			ActorProps: make(map[string]types.ProjectActorProps, 0),
		}
		for pid, actorProps := range p.actors {
			props.ActorProps[pid] = actorProps
		}
		ctx.Respond(RspGetProjectProps{
			Props: props,
		})
	case MsgGetProjectActorProps:
		props, ok := p.actors[msg.ActorPID]
		if ok {
			ctx.Respond(RspGetProjectActorProps{
				Props: props,
			})
		} else {
			return errors.New("actor pid does not exist")
		}
	case MsgGetProjectEvents:
		actorProps, ok := p.actors[msg.ActorPID]
		if ok {
			events, err := p.app.DB().QueryEvents(eventsQuery{
				projectID: p.id,
				actorPID:  msg.ActorPID,
				startIdx:  msg.StartIdx,
				endIdx:    msg.EndIdx,
			})
			if err != nil {
				return err
			}
			ctx.Respond(RspGetProjectEvents{
				Events:         events,
				NumTotalEvents: actorProps.NumEvents,
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

	actorProps, err := p.app.DB().QueryActorProps(id)
	if err != nil {
		return err
	}

	for _, props := range actorProps {
		p.actors[props.PID] = props
	}

	return nil
}

func (p *project) create(id uuid.UUID) error {
	p.id = id
	return nil
}

func (p *project) handleEvent(ctx *actor.Context, sdkEvent sdk.Event) error {
	id, err := ulid.New(ulid.Timestamp(sdkEvent.Timestamp), p.entropy)
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        id,
		ProjectID: p.id,
		Event:     sdkEvent,
	}

	actorProps, ok := p.actors[event.PID]
	if !ok {
		actorProps = types.ProjectActorProps{
			PID:       event.PID,
			NumEvents: 0,
			Alive:     false,
		}
	}

	event.Index = int32(actorProps.NumEvents)

	actorProps.NumEvents++
	if event.ID.Compare(actorProps.LastEventID) > 0 {
		actorProps.Alive = isActorAlive(event.Type)
		actorProps.LastEventID = event.ID
	}
	p.actors[event.PID] = actorProps

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
