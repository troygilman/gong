package domain

import (
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
	ActorKindProcessor  = "processor"
	sendUpdatesInterval = 100 * time.Millisecond
)

type processor struct {
	commonActorData
	actors   map[string]types.ProjectActorProps
	entropy  ulid.MonotonicReader
	updates  map[string]struct{}
	repeater actor.SendRepeater
}

func NewProcessorProducer() actor.Producer {
	return func() actor.Receiver {
		return &processor{}
	}
}

func NewProcessorProducerWithApp(app App) actor.Producer {
	return func() actor.Receiver {
		return &processor{
			commonActorData: commonActorData{
				app: app,
			},
		}
	}
}

func (p *processor) Receive(ctx *actor.Context) {
	actor_utils.LogActorMsg(ctx)
	if err := p.receive(ctx); err != nil {
		actor_utils.LogActorError(ctx, err)
		if ctx.Sender() != nil {
			ctx.Respond(err)
		}
	}
}

func (p *processor) receive(ctx *actor.Context) error {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		p.actors = make(map[string]types.ProjectActorProps)
		p.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		p.updates = make(map[string]struct{})
	case actor.Started:
		p.repeater = ctx.SendRepeat(ctx.PID(), MsgSendUpdates{}, sendUpdatesInterval)
	case actor.Stopped:
		p.repeater.Stop()
	case MsgSdkEvent:
		if err := p.handleEvent(ctx, msg.Event); err != nil {
			return err
		}
	case MsgSendUpdates:
		projectPID := ctx.Engine().Registry.GetPID(ActorKindProject, p.id.String())
		props := make([]types.ProjectActorProps, 0)
		for pid := range p.updates {
			props = append(props, p.actors[pid])
			delete(p.updates, pid)
		}
		if len(props) > 0 {
			ctx.Send(projectPID, MsgUpdateProjectActorProps{Props: props})
		}
	}
	return nil
}

func (p *processor) handleEvent(ctx *actor.Context, sdkEvent sdk.Event) error {
	eventID, err := ulid.New(ulid.Timestamp(sdkEvent.Timestamp), p.entropy)
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        eventID,
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

	p.updates[event.PID] = struct{}{}

	if err := p.app.DB().CreateEvent(&event); err != nil {
		return err
	}

	return nil
}

func (p *processor) load(id uuid.UUID) error {
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

func (p *processor) create(id uuid.UUID) error {
	p.id = id
	return nil
}
